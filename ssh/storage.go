package ssh

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// ConnectionInfo SSH 连接信息。
// 落盘时 Password / PrivateKey 始终以加密形式存储在对应 Cipher 字段中。
type ConnectionInfo struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"-"` // 内存明文，永不序列化
	KeyPath    string `json:"keyPath,omitempty"`
	PrivateKey string `json:"-"` // 内存明文，永不序列化

	PasswordCiphertext   string `json:"passwordCiphertext,omitempty"`
	PrivateKeyCiphertext string `json:"privateKeyCiphertext,omitempty"`
	CipherKind           string `json:"cipherKind,omitempty"`

	// 跳板机配置
	JumpHost     string `json:"jumpHost,omitempty"`
	JumpUsername string `json:"jumpUsername,omitempty"`
	JumpPassword string `json:"-"` // 内存明文，永不序列化
	JumpKeyPath  string `json:"jumpKeyPath,omitempty"`

	JumpPasswordCiphertext string `json:"jumpPasswordCiphertext,omitempty"`

	Status  string `json:"status"`
	Saved   bool   `json:"saved"`
	GroupID string `json:"group_id,omitempty"`
}

// SecretCipher 抽象加密器
type SecretCipher interface {
	Kind() string
	Encrypt(plain string) (string, error)
	Decrypt(cipher string) (string, error)
}

// StorageManager 连接存储管理器（统一单文件 + 全量加密）。
type StorageManager struct {
	mu          sync.RWMutex
	connections map[string]*ConnectionInfo
	dataFile    string
	cipher      SecretCipher
}

// NewStorageManager 创建存储管理器。
// 数据文件路径: {GetDataDir()}/connections.json
func NewStorageManager() *StorageManager {
	dataDir := GetDataDir()
	dataFile := filepath.Join(dataDir, "connections.json")
	cipher := newPlatformCipher()

	sm := &StorageManager{
		connections: make(map[string]*ConnectionInfo),
		dataFile:    dataFile,
		cipher:      cipher,
	}

	sm.loadConnections()
	return sm
}

// loadConnections 从文件加载连接（自动解密）。
func (sm *StorageManager) loadConnections() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 1. 尝试从主文件加载
	data, err := os.ReadFile(sm.dataFile)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Error("读取连接文件失败", "error", err)
		}
		// 2. 主文件不存在，尝试从旧文件迁移
		sm.migrateFromLegacy()
		return
	}

	var conns []*ConnectionInfo
	if err := json.Unmarshal(data, &conns); err != nil {
		log.Error("解析连接文件失败，尝试旧格式迁移", "error", err)
		sm.migrateFromLegacy()
		return
	}

	// 检查是否为旧格式（ID 为空 = SavedConnection 格式）
	hasValidIDs := false
	for _, c := range conns {
		if c.ID != "" {
			hasValidIDs = true
			break
		}
	}

	if !hasValidIDs && len(conns) > 0 {
		// 旧格式：SavedConnection 没有 ID 字段，需要迁移
		log.Info("检测到旧格式连接文件，执行迁移", "count", len(conns))
		for _, c := range conns {
			c.ID = generateConnectionID()
			c.Status = "disconnected"
			c.Saved = true
			sm.decryptInPlace(c)
			sm.connections[c.ID] = c
		}
		// 迁移后写入新格式
		sm.saveLocked()
		return
	}

	for _, c := range conns {
		c.Status = "disconnected"
		sm.decryptInPlace(c)
		sm.connections[c.ID] = c
	}
	log.Info("加载连接成功", "count", len(conns))
}

// migrateFromLegacy 从旧的 cache/permanent 双文件格式迁移。
func (sm *StorageManager) migrateFromLegacy() {
	exePath, err := os.Executable()
	if err != nil {
		return
	}
	exeDir := filepath.Dir(exePath)

	// 优先读 permanent 文件（加密的），再读 cache 文件（明文的）
	paths := []string{
		filepath.Join(exeDir, "data", "persistent", "connections.json"),
		filepath.Join(exeDir, "data", "cache", "connections.json"),
	}

	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}

		// 尝试解析为 []ConnectionInfo（旧格式）
		var legacy []*ConnectionInfo
		if err := json.Unmarshal(data, &legacy); err != nil || len(legacy) == 0 {
			continue
		}

		// 迁移：解密明文字段，标记为已保存
		for _, c := range legacy {
			if c.ID == "" {
				c.ID = generateConnectionID()
			}
			c.Status = "disconnected"
			c.Saved = true
			sm.decryptInPlace(c)
			sm.connections[c.ID] = c
		}

		log.Info("从旧格式迁移连接", "source", p, "count", len(legacy))
		// 迁移后立即写入新格式
		sm.saveLocked()
		return
	}
}

// decryptInPlace 解密 Enc 字段回填到内存明文字段。
func (sm *StorageManager) decryptInPlace(c *ConnectionInfo) {
	if c.CipherKind == "" {
		return // 旧明文数据，无加密字段
	}
	if c.PasswordCiphertext != "" {
		if plain, err := sm.cipher.Decrypt(c.PasswordCiphertext); err == nil {
			c.Password = plain
		} else {
			log.Error("解密密码失败", "id", c.ID, "error", err)
		}
	}
	if c.PrivateKeyCiphertext != "" {
		if plain, err := sm.cipher.Decrypt(c.PrivateKeyCiphertext); err == nil {
			c.PrivateKey = plain
		} else {
			log.Error("解密私钥失败", "id", c.ID, "error", err)
		}
	}
	if c.JumpPasswordCiphertext != "" {
		if plain, err := sm.cipher.Decrypt(c.JumpPasswordCiphertext); err == nil {
			c.JumpPassword = plain
		} else {
			log.Error("解密跳板机密码失败", "id", c.ID, "error", err)
		}
	}
}

// encryptForWrite 加密内存明文字段到 Cipher 字段（用于写盘）。
func (sm *StorageManager) encryptForWrite(c *ConnectionInfo) *ConnectionInfo {
	cp := *c
	cp.Password = ""
	cp.PrivateKey = ""
	cp.JumpPassword = ""

	if c.Password != "" {
		if enc, err := sm.cipher.Encrypt(c.Password); err == nil {
			cp.PasswordCiphertext = enc
			cp.CipherKind = sm.cipher.Kind()
		} else {
			log.Error("加密密码失败", "id", c.ID, "error", err)
		}
	}
	if c.PrivateKey != "" {
		if enc, err := sm.cipher.Encrypt(c.PrivateKey); err == nil {
			cp.PrivateKeyCiphertext = enc
			cp.CipherKind = sm.cipher.Kind()
		} else {
			log.Error("加密私钥失败", "id", c.ID, "error", err)
		}
	}
	if c.JumpPassword != "" {
		if enc, err := sm.cipher.Encrypt(c.JumpPassword); err == nil {
			cp.JumpPasswordCiphertext = enc
			cp.CipherKind = sm.cipher.Kind()
		} else {
			log.Error("加密跳板机密码失败", "id", c.ID, "error", err)
		}
	}
	return &cp
}

// saveLocked 将当前所有连接加密后写入文件（调用者需持有写锁）。
func (sm *StorageManager) saveLocked() error {
	dir := filepath.Dir(sm.dataFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建数据目录失败: %w", err)
	}

	var list []*ConnectionInfo
	for _, c := range sm.connections {
		if !c.Saved {
			continue // 未保存的临时连接不落盘
		}
		list = append(list, sm.encryptForWrite(c))
	}

	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化连接数据失败: %w", err)
	}

	if err := os.WriteFile(sm.dataFile, data, 0644); err != nil {
		return fmt.Errorf("写入连接文件失败: %w", err)
	}
	return nil
}

// AddConnection 添加新连接（默认未保存）。
func (sm *StorageManager) AddConnection(conn *ConnectionInfo) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if conn.ID == "" {
		return fmt.Errorf("连接 ID 不能为空")
	}

	conn.Status = "disconnected"
	conn.Saved = false
	sm.connections[conn.ID] = conn
	return sm.saveLocked()
}

// UpdateConnection 更新连接信息。
func (sm *StorageManager) UpdateConnection(conn *ConnectionInfo) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	existing, ok := sm.connections[conn.ID]
	if !ok {
		return fmt.Errorf("连接不存在: %s", conn.ID)
	}

	conn.Status = existing.Status
	conn.Saved = existing.Saved
	sm.connections[conn.ID] = conn
	return sm.saveLocked()
}

// SyncImportConnection 云端同步导入（按 host:port:username 去重）。
func (sm *StorageManager) SyncImportConnection(conn *ConnectionInfo) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, existing := range sm.connections {
		if existing.Host == conn.Host && existing.Port == conn.Port && existing.Username == conn.Username {
			existing.Name = conn.Name
			if conn.Password != "" {
				existing.Password = conn.Password
			}
			if conn.PrivateKey != "" {
				existing.PrivateKey = conn.PrivateKey
			}
			existing.Saved = true
			return sm.saveLocked()
		}
	}

	conn.ID = generateConnectionID()
	conn.Status = "disconnected"
	conn.Saved = true
	sm.connections[conn.ID] = conn
	return sm.saveLocked()
}

// DeleteConnection 删除连接。
func (sm *StorageManager) DeleteConnection(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, ok := sm.connections[id]; !ok {
		return fmt.Errorf("连接不存在: %s", id)
	}
	delete(sm.connections, id)
	return sm.saveLocked()
}

// DeleteFromCache 从缓存中移除（未保存的连接从内存删除，已保存的仅重置状态）。
func (sm *StorageManager) DeleteFromCache(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	conn, ok := sm.connections[id]
	if !ok {
		return fmt.Errorf("连接不存在: %s", id)
	}

	if conn.Saved {
		conn.Status = "disconnected"
	} else {
		delete(sm.connections, id)
	}
	return sm.saveLocked()
}

// GetConnection 获取单个连接（返回副本）。
func (sm *StorageManager) GetConnection(id string) (*ConnectionInfo, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	conn, ok := sm.connections[id]
	if !ok {
		return nil, fmt.Errorf("连接不存在: %s", id)
	}
	cp := *conn
	return &cp, nil
}

// GetAllConnections 获取所有连接（返回副本列表）。
func (sm *StorageManager) GetAllConnections() []*ConnectionInfo {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make([]*ConnectionInfo, 0, len(sm.connections))
	for _, c := range sm.connections {
		cp := *c
		result = append(result, &cp)
	}
	return result
}

// UpdateConnectionStatus 更新连接状态。
func (sm *StorageManager) UpdateConnectionStatus(id string, status string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if c, ok := sm.connections[id]; ok {
		c.Status = status
	}
}

// MarkAsSaved 标记为已保存并写盘。
func (sm *StorageManager) MarkAsSaved(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	c, ok := sm.connections[id]
	if !ok {
		return fmt.Errorf("连接不存在: %s", id)
	}
	c.Saved = true
	return sm.saveLocked()
}

// MarkAsUnsaved 标记为未保存。
func (sm *StorageManager) MarkAsUnsaved(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if c, ok := sm.connections[id]; ok {
		c.Saved = false
	}
}

// SaveToPermanent 保持向后兼容（实际已统一为 saveLocked）。
func (sm *StorageManager) SaveToPermanent() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.saveLocked()
}

// LoadFromPermanent 保持向后兼容（实际已统一为 loadConnections）。
func (sm *StorageManager) LoadFromPermanent() error {
	return nil
}
