package ssh

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// ConnectionInfo SSH连接信息
type ConnectionInfo struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password,omitempty"`
	KeyPath    string `json:"keyPath,omitempty"`
	PrivateKey string `json:"privateKey,omitempty"`
	Status     string `json:"status"`
	Saved      bool   `json:"saved"`
	GroupID    string `json:"group_id,omitempty"`
}

// StorageManager 连接存储管理器
type StorageManager struct {
	mu          sync.RWMutex
	connections map[string]*ConnectionInfo
	dataFile    string
	permanentDataFile string // 永久保存的文件路径
	encryptKey  []byte // 加密密钥
}

// encryptString 加密字符串
func encryptString(plainText string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptString 解密字符串
func decryptString(encryptedText string, key []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	plaintext, err := aesGCM.Open(nil, data[:nonceSize], data[nonceSize:], nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// NewStorageManager 创建存储管理器
func NewStorageManager() *StorageManager {
	// 获取程序运行目录
	exePath, err := os.Executable()
	if err != nil {
		// 如果获取失败，使用当前工作目录
		exePath = "."
	}

	// 获取程序所在目录
	exeDir := filepath.Dir(exePath)

	// 创建 data/cache 目录（先清空旧缓存）
	cacheDataDir := filepath.Join(exeDir, "data", "cache")
	fmt.Printf("[StorageManager] 缓存数据目录: %s\n", cacheDataDir)
	// 清空缓存目录，确保旧数据不干扰
	if err := os.RemoveAll(cacheDataDir); err != nil {
		fmt.Printf("[StorageManager] 清空缓存目录失败: %v\n", err)
	}
	if err := os.MkdirAll(cacheDataDir, 0755); err != nil {
		fmt.Printf("[StorageManager] 创建 data/cache 目录失败: %v\n", err)
		cacheDataDir = filepath.Join(exeDir, "data") // 降级到 data 目录
		if err := os.MkdirAll(cacheDataDir, 0755); err != nil {
			fmt.Printf("[StorageManager] 创建 data 目录失败: %v\n", err)
			cacheDataDir = exeDir // 降级到程序目录
		}
	}

	cacheDataFile := filepath.Join(cacheDataDir, "connections.json")
	fmt.Printf("[StorageManager] 缓存数据文件: %s\n", cacheDataFile)
	
	// 创建 data/persistent 目录用于永久保存
	permanentDataDir := filepath.Join(exeDir, "data", "persistent")
	fmt.Printf("[StorageManager] 永久数据目录: %s\n", permanentDataDir)
	if err := os.MkdirAll(permanentDataDir, 0755); err != nil {
		fmt.Printf("[StorageManager] 创建 data/persistent 目录失败: %v\n", err)
		permanentDataDir = filepath.Join(exeDir, "data") // 降级到 data 目录
		if err := os.MkdirAll(permanentDataDir, 0755); err != nil {
			fmt.Printf("[StorageManager] 创建 data 目录失败: %v\n", err)
			permanentDataDir = exeDir // 降级到程序目录
		}
	}

	permanentDataFile := filepath.Join(permanentDataDir, "connections.json")
	fmt.Printf("[StorageManager] 永久数据文件: %s\n", permanentDataFile)

	// 生成加密密钥（32字节用于AES-256）
	encryptKey := make([]byte, 32)
	if _, err := rand.Read(encryptKey); err != nil {
		fmt.Printf("[StorageManager] 生成加密密钥失败: %v\n", err)
		// 使用固定密钥作为备用（不安全，仅用于开发）
		encryptKey = []byte("pzssh-default-key-32bytes!!!!!!")
	}

	sm := &StorageManager{
		connections: make(map[string]*ConnectionInfo),
		dataFile:    cacheDataFile,        // 默认使用缓存文件
		permanentDataFile: permanentDataFile, // 永久保存文件
		encryptKey:  encryptKey,
	}

	// 加载已保存的连接
	sm.loadConnections()

	return sm
}

// SaveToPermanent 保存连接到永久文件（不加密，确保重启后密码可用）
func (sm *StorageManager) SaveToPermanent() error {
	fmt.Printf("[StorageManager] SaveToPermanent 开始\n")

	var connections []*ConnectionInfo
	for _, conn := range sm.connections {
		if conn.Saved {
			connCopy := *conn
			connections = append(connections, &connCopy)
		}
	}
	fmt.Printf("[StorageManager] 准备序列化 %d 个永久连接\n", len(connections))

	data, err := json.MarshalIndent(connections, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化永久连接数据失败: %v", err)
	}
	fmt.Printf("[StorageManager] 序列化成功，数据大小: %d bytes\n", len(data))

	if err := os.WriteFile(sm.permanentDataFile, data, 0644); err != nil {
		return fmt.Errorf("写入永久连接数据失败: %v", err)
	}
	fmt.Printf("[StorageManager] 永久文件写入成功: %s\n", sm.permanentDataFile)

	return nil
}

// LoadFromPermanent 从永久文件加载连接（兼容旧格式和新格式）
func (sm *StorageManager) LoadFromPermanent() error {
	data, err := os.ReadFile(sm.permanentDataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// 尝试解析为新格式 []SavedConnection
	var savedConns []SavedConnection
	if err := json.Unmarshal(data, &savedConns); err == nil && len(savedConns) > 0 && savedConns[0].Host != "" {
		fmt.Printf("[StorageManager] 解析到 %d 个新格式连接\n", len(savedConns))
		sm.mu.Lock()
		for i, sc := range savedConns {
			conn := &ConnectionInfo{
				ID:         generateConnectionID(),
				Name:       sc.Name,
				Host:       sc.Host,
				Port:       sc.Port,
				Username:   sc.Username,
				Password:   sc.Password,
				PrivateKey: sc.PrivateKey,
				Status:     "disconnected",
				Saved:      true,
			}
			sm.connections[conn.ID] = conn
			fmt.Printf("[StorageManager]   [%d] %s@%s:%d (ID=%s)\n", i+1, sc.Username, sc.Host, sc.Port, conn.ID)
		}
		sm.mu.Unlock()
		fmt.Printf("[StorageManager] 从永久文件加载了 %d 个连接，内存中共 %d 个\n", len(savedConns), len(sm.connections))
		return nil
	}

	// 兼容旧格式 []ConnectionInfo
	var oldConns []*ConnectionInfo
	if err := json.Unmarshal(data, &oldConns); err != nil {
		return err
	}
	sm.mu.Lock()
	for _, conn := range oldConns {
		conn.Status = "disconnected"
		conn.Saved = true
		sm.connections[conn.ID] = conn
	}
	sm.mu.Unlock()
	fmt.Printf("[StorageManager] 从永久文件加载了 %d 个连接（旧格式）\n", len(oldConns))
	return nil
}

// loadConnections 从文件加载连接信息
func (sm *StorageManager) loadConnections() {
	// 首先尝试从永久文件加载
	sm.LoadFromPermanent()
	
	// 然后从缓存文件加载（覆盖永久数据，因为缓存数据是最新的）
	sm.mu.Lock()
	defer sm.mu.Unlock()

	data, err := os.ReadFile(sm.dataFile)
	if err != nil {
		// 文件不存在或读取失败，使用已加载的永久数据
		fmt.Printf("[StorageManager] 读取缓存连接文件失败: %v\n", err)
		return
	}

	var connections []*ConnectionInfo
	if err := json.Unmarshal(data, &connections); err != nil {
		fmt.Printf("[StorageManager] 解析缓存连接数据失败: %v\n", err)
		return
	}

	for _, conn := range connections {
		conn.Status = "disconnected"
		sm.connections[conn.ID] = conn
	}
	fmt.Printf("[StorageManager] 加载了 %d 个缓存连接\n", len(connections))
}

// saveConnections 保存连接信息到文件（不加密）
func (sm *StorageManager) saveConnections() error {
	fmt.Printf("[StorageManager] saveConnections 开始\n")

	var connections []*ConnectionInfo
	for _, conn := range sm.connections {
		connCopy := *conn
		connections = append(connections, &connCopy)
	}
	fmt.Printf("[StorageManager] 准备序列化 %d 个连接\n", len(connections))

	data, err := json.MarshalIndent(connections, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化连接数据失败: %v", err)
	}
	fmt.Printf("[StorageManager] 序列化成功，数据大小: %d bytes\n", len(data))

	if err := os.WriteFile(sm.dataFile, data, 0644); err != nil {
		return fmt.Errorf("写入连接数据失败: %v", err)
	}
	fmt.Printf("[StorageManager] 文件写入成功: %s\n", sm.dataFile)

	return nil
}

// saveConnectionsLocked 保存连接信息到文件（外部版本，需要已持有锁）
func (sm *StorageManager) saveConnectionsLocked() error {
	return sm.saveConnections()
}

// AddConnection 添加新连接（默认未保存，用户手动保存后才标记）
func (sm *StorageManager) AddConnection(conn *ConnectionInfo) error {
	fmt.Printf("[StorageManager] AddConnection 开始: ID=%s\n", conn.ID)
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if conn.ID == "" {
		return fmt.Errorf("连接ID不能为空")
	}

	conn.Status = "disconnected"
	conn.Saved = false // 默认未保存，用户手动保存
	sm.connections[conn.ID] = conn
	fmt.Printf("[StorageManager] 已添加到内存\n")

	// 保存到文件
	fmt.Printf("[StorageManager] 开始保存到文件: %s\n", sm.dataFile)
	if err := sm.saveConnectionsLocked(); err != nil {
		fmt.Printf("[StorageManager] 保存失败: %v\n", err)
		delete(sm.connections, conn.ID)
		return err
	}
	fmt.Printf("[StorageManager] 保存成功\n")

	return nil
}

// UpdateConnection 更新连接信息
func (sm *StorageManager) UpdateConnection(conn *ConnectionInfo) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.connections[conn.ID]; !exists {
		return fmt.Errorf("连接不存在: %s", conn.ID)
	}

	// 保留当前状态和保存标记
	currentStatus := sm.connections[conn.ID].Status
	currentSaved := sm.connections[conn.ID].Saved
	conn.Status = currentStatus
	conn.Saved = currentSaved

	sm.connections[conn.ID] = conn

	// 保存缓存文件
	if err := sm.saveConnectionsLocked(); err != nil {
		return err
	}

	// 如果是已保存的连接，同步更新永久文件
	if conn.Saved {
		sm.saveToPermanentLocked()
	}

	return nil
}

// SyncImportConnection 云端同步导入（按 host:port 去重：存在则更新，不存在则新增）
func (sm *StorageManager) SyncImportConnection(conn *ConnectionInfo) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 查找是否已存在同 host:port 的连接
	var existingID string
	for id, existing := range sm.connections {
		if existing.Host == conn.Host && existing.Port == conn.Port {
			existingID = id
			break
		}
	}

	if existingID != "" {
		// 已存在：更新
		existing := sm.connections[existingID]
		existing.Name = conn.Name
		existing.Username = conn.Username
		if conn.Password != "" {
			existing.Password = conn.Password
		}
		if conn.KeyPath != "" {
			existing.KeyPath = conn.KeyPath
		}
		existing.Saved = true
		fmt.Printf("[StorageManager] 同步更新连接: %s (%s:%d)\n", existing.Name, existing.Host, existing.Port)
	} else {
		// 不存在：新增
		conn.ID = generateConnectionID()
		conn.Status = "disconnected"
		conn.Saved = true
		sm.connections[conn.ID] = conn
		fmt.Printf("[StorageManager] 同步新增连接: %s (%s:%d)\n", conn.Name, conn.Host, conn.Port)
	}

	// 保存到缓存和永久文件
	sm.saveConnectionsLocked()
	sm.saveToPermanentLocked()

	return nil
}

// DeleteConnection 删除连接（同时从永久文件中移除）
func (sm *StorageManager) DeleteConnection(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	conn, exists := sm.connections[id]
	if !exists {
		return fmt.Errorf("连接不存在: %s", id)
	}

	// 如果是已保存的连接，从永久文件中移除
	if conn.Saved {
		sm.removeFromPermanent(conn.Host, conn.Port, conn.Username)
	}

	delete(sm.connections, id)

	// 保存缓存文件
	return sm.saveConnectionsLocked()
}

// DeleteFromCache 从缓存中删除连接（保留永久记录，已保存的连接保留在内存中）
func (sm *StorageManager) DeleteFromCache(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	conn, exists := sm.connections[id]
	if !exists {
		return fmt.Errorf("连接不存在: %s", id)
	}

	if conn.Saved {
		// 已保存的连接：重置状态为断开，但保留在内存中
		conn.Status = "disconnected"
	} else {
		// 未保存的连接：从内存中移除
		delete(sm.connections, id)
	}

	// 只保存缓存文件，不修改永久文件
	return sm.saveConnectionsLocked()
}

// removeFromPermanent 从永久文件中移除指定连接
func (sm *StorageManager) removeFromPermanent(host string, port int, username string) {
	existing := sm.loadSavedConns()
	key := fmt.Sprintf("%s:%d:%s", host, port, username)
	var filtered []SavedConnection
	for _, c := range existing {
		cKey := fmt.Sprintf("%s:%d:%s", c.Host, c.Port, c.Username)
		if cKey != key {
			filtered = append(filtered, c)
		}
	}
	data, err := json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(sm.permanentDataFile, data, 0644)
	fmt.Printf("[StorageManager] 已从永久文件移除: %s\n", key)
}

// GetConnection 获取单个连接
func (sm *StorageManager) GetConnection(id string) (*ConnectionInfo, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	conn, exists := sm.connections[id]
	if !exists {
		return nil, fmt.Errorf("连接不存在: %s", id)
	}

	// 返回副本
	connCopy := *conn
	return &connCopy, nil
}

// GetAllConnections 获取所有连接
func (sm *StorageManager) GetAllConnections() []*ConnectionInfo {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var result []*ConnectionInfo
	for _, conn := range sm.connections {
		connCopy := *conn
		result = append(result, &connCopy)
	}

	return result
}

// UpdateConnectionStatus 更新连接状态
func (sm *StorageManager) UpdateConnectionStatus(id string, status string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if conn, exists := sm.connections[id]; exists {
		conn.Status = status
	}
}

// MarkAsSaved 标记为已保存（同时保存到缓存和永久文件）
func (sm *StorageManager) MarkAsSaved(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if conn, exists := sm.connections[id]; exists {
		conn.Saved = true
		
		// 1. 先保存到缓存文件
		if err := sm.saveConnectionsLocked(); err != nil {
			return fmt.Errorf("保存缓存文件失败: %v", err)
		}
		fmt.Printf("[StorageManager] ✓ 缓存文件已更新\n")
		
		// 2. 再保存到永久文件
		if err := sm.saveToPermanentLocked(); err != nil {
			return fmt.Errorf("保存永久文件失败: %v", err)
		}
		fmt.Printf("[StorageManager] ✓ 永久文件已更新\n")
		
		return nil
	}

	return fmt.Errorf("连接不存在: %s", id)
}

// SavedConnection 永久保存的连接格式（只保留必要字段）
type SavedConnection struct {
	Name       string `json:"name"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password,omitempty"`
	PrivateKey string `json:"privateKey,omitempty"`
}

// saveToPermanentLocked 保存已标记的连接到永久文件（合并已有数据，去重）
func (sm *StorageManager) saveToPermanentLocked() error {
	// 1. 读取已保存的连接
	existing := sm.loadSavedConns()

	// 2. 按 host:port:username 建索引
	seen := make(map[string]int) // key → index in existing
	for i, c := range existing {
		key := fmt.Sprintf("%s:%d:%s", c.Host, c.Port, c.Username)
		seen[key] = i
	}

	// 3. 把当前会话中标记为 Saved 的连接合并进去（新的覆盖旧的）
	for _, conn := range sm.connections {
		if !conn.Saved {
			continue
		}
		key := fmt.Sprintf("%s:%d:%s", conn.Host, conn.Port, conn.Username)
		sc := SavedConnection{
			Name:       conn.Name,
			Host:       conn.Host,
			Port:       conn.Port,
			Username:   conn.Username,
			Password:   conn.Password,
			PrivateKey: conn.PrivateKey,
		}
		if idx, ok := seen[key]; ok {
			existing[idx] = sc // 更新已有
		} else {
			existing = append(existing, sc) // 新增
		}
	}

	// 4. 写入文件
	data, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化永久连接数据失败: %v", err)
	}
	if err := os.WriteFile(sm.permanentDataFile, data, 0644); err != nil {
		return fmt.Errorf("写入永久连接数据失败: %v", err)
	}
	fmt.Printf("[StorageManager] 永久文件写入成功: %s (%d个连接)\n", sm.permanentDataFile, len(existing))
	return nil
}

// loadSavedConns 从永久文件读取已保存的连接
func (sm *StorageManager) loadSavedConns() []SavedConnection {
	data, err := os.ReadFile(sm.permanentDataFile)
	if err != nil {
		return nil
	}
	var conns []SavedConnection
	json.Unmarshal(data, &conns)
	return conns
}

// MarkAsUnsaved 标记为未保存（仅内存中）
func (sm *StorageManager) MarkAsUnsaved(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if conn, exists := sm.connections[id]; exists {
		conn.Saved = false
	}
}
