package ssh

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	dataDirOnce sync.Once
	dataDirPath string
)

// GetDataDir 返回应用数据目录（统一实现，供 ssh 和 ai 模块共用）。
//
// 优先级：
//  1. 可执行文件所在目录下的 data/
//  2. 当前工作目录下的 data/（dev 模式）
//  3. 当前目录（最终 fallback）
func GetDataDir() string {
	dataDirOnce.Do(func() {
		// 优先用可执行文件目录（生产模式）
		if exePath, err := os.Executable(); err == nil {
			candidate := filepath.Join(filepath.Dir(exePath), "data")
			if mkErr := os.MkdirAll(candidate, 0755); mkErr == nil {
				dataDirPath = candidate
				return
			}
		}

		// 回退到当前工作目录（dev 模式）
		if cwd, err := os.Getwd(); err == nil {
			candidate := filepath.Join(cwd, "data")
			if mkErr := os.MkdirAll(candidate, 0755); mkErr == nil {
				dataDirPath = candidate
				return
			}
		}

		dataDirPath = "."
		fmt.Fprintln(os.Stderr, "[pathutil] 警告：无法确定数据目录，使用当前目录")
	})
	return dataDirPath
}
