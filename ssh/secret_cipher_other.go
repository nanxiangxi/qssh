//go:build !windows

package ssh

// newPlatformCipher 在非 Windows 平台用 AES-GCM + master key 文件
//
// macOS 理想方案是 Keychain，Linux 理想方案是 Secret Service（GNOME Keyring / KWallet）。
// 这两个都需要 CGO 或纯 Go 替代品（如 github.com/zalando/go-keyring），
// 为保持 PR 简洁暂时用 fallback 实现，后续可以独立 PR 升级。
func newPlatformCipher() SecretCipher {
	return newAESCipher(getMasterKeyPath())
}