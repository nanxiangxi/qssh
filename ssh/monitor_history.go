package ssh

import (
	"sync"
	"time"
)

// MonitorSnapshot 监控快照
type MonitorSnapshot struct {
	Timestamp   int64   `json:"timestamp"`
	CPUPercent  float64 `json:"cpuPercent"`
	MemPercent  float64 `json:"memPercent"`
	MemUsed     uint64  `json:"memUsed"`
	MemTotal    uint64  `json:"memTotal"`
	DiskPercent float64 `json:"diskPercent"`
	DiskUsed    uint64  `json:"diskUsed"`
	DiskTotal   uint64  `json:"diskTotal"`
	NetRx       uint64  `json:"netRx"`
	NetTx       uint64  `json:"netTx"`
}

// MonitorHistory 监控历史数据
type MonitorHistory struct {
	mu       sync.RWMutex
	data     map[string][]MonitorSnapshot // connID → snapshots
	maxAge   time.Duration
	maxItems int
}

// NewMonitorHistory 创建监控历史
func NewMonitorHistory() *MonitorHistory {
	mh := &MonitorHistory{
		data:     make(map[string][]MonitorSnapshot),
		maxAge:   1 * time.Hour,
		maxItems: 360, // 每 10 秒一个点，1 小时 = 360 个点
	}

	// 启动清理定时器
	go mh.cleanupLoop()

	return mh
}

// Add 添加监控快照
func (mh *MonitorHistory) Add(connID string, snapshot MonitorSnapshot) {
	mh.mu.Lock()
	defer mh.mu.Unlock()

	snapshots := mh.data[connID]
	snapshots = append(snapshots, snapshot)

	// 限制数量
	if len(snapshots) > mh.maxItems {
		snapshots = snapshots[len(snapshots)-mh.maxItems:]
	}

	mh.data[connID] = snapshots
}

// Get 获取最近 N 个快照
func (mh *MonitorHistory) Get(connID string, count int) []MonitorSnapshot {
	mh.mu.RLock()
	defer mh.mu.RUnlock()

	snapshots := mh.data[connID]
	if count <= 0 || count > len(snapshots) {
		count = len(snapshots)
	}

	// 返回最近 N 个
	start := len(snapshots) - count
	if start < 0 {
		start = 0
	}

	result := make([]MonitorSnapshot, count)
	copy(result, snapshots[start:])
	return result
}

// GetSince 获取指定时间以来的快照
func (mh *MonitorHistory) GetSince(connID string, since int64) []MonitorSnapshot {
	mh.mu.RLock()
	defer mh.mu.RUnlock()

	snapshots := mh.data[connID]
	var result []MonitorSnapshot
	for _, s := range snapshots {
		if s.Timestamp >= since {
			result = append(result, s)
		}
	}
	return result
}

// Clear 清除指定连接的历史数据
func (mh *MonitorHistory) Clear(connID string) {
	mh.mu.Lock()
	defer mh.mu.Unlock()
	delete(mh.data, connID)
}

// cleanupLoop 定期清理过期数据
func (mh *MonitorHistory) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		mh.cleanup()
	}
}

// cleanup 清理过期数据
func (mh *MonitorHistory) cleanup() {
	mh.mu.Lock()
	defer mh.mu.Unlock()

	cutoff := time.Now().Add(-mh.maxAge).UnixMilli()
	for connID, snapshots := range mh.data {
		valid := 0
		for _, s := range snapshots {
			if s.Timestamp >= cutoff {
				snapshots[valid] = s
				valid++
			}
		}
		if valid == 0 {
			delete(mh.data, connID)
		} else {
			mh.data[connID] = snapshots[:valid]
		}
	}
}
