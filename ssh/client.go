package ssh

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/sftp"
	"github.com/skeema/knownhosts"
	"golang.org/x/crypto/ssh"
)

// HostKeyCallback SSH 主机密钥校验策略
//   - "strict": 拒绝未知主机（生产环境推荐）
//   - "tofu":   首次连接自动信任并写入 known_hosts（最方便，向后兼容默认）
//   - "off":    不校验（仅用于调试，强烈不推荐）
type HostKeyCallback string

const (
	HostKeyCallbackStrict HostKeyCallback = "strict"
	HostKeyCallbackTOFU   HostKeyCallback = "tofu"
	HostKeyCallbackOff    HostKeyCallback = "off"
)

// SSHConfig SSH连接配置
type SSHConfig struct {
	Name       string `json:"name"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password,omitempty"`
	KeyPath    string `json:"keyPath,omitempty"`    // 密钥文件路径（兼容旧版）
	PrivateKey string `json:"privateKey,omitempty"` // 密钥内容（直接输入）
	Timeout    int    `json:"timeout,omitempty"`

	// HostKeyPolicy 主机密钥校验策略
	// 默认 "tofu"（向后兼容）：首次连接自动信任并写入 known_hosts
	// 推荐生产环境设为 "strict"
	HostKeyPolicy HostKeyCallback `json:"hostKeyPolicy,omitempty"`

	// KnownHostsFile 自定义 known_hosts 文件路径
	// 默认空 = 使用 ~/.ssh/known_hosts
	KnownHostsFile string `json:"knownHostsFile,omitempty"`
}

// shellSession 单个 shell 会话
type shellSession struct {
	session   *ssh.Session
	stdin     io.WriteCloser
	stdout    io.Reader
	readBuf   []byte
	bufMutex  sync.Mutex
	closing   int32 // atomic: 0=active, 1=closing
}

// SSHClient SSH客户端结构体
type SSHClient struct {
	config      *SSHConfig
	client      *ssh.Client
	sftpClient  *sftp.Client
	isConnected bool
	closing     bool

	// 多 shell 会话支持
	sessions   map[string]*shellSession // sessionID -> session
	sessMu     sync.RWMutex

	lastPingTime time.Time
	latency      int64

	// 搜索取消标志
	searchCancelMap map[string]bool
	searchMutex     sync.RWMutex

	// 连接状态回调
	onDisconnect func(connID string) // 断线回调
	connID       string              // 当前连接 ID（用于回调）
}

// CommandResult 命令执行结果
type CommandResult struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exitCode"`
	Success  bool   `json:"success"`
}

// FileInfo 文件信息（定义在 sftp_service.go）

// PortForwardConfig 端口转发配置
type PortForwardConfig struct {
	LocalPort  int    `json:"localPort"`
	RemoteHost string `json:"remoteHost"`
	RemotePort int    `json:"remotePort"`
}

// NewSSHClient 创建新的SSH客户端
func NewSSHClient(config *SSHConfig) *SSHClient {
	if config.Port == 0 {
		config.Port = 22
	}
	if config.Timeout == 0 {
		config.Timeout = 30
	}
	return &SSHClient{
		config:      config,
		isConnected: false,
		latency:     0,
		sessions:    make(map[string]*shellSession),
		searchCancelMap: make(map[string]bool),
	}
}

// SetDisconnectCallback 设置断线回调
func (s *SSHClient) SetDisconnectCallback(connID string, callback func(string)) {
	s.connID = connID
	s.onDisconnect = callback
}

// handleDisconnect 处理断线
func (s *SSHClient) handleDisconnect() {
	if s.closing || !s.isConnected {
		return
	}

	fmt.Printf("[SSHClient] ⚠️ 检测到连接断开: %s\n", s.connID)
	s.isConnected = false

	// 通知外部
	if s.onDisconnect != nil {
		go s.onDisconnect(s.connID)
	}
}

// Connect 连接到SSH服务器
func (s *SSHClient) Connect() error {
	authMethods := []ssh.AuthMethod{}

	// 密码认证
	if s.config.Password != "" {
		authMethods = append(authMethods, ssh.Password(s.config.Password))
	}

	// 密钥认证（优先使用直接输入的密钥内容）
	if s.config.PrivateKey != "" {
		key, err := parsePrivateKeyContent([]byte(s.config.PrivateKey))
		if err != nil {
			return fmt.Errorf("解析私钥失败: %v", err)
		}
		authMethods = append(authMethods, key)
	} else if s.config.KeyPath != "" {
		key, err := loadPrivateKey(s.config.KeyPath)
		if err != nil {
			return fmt.Errorf("加载私钥失败: %v", err)
		}
		authMethods = append(authMethods, key)
	}

	if len(authMethods) == 0 {
		return fmt.Errorf("未提供认证方式（密码或密钥）")
	}

	// 构造主机密钥校验回调（修复 MITM 漏洞）
	hostKeyCallback, err := buildHostKeyCallback(s.config)
	if err != nil {
		return fmt.Errorf("构造主机密钥校验失败: %v", err)
	}

	sshConfig := &ssh.ClientConfig{
		User:            s.config.Username,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         time.Duration(s.config.Timeout) * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	start := time.Now()
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("连接SSH服务器失败: %v", err)
	}
	elapsed := time.Since(start)
	s.latency = elapsed.Milliseconds()
	s.lastPingTime = start

	s.client = client
	s.isConnected = true

	return nil
}

// buildHostKeyCallback 根据策略构造 ssh.HostKeyCallback
//
// 安全修复：之前使用 ssh.InsecureIgnoreHostKey() 会让中间人攻击者
// 轻易伪造服务器。此函数提供三种模式：
//   - HostKeyCallbackOff:   不校验（仅用于调试，不推荐）
//   - HostKeyCallbackTOFU:  首次信任 + 写入 known_hosts（默认，向后兼容）
//   - HostKeyCallbackStrict: 拒绝任何未知主机（生产推荐）
func buildHostKeyCallback(cfg *SSHConfig) (ssh.HostKeyCallback, error) {
	// 1. 解析 known_hosts 文件路径
	khPath := cfg.KnownHostsFile
	if khPath == "" {
		// 默认使用 ~/.ssh/known_hosts，多用户系统下用 home 目录更安全
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("获取用户主目录失败: %w", err)
		}
		khPath = filepath.Join(home, ".ssh", "known_hosts")
	}

	// 2. 如果文件不存在，提前创建好（knownhosts.New 不自动创建）
	if _, err := os.Stat(khPath); os.IsNotExist(err) {
		// 确保 ~/.ssh 目录存在
		if err := os.MkdirAll(filepath.Dir(khPath), 0700); err != nil {
			return nil, fmt.Errorf("创建 known_hosts 目录失败: %w", err)
		}
		// 创建空文件
		f, err := os.OpenFile(khPath, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return nil, fmt.Errorf("创建 known_hosts 文件失败: %w", err)
		}
		f.Close()
	}

	// 3. 加载 known_hosts（可能为空文件，这是合法的）
	kh, err := knownhosts.New(khPath)
	if err != nil {
		return nil, fmt.Errorf("加载 known_hosts 失败: %w", err)
	}

	// 4. 根据策略返回不同的 callback
	policy := cfg.HostKeyPolicy
	if policy == "" {
		policy = HostKeyCallbackTOFU // 默认 ToFU，向后兼容老用户
	}

	switch policy {
	case HostKeyCallbackOff:
		// 显式 opt-in：不校验。仅用于开发调试。
		// 给个清晰的警告日志，提醒用户这不是好主意。
		fmt.Printf("[SSHClient] ⚠️ 警告：主机密钥校验已关闭，存在 MITM 攻击风险\n")
		return ssh.InsecureIgnoreHostKey(), nil

	case HostKeyCallbackStrict:
		// 严格模式：未知主机直接拒绝
		return kh, nil

	case HostKeyCallbackTOFU:
		// ToFU 模式：首次连接自动信任并写入 known_hosts
		// 注意：knownhosts.New 返回的 callback 在遇到未知主机时会返回错误，
		// 我们需要包装一下让它支持 ToFU（写入 + 通过）。
		return wrapKnownHostsForTOFU(khPath, kh), nil

	default:
		return nil, fmt.Errorf("未知的主机密钥校验策略: %s", policy)
	}
}

// wrapKnownHostsForTOFU 把 knownhosts 的"严格校验"包装成 ToFU 模式：
//   - 已知主机 → 通过
//   - 未知主机 → 写入 known_hosts 文件，然后通过
//
// 这样首次连接仍然能正常工作，但不会让用户面对一个无脑放行的 InsecureIgnore。
func wrapKnownHostsForTOFU(khPath string, strict ssh.HostKeyCallback) ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		// 先用严格模式校验
		err := strict(hostname, remote, key)
		if err == nil {
			return nil // 已知主机，通过
		}

		// 严格校验失败 → 可能是未知主机，尝试作为 ToFU 处理
		// knownhosts 的错误可能是 *knownhosts.KeyError，可提取实际 key
		var keyErr *knownhosts.KeyError
		if errors.As(err, &keyErr) && len(keyErr.Want) == 0 {
			// Want 为空 = 完全没有这个主机的记录 = 全新主机 → 写入 known_hosts
			if writeErr := knownhosts.WriteKnownHost(khPath, hostname, remote, key); writeErr != nil {
				return fmt.Errorf("写入 known_hosts 失败（拒绝信任未知主机）: %w", writeErr)
			}
			fmt.Printf("[SSHClient] ✓ ToFU：已信任新主机 %s 并写入 known_hosts\n", hostname)
			return nil
		}

		// 其他错误（如 knownhosts.KeyError.Want 不为空 = 主机 key 不匹配！）
		// 这就是经典的 MITM 攻击信号，绝不能放行
		return err
	}
}

// Close 关闭SSH连接
func (s *SSHClient) Close() {
	fmt.Printf("[SSHClient] ========== 开始关闭连接 ==========\n")
	
	// 标记为正在关闭
	s.closing = true

	// 关闭所有 Shell 会话
	s.CloseAllShells()

	// 使用 channel 来跟踪关闭状态
	done := make(chan struct{})

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("[SSHClient] ⚠️ Close panic recovered: %v\n", r)
			}
			close(done)
		}()

		// 关闭 SFTP
		if s.sftpClient != nil {
			s.sftpClient.Close()
		}

		// 关闭 SSH Client
		if s.client != nil {
			s.client.Close()
		}
	}()
	
	// 等待关闭完成，最多等待2秒
	select {
	case <-done:
		fmt.Printf("[SSHClient] ========== 连接已完全关闭 ==========\n")
	case <-time.After(2 * time.Second):
		fmt.Printf("[SSHClient] ⚠️ 关闭超时，强制继续\n")
	}
	
	s.isConnected = false
}

// IsConnected 检查是否已连接
func (s *SSHClient) IsConnected() bool {
	return s.isConnected
}

// ExecuteCommand 执行远程命令
func (s *SSHClient) ExecuteCommand(command string) (*CommandResult, error) {
	if !s.isConnected {
		return nil, fmt.Errorf("未连接到SSH服务器")
	}

	session, err := s.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("创建会话失败: %v", err)
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("获取标准输出管道失败: %v", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("获取标准错误管道失败: %v", err)
	}

	if err := session.Start(command); err != nil {
		return nil, fmt.Errorf("启动命令失败: %v", err)
	}

	var stdoutBuf, stderrBuf []byte
	stdoutBuf, _ = io.ReadAll(stdout)
	stderrBuf, _ = io.ReadAll(stderr)

	err = session.Wait()
	exitCode := 0
	success := true
	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			exitCode = exitErr.ExitStatus()
			success = false
		} else {
			return nil, fmt.Errorf("等待命令执行失败: %v", err)
		}
	}

	return &CommandResult{
		Stdout:   string(stdoutBuf),
		Stderr:   string(stderrBuf),
		ExitCode: exitCode,
		Success:  success,
	}, nil
}

// StartPortForward 启动本地端口转发
func (s *SSHClient) StartPortForward(config *PortForwardConfig) error {
	if !s.isConnected {
		return fmt.Errorf("未连接到SSH服务器")
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", config.LocalPort))
	if err != nil {
		return fmt.Errorf("监听本地端口失败: %v", err)
	}

	go func() {
		defer listener.Close()
		for {
			localConn, err := listener.Accept()
			if err != nil {
				return
			}

			remoteAddr := fmt.Sprintf("%s:%d", config.RemoteHost, config.RemotePort)
			remoteConn, err := s.client.Dial("tcp", remoteAddr)
			if err != nil {
				localConn.Close()
				continue
			}

			// 双向数据传输
			go func() {
				io.Copy(localConn, remoteConn)
				localConn.Close()
				remoteConn.Close()
			}()
			go func() {
				io.Copy(remoteConn, localConn)
				localConn.Close()
				remoteConn.Close()
			}()
		}
	}()

	return nil
}

// CreateShell 创建独立的 Shell 会话
func (s *SSHClient) CreateShell(sessionID string) (*ssh.Session, error) {
	if !s.isConnected {
		return nil, fmt.Errorf("未连接到SSH服务器")
	}

	// 关闭同 ID 的旧会话
	s.CloseShell(sessionID)

	fmt.Printf("[SSHClient] 创建 Shell 会话: %s\n", sessionID)
	sess, err := s.client.NewSession()
	if err != nil {
		fmt.Printf("[SSHClient] 创建会话失败: %s, %v\n", sessionID, err)
		return nil, fmt.Errorf("创建会话失败: %v", err)
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err = sess.RequestPty("xterm-256color", 80, 40, modes); err != nil {
		sess.Close()
		return nil, fmt.Errorf("请求PTY失败: %v", err)
	}

	stdin, err := sess.StdinPipe()
	if err != nil {
		sess.Close()
		return nil, fmt.Errorf("获取stdin失败: %v", err)
	}
	stdout, err := sess.StdoutPipe()
	if err != nil {
		sess.Close()
		return nil, fmt.Errorf("获取stdout失败: %v", err)
	}
	if err = sess.Shell(); err != nil {
		sess.Close()
		return nil, fmt.Errorf("启动Shell失败: %v", err)
	}

	sh := &shellSession{session: sess, stdin: stdin, stdout: stdout}
	s.sessMu.Lock()
	s.sessions[sessionID] = sh
	s.sessMu.Unlock()

	go s.readShellOutput(sessionID, sh)
	fmt.Printf("[SSHClient] ✓ Shell 会话已启动: %s (总 sessions: %d)\n", sessionID, len(s.sessions))
	return sess, nil
}

// readShellOutput 持续读取单个 session 的输出
func (s *SSHClient) readShellOutput(sessionID string, sh *shellSession) {
	buf := make([]byte, 4096)
	for atomic.LoadInt32(&sh.closing) == 0 {
		n, err := sh.stdout.Read(buf)
		if err != nil {
			if atomic.LoadInt32(&sh.closing) == 0 {
				fmt.Printf("[SSHClient] Session %s 输出读取退出: %v\n", sessionID, err)
				// 检测到连接断开（EOF 或网络错误）
				if err == io.EOF || !s.closing {
					s.handleDisconnect()
				}
			}
			break
		}
		if n > 0 {
			sh.bufMutex.Lock()
			sh.readBuf = append(sh.readBuf, buf[:n]...)
			sh.bufMutex.Unlock()
		}
	}
}

// WriteToShell 向指定 session 写入数据
func (s *SSHClient) WriteToShell(sessionID string, data []byte) error {
	s.sessMu.RLock()
	sh, ok := s.sessions[sessionID]
	s.sessMu.RUnlock()
	if !ok || sh.stdin == nil {
		return fmt.Errorf("Shell会话未初始化: %s", sessionID)
	}
	_, err := sh.stdin.Write(data)
	return err
}

// ReadFromShell 从指定 session 读取数据
func (s *SSHClient) ReadFromShell(sessionID string, buf []byte) (int, error) {
	s.sessMu.RLock()
	sh, ok := s.sessions[sessionID]
	s.sessMu.RUnlock()
	if !ok {
		return 0, fmt.Errorf("Shell会话不存在: %s", sessionID)
	}
	sh.bufMutex.Lock()
	defer sh.bufMutex.Unlock()
	if len(sh.readBuf) == 0 {
		return 0, nil
	}
	n := copy(buf, sh.readBuf)
	sh.readBuf = sh.readBuf[n:]
	return n, nil
}

// IsShellActive 检查 session 是否活跃
func (s *SSHClient) IsShellActive(sessionID string) bool {
	s.sessMu.RLock()
	sh, ok := s.sessions[sessionID]
	s.sessMu.RUnlock()
	return ok && atomic.LoadInt32(&sh.closing) == 0
}

// CloseShell 关闭指定 session
func (s *SSHClient) CloseShell(sessionID string) {
	s.sessMu.Lock()
	sh, ok := s.sessions[sessionID]
	if ok {
		atomic.StoreInt32(&sh.closing, 1)
		delete(s.sessions, sessionID)
	}
	s.sessMu.Unlock()
	if ok && sh.session != nil {
		sh.session.Close()
	}
}

// CloseAllShells 关闭所有 session
func (s *SSHClient) CloseAllShells() {
	s.sessMu.Lock()
	sessions := make(map[string]*shellSession, len(s.sessions))
	for k, v := range s.sessions {
		sessions[k] = v
	}
	s.sessions = make(map[string]*shellSession)
	s.sessMu.Unlock()
	for _, sh := range sessions {
		atomic.StoreInt32(&sh.closing, 1)
		if sh.session != nil {
			sh.session.Close()
		}
	}
}

// GetSessionIDs 获取所有活跃 session ID
func (s *SSHClient) GetSessionIDs() []string {
	s.sessMu.RLock()
	defer s.sessMu.RUnlock()
	ids := make([]string, 0, len(s.sessions))
	for id := range s.sessions {
		ids = append(ids, id)
	}
	return ids
}

// UpdateLatency 更新延迟（通过执行简单命令测量）
func (s *SSHClient) UpdateLatency() int64 {
	if !s.isConnected || s.client == nil {
		return 0
	}
	
	// 通过执行一个简单命令来测量延迟
	start := time.Now()
	session, err := s.client.NewSession()
	if err != nil {
		return s.latency // 返回旧值
	}
	defer session.Close()
	
	// 执行一个简单的命令（true 是 bash 内置命令，非常快）
	err = session.Run("true")
	elapsed := time.Since(start)
	
	if err == nil {
		s.latency = elapsed.Milliseconds()
		s.lastPingTime = start
	}
	
	return s.latency
}

// ResizeTerminal 调整终端大小（默认 session）
func (s *SSHClient) ResizeTerminal(cols, rows int) error {
	return s.ResizeTerminalByID("default", cols, rows)
}

// ResizeTerminalByID 调整指定 session 的终端大小
func (s *SSHClient) ResizeTerminalByID(sessionID string, cols, rows int) error {
	s.sessMu.RLock()
	sh, ok := s.sessions[sessionID]
	s.sessMu.RUnlock()
	if !ok || sh.session == nil {
		return fmt.Errorf("会话未初始化: %s", sessionID)
	}
	return sh.session.WindowChange(rows, cols)
}

// GetLatency 获取连接延迟（实时测量）
func (s *SSHClient) GetLatency() int64 {
	return s.latency
}

// loadPrivateKey 加载私钥文件
func loadPrivateKey(keyPath string) (ssh.AuthMethod, error) {
	buffer, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("读取密钥文件失败: %v", err)
	}
	return parsePrivateKeyContent(buffer)
}

// parsePrivateKeyContent 解析私钥内容
func parsePrivateKeyContent(buffer []byte) (ssh.AuthMethod, error) {
	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil, fmt.Errorf("解析私钥失败: %v", err)
	}
	return ssh.PublicKeys(key), nil
}
