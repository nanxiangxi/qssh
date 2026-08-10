package ssh

import (
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
type HostKeyCallback string

const (
	HostKeyCallbackStrict HostKeyCallback = "strict"
	HostKeyCallbackTOFU   HostKeyCallback = "tofu"
	HostKeyCallbackOff    HostKeyCallback = "off"
)

// SSHConfig SSH 连接配置
type SSHConfig struct {
	Name           string          `json:"name"`
	Host           string          `json:"host"`
	Port           int             `json:"port"`
	Username       string          `json:"username"`
	Password       string          `json:"password,omitempty"`
	KeyPath        string          `json:"keyPath,omitempty"`
	PrivateKey     string          `json:"privateKey,omitempty"`
	Timeout        int             `json:"timeout,omitempty"`
	HostKeyPolicy  HostKeyCallback `json:"hostKeyPolicy,omitempty"`
	KnownHostsFile string          `json:"knownHostsFile,omitempty"`
	JumpHost       string          `json:"jumpHost,omitempty"`       // 跳板机 host:port
	JumpUsername   string          `json:"jumpUsername,omitempty"`   // 跳板机用户名
	JumpPassword   string          `json:"jumpPassword,omitempty"`   // 跳板机密码
	JumpKeyPath    string          `json:"jumpKeyPath,omitempty"`    // 跳板机密钥路径
}

// shellSession 单个 shell 会话
type shellSession struct {
	session  *ssh.Session
	stdin    io.WriteCloser
	stdout   io.Reader
	readBuf  []byte
	bufMutex sync.Mutex
	closing  atomic.Int32
}

// SSHClient SSH 客户端
type SSHClient struct {
	config     *SSHConfig
	client     *ssh.Client
	sftpClient *sftp.Client
	isConnected atomic.Int32 // 0=断开, 1=已连接
	closing     atomic.Int32 // 0=正常, 1=正在关闭

	sessions map[string]*shellSession
	sessMu   sync.RWMutex

	lastPingTime time.Time
	latency      int64

	searchCancelMap map[string]bool
	searchMutex     sync.RWMutex

	onDisconnect func(connID string)
	connID       string
}

// CommandResult 命令执行结果
type CommandResult struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exitCode"`
	Success  bool   `json:"success"`
}

// PortForwardConfig 端口转发配置
type PortForwardConfig struct {
	LocalPort  int    `json:"localPort"`
	RemoteHost string `json:"remoteHost"`
	RemotePort int    `json:"remotePort"`
}

// NewSSHClient 创建新的 SSH 客户端
func NewSSHClient(config *SSHConfig) *SSHClient {
	if config.Port == 0 {
		config.Port = 22
	}
	if config.Timeout == 0 {
		config.Timeout = 30
	}
	return &SSHClient{
		config:          config,
		sessions:        make(map[string]*shellSession),
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
	if s.closing.Load() == 1 || s.isConnected.Load() == 0 {
		return
	}
	s.isConnected.Store(0)
	if s.onDisconnect != nil {
		go s.onDisconnect(s.connID)
	}
}

// keepaliveLoop 定期发送 SSH keepalive 检测连接存活
func (s *SSHClient) keepaliveLoop() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	consecutiveFailures := 0
	const maxFailures = 3 // 连续 3 次失败判定断线（45秒）

	for range ticker.C {
		if s.closing.Load() == 1 || s.isConnected.Load() == 0 {
			return
		}
		if s.client == nil {
			return
		}

		_, _, err := s.client.SendRequest("keepalive@openssh.com", true, nil)
		if err != nil {
			consecutiveFailures++
			if consecutiveFailures >= maxFailures {
				s.handleDisconnect()
				return
			}
		} else {
			consecutiveFailures = 0
		}
	}
}

// Connect 连接到 SSH 服务器
func (s *SSHClient) Connect() error {
	var authMethods []ssh.AuthMethod

	if s.config.Password != "" {
		authMethods = append(authMethods, ssh.Password(s.config.Password))
	}

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

	var client *ssh.Client

	// 跳板机代理连接
	if s.config.JumpHost != "" {
		client, err = s.connectViaJumpHost(addr, sshConfig)
	} else {
		client, err = ssh.Dial("tcp", addr, sshConfig)
	}

	if err != nil {
		return fmt.Errorf("连接SSH服务器失败: %v", err)
	}

	s.latency = time.Since(start).Milliseconds()
	s.lastPingTime = start
	s.client = client
	s.isConnected.Store(1)

	// 启动 SSH keepalive 心跳（检测网络断开）
	go s.keepaliveLoop()

	return nil
}

// connectViaJumpHost 通过跳板机连接目标服务器
func (s *SSHClient) connectViaJumpHost(targetAddr string, targetConfig *ssh.ClientConfig) (*ssh.Client, error) {
	// 1. 构建跳板机认证
	jumpAuth := []ssh.AuthMethod{}
	if s.config.JumpPassword != "" {
		jumpAuth = append(jumpAuth, ssh.Password(s.config.JumpPassword))
	}
	if s.config.JumpKeyPath != "" {
		key, err := loadPrivateKey(s.config.JumpKeyPath)
		if err == nil {
			jumpAuth = append(jumpAuth, key)
		}
	}
	// 如果跳板机没有单独配置认证，使用目标服务器的认证
	if len(jumpAuth) == 0 {
		jumpAuth = targetConfig.Auth
	}

	jumpUser := s.config.JumpUsername
	if jumpUser == "" {
		jumpUser = s.config.Username
	}

	jumpConfig := &ssh.ClientConfig{
		User:            jumpUser,
		Auth:            jumpAuth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 跳板机通常不校验
		Timeout:         time.Duration(s.config.Timeout) * time.Second,
	}

	// 2. 连接跳板机
	jumpConn, err := ssh.Dial("tcp", s.config.JumpHost, jumpConfig)
	if err != nil {
		return nil, fmt.Errorf("连接跳板机失败: %v", err)
	}

	// 3. 通过跳板机连接目标服务器
	targetConn, err := jumpConn.Dial("tcp", targetAddr)
	if err != nil {
		jumpConn.Close()
		return nil, fmt.Errorf("通过跳板机拨号目标服务器失败: %v", err)
	}

	// 4. 建立 SSH 连接
	ncc, chans, reqs, err := ssh.NewClientConn(targetConn, targetAddr, targetConfig)
	if err != nil {
		targetConn.Close()
		jumpConn.Close()
		return nil, fmt.Errorf("通过跳板机建立SSH连接失败: %v", err)
	}

	return ssh.NewClient(ncc, chans, reqs), nil
}

// buildHostKeyCallback 根据策略构造 ssh.HostKeyCallback
func buildHostKeyCallback(cfg *SSHConfig) (ssh.HostKeyCallback, error) {
	khPath := cfg.KnownHostsFile
	if khPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("获取用户主目录失败: %w", err)
		}
		khPath = filepath.Join(home, ".ssh", "known_hosts")
	}

	if _, err := os.Stat(khPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(khPath), 0700); err != nil {
			return nil, fmt.Errorf("创建 known_hosts 目录失败: %w", err)
		}
		f, err := os.OpenFile(khPath, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return nil, fmt.Errorf("创建 known_hosts 文件失败: %w", err)
		}
		f.Close()
	}

	kh, err := knownhosts.New(khPath)
	if err != nil {
		return nil, fmt.Errorf("加载 known_hosts 失败: %w", err)
	}

	policy := cfg.HostKeyPolicy
	if policy == "" {
		policy = HostKeyCallbackTOFU
	}

	switch policy {
	case HostKeyCallbackOff:
		log.Warn("主机密钥校验已关闭，存在 MITM 攻击风险")
		return ssh.InsecureIgnoreHostKey(), nil

	case HostKeyCallbackStrict:
		return ssh.HostKeyCallback(kh), nil

	case HostKeyCallbackTOFU:
		return wrapKnownHostsForTOFU(khPath, ssh.HostKeyCallback(kh)), nil

	default:
		return nil, fmt.Errorf("未知的主机密钥校验策略: %s", policy)
	}
}

// wrapKnownHostsForTOFU 把严格校验包装成 ToFU 模式
func wrapKnownHostsForTOFU(khPath string, strict ssh.HostKeyCallback) ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		err := strict(hostname, remote, key)
		if err == nil {
			return nil
		}

		if knownhosts.IsHostUnknown(err) {
			f, openErr := os.OpenFile(khPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
			if openErr != nil {
				return fmt.Errorf("打开 known_hosts 失败: %w", openErr)
			}
			defer f.Close()
			if writeErr := knownhosts.WriteKnownHost(f, hostname, remote, key); writeErr != nil {
				return fmt.Errorf("写入 known_hosts 失败: %w", writeErr)
			}
			log.Info("ToFU: 已信任新主机", "host", hostname)
			return nil
		}

		return err
	}
}

// Close 关闭 SSH 连接
func (s *SSHClient) Close() {
	s.closing.Store(1)
	s.CloseAllShells()

	if s.sftpClient != nil {
		s.sftpClient.Close()
	}
	if s.client != nil {
		s.client.Close()
	}

	s.isConnected.Store(0)
}

// IsConnected 检查是否已连接
func (s *SSHClient) IsConnected() bool {
	return s.isConnected.Load() == 1
}

// ExecuteCommand 执行远程命令
func (s *SSHClient) ExecuteCommand(command string) (*CommandResult, error) {
	if s.isConnected.Load() == 0 {
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

	stdoutBuf, _ := io.ReadAll(stdout)
	stderrBuf, _ := io.ReadAll(stderr)

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

// CreateShell 创建独立的 Shell 会话
func (s *SSHClient) CreateShell(sessionID string) (*ssh.Session, error) {
	if s.isConnected.Load() == 0 {
		return nil, fmt.Errorf("未连接到SSH服务器")
	}

	s.CloseShell(sessionID)

	sess, err := s.client.NewSession()
	if err != nil {
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
	return sess, nil
}


// readShellOutput 持续读取单个 session 的输出
func (s *SSHClient) readShellOutput(sessionID string, sh *shellSession) {
	buf := make([]byte, 4096)
	for sh.closing.Load() == 0 {
		n, err := sh.stdout.Read(buf)
		if err != nil {
			if sh.closing.Load() == 0 {
				if err == io.EOF || s.closing.Load() == 0 {
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
	return ok && sh.closing.Load() == 0
}

// CloseShell 关闭指定 session
func (s *SSHClient) CloseShell(sessionID string) {
	s.sessMu.Lock()
	sh, ok := s.sessions[sessionID]
	if ok {
		sh.closing.Store(1)
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
		sh.closing.Store(1)
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

// UpdateLatency 更新延迟（使用轻量 SSH request 测量，不创建新会话）
func (s *SSHClient) UpdateLatency() int64 {
	if s.isConnected.Load() == 0 || s.client == nil {
		return 0
	}

	start := time.Now()
	_, _, err := s.client.SendRequest("keepalive@openssh.com", true, nil)
	if err == nil {
		s.latency = time.Since(start).Milliseconds()
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

// GetLatency 获取连接延迟
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
