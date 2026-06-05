package ai

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AILogger 带缓冲的 AI 操作日志
type AILogger struct {
	mu      sync.Mutex
	baseDir string
	file    *os.File
	buf     []string
	flushAt int
}

// NewAILogger 创建日志器
func NewAILogger(baseDir string) *AILogger {
	dir := filepath.Join(baseDir, "logs")
	os.MkdirAll(dir, 0755)
	return &AILogger{baseDir: dir, flushAt: 10}
}

// Log 记录一条日志
func (l *AILogger) Log(connID, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	line := fmt.Sprintf("[%s] [%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), connID, msg)
	fmt.Printf("[AI] %s", line)

	l.mu.Lock()
	defer l.mu.Unlock()

	l.buf = append(l.buf, line)
	if len(l.buf) >= l.flushAt {
		l.flush()
	}
}

// Flush 将缓冲写入文件
func (l *AILogger) Flush() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.flush()
}

func (l *AILogger) flush() {
	if len(l.buf) == 0 {
		return
	}

	logFile := filepath.Join(l.baseDir, fmt.Sprintf("%s.log", time.Now().Format("2006-01-02")))
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		l.buf = nil
		return
	}
	defer f.Close()

	for _, line := range l.buf {
		f.WriteString(line)
	}
	l.buf = nil
}
