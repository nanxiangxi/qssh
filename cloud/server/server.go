package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"changeme/cloud/server/crypto"
)

// RegisterRequest 设备注册请求
type RegisterRequest struct {
	Name      string `json:"name"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	OS        string `json:"os"`
	Version   string `json:"version"`
	Token     string `json:"token"`
	Timestamp string `json:"timestamp"` // RFC3339 格式，防重放
	Nonce     string `json:"nonce"`     // 随机值，防重放
}

// Device 已注册设备
type Device struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Host     string    `json:"host"`
	Port     int       `json:"port"`
	OS       string    `json:"os"`
	Version  string    `json:"version"`
	Status   string    `json:"status"`
	LastSeen time.Time `json:"lastSeen"`
}

// SyncData 同步数据
type SyncData struct {
	Connections []SyncConnection `json:"connections"`
	UpdatedAt   time.Time        `json:"updatedAt"`
	DeviceID    string           `json:"deviceId"`
}

// SyncConnection 同步的连接配置（密码已加密）
type SyncConnection struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	Username  string    `json:"username"`
	Password  string    `json:"password,omitempty"` // AES-256-GCM 加密
	KeyPath   string    `json:"keyPath,omitempty"`
	Source    string    `json:"source,omitempty"`    // 来源设备 ID
	UpdatedAt time.Time `json:"updatedAt,omitempty"` // 最后更新时间
}

// Status 服务状态
type Status struct {
	Running     bool      `json:"running"`
	Port        int       `json:"port"`
	DeviceCount int       `json:"deviceCount"`
	SyncCount   int       `json:"syncCount"`
	StartedAt   time.Time `json:"startedAt"`
	Version     string    `json:"version"`
}

// Server 云端服务器
type Server struct {
	mu           sync.RWMutex
	config       *Config
	dataDir      string
	devices      map[string]*Device
	syncData     *SyncData
	clients      map[string]net.Conn
	startedAt    time.Time
	tokenMgr     *crypto.TokenManager
	encryptor    *crypto.Encryptor
	usedNonces   map[string]time.Time // 防重放
	nonceMutex   sync.RWMutex
}

// New 创建服务器实例
func New(config *Config, dataDir string) *Server {
	s := &Server{
		config:     config,
		dataDir:    dataDir,
		devices:    make(map[string]*Device),
		clients:    make(map[string]net.Conn),
		startedAt:  time.Now(),
		tokenMgr:   crypto.NewTokenManager(config.Token),
		encryptor:  crypto.NewEncryptor([]byte(config.Token)),
		usedNonces: make(map[string]time.Time),
	}

	// 加载同步数据
	s.loadSyncData()

	// 启动心跳检测
	go s.heartbeatLoop()

	// 启动 nonce 清理
	go s.nonceCleanupLoop()

	return s
}

// VerifyToken 验证令牌
func VerifyToken(expected, provided string) bool {
	tm := crypto.NewTokenManager(expected)
	ok, _ := tm.Verify(provided, "api")
	return ok
}

// GetStatus 获取服务状态
func (s *Server) GetStatus() Status {
	s.mu.RLock()
	defer s.mu.RUnlock()

	onlineCount := 0
	for _, d := range s.devices {
		if d.Status == "online" {
			onlineCount++
		}
	}

	syncCount := 0
	if s.syncData != nil {
		syncCount = len(s.syncData.Connections)
	}

	return Status{
		Running:     true,
		Port:        s.config.Port,
		DeviceCount: onlineCount,
		SyncCount:   syncCount,
		StartedAt:   s.startedAt,
	}
}

// RegisterDevice 注册设备
func (s *Server) RegisterDevice(req RegisterRequest) (*Device, error) {
	// 验证时间戳（防重放攻击）
	if !crypto.ValidateTimestamp(req.Timestamp) {
		return nil, fmt.Errorf("请求已过期或时间戳无效")
	}

	// 验证 nonce（防重放攻击）
	if req.Nonce == "" {
		return nil, fmt.Errorf("缺少 nonce")
	}
	if s.isNonceUsed(req.Nonce) {
		return nil, fmt.Errorf("请求已处理（nonce 重复）")
	}
	s.markNonceUsed(req.Nonce)

	// 验证令牌
	ok, lockout := s.tokenMgr.Verify(req.Token, req.Host)
	if !ok {
		if lockout > 0 {
			return nil, fmt.Errorf("认证失败，IP 已锁定 %v", lockout.Round(time.Minute))
		}
		return nil, fmt.Errorf("认证失败")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 查找或创建设备
	var device *Device
	for _, d := range s.devices {
		if d.Host == req.Host && d.Name == req.Name {
			device = d
			break
		}
	}

	if device == nil {
		device = &Device{
			ID: crypto.GenerateDeviceID(),
		}
		s.devices[device.ID] = device
	}

	device.Name = req.Name
	device.Host = req.Host
	device.Port = req.Port
	device.OS = req.OS
	device.Version = req.Version
	device.Status = "online"
	device.LastSeen = time.Now()

	fmt.Printf("[Server] 设备注册: %s (%s)\n", device.Name, device.Host)
	return device, nil
}

// GetDevices 获取设备列表
func (s *Server) GetDevices() []Device {
	s.mu.RLock()
	defer s.mu.RUnlock()

	devices := make([]Device, 0, len(s.devices))
	for _, d := range s.devices {
		devices = append(devices, *d)
	}
	return devices
}

// GetConfig 获取配置
func (s *Server) GetConfig() *Config {
	return s.config
}

// UpdateDeviceHeartbeat 更新设备心跳
func (s *Server) UpdateDeviceHeartbeat(deviceID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if device, ok := s.devices[deviceID]; ok {
		device.Status = "online"
		device.LastSeen = time.Now()
	}
}

// MarkDeviceOffline 标记设备离线
func (s *Server) MarkDeviceOffline(deviceID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if device, ok := s.devices[deviceID]; ok {
		device.Status = "offline"
		fmt.Printf("[Server] 设备离线: %s (%s)\n", device.Name, deviceID)
	}
}

// GetSyncData 获取同步数据（密码已解密）
func (s *Server) GetSyncData() *SyncData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.syncData == nil {
		return &SyncData{Connections: []SyncConnection{}}
	}

	// 解密密码
	result := *s.syncData
	result.Connections = make([]SyncConnection, len(s.syncData.Connections))
	for i, conn := range s.syncData.Connections {
		result.Connections[i] = conn
		if conn.Password != "" {
			decrypted, err := s.encryptor.Decrypt(conn.Password)
			if err == nil {
				result.Connections[i].Password = string(decrypted)
			}
		}
	}

	return &result
}

// UpdateSyncData 更新同步数据（智能合并：保留所有设备的连接，按时间戳解决冲突）
func (s *Server) UpdateSyncData(data SyncData) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	// 构建现有连接的 map（key = host:port）
	existing := make(map[string]SyncConnection)
	if s.syncData != nil {
		for _, c := range s.syncData.Connections {
			key := fmt.Sprintf("%s:%d", c.Host, c.Port)
			existing[key] = c
		}
	}

	// 合并新数据
	for _, c := range data.Connections {
		key := fmt.Sprintf("%s:%d", c.Host, c.Port)

		// 设置来源和时间
		if c.Source == "" {
			c.Source = data.DeviceID
		}
		if c.UpdatedAt.IsZero() {
			c.UpdatedAt = now
		}

		// 冲突解决：只在新数据更新时覆盖
		if old, exists := existing[key]; exists {
			// 同一来源：直接覆盖
			if old.Source == c.Source {
				existing[key] = c
			} else if c.UpdatedAt.After(old.UpdatedAt) {
				// 不同来源：更新的胜出
				existing[key] = c
			}
			// 否则保留旧数据
		} else {
			// 新连接：直接添加
			existing[key] = c
		}

		// 加密密码
		if existing[key].Password != "" {
			encrypted, err := s.encryptor.Encrypt([]byte(existing[key].Password))
			if err == nil {
				entry := existing[key]
				entry.Password = encrypted
				existing[key] = entry
			}
		}
	}

	// 转换为切片
	connections := make([]SyncConnection, 0, len(existing))
	for _, c := range existing {
		connections = append(connections, c)
	}

	s.syncData = &SyncData{
		Connections: connections,
		UpdatedAt:   now,
		DeviceID:    "server",
	}

	s.saveSyncData()
	fmt.Printf("[Server] 同步数据已合并: %d 个连接（来自 %d 个设备）\n", len(connections), countSources(connections))
}

// countSources 统计来源设备数
func countSources(conns []SyncConnection) int {
	sources := make(map[string]bool)
	for _, c := range conns {
		if c.Source != "" {
			sources[c.Source] = true
		}
	}
	return len(sources)
}

// HandleWebSocket WebSocket 处理
func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "WebSocket not implemented yet", http.StatusNotImplemented)
}

// heartbeatLoop 心跳检测
func (s *Server) heartbeatLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		for id, device := range s.devices {
			if time.Since(device.LastSeen) > 2*time.Minute {
				device.Status = "offline"
				fmt.Printf("[Server] 设备离线: %s\n", id)
			}
		}
		s.mu.Unlock()
	}
}

// isNonceUsed 检查 nonce 是否已使用
func (s *Server) isNonceUsed(nonce string) bool {
	s.nonceMutex.RLock()
	defer s.nonceMutex.RUnlock()
	_, exists := s.usedNonces[nonce]
	return exists
}

// markNonceUsed 标记 nonce 已使用
func (s *Server) markNonceUsed(nonce string) {
	s.nonceMutex.Lock()
	defer s.nonceMutex.Unlock()
	s.usedNonces[nonce] = time.Now()
}

// nonceCleanupLoop 清理过期 nonce
func (s *Server) nonceCleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.nonceMutex.Lock()
		for nonce, t := range s.usedNonces {
			if time.Since(t) > 10*time.Minute {
				delete(s.usedNonces, nonce)
			}
		}
		s.nonceMutex.Unlock()
	}
}

// loadSyncData 加载同步数据
func (s *Server) loadSyncData() {
	path := s.dataDir + "/sync.json"
	data, err := os.ReadFile(path)
	if err != nil {
		s.syncData = &SyncData{
			Connections: []SyncConnection{},
			UpdatedAt:   time.Now(),
		}
		return
	}

	var syncData SyncData
	if err := json.Unmarshal(data, &syncData); err != nil {
		s.syncData = &SyncData{
			Connections: []SyncConnection{},
			UpdatedAt:   time.Now(),
		}
		return
	}

	s.syncData = &syncData
	fmt.Printf("[Server] 已加载同步数据: %d 个连接\n", len(syncData.Connections))
}

// saveSyncData 保存同步数据
func (s *Server) saveSyncData() {
	path := s.dataDir + "/sync.json"
	data, err := json.MarshalIndent(s.syncData, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(path, data, 0644)
}

// generateID 生成唯一ID
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
