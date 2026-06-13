package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// ==================== AES-256-GCM 加密 ====================

// Encryptor 数据加密器
type Encryptor struct {
	key []byte
}

// NewEncryptor 创建加密器
// key 建议 32+ 字节，内部会 SHA-256 哈希确保长度
func NewEncryptor(key []byte) *Encryptor {
	hash := sha256.Sum256(key)
	return &Encryptor{key: hash[:]}
}

// Encrypt 加密（返回 base64 编码的 nonce + ciphertext）
func (e *Encryptor) Encrypt(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密
func (e *Encryptor) Decrypt(encoded string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("密文太短")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return aesGCM.Open(nil, nonce, ciphertext, nil)
}

// ==================== 安全令牌认证 ====================

// TokenManager 令牌管理器（防暴力破解）
type TokenManager struct {
	mu          sync.RWMutex
	tokenHash   string            // 令牌的 SHA-512 哈希
	attempts    map[string]*attempt // IP -> 尝试记录
	maxAttempts int               // 最大尝试次数
	lockoutTime time.Duration     // 锁定时间
}

type attempt struct {
	count    int
	lastTry  time.Time
	lockout  time.Time
}

// NewTokenManager 创建令牌管理器
func NewTokenManager(token string) *TokenManager {
	tm := &TokenManager{
		tokenHash:   SHA512Hash(token),
		attempts:    make(map[string]*attempt),
		maxAttempts: 5,
		lockoutTime: 15 * time.Minute,
	}

	// 定期清理过期记录
	go tm.cleanupLoop()

	return tm
}

// Verify 验证令牌（带防暴力破解）
// 返回: 是否通过, 剩余等待时间
func (tm *TokenManager) Verify(token, clientIP string) (bool, time.Duration) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// 检查是否被锁定
	if a, exists := tm.attempts[clientIP]; exists {
		if !a.lockout.IsZero() && time.Now().Before(a.lockout) {
			remaining := time.Until(a.lockout)
			return false, remaining
		}
	}

	// 验证令牌（使用 HMAC 恒定时间比较）
	providedHash := SHA512Hash(token)
	if !HMACEqual(providedHash, tm.tokenHash) {
		// 记录失败尝试
		tm.recordFailure(clientIP)
		return false, 0
	}

	// 验证成功，清除失败记录
	delete(tm.attempts, clientIP)
	return true, 0
}

// recordFailure 记录失败尝试
func (tm *TokenManager) recordFailure(clientIP string) {
	a, exists := tm.attempts[clientIP]
	if !exists {
		a = &attempt{}
		tm.attempts[clientIP] = a
	}

	a.count++
	a.lastTry = time.Now()

	if a.count >= tm.maxAttempts {
		a.lockout = time.Now().Add(tm.lockoutTime)
		fmt.Printf("[Security] IP %s 因多次失败被锁定 %v\n", clientIP, tm.lockoutTime)
	}
}

// cleanupLoop 定期清理过期记录
func (tm *TokenManager) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		tm.mu.Lock()
		for ip, a := range tm.attempts {
			if !a.lockout.IsZero() && time.Now().After(a.lockout) {
				delete(tm.attempts, ip)
			} else if time.Since(a.lastTry) > 30*time.Minute {
				delete(tm.attempts, ip)
			}
		}
		tm.mu.Unlock()
	}
}

// ==================== 请求签名 ====================

// SignRequest 签名请求（防篡改）
// 签名 = HMAC-SHA256(secret, method + path + timestamp + body)
func SignRequest(secret, method, path, timestamp string, body []byte) string {
	message := method + path + timestamp + string(body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyRequestSignature 验证请求签名
func VerifyRequestSignature(secret, method, path, timestamp, signature string, body []byte) bool {
	expected := SignRequest(secret, method, path, timestamp, body)
	return HMACEqual(expected, signature)
}

// ==================== 工具函数 ====================

// SHA512Hash SHA-512 哈希
func SHA512Hash(data string) string {
	h := sha512.Sum512([]byte(data))
	return hex.EncodeToString(h[:])
}

// HMACEqual HMAC 恒定时间比较（防时序攻击）
func HMACEqual(a, b string) bool {
	return hmac.Equal([]byte(a), []byte(b))
}

// GenerateToken 生成加密安全的随机令牌（64 字符十六进制）
func GenerateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GenerateDeviceID 生成设备 ID
func GenerateDeviceID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GenerateNonce 生成随机 nonce（用于防重放攻击）
func GenerateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// ValidateTimestamp 验证时间戳（防重放攻击，允许 ±5 分钟误差）
func ValidateTimestamp(timestamp string) bool {
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return false
	}
	diff := time.Since(t)
	return diff > -5*time.Minute && diff < 5*time.Minute
}

// MaskToken 脱敏显示令牌
func MaskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + strings.Repeat("*", len(token)-8) + token[len(token)-4:]
}
