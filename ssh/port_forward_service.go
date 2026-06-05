package ssh

import (
	"fmt"
	"io"
	"net"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// PortForward 端口转发规则
type PortForward struct {
	ID         string `json:"id"`
	ConnID     string `json:"connId"`
	Type       string `json:"type"`       // "local" / "remote"
	BindAddr   string `json:"bindAddr"`
	BindPort   int    `json:"bindPort"`
	RemoteHost string `json:"remoteHost"`
	RemotePort int    `json:"remotePort"`
	Status     string `json:"status"` // "running" / "stopped" / "error"
	Error      string `json:"error,omitempty"`
	CreatedAt  int64  `json:"createdAt"`

	// 流量统计
	ActiveConns int64 `json:"activeConns"` // 当前活跃连接数
	TotalConns  int64 `json:"totalConns"`  // 历史总连接数
	BytesSent   int64 `json:"bytesSent"`   // 发送字节数（上行）
	BytesRecv   int64 `json:"bytesRecv"`   // 接收字节数（下行）
}

// countingConn 包装 net.Conn，统计流量
type countingConn struct {
	net.Conn
	bytesSent *int64
	bytesRecv *int64
}

func (c *countingConn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	if n > 0 {
		atomic.AddInt64(c.bytesRecv, int64(n))
	}
	return n, err
}

func (c *countingConn) Write(b []byte) (int, error) {
	n, err := c.Conn.Write(b)
	if n > 0 {
		atomic.AddInt64(c.bytesSent, int64(n))
	}
	return n, err
}

// forwardState 每个转发的运行时状态
type forwardState struct {
	mu        sync.Mutex
	activeConns int64
	totalConns  int64
	bytesSent   int64
	bytesRecv   int64
	done        chan struct{}
}

// PortForwardService 端口转发服务
type PortForwardService struct {
	mu        sync.RWMutex
	forwards  map[string]*PortForward  // id → forward
	listeners map[string]net.Listener  // id → listener
	states    map[string]*forwardState // id → runtime state
	sshSvc    *SSHService
	app       *application.App
}

// NewPortForwardService 创建端口转发服务
func NewPortForwardService(sshSvc *SSHService) *PortForwardService {
	svc := &PortForwardService{
		forwards:  make(map[string]*PortForward),
		listeners: make(map[string]net.Listener),
		states:    make(map[string]*forwardState),
		sshSvc:    sshSvc,
		app:       sshSvc.GetApp(),
	}
	// 启动定时统计广播
	go svc.statsBroadcastLoop()
	return svc
}

// statsBroadcastLoop 每 2 秒广播一次流量统计
func (s *PortForwardService) statsBroadcastLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		s.broadcastAllStats()
	}
}

// broadcastAllStats 广播所有活跃转发的统计
func (s *PortForwardService) broadcastAllStats() {
	if s.app == nil {
		return
	}

	s.mu.RLock()
	connIDs := make(map[string]bool)
	for _, fwd := range s.forwards {
		if fwd.Status == "running" {
			connIDs[fwd.ConnID] = true
		}
	}
	s.mu.RUnlock()

	for connID := range connIDs {
		forwards := s.getForwardsWithStats(connID)
		if len(forwards) > 0 {
			s.app.Event.Emit("port-forward:status", map[string]interface{}{
				"connId":   connID,
				"forwards": forwards,
			})
		}
	}
}

// getForwardsWithStats 获取转发列表（含实时统计，按创建时间排序）
func (s *PortForwardService) getForwardsWithStats(connID string) []PortForward {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []PortForward
	for _, fwd := range s.forwards {
		if fwd.ConnID == connID {
			copy := *fwd
			if st, ok := s.states[fwd.ID]; ok {
				copy.ActiveConns = atomic.LoadInt64(&st.activeConns)
				copy.TotalConns = atomic.LoadInt64(&st.totalConns)
				copy.BytesSent = atomic.LoadInt64(&st.bytesSent)
				copy.BytesRecv = atomic.LoadInt64(&st.bytesRecv)
			}
			result = append(result, copy)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt < result[j].CreatedAt
	})

	return result
}

// generateForwardID 生成转发 ID
func generateForwardID() string {
	return fmt.Sprintf("fwd_%d", time.Now().UnixNano())
}

// GetForwards 获取指定连接的所有转发规则
func (s *PortForwardService) GetForwards(connID string) []PortForward {
	return s.getForwardsWithStats(connID)
}

// AddLocalForward 添加本地端口转发
// bindPort=0 时由操作系统自动分配可用端口
func (s *PortForwardService) AddLocalForward(connID, bindAddr string, bindPort int, remoteHost string, remotePort int) (*PortForward, error) {
	if bindPort < 0 || bindPort > 65535 {
		return nil, fmt.Errorf("本地端口无效: %d", bindPort)
	}
	if remotePort <= 0 || remotePort > 65535 {
		return nil, fmt.Errorf("远程端口无效: %d", remotePort)
	}
	if bindAddr == "" {
		bindAddr = "127.0.0.1"
	}
	if remoteHost == "" {
		remoteHost = "127.0.0.1"
	}

	// 检查端口冲突
	s.mu.RLock()
	for _, fwd := range s.forwards {
		if fwd.ConnID == connID && fwd.BindAddr == bindAddr && fwd.BindPort == bindPort && fwd.Status == "running" {
			s.mu.RUnlock()
			return nil, fmt.Errorf("端口 %s:%d 已被占用", bindAddr, bindPort)
		}
	}
	s.mu.RUnlock()

	id := generateForwardID()
	fwd := &PortForward{
		ID:         id,
		ConnID:     connID,
		Type:       "local",
		BindAddr:   bindAddr,
		BindPort:   bindPort,
		RemoteHost: remoteHost,
		RemotePort: remotePort,
		Status:     "stopped",
		CreatedAt:  time.Now().UnixMilli(),
	}

	s.mu.Lock()
	s.forwards[id] = fwd
	s.mu.Unlock()

	s.emitStatus(connID)
	return fwd, nil
}

// AddRemoteForward 添加远程端口转发
func (s *PortForwardService) AddRemoteForward(connID, bindAddr string, bindPort int, remoteHost string, remotePort int) (*PortForward, error) {
	if bindPort <= 0 || bindPort > 65535 {
		return nil, fmt.Errorf("远程监听端口无效: %d", bindPort)
	}
	if remotePort <= 0 || remotePort > 65535 {
		return nil, fmt.Errorf("本地目标端口无效: %d", remotePort)
	}
	if bindAddr == "" {
		bindAddr = "0.0.0.0"
	}
	if remoteHost == "" {
		remoteHost = "127.0.0.1"
	}

	id := generateForwardID()
	fwd := &PortForward{
		ID:         id,
		ConnID:     connID,
		Type:       "remote",
		BindAddr:   bindAddr,
		BindPort:   bindPort,
		RemoteHost: remoteHost,
		RemotePort: remotePort,
		Status:     "stopped",
		CreatedAt:  time.Now().UnixMilli(),
	}

	s.mu.Lock()
	s.forwards[id] = fwd
	s.mu.Unlock()

	s.emitStatus(connID)
	return fwd, nil
}

// StartForward 启动转发
func (s *PortForwardService) StartForward(forwardID string) error {
	s.mu.RLock()
	fwd, exists := s.forwards[forwardID]
	s.mu.RUnlock()
	if !exists {
		return fmt.Errorf("转发规则不存在: %s", forwardID)
	}

	if fwd.Status == "running" {
		return nil
	}

	client, err := s.sshSvc.GetClient(fwd.ConnID)
	if err != nil {
		s.updateStatus(forwardID, "error", err.Error())
		return err
	}

	if !client.IsConnected() {
		s.updateStatus(forwardID, "error", "SSH连接已断开")
		return fmt.Errorf("SSH连接已断开")
	}

	// 创建运行时状态
	st := &forwardState{done: make(chan struct{})}
	s.mu.Lock()
	s.states[forwardID] = st
	s.mu.Unlock()

	switch fwd.Type {
	case "local":
		err = s.startLocalForward(fwd, client, st)
	case "remote":
		err = s.startRemoteForward(fwd, client, st)
	default:
		err = fmt.Errorf("未知转发类型: %s", fwd.Type)
	}

	if err != nil {
		s.mu.Lock()
		delete(s.states, forwardID)
		s.mu.Unlock()
		s.updateStatus(forwardID, "error", err.Error())
		return err
	}

	s.updateStatus(forwardID, "running", "")
	s.emitStatus(fwd.ConnID)
	return nil
}

// startLocalForward 启动本地转发
// 流量方向：本地客户端 → 本地监听端口 → SSH隧道 → 远程服务
// bindPort=0 时由操作系统自动分配可用端口
func (s *PortForwardService) startLocalForward(fwd *PortForward, client *SSHClient, st *forwardState) error {
	addr := fmt.Sprintf("%s:%d", fwd.BindAddr, fwd.BindPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("监听 %s 失败: %v", addr, err)
	}

	// 如果端口是 0（自动分配），更新为实际分配的端口
	if fwd.BindPort == 0 {
		actualPort := listener.Addr().(*net.TCPAddr).Port
		s.mu.Lock()
		fwd.BindPort = actualPort
		s.mu.Unlock()
		fmt.Printf("[PortForward] 自动分配本地端口: %d\n", actualPort)
	}

	s.mu.Lock()
	s.listeners[fwd.ID] = listener
	s.mu.Unlock()

	fmt.Printf("[PortForward] 本地监听已启动: %s → %s:%d (SSH隧道)\n",
		listener.Addr().String(), fwd.RemoteHost, fwd.RemotePort)

	go func() {
		for {
			localConn, err := listener.Accept()
			if err != nil {
				return
			}

			remoteAddr := fmt.Sprintf("%s:%d", fwd.RemoteHost, fwd.RemotePort)
			remoteConn, err := client.client.Dial("tcp", remoteAddr)
			if err != nil {
				fmt.Printf("[PortForward] SSH隧道拨号失败 (%s): %v\n", remoteAddr, err)
				localConn.Close()
				continue
			}

			atomic.AddInt64(&st.activeConns, 1)
			atomic.AddInt64(&st.totalConns, 1)

			wrappedLocal := &countingConn{Conn: localConn, bytesSent: &st.bytesSent, bytesRecv: &st.bytesRecv}
			wrappedRemote := &countingConn{Conn: remoteConn, bytesSent: &st.bytesSent, bytesRecv: &st.bytesRecv}

			go pipeConns(wrappedLocal, wrappedRemote, func() {
				atomic.AddInt64(&st.activeConns, -1)
			})
		}
	}()

	return nil
}

// startRemoteForward 启动远程转发（内网穿透）
// 流量方向：远程客户端 → SSH服务器监听端口 → SSH隧道 → 本地服务
func (s *PortForwardService) startRemoteForward(fwd *PortForward, client *SSHClient, st *forwardState) error {
	// 在 SSH 服务器上监听端口（使用 :port 格式，让 SSH 服务器决定绑定地址）
	listenAddr := fmt.Sprintf("0.0.0.0:%d", fwd.BindPort)
	listener, err := client.client.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("远程监听失败 (%s): %v", listenAddr, err)
	}

	s.mu.Lock()
	s.listeners[fwd.ID] = listener
	s.mu.Unlock()

	fmt.Printf("[PortForward] 远程监听已启动: %s → %s:%d (SSH服务器)\n",
		listenAddr, fwd.RemoteHost, fwd.RemotePort)

	go func() {
		for {
			remoteConn, err := listener.Accept()
			if err != nil {
				fmt.Printf("[PortForward] 远程监听 Accept 退出: %v\n", err)
				return
			}

			// 连接到本地服务
			localAddr := fmt.Sprintf("%s:%d", fwd.RemoteHost, fwd.RemotePort)
			localConn, err := net.Dial("tcp", localAddr)
			if err != nil {
				fmt.Printf("[PortForward] 连接本地服务失败 (%s): %v\n", localAddr, err)
				remoteConn.Close()
				continue
			}

			atomic.AddInt64(&st.activeConns, 1)
			atomic.AddInt64(&st.totalConns, 1)

			fmt.Printf("[PortForward] 远程转发连接建立: %s ←→ %s\n", listenAddr, localAddr)

			wrappedLocal := &countingConn{Conn: localConn, bytesSent: &st.bytesSent, bytesRecv: &st.bytesRecv}
			wrappedRemote := &countingConn{Conn: remoteConn, bytesSent: &st.bytesSent, bytesRecv: &st.bytesRecv}

			go pipeConns(wrappedRemote, wrappedLocal, func() {
				atomic.AddInt64(&st.activeConns, -1)
			})
		}
	}()

	return nil
}

// pipeConns 双向转发两个连接，任一方向结束时关闭两端
func pipeConns(a, b net.Conn, onDone func()) {
	var once sync.Once
	cleanup := func() {
		once.Do(func() {
			a.Close()
			b.Close()
			if onDone != nil {
				onDone()
			}
		})
	}

	go func() {
		defer cleanup()
		io.Copy(a, b)
	}()
	go func() {
		defer cleanup()
		io.Copy(b, a)
	}()
}

// StopForward 停止转发
func (s *PortForwardService) StopForward(forwardID string) error {
	s.mu.RLock()
	fwd, exists := s.forwards[forwardID]
	listener, hasListener := s.listeners[forwardID]
	st, hasState := s.states[forwardID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("转发规则不存在: %s", forwardID)
	}

	fmt.Printf("[PortForward] 停止转发: %s (%s %s:%d)\n", forwardID, fwd.Type, fwd.BindAddr, fwd.BindPort)

	if hasListener && listener != nil {
		listener.Close()
		fmt.Printf("[PortForward] 监听器已关闭: %s\n", forwardID)
	}

	// 关闭状态 goroutine
	if hasState && st.done != nil {
		close(st.done)
	}

	// 将最终统计写回 fwd
	if hasState {
		s.mu.Lock()
		if f, ok := s.forwards[forwardID]; ok {
			f.ActiveConns = atomic.LoadInt64(&st.activeConns)
			f.TotalConns = atomic.LoadInt64(&st.totalConns)
			f.BytesSent = atomic.LoadInt64(&st.bytesSent)
			f.BytesRecv = atomic.LoadInt64(&st.bytesRecv)
		}
		delete(s.listeners, forwardID)
		delete(s.states, forwardID)
		s.mu.Unlock()
	} else {
		s.mu.Lock()
		delete(s.listeners, forwardID)
		delete(s.states, forwardID)
		s.mu.Unlock()
	}

	s.updateStatus(forwardID, "stopped", "")
	s.emitStatus(fwd.ConnID)
	return nil
}

// RemoveForward 删除转发规则
func (s *PortForwardService) RemoveForward(forwardID string) error {
	s.mu.RLock()
	fwd, exists := s.forwards[forwardID]
	if !exists {
		s.mu.RUnlock()
		return fmt.Errorf("转发规则不存在: %s", forwardID)
	}
	connID := fwd.ConnID
	needStop := fwd.Status == "running"
	s.mu.RUnlock()

	// 先停止转发（关闭监听器）
	if needStop {
		s.StopForward(forwardID)
	}

	// 删除规则
	s.mu.Lock()
	delete(s.forwards, forwardID)
	delete(s.listeners, forwardID)
	delete(s.states, forwardID)
	s.mu.Unlock()

	s.emitStatus(connID)
	return nil
}

// GetForwardStatus 获取单个转发状态
func (s *PortForwardService) GetForwardStatus(forwardID string) *PortForward {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.forwards[forwardID]
}

// StopAllByConnID 停止指定连接的所有转发
func (s *PortForwardService) StopAllByConnID(connID string) {
	s.mu.RLock()
	var ids []string
	for id, fwd := range s.forwards {
		if fwd.ConnID == connID && fwd.Status == "running" {
			ids = append(ids, id)
		}
	}
	s.mu.RUnlock()

	for _, id := range ids {
		s.StopForward(id)
	}

	s.mu.Lock()
	for id, fwd := range s.forwards {
		if fwd.ConnID == connID {
			delete(s.forwards, id)
		}
	}
	s.mu.Unlock()
}

// updateStatus 更新转发状态
func (s *PortForwardService) updateStatus(forwardID, status, errMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if fwd, ok := s.forwards[forwardID]; ok {
		fwd.Status = status
		fwd.Error = errMsg
	}
}

// emitStatus 发送状态变更事件到前端
func (s *PortForwardService) emitStatus(connID string) {
	if s.app == nil {
		return
	}
	forwards := s.getForwardsWithStats(connID)
	s.app.Event.Emit("port-forward:status", map[string]interface{}{
		"connId":   connID,
		"forwards": forwards,
	})
}
