package ssh

import (
	"os"
	"path/filepath"
	"testing"
)

// TestAESCipherRoundTrip 验证 AES cipher 的加解密往返
func TestAESCipherRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	masterKeyPath := filepath.Join(tmpDir, "master.key")

	cipher := newAESCipher(masterKeyPath)

	plain := "my-super-secret-password-!@#$%^&*()_+"

	// 加密
	enc, err := cipher.Encrypt(plain)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}
	if enc == "" {
		t.Fatal("加密结果为空")
	}
	if enc == plain {
		t.Fatal("加密结果与明文相同，cipher 没工作！")
	}

	// 解密
	dec, err := cipher.Decrypt(enc)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}
	if dec != plain {
		t.Errorf("解密结果不一致:\n  got:  %q\n  want: %q", dec, plain)
	}
}

// TestAESCipherWrongKey 验证用错的 master key 加密的密文解不出来
//
// 这保证 attacker 即使拿到密文，置换一个 key 也解不出来。
func TestAESCipherWrongKey(t *testing.T) {
	tmpDir := t.TempDir()

	// 用 key A 加密
	cipherA := newAESCipher(filepath.Join(tmpDir, "keyA.key"))
	enc, err := cipherA.Encrypt("secret payload")
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	// 用 key B 解密 → 必然失败
	cipherB := newAESCipher(filepath.Join(tmpDir, "keyB.key"))
	_, err = cipherB.Decrypt(enc)
	if err == nil {
		t.Fatal("用错误 key 解密竟然成功了！cipher 没有真正隔离 key")
	}
}

// TestAESCipherNonceUniqueness 验证多次加密同一个明文得到不同密文（GCM nonce 必须唯一）
func TestAESCipherNonceUniqueness(t *testing.T) {
	tmpDir := t.TempDir()
	cipher := newAESCipher(filepath.Join(tmpDir, "master.key"))

	plain := "same input"
	enc1, _ := cipher.Encrypt(plain)
	enc2, _ := cipher.Encrypt(plain)

	if enc1 == enc2 {
		t.Fatal("两次加密同明文得到相同密文，nonce 重复 = 严重安全漏洞！")
	}
}

// TestAESCipherTamperDetection 验证密文被篡改后能检测出来
func TestAESCipherTamperDetection(t *testing.T) {
	tmpDir := t.TempDir()
	cipher := newAESCipher(filepath.Join(tmpDir, "master.key"))

	enc, _ := cipher.Encrypt("important data")

	// 翻转密文的最后一个字符（GCM 认证 tag 应能捕获这个篡改）
	tampered := enc[:len(enc)-1]
	if enc[len(enc)-1] == 'A' {
		tampered += "B"
	} else {
		tampered += "A"
	}

	_, err := cipher.Decrypt(tampered)
	if err == nil {
		t.Fatal("被篡改的密文竟然解密成功了！AES-GCM 认证失败！")
	}
}

// TestMasterKeyPersistence 验证 master key 在重启后能继续解密同一文件
func TestMasterKeyPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	masterKeyPath := filepath.Join(tmpDir, "master.key")

	// 第一次启动：生成 master key 并加密
	cipher1 := newAESCipher(masterKeyPath)
	enc, err := cipher1.Encrypt("persistent secret")
	if err != nil {
		t.Fatalf("首次加密失败: %v", err)
	}

	// 检查 master key 文件存在且权限正确（0600）
	info, err := os.Stat(masterKeyPath)
	if err != nil {
		t.Fatalf("master key 文件未生成: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("master key 文件权限应该是 0600，实际是 %o", perm)
	}

	// 第二次启动：重新构造 cipher，应该能解密之前的数据
	cipher2 := newAESCipher(masterKeyPath)
	dec, err := cipher2.Decrypt(enc)
	if err != nil {
		t.Fatalf("重启后解密失败: %v", err)
	}
	if dec != "persistent secret" {
		t.Errorf("重启后解密结果不对: got %q", dec)
	}
}

// TestGetMasterKeyPath 验证 master key 路径函数
func TestGetMasterKeyPath(t *testing.T) {
	path := getMasterKeyPath()
	if path == "" {
		t.Fatal("master key 路径不应为空")
	}
	// 路径应该包含 qssh 子目录
	if !filepath.IsAbs(path) && path[0] != '.' {
		t.Errorf("master key 路径应该为绝对路径或当前目录相对路径：%s", path)
	}
}