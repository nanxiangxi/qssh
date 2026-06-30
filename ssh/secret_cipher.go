package ssh

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// aesCipher 用 AES-256-GCM 加密，密钥从持久化的 master key 文件加载
//
// 这是一个 fallback cipher，主要给 macOS / Linux 用。
// 不是最理想的（理想方案是 Keychain / Secret Service），但比之前的明文强 N 倍。
//
// 安全注意点：
//   - master key 文件权限 0600（仅当前用户可读写）
//   - 文件路径在用户配置目录下，跨用户隔离
//   - 一旦泄露，攻击者能解密所有连接 → 平时要给文件加 ACL / 用 umask
type aesCipher struct {
	key        []byte
	masterPath string
}

// newAESCipher 构造 AES cipher，会按需生成或读取 master key
func newAESCipher(masterKeyPath string) *aesCipher {
	key, err := loadOrCreateMasterKey(masterKeyPath)
	if err != nil {
		// 最后的 fallback：随机生成一个（仅本次进程有效，重启后无法解密）
		// 这种情况下相当于降级到无加密状态，必须大声报错提醒用户
		fmt.Printf("[SecretCipher] ⚠️ 严重警告：master key 加载失败 (%v)，本会话密码将无法持久化\n", err)
		key = make([]byte, 32)
		_, _ = rand.Read(key)
	}
	return &aesCipher{key: key, masterPath: masterKeyPath}
}

func (a *aesCipher) Kind() string {
	return "aes-gcm"
}

func (a *aesCipher) Encrypt(plain string) (string, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	// 把 nonce 拼接到密文前面，方便解密时取出来
	ciphertext := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (a *aesCipher) Decrypt(cipherText string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("密文过短")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// loadOrCreateMasterKey 从文件加载 master key，文件不存在则生成一个新的
//
// 这个 key 决定所有密文能否被解密，一旦丢失意味着所有已保存连接失效（需用户重新输入密码）
func loadOrCreateMasterKey(path string) ([]byte, error) {
	// 1. 确保目录存在且权限正确
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, fmt.Errorf("创建 master key 目录失败: %w", err)
	}

	// 2. 文件存在 → 读取
	if data, err := os.ReadFile(path); err == nil {
		if len(data) == 32 {
			return data, nil
		}
		// 文件存在但长度不对，丢弃并重新生成（这是个明显错误状态）
		fmt.Printf("[SecretCipher] ⚠️ master key 长度异常 (%d 字节)，重新生成（已加密数据将无法解密！）\n", len(data))
	}

	// 3. 生成新 key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("生成 master key 失败: %w", err)
	}
	if err := os.WriteFile(path, key, 0600); err != nil {
		return nil, fmt.Errorf("写入 master key 失败: %w", err)
	}
	fmt.Printf("[SecretCipher] ✓ 已生成新的 master key: %s\n", path)
	return key, nil
}

// getMasterKeyPath 返回 master key 文件的路径（macOS / Linux 用）
//
// 优先级：$XDG_CONFIG_HOME/qssh/master.key > $HOME/.config/qssh/master.key
func getMasterKeyPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "qssh", "master.key")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		// 最后 fallback：当前目录的临时文件
		return filepath.Join(".qssh", "master.key")
	}
	return filepath.Join(home, ".config", "qssh", "master.key")
}