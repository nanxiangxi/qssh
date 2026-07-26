//go:build windows

package ssh

import (
	"encoding/base64"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	crypt32                = windows.NewLazySystemDLL("crypt32")
	procCryptProtectData   = crypt32.NewProc("CryptProtectData")
	procCryptUnprotectData = crypt32.NewProc("CryptUnprotectData")
	procLocalFree          = windows.NewLazySystemDLL("kernel32").NewProc("LocalFree")
)

type _DATA_BLOB struct {
	cbData uint32
	pbData unsafe.Pointer
}

func newPlatformCipher() SecretCipher {
	return newDPAPICipher()
}

type dpapiCipher struct{}

func newDPAPICipher() *dpapiCipher {
	return &dpapiCipher{}
}

func (d *dpapiCipher) Kind() string {
	return "dpapi"
}

const dpapiEntropy = "qssh-connection-secret-v1"

func (d *dpapiCipher) Encrypt(plain string) (string, error) {
	data := []byte(plain)
	inBlob := &_DATA_BLOB{cbData: uint32(len(data)), pbData: unsafe.Pointer(&data[0])}
	entropyBytes := []byte(dpapiEntropy)
	entropyBlob := &_DATA_BLOB{cbData: uint32(len(entropyBytes)), pbData: unsafe.Pointer(&entropyBytes[0])}
	var outBlob _DATA_BLOB
	ret, _, err := procCryptProtectData.Call(
		uintptr(unsafe.Pointer(inBlob)), 0,
		uintptr(unsafe.Pointer(entropyBlob)), 0, 0, 0,
		uintptr(unsafe.Pointer(&outBlob)),
	)
	if ret == 0 {
		return "", fmt.Errorf("CryptProtectData 失败: %w", err)
	}
	defer procLocalFree.Call(uintptr(outBlob.pbData))
	enc := make([]byte, outBlob.cbData)
	copy(enc, unsafe.Slice((*byte)(outBlob.pbData), outBlob.cbData))
	return base64.StdEncoding.EncodeToString(enc), nil
}

func (d *dpapiCipher) Decrypt(cipherText string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", fmt.Errorf("base64 解码失败: %w", err)
	}
	inBlob := &_DATA_BLOB{cbData: uint32(len(data)), pbData: unsafe.Pointer(&data[0])}
	entropyBytes := []byte(dpapiEntropy)
	entropyBlob := &_DATA_BLOB{cbData: uint32(len(entropyBytes)), pbData: unsafe.Pointer(&entropyBytes[0])}
	var outBlob _DATA_BLOB
	ret, _, err := procCryptUnprotectData.Call(
		uintptr(unsafe.Pointer(inBlob)), 0,
		uintptr(unsafe.Pointer(entropyBlob)), 0, 0, 0,
		uintptr(unsafe.Pointer(&outBlob)),
	)
	if ret == 0 {
		return "", fmt.Errorf("CryptUnprotectData 失败: %w", err)
	}
	defer procLocalFree.Call(uintptr(outBlob.pbData))
	dec := make([]byte, outBlob.cbData)
	copy(dec, unsafe.Slice((*byte)(outBlob.pbData), outBlob.cbData))
	return string(dec), nil

}
