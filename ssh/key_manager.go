package ssh

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHKeyInfo SSH 密钥信息
type SSHKeyInfo struct {
	Name      string `json:"name"`       // 文件名
	Path      string `json:"path"`       // 完整路径
	Type      string `json:"type"`       // ed25519, rsa, ecdsa
	Bits      int    `json:"bits"`       // 密钥位数
	Comment   string `json:"comment"`    // 注释
	CreatedAt string `json:"createdAt"`  // 创建时间
	HasPublic bool   `json:"hasPublic"`  // 是否有公钥
}

// KeyManager SSH 密钥管理器
type KeyManager struct {
	sshDir string
}

// NewKeyManager 创建密钥管理器
func NewKeyManager() *KeyManager {
	home, err := os.UserHomeDir()
	if err != nil {
		return &KeyManager{sshDir: ".ssh"}
	}
	return &KeyManager{sshDir: filepath.Join(home, ".ssh")}
}

// ListKeys 列出所有 SSH 密钥
func (km *KeyManager) ListKeys() ([]SSHKeyInfo, error) {
	entries, err := os.ReadDir(km.sshDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []SSHKeyInfo{}, nil
		}
		return nil, fmt.Errorf("读取 .ssh 目录失败: %w", err)
	}

	var keys []SSHKeyInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// 跳过非密钥文件
		if !isPrivateKeyFile(name) {
			continue
		}

		path := filepath.Join(km.sshDir, name)
		info, err := km.analyzeKey(path, name)
		if err != nil {
			continue
		}

		keys = append(keys, *info)
	}

	return keys, nil
}

// GenerateKey 生成新密钥
func (km *KeyManager) GenerateKey(name, keyType, comment string, bits int) error {
	if name == "" {
		return fmt.Errorf("密钥名称不能为空")
	}

	// 确保 .ssh 目录存在
	if err := os.MkdirAll(km.sshDir, 0700); err != nil {
		return fmt.Errorf("创建 .ssh 目录失败: %w", err)
	}

	privatePath := filepath.Join(km.sshDir, name)
	publicPath := privatePath + ".pub"

	// 检查是否已存在
	if _, err := os.Stat(privatePath); err == nil {
		return fmt.Errorf("密钥 %s 已存在", name)
	}

	if comment == "" {
		comment = fmt.Sprintf("generated@pzssh-%s", time.Now().Format("20060102"))
	}

	var privateKeyPEM []byte
	var publicKeyBytes []byte

	switch keyType {
	case "ed25519":
		pub, priv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return fmt.Errorf("生成 ed25519 密钥失败: %w", err)
		}
		privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
		if err != nil {
			return fmt.Errorf("序列化私钥失败: %w", err)
		}
		privateKeyPEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})

		sshPub, err := ssh.NewPublicKey(pub)
		if err != nil {
			return fmt.Errorf("创建公钥失败: %w", err)
		}
		publicKeyBytes = append(ssh.MarshalAuthorizedKey(sshPub), []byte(" "+comment+"\n")...)

	case "rsa":
		if bits == 0 {
			bits = 4096
		}
		key, err := rsa.GenerateKey(rand.Reader, bits)
		if err != nil {
			return fmt.Errorf("生成 RSA 密钥失败: %w", err)
		}
		privBytes, err := x509.MarshalPKCS8PrivateKey(key)
		if err != nil {
			return fmt.Errorf("序列化私钥失败: %w", err)
		}
		privateKeyPEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})

		sshPub, err := ssh.NewPublicKey(&key.PublicKey)
		if err != nil {
			return fmt.Errorf("创建公钥失败: %w", err)
		}
		publicKeyBytes = append(ssh.MarshalAuthorizedKey(sshPub), []byte(" "+comment+"\n")...)

	default:
		return fmt.Errorf("不支持的密钥类型: %s（支持 ed25519, rsa）", keyType)
	}

	// 写入私钥（权限 0600）
	if err := os.WriteFile(privatePath, privateKeyPEM, 0600); err != nil {
		return fmt.Errorf("写入私钥失败: %w", err)
	}

	// 写入公钥（权限 0644）
	if err := os.WriteFile(publicPath, publicKeyBytes, 0644); err != nil {
		return fmt.Errorf("写入公钥失败: %w", err)
	}

	return nil
}

// DeployPublicKey 部署公钥到远程服务器
func (km *KeyManager) DeployPublicKey(keyName string, client *SSHClient) error {
	publicPath := filepath.Join(km.sshDir, keyName+".pub")
	pubBytes, err := os.ReadFile(publicPath)
	if err != nil {
		return fmt.Errorf("读取公钥失败: %w", err)
	}

	// 获取远程用户的 home 目录
	homeResult, err := client.ExecuteCommand("echo $HOME")
	if err != nil || !homeResult.Success {
		return fmt.Errorf("获取远程 home 目录失败: %v", err)
	}
	remoteHome := strings.TrimSpace(homeResult.Stdout)

	// 确保 .ssh 目录存在
	remoteSSHDir := remoteHome + "/.ssh"
	client.ExecuteCommand(fmt.Sprintf("mkdir -p %s && chmod 700 %s", remoteSSHDir, remoteSSHDir))

	// 追加公钥到 authorized_keys
	authorizedKeys := remoteSSHDir + "/authorized_keys"
	pubKeyStr := strings.TrimSpace(string(pubBytes))

	// 检查是否已存在
	checkCmd := fmt.Sprintf("grep -qF '%s' %s 2>/dev/null && echo 'exists' || echo 'not_found'", pubKeyStr, authorizedKeys)
	checkResult, err := client.ExecuteCommand(checkCmd)
	if err == nil && strings.TrimSpace(checkResult.Stdout) == "exists" {
		return nil // 已存在，无需重复部署
	}

	// 追加公钥
	appendCmd := fmt.Sprintf("echo '%s' >> %s && chmod 600 %s", pubKeyStr, authorizedKeys, authorizedKeys)
	result, err := client.ExecuteCommand(appendCmd)
	if err != nil {
		return fmt.Errorf("追加公钥失败: %v", err)
	}
	if !result.Success {
		return fmt.Errorf("追加公钥失败: %s", result.Stderr)
	}

	return nil
}

// GetPublicKey 获取公钥内容
func (km *KeyManager) GetPublicKey(keyName string) (string, error) {
	publicPath := filepath.Join(km.sshDir, keyName+".pub")
	data, err := os.ReadFile(publicPath)
	if err != nil {
		return "", fmt.Errorf("读取公钥失败: %w", err)
	}
	return string(data), nil
}

// DeleteKey 删除密钥
func (km *KeyManager) DeleteKey(keyName string) error {
	privatePath := filepath.Join(km.sshDir, keyName)
	publicPath := privatePath + ".pub"

	os.Remove(publicPath)
	if err := os.Remove(privatePath); err != nil {
		return fmt.Errorf("删除私钥失败: %w", err)
	}

	return nil
}

// analyzeKey 分析密钥文件
func (km *KeyManager) analyzeKey(path, name string) (*SSHKeyInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("无法解析 PEM")
	}

	keyType := "unknown"
	bits := 0
	comment := ""

	// 尝试解析密钥获取类型和位数
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		switch k := key.(type) {
		case *rsa.PrivateKey:
			keyType = "rsa"
			bits = k.N.BitLen()
		case ed25519.PrivateKey:
			keyType = "ed25519"
			bits = 256
		}
	}

	// 从公钥文件获取注释
	pubPath := path + ".pub"
	if pubData, err := os.ReadFile(pubPath); err == nil {
		parts := strings.Fields(string(pubData))
		if len(parts) >= 3 {
			comment = parts[2]
		}
	}

	// 获取文件信息
	stat, _ := os.Stat(path)
	createdAt := ""
	if stat != nil {
		createdAt = stat.ModTime().Format("2006-01-02 15:04")
	}

	return &SSHKeyInfo{
		Name:      name,
		Path:      path,
		Type:      keyType,
		Bits:      bits,
		Comment:   comment,
		CreatedAt: createdAt,
		HasPublic: fileExists(pubPath),
	}, nil
}

// isPrivateKeyFile 判断是否是私钥文件
func isPrivateKeyFile(name string) bool {
	if strings.HasSuffix(name, ".pub") {
		return false
	}
	skipFiles := []string{
		"known_hosts", "authorized_keys", "config",
		"known_hosts.old", "config.old",
	}
	for _, skip := range skipFiles {
		if name == skip {
			return false
		}
	}
	// 私钥文件通常没有扩展名，或者是 .pem
	return true
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
