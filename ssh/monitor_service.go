package ssh

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// SystemStats 系统整体统计
type SystemStats struct {
	Timestamp   int64        `json:"timestamp"`
	Uptime      string       `json:"uptime"`       // 系统运行时长
	CPU         CPUStats     `json:"cpu"`
	Memory      MemoryStats  `json:"memory"`
	Disk        DiskStats    `json:"disk"`
	Network     NetworkStats `json:"network"`
}

// CPUStats CPU 统计
type CPUStats struct {
	UsagePercent float64   `json:"usagePercent"` // 使用率 %
	Cores        int       `json:"cores"`        // 核心数
	PerCPUUsage  []float64 `json:"perCpuUsage"`  // 每个核心的使用率
	LoadAvg      LoadAvg   `json:"loadAvg"`      // 平均负载
}

// LoadAvg 平均负载
type LoadAvg struct {
	Load1  float64 `json:"load1"`  // 1分钟
	Load5  float64 `json:"load5"`  // 5分钟
	Load15 float64 `json:"load15"` // 15分钟
}

// MemoryStats 内存统计
type MemoryStats struct {
	Total       uint64  `json:"total"`       // 总内存 (bytes)
	Used        uint64  `json:"used"`        // 已使用 (bytes)
	Free        uint64  `json:"free"`        // 空闲 (bytes)
	UsedPercent float64 `json:"usedPercent"` // 使用率 %
	Cached      uint64  `json:"cached"`      // 缓存 (bytes)
	SwapTotal   uint64  `json:"swapTotal"`   // 交换区总大小
	SwapUsed    uint64  `json:"swapUsed"`    // 交换区已使用
}

// DiskStats 磁盘统计
type DiskStats struct {
	Partitions []DiskPartition `json:"partitions"` // 分区列表
	IOStats    DiskIOStats     `json:"ioStats"`    // IO 统计
}

// DiskPartition 磁盘分区
type DiskPartition struct {
	Device      string  `json:"device"`      // 设备名
	Mountpoint  string  `json:"mountpoint"`  // 挂载点
	Fstype      string  `json:"fstype"`      // 文件系统类型
	Total       uint64  `json:"total"`       // 总容量 (bytes)
	Used        uint64  `json:"used"`        // 已使用 (bytes)
	Free        uint64  `json:"free"`        // 空闲 (bytes)
	UsedPercent float64 `json:"usedPercent"` // 使用率 %
}

// DiskIOStats 磁盘 IO 统计
type DiskIOStats struct {
	ReadBytes  uint64 `json:"readBytes"`  // 读取字节数
	WriteBytes uint64 `json:"writeBytes"` // 写入字节数
	ReadCount  uint64 `json:"readCount"`  // 读取次数
	WriteCount uint64 `json:"writeCount"` // 写入次数
}

// NetworkStats 网络统计
type NetworkStats struct {
	Interfaces []NetInterface `json:"interfaces"` // 网络接口列表
	TotalRx    uint64         `json:"totalRx"`    // 总接收字节
	TotalTx    uint64         `json:"totalTx"`    // 总发送字节
}

// NetInterface 网络接口
type NetInterface struct {
	Name        string `json:"name"`        // 接口名称
	BytesSent   uint64 `json:"bytesSent"`   // 发送字节
	BytesRecv   uint64 `json:"bytesRecv"`   // 接收字节
	PacketsSent uint64 `json:"packetsSent"` // 发送包数
	PacketsRecv uint64 `json:"packetsRecv"` // 接收包数
}

// ProcessInfo 进程信息
type ProcessInfo struct {
	PID        int32   `json:"pid"`        // 进程 ID
	Name       string  `json:"name"`       // 进程名称
	CPUPercent float64 `json:"cpuPercent"` // CPU 使用率 %
	MemPercent float64 `json:"memPercent"` // 内存使用率 %
	MemRSS     uint64  `json:"memRss"`     // 物理内存使用 (bytes)
	Status     string  `json:"status"`     // 进程状态
	Username   string  `json:"username"`   // 用户名
	StartTime  string  `json:"startTime"`  // 启动时间
	ElapsedTime string `json:"elapsedTime"` // 运行时长
	Cmdline    string  `json:"cmdline"`    // 命令行
	NumThreads int32   `json:"numThreads"` // 线程数
	Priority   int     `json:"priority"`   // 优先级
	Nice       int     `json:"nice"`       // Nice 值
}

// GetSystemStats 获取系统资源统计（通过 SSH 在远程服务器执行命令）
func (c *SSHClient) GetSystemStats(ctx context.Context) (*SystemStats, error) {
	if !c.isConnected {
		return nil, fmt.Errorf("SSH 未连接")
	}

	stats := &SystemStats{
		Timestamp: time.Now().UnixMilli(),
	}

	// 0. 获取系统运行时长
	// 直接读取 /proc/uptime 并格式化为中文
	uptimeRaw, err := c.ExecuteCommand("cat /proc/uptime | awk '{print $1}'")
	if err == nil && uptimeRaw.Success {
		seconds := parseFloat64(strings.TrimSpace(uptimeRaw.Stdout))
		stats.Uptime = formatUptime(seconds)
	} else {
		stats.Uptime = "未知"
	}

	// 1. 获取 CPU 使用率
	cpuResult, err := c.ExecuteCommand("top -bn1 | grep 'Cpu(s)' | awk '{print $2}'")
	if err == nil && cpuResult.Success {
		cpuPercent, parseErr := strconv.ParseFloat(strings.TrimSpace(cpuResult.Stdout), 64)
		if parseErr == nil {
			stats.CPU.UsagePercent = cpuPercent
		}
	}

	// 2. 获取 CPU 核心数
	coresResult, err := c.ExecuteCommand("nproc")
	if err == nil && coresResult.Success {
		cores, parseErr := strconv.Atoi(strings.TrimSpace(coresResult.Stdout))
		if parseErr == nil {
			stats.CPU.Cores = cores
		}
	}

	// 3. 获取平均负载
	loadResult, err := c.ExecuteCommand("cat /proc/loadavg | awk '{print $1,$2,$3}'")
	if err == nil && loadResult.Success {
		fields := strings.Fields(strings.TrimSpace(loadResult.Stdout))
		if len(fields) >= 3 {
			stats.CPU.LoadAvg.Load1, _ = strconv.ParseFloat(fields[0], 64)
			stats.CPU.LoadAvg.Load5, _ = strconv.ParseFloat(fields[1], 64)
			stats.CPU.LoadAvg.Load15, _ = strconv.ParseFloat(fields[2], 64)
		}
	}

	// 4. 获取内存信息
	memResult, err := c.ExecuteCommand("free -b | grep '^Mem:'")
	if err == nil && memResult.Success {
		fields := strings.Fields(strings.TrimSpace(memResult.Stdout))
		if len(fields) >= 7 {
			stats.Memory.Total, _ = strconv.ParseUint(fields[1], 10, 64)
			stats.Memory.Used, _ = strconv.ParseUint(fields[2], 10, 64)
			stats.Memory.Free, _ = strconv.ParseUint(fields[3], 10, 64)
			stats.Memory.Cached, _ = strconv.ParseUint(fields[6], 10, 64)
			if stats.Memory.Total > 0 {
				stats.Memory.UsedPercent = float64(stats.Memory.Used) / float64(stats.Memory.Total) * 100
			}
		}
	}

	// 5. 获取交换区信息
	swapResult, err := c.ExecuteCommand("free -b | grep '^Swap:'")
	if err == nil && swapResult.Success {
		fields := strings.Fields(strings.TrimSpace(swapResult.Stdout))
		if len(fields) >= 3 {
			stats.Memory.SwapTotal, _ = strconv.ParseUint(fields[1], 10, 64)
			stats.Memory.SwapUsed, _ = strconv.ParseUint(fields[2], 10, 64)
		}
	}

	// 6. 获取磁盘分区信息
	dfResult, err := c.ExecuteCommand("df -B1 --output=target,size,used,avail,pcent,fstype | tail -n +2")
	if err == nil && dfResult.Success {
		lines := strings.Split(strings.TrimSpace(dfResult.Stdout), "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) >= 6 {
				mountpoint := fields[0]
				total, _ := strconv.ParseUint(fields[1], 10, 64)
				used, _ := strconv.ParseUint(fields[2], 10, 64)
				free, _ := strconv.ParseUint(fields[3], 10, 64)
				usedPercentStr := strings.TrimSuffix(fields[4], "%")
				usedPercent, _ := strconv.ParseFloat(usedPercentStr, 64)
				fstype := fields[5]

				// 跳过虚拟文件系统
				if fstype == "tmpfs" || fstype == "devtmpfs" || mountpoint == "/run" || mountpoint == "/dev/shm" {
					continue
				}

				stats.Disk.Partitions = append(stats.Disk.Partitions, DiskPartition{
					Device:      "",
					Mountpoint:  mountpoint,
					Fstype:      fstype,
					Total:       total,
					Used:        used,
					Free:        free,
					UsedPercent: usedPercent,
				})
			}
		}
	}

	// 7. 获取磁盘 IO 统计
	ioResult, err := c.ExecuteCommand("cat /proc/diskstats | awk '{read+=$4; write+=$8} END {print read, write}'")
	if err == nil && ioResult.Success {
		fields := strings.Fields(strings.TrimSpace(ioResult.Stdout))
		if len(fields) >= 2 {
			// /proc/diskstats 的单位是扇区（512字节）
			readSectors, _ := strconv.ParseUint(fields[0], 10, 64)
			writeSectors, _ := strconv.ParseUint(fields[1], 10, 64)
			stats.Disk.IOStats.ReadBytes = readSectors * 512
			stats.Disk.IOStats.WriteBytes = writeSectors * 512
		}
	}

	// 8. 获取网络接口信息
	netResult, err := c.ExecuteCommand("cat /proc/net/dev | tail -n +3")
	if err == nil && netResult.Success {
		lines := strings.Split(strings.TrimSpace(netResult.Stdout), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			// 移除冒号并分割
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}
			ifaceName := strings.TrimSpace(parts[0])
			fields := strings.Fields(parts[1])
			if len(fields) >= 10 {
				bytesRecv, _ := strconv.ParseUint(fields[0], 10, 64)
				packetsRecv, _ := strconv.ParseUint(fields[1], 10, 64)
				bytesSent, _ := strconv.ParseUint(fields[8], 10, 64)
				packetsSent, _ := strconv.ParseUint(fields[9], 10, 64)

				stats.Network.Interfaces = append(stats.Network.Interfaces, NetInterface{
					Name:        ifaceName,
					BytesSent:   bytesSent,
					BytesRecv:   bytesRecv,
					PacketsSent: packetsSent,
					PacketsRecv: packetsRecv,
				})
				stats.Network.TotalRx += bytesRecv
				stats.Network.TotalTx += bytesSent
			}
		}
	}

	return stats, nil
}

// GetProcessList 获取进程列表
func (c *SSHClient) GetProcessList(ctx context.Context) ([]ProcessInfo, error) {
	if !c.isConnected {
		return nil, fmt.Errorf("SSH 会话未建立")
	}

	// 通过 SSH 执行命令获取进程信息（包含运行时长）
	cmd := "ps -eo pid,user,%cpu,%mem,rss,stat,start,etime,args --sort=-%cpu"
	result, err := c.ExecuteCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("执行 ps 命令失败: %v", err)
	}

	processes, err := parsePsOutput(result.Stdout)
	if err != nil {
		return nil, fmt.Errorf("解析进程信息失败: %v", err)
	}

	return processes, nil
}

// KillProcess 终止进程
func (c *SSHClient) KillProcess(ctx context.Context, pid int32) error {
	if !c.isConnected {
		return fmt.Errorf("SSH 未连接")
	}

	// 执行 kill 命令
	cmd := fmt.Sprintf("kill -9 %d", pid)
	result, err := c.ExecuteCommand(cmd)
	if err != nil {
		return fmt.Errorf("执行 kill 命令失败: %v", err)
	}

	if !result.Success {
		return fmt.Errorf("终止进程失败: %s", result.Stderr)
	}

	return nil
}

// SendSignal 向进程发送信号
func (c *SSHClient) SendSignal(ctx context.Context, pid int32, signal string) error {
	if !c.isConnected {
		return fmt.Errorf("SSH 未连接")
	}

	// 执行 kill 命令发送信号
	cmd := fmt.Sprintf("kill -%s %d", signal, pid)
	result, err := c.ExecuteCommand(cmd)
	if err != nil {
		return fmt.Errorf("发送信号失败: %v", err)
	}

	if !result.Success {
		return fmt.Errorf("发送信号失败: %s", result.Stderr)
	}

	return nil
}

// GetProcessDetail 获取进程详细信息
func (c *SSHClient) GetProcessDetail(ctx context.Context, pid int32) (string, error) {
	if !c.isConnected {
		return "", fmt.Errorf("SSH 未连接")
	}

	// 获取进程详细信息
	cmd := fmt.Sprintf("ps -p %d -f", pid)
	result, err := c.ExecuteCommand(cmd)
	if err != nil {
		return "", fmt.Errorf("获取进程信息失败: %v", err)
	}

	if !result.Success {
		return "", fmt.Errorf("进程不存在或无权访问")
	}

	return result.Stdout, nil
}

// parsePsOutput 解析 ps 命令输出
func parsePsOutput(output string) ([]ProcessInfo, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("无效的 ps 输出")
	}

	var processes []ProcessInfo

	// 跳过标题行
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}

		pid := parseInt32(fields[0])
		username := fields[1]
		cpuPercent := parseFloat64(fields[2])
		memPercent := parseFloat64(fields[3])
		memRSS := parseUint64(fields[4]) * 1024 // KB 转 bytes
		statusRaw := fields[5]     // STAT 列
		startTime := fields[6]     // START 列
		elapsedTime := fields[7]   // ELAPSED 列（运行时长）
		cmdline := strings.Join(fields[8:], " ")

		// 转换状态为中文
		status := convertProcessStatus(statusRaw)

		// 提取进程名称（命令的第一部分）
		name := extractProcessName(cmdline)

		processes = append(processes, ProcessInfo{
			PID:         pid,
			Name:        name,
			CPUPercent:  cpuPercent,
			MemPercent:  memPercent,
			MemRSS:      memRSS,
			Status:      status,
			StartTime:   startTime,
			ElapsedTime: elapsedTime,
			Username:    username,
			Cmdline:     cmdline,
		})
	}

	return processes, nil
}

// convertProcessStatus 转换进程状态为中文
func convertProcessStatus(status string) string {
	if len(status) == 0 {
		return "未知"
	}

	// 取第一个字符作为主要状态
	mainStatus := status[0]

	switch mainStatus {
	case 'R':
		return "运行中"
	case 'S':
		return "休眠"
	case 'D':
		return "不可中断"
	case 'Z':
		return "僵尸"
	case 'T':
		return "已停止"
	case 't':
		return "追踪停止"
	case 'X':
		return "死亡"
	case 'I':
		return "空闲"
	default:
		return string(mainStatus)
	}
}

// formatUptime 格式化系统运行时间
func formatUptime(seconds float64) string {
	days := int(seconds) / 86400
	hours := (int(seconds) % 86400) / 3600
	minutes := (int(seconds) % 3600) / 60

	if days > 0 {
		return fmt.Sprintf("%d天%d小时%d分钟", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%d小时%d分钟", hours, minutes)
	} else {
		return fmt.Sprintf("%d分钟", minutes)
	}
}

// 辅助函数
func parseInt32(s string) int32 {
	var n int32
	fmt.Sscanf(s, "%d", &n)
	return n
}

func parseFloat64(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

func parseUint64(s string) uint64 {
	var n uint64
	fmt.Sscanf(s, "%d", &n)
	return n
}

func extractProcessName(cmdline string) string {
	// 提取命令的第一部分作为进程名
	parts := strings.Fields(cmdline)
	if len(parts) > 0 {
		name := parts[0]
		// 去除路径前缀
		if idx := strings.LastIndex(name, "/"); idx >= 0 {
			name = name[idx+1:]
		}
		return name
	}
	return cmdline
}
