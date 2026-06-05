package ssh

import (
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
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
}

// shellSession 单个 shell 会话
type shellSession struct {
	session   *ssh.Session
	stdin     io.WriteCloser
	stdout    io.Reader
	readBuf   []byte
	bufMutex  sync.Mutex
	closing   bool
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

	sshConfig := &ssh.ClientConfig{
		User:            s.config.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 生产环境应使用严格的主机密钥验证
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
	for !sh.closing {
		n, err := sh.stdout.Read(buf)
		if err != nil {
			if !sh.closing {
				fmt.Printf("[SSHClient] Session %s 输出读取退出: %v\n", sessionID, err)
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
	return ok && !sh.closing
}

// CloseShell 关闭指定 session
func (s *SSHClient) CloseShell(sessionID string) {
	s.sessMu.Lock()
	sh, ok := s.sessions[sessionID]
	if ok {
		sh.closing = true
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
		sh.closing = true
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
