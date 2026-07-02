//go:build windows

package ssh

import (
	"encoding/base64"
	"fmt"

	"github.com/Microsoft/go-winio/pkg/dpapi"
)

// newPlatformCipher 在 Windows 下用 DPAPI
func newPlatformCipher() SecretCipher {
	return newDPAPICipher()
}

// dpapiCipher 使用 Windows DPAPI 加密
//
// DPAPI（Data Protection API）是 Windows 内置的凭据保护 API：
//   - 加密密钥由 Windows 凭据体系管理，与当前 Windows 用户登录会话绑定
//   - 同一个用户在同一台机器上能解密；其他用户、其他机器无法解密
//   - 即使磁盘文件被拷贝走，攻击者也不能解密
//
// 参考：https://learn.microsoft.com/en-us/windows/win32/api/dpapi/
type dpapiCipher struct{}

// newDPAPICipher 构造 DPAPI cipher
func newDPAPICipher() *dpapiCipher {
	return &dpapiCipher{}
}

func (d *dpapiCipher) Kind() string {
	return "dpapi"
}

// encryptSecret 是 qssh 用作 DPAPI context 的固定字符串
//
// DPAPI 的 "context" 参数不是密钥本身，但会作为额外熵参与派生。
// 固定值 + 代码签名 = 攻击者不能在不替换程序的前提下伪造 context。
const dpapiEntropy = "qssh-connection-secret-v1"

func (d *dpapiCipher) Encrypt(plain string) (string, error) {
	enc, err := dpapi.Encrypt([]byte(plain), dpapiEntropy)
	if err != nil {
		return "", fmt.Errorf("DPAPI 加密失败: %w", err)
	}
	return base64.StdEncoding.EncodeToString(enc), nil
}

func (d *dpapiCipher) Decrypt(cipherText string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", fmt.Errorf("base64 解码失败: %w", err)
	}
	plain, err := dpapi.Decrypt(data, dpapiEntropy)
	if err != nil {
		return "", fmt.Errorf("DPAPI 解密失败（密钥可能已损坏或在其他用户的上下文加密）: %w", err)
	}
	return string(plain), nil
}