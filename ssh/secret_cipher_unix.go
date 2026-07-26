//go:build !windows

package ssh

func newPlatformCipher() SecretCipher {
	return newAESCipher(getMasterKeyPath())
}
