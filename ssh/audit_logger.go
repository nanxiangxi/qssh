package ssh

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AuditEntry 审计日志条目
type AuditEntry struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	ConnID    string `json:"connId"`
	Host      string `json:"host"`
	Username  string `json:"username"`
	Action    string `json:"action"`    // command, upload, download, delete, rename, connect, disconnect
	Detail    string `json:"detail"`    // 具体内容
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}

// AuditLogger 操作审计日志
type AuditLogger struct {
	mu       sync.Mutex
	baseDir  string
	buf      []AuditEntry
	flushAt  int
}

// NewAuditLogger 创建审计日志器
func NewAuditLogger() *AuditLogger {
	baseDir := filepath.Join(GetDataDir(), "audit")
	os.MkdirAll(baseDir, 0755)
	return &AuditLogger{
		baseDir: baseDir,
		flushAt: 10,
	}
}

// Log 记录一条审计日志
func (al *AuditLogger) Log(connID, host, username, action, detail string, success bool, errMsg string) {
	entry := AuditEntry{
		ID:        fmt.Sprintf("audit_%d", time.Now().UnixNano()),
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		ConnID:    connID,
		Host:      host,
		Username:  username,
		Action:    action,
		Detail:    detail,
		Success:   success,
		Error:     errMsg,
	}

	al.mu.Lock()
	al.buf = append(al.buf, entry)
	if len(al.buf) >= al.flushAt {
		al.flush()
	}
	al.mu.Unlock()
}

// GetLogs 获取审计日志（最近 N 条）
func (al *AuditLogger) GetLogs(limit int) []AuditEntry {
	al.mu.Lock()
	defer al.mu.Unlock()

	// 先刷新缓冲区
	al.flush()

	// 读取今天的日志文件
	logFile := al.getLogFile(time.Now())
	data, err := os.ReadFile(logFile)
	if err != nil {
		return []AuditEntry{}
	}

	var entries []AuditEntry
	json.Unmarshal(data, &entries)

	// 返回最近 N 条
	if limit > 0 && len(entries) > limit {
		entries = entries[len(entries)-limit:]
	}

	return entries
}

// GetLogsByDate 按日期获取审计日志
func (al *AuditLogger) GetLogsByDate(date string) []AuditEntry {
	al.mu.Lock()
	defer al.mu.Unlock()

	al.flush()

	logFile := filepath.Join(al.baseDir, date+".json")
	data, err := os.ReadFile(logFile)
	if err != nil {
		return []AuditEntry{}
	}

	var entries []AuditEntry
	json.Unmarshal(data, &entries)
	return entries
}

// ExportLogs 导出审计日志为 JSON 字符串
func (al *AuditLogger) ExportLogs(startDate, endDate string) (string, error) {
	var allEntries []AuditEntry

	// 遍历日期范围内的日志文件
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return "", fmt.Errorf("无效的开始日期: %v", err)
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return "", fmt.Errorf("无效的结束日期: %v", err)
	}

	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		entries := al.GetLogsByDate(d.Format("2006-01-02"))
		allEntries = append(allEntries, entries...)
	}

	data, err := json.MarshalIndent(allEntries, "", "  ")
	if err != nil {
		return "", fmt.Errorf("序列化失败: %v", err)
	}

	return string(data), nil
}

// flush 将缓冲写入文件
func (al *AuditLogger) flush() {
	if len(al.buf) == 0 {
		return
	}

	logFile := al.getLogFile(time.Now())

	// 读取已有数据
	var existing []AuditEntry
	if data, err := os.ReadFile(logFile); err == nil {
		json.Unmarshal(data, &existing)
	}

	// 合并
	existing = append(existing, al.buf...)

	// 写入
	data, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		al.buf = nil
		return
	}

	os.WriteFile(logFile, data, 0644)
	al.buf = nil
}

// getLogFile 获取日志文件路径
func (al *AuditLogger) getLogFile(t time.Time) string {
	return filepath.Join(al.baseDir, t.Format("2006-01-02")+".json")
}
