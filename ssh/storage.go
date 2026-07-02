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
	"runtime"
	"sync"
)

// ConnectionInfo SSH连接信息
//
// 安全修复：之前 Password / PrivateKey 以明文 JSON 落盘，
// 现在改为以 SecretCipher 加密后的 ciphertext（base64）保存到 Enc 字段。
// 旧字段保留以便迁移期兼容，新写入只设置 Enc，原字段清空。
type ConnectionInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Host   string `json:"host"`
	Port   int    `json:"port"`
	Username string `json:"username"`

	// === 明文字段（仅作迁移期兼容用，新写入会留空） ===
	Password   string `json:"password,omitempty"`
	KeyPath    string `json:"keyPath,omitempty"`
	PrivateKey string `json:"privateKey,omitempty"`

	// === 加密字段（推荐使用） ===
	PasswordCiphertext   string `json:"passwordCiphertext,omitempty"`
	PrivateKeyCiphertext string `json:"privateKeyCiphertext,omitempty"`

	// CipherKind 标识加密方式，便于以后平滑升级
	// "dpapi" = Windows DPAPI，"aes-gcm" = 跨平台 fallback
	CipherKind string `json:"cipherKind,omitempty"`

	Status  string `json:"status"`
	Saved   bool   `json:"saved"`
	GroupID string `json:"group_id,omitempty"`
}

// StorageManager 连接存储管理器
type StorageManager struct {
	mu                sync.RWMutex
	connections       map[string]*ConnectionInfo
	dataFile          string
	permanentDataFile string // 永久保存的文件路径
	cipher            SecretCipher
}

// SecretCipher 抽象加密器，让 storage.go 不直接依赖系统 API
type SecretCipher interface {
	// Kind 返回加密器标识（写入 ConnectionInfo.CipherKind）
	Kind() string
	// Encrypt 加密明文，返回 base64 字符串
	Encrypt(plain string) (string, error)
	// Decrypt 解密 base64 字符串，返回明文
	Decrypt(cipher string) (string, error)
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

	// 安全修复：使用跨平台 SecretCipher 替换原来的硬编码 fallback 密钥
	// - Windows 优先用 DPAPI（密钥绑定到当前用户，迁移成本最低）
	// - 其他平台 fallback 到 AES-GCM + 持久化 master key 文件
	cipher := newPlatformCipher()

	sm := &StorageManager{
		connections:       make(map[string]*ConnectionInfo),
		dataFile:          cacheDataFile,         // 默认使用缓存文件
		permanentDataFile: permanentDataFile,     // 永久保存文件
		cipher:            cipher,
	}

	// 加载已保存的连接
	sm.loadConnections()

	// 启动时尝试把旧明文连接迁移到加密格式（静默失败不阻塞启动）
	if migrated, err := sm.migrateLegacyPlaintext(); err != nil {
		fmt.Printf("[StorageManager] 旧明文迁移失败（不影响启动）: %v\n", err)
	} else if migrated > 0 {
		fmt.Printf("[StorageManager] ✓ 已迁移 %d 个旧明文连接到加密格式\n", migrated)
	}

	return sm
}

// newPlatformCipher 根据当前操作系统构造合适的 SecretCipher
//
// 设计取舍：
//   - Windows 用 DPAPI 是最优解：密钥由 Windows 凭据体系管理，
//     加密数据只能由同一用户在同一台机器上解密，跨用户/跨机器迁移无意义（也防泄漏）
//   - macOS / Linux 暂用 AES-GCM + 持久化 master key 文件（落到用户配置目录）
//     这不是最理想的（理想是 Keychain / Secret Service），但比之前的明文强 N 倍
func newPlatformCipher() SecretCipher {
	if runtime.GOOS == "windows" {
		return newDPAPICipher()
	}
	return newAESCipher(getMasterKeyPath())
}

// migrateLegacyPlaintext 把磁盘上仍以明文形式存在的 Password / PrivateKey
// 迁移到加密字段，并清空明文字段。
//
// 触发条件：ConnectionInfo.CipherKind == "" 且 Password/PrivateKey != ""
//
// 只迁移"已保存"的连接（Saved == true），临时连接不动。
// 迁移成功立即写盘，失败只记日志不阻塞。
func (sm *StorageManager) migrateLegacyPlaintext() (int, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	migrated := 0
	for _, conn := range sm.connections {
		if !conn.Saved {
			continue
		}
		if conn.CipherKind != "" {
			continue // 已经加密过
		}
		if conn.Password == "" && conn.PrivateKey == "" {
			continue // 没东西可迁移
		}

		// 加密现有明文
		if conn.Password != "" {
			enc, err := sm.cipher.Encrypt(conn.Password)
			if err != nil {
				return migrated, fmt.Errorf("加密密码失败 (ID=%s): %w", conn.ID, err)
			}
			conn.PasswordCiphertext = enc
			conn.Password = "" // 清空明文
		}
		if conn.PrivateKey != "" {
			enc, err := sm.cipher.Encrypt(conn.PrivateKey)
			if err != nil {
				return migrated, fmt.Errorf("加密私钥失败 (ID=%s): %w", conn.ID, err)
			}
			conn.PrivateKeyCiphertext = enc
			conn.PrivateKey = "" // 清空明文
		}
		conn.CipherKind = sm.cipher.Kind()
		migrated++
	}

	if migrated > 0 {
		// 写回缓存 + 永久文件
		if err := sm.saveConnectionsLocked(); err != nil {
			return migrated, fmt.Errorf("写回缓存失败: %w", err)
		}
		if err := sm.saveToPermanentLocked(); err != nil {
			return migrated, fmt.Errorf("写回永久文件失败: %w", err)
		}
	}
	return migrated, nil
}

// SaveToPermanent 保存连接到永久文件（已加密，确保重启后密码可用）
func (sm *StorageManager) SaveToPermanent() error {
	fmt.Printf("[StorageManager] SaveToPermanent 开始\n")

	var connections []*ConnectionInfo
	for _, conn := range sm.connections {
		if conn.Saved {
			// 拷贝前先加密所有明文，确保落盘是密文
			encryptedCopy, err := sm.encryptForPersist(conn)
			if err != nil {
				return fmt.Errorf("加密连接失败 (ID=%s): %w", conn.ID, err)
			}
			connections = append(connections, encryptedCopy)
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

// encryptForPersist 拷贝连接并把明文密码/私钥加密到 Enc 字段、清空原文
// 用于 SaveToPermanent（写永久文件前），确保磁盘上永远是密文
func (sm *StorageManager) encryptForPersist(conn *ConnectionInfo) (*ConnectionInfo, error) {
	connCopy := *conn
	connCopy.Password = ""
	connCopy.PrivateKey = ""

	if conn.Password != "" {
		enc, err := sm.cipher.Encrypt(conn.Password)
		if err != nil {
			return nil, err
		}
		connCopy.PasswordCiphertext = enc
	}
	if conn.PrivateKey != "" {
		enc, err := sm.cipher.Encrypt(conn.PrivateKey)
		if err != nil {
			return nil, err
		}
		connCopy.PrivateKeyCiphertext = enc
	}
	connCopy.CipherKind = sm.cipher.Kind()
	return &connCopy, nil
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
		// 安全修复：从磁盘读取时自动解密 Enc 字段回填到 Password/PrivateKey，
		// 让上层调用方完全不用感知"加解密"这件事。
		if err := sm.decryptInPlace(conn); err != nil {
			fmt.Printf("[StorageManager] ⚠️ 解密连接失败 (ID=%s): %v\n", conn.ID, err)
		}
		conn.Status = "disconnected"
		conn.Saved = true
		sm.connections[conn.ID] = conn
	}
	sm.mu.Unlock()
	fmt.Printf("[StorageManager] 从永久文件加载了 %d 个连接（旧格式）\n", len(oldConns))
	return nil
}

// decryptInPlace 把 conn 的 Enc 字段解密后填回明文字段
// 如果 conn 没有 Enc 字段（旧明文数据），什么都不做（让 migrateLegacyPlaintext 后续处理）
func (sm *StorageManager) decryptInPlace(conn *ConnectionInfo) error {
	if conn.CipherKind == "" {
		return nil // 旧明文，不归这里管
	}
	if conn.PasswordCiphertext != "" {
		plain, err := sm.cipher.Decrypt(conn.PasswordCiphertext)
		if err != nil {
			return fmt.Errorf("解密密码失败: %w", err)
		}
		conn.Password = plain
	}
	if conn.PrivateKeyCiphertext != "" {
		plain, err := sm.cipher.Decrypt(conn.PrivateKeyCiphertext)
		if err != nil {
			return fmt.Errorf("解密私钥失败: %w", err)
		}
		conn.PrivateKey = plain
	}
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
