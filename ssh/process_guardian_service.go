package ssh

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// GuardianProcess 守护进程信息
type GuardianProcess struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Command     string `json:"command"`
	WorkDir     string `json:"workDir"`
	Status      string `json:"status"` // running / stopped / failed / unknown
	PID         int    `json:"pid"`
	AutoRestart bool   `json:"autoRestart"`
	LogPath     string `json:"logPath"`
	CreatedAt   int64  `json:"createdAt"`
	Restarts    int    `json:"restarts"`
}

// ProcessGuardianService 进程守护服务（基于 systemd，兼容所有现代 Linux）
type ProcessGuardianService struct {
	sshSvc *SSHService
}

func NewProcessGuardianService(sshSvc *SSHService) *ProcessGuardianService {
	return &ProcessGuardianService{sshSvc: sshSvc}
}

// runCmd 执行远程命令
func (s *ProcessGuardianService) runCmd(connID, cmd string) (string, error) {
	client, err := s.sshSvc.GetClient(connID)
	if err != nil {
		return "", fmt.Errorf("连接不存在: %v", err)
	}
	if !client.IsConnected() {
		return "", fmt.Errorf("SSH连接已断开")
	}
	result, err := client.ExecuteCommand(cmd)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(result.Stdout), nil
}

// detectInit 检测 init 系统
func (s *ProcessGuardianService) detectInit(connID string) string {
	// systemd
	if out, _ := s.runCmd(connID, "pidof systemd 2>/dev/null"); out != "" {
		return "systemd"
	}
	// OpenRC
	if _, err := s.runCmd(connID, "rc-status --help 2>/dev/null"); err == nil {
		return "openrc"
	}
	// sysvinit
	if _, err := s.runCmd(connID, "service --status-all 2>/dev/null"); err == nil {
		return "sysvinit"
	}
	return "unknown"
}

// GetGuardians 获取所有守护进程
func (s *ProcessGuardianService) GetGuardians(connID string) []GuardianProcess {
	initSystem := s.detectInit(connID)
	switch initSystem {
	case "systemd":
		return s.getSystemdServices(connID)
	default:
		return s.getSystemdServices(connID) // 默认尝试 systemd
	}
}

// getSystemdServices 获取自定义的 pzssh 守护服务
func (s *ProcessGuardianService) getSystemdServices(connID string) []GuardianProcess {
	// 列出所有 pzssh-managed 服务
	out, _ := s.runCmd(connID, "systemctl list-units --type=service --all --no-pager 2>/dev/null | grep 'pzssh-' | awk '{print $1}'")
	if out == "" {
		return []GuardianProcess{}
	}

	var processes []GuardianProcess
	for _, unit := range strings.Split(out, "\n") {
		unit = strings.TrimSpace(unit)
		if unit == "" {
			continue
		}
		// 获取服务详情
		name := strings.TrimPrefix(unit, "pzssh-")
		name = strings.TrimSuffix(name, ".service")

		status := s.getSystemdStatus(connID, unit)
		pid := s.getSystemdPID(connID, unit)
		cmd := s.getSystemdCommand(connID, name)
		restarts := s.getSystemdRestarts(connID, unit)

		processes = append(processes, GuardianProcess{
			ID:          name,
			Name:        name,
			Command:     cmd,
			Status:      status,
			PID:         pid,
			AutoRestart: true,
			Restarts:    restarts,
		})
	}

	return processes
}

func (s *ProcessGuardianService) getSystemdStatus(connID, unit string) string {
	out, _ := s.runCmd(connID, fmt.Sprintf("systemctl is-active %s 2>/dev/null", unit))
	switch strings.TrimSpace(out) {
	case "active":
		return "running"
	case "activating":
		return "running"
	case "reloading":
		return "running"
	case "inactive":
		return "stopped"
	case "deactivating":
		return "stopped"
	case "failed":
		return "failed"
	default:
		// 尝试用 status 命令获取更详细信息
		statusOut, _ := s.runCmd(connID, fmt.Sprintf("systemctl show %s --property=ActiveState --value 2>/dev/null", unit))
		switch strings.TrimSpace(statusOut) {
		case "active", "activating", "reloading":
			return "running"
		case "inactive", "deactivating":
			return "stopped"
		case "failed":
			return "failed"
		default:
			return "stopped"
		}
	}
}

func (s *ProcessGuardianService) getSystemdPID(connID, unit string) int {
	out, _ := s.runCmd(connID, fmt.Sprintf("systemctl show %s --property=MainPID --value 2>/dev/null", unit))
	var pid int
	fmt.Sscanf(strings.TrimSpace(out), "%d", &pid)
	return pid
}

func (s *ProcessGuardianService) getSystemdCommand(connID, name string) string {
	out, _ := s.runCmd(connID, fmt.Sprintf("cat /etc/systemd/system/pzssh-%s.service 2>/dev/null | grep ExecStart | head -1 | sed 's/ExecStart=//'", name))
	cmd := strings.TrimSpace(out)
	// 去掉 /bin/sh -c '...' 包装
	if strings.HasPrefix(cmd, "/bin/sh -c '") && strings.HasSuffix(cmd, "'") {
		cmd = cmd[len("/bin/sh -c '") : len(cmd)-1]
	}
	return cmd
}

func (s *ProcessGuardianService) getSystemdRestarts(connID, unit string) int {
	out, _ := s.runCmd(connID, fmt.Sprintf("systemctl show %s --property=NRestarts --value 2>/dev/null", unit))
	var n int
	fmt.Sscanf(strings.TrimSpace(out), "%d", &n)
	return n
}

// CreateGuardian 创建守护进程
func (s *ProcessGuardianService) CreateGuardian(connID, name, command, workDir string, autoRestart bool) error {
	if name == "" || command == "" {
		return fmt.Errorf("名称和命令不能为空")
	}

	// 清理名称（只允许字母数字和连字符）
	safeName := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, name)

	if workDir == "" {
		workDir = "/tmp"
	}

	restartPolicy := "on-failure"
	if !autoRestart {
		restartPolicy = "no"
	}

	// 生成 systemd service 文件（日志由 systemd journal 自动管理）
	serviceContent := fmt.Sprintf(`[Unit]
Description=pzssh guardian: %s
After=network.target

[Service]
Type=simple
ExecStart=/bin/sh -c '%s'
WorkingDirectory=%s
Restart=%s
RestartSec=3

[Install]
WantedBy=multi-user.target
`, name, command, workDir, restartPolicy)

	// 写入 service 文件（使用 base64 避免特殊字符问题）
	unitName := fmt.Sprintf("pzssh-%s.service", safeName)
	servicePath := fmt.Sprintf("/etc/systemd/system/%s", unitName)

	encoded := base64.StdEncoding.EncodeToString([]byte(serviceContent))
	writeCmd := fmt.Sprintf("echo %s | base64 -d > %s", encoded, servicePath)
	if _, err := s.runCmd(connID, writeCmd); err != nil {
		return fmt.Errorf("写入服务文件失败: %v", err)
	}

	// 重载 systemd
	s.runCmd(connID, "systemctl daemon-reload")

	// 启用服务
	s.runCmd(connID, fmt.Sprintf("systemctl enable %s 2>/dev/null", unitName))

	// 启动服务
	if _, err := s.runCmd(connID, fmt.Sprintf("systemctl start %s", unitName)); err != nil {
		return fmt.Errorf("启动服务失败: %v", err)
	}

	return nil
}

// StartGuardian 启动守护进程
func (s *ProcessGuardianService) StartGuardian(connID, name string) error {
	unitName := fmt.Sprintf("pzssh-%s.service", name)
	_, err := s.runCmd(connID, fmt.Sprintf("systemctl start %s", unitName))
	return err
}

// StopGuardian 停止守护进程
func (s *ProcessGuardianService) StopGuardian(connID, name string) error {
	unitName := fmt.Sprintf("pzssh-%s.service", name)
	_, err := s.runCmd(connID, fmt.Sprintf("systemctl stop %s", unitName))
	return err
}

// RestartGuardian 重启守护进程
func (s *ProcessGuardianService) RestartGuardian(connID, name string) error {
	unitName := fmt.Sprintf("pzssh-%s.service", name)
	_, err := s.runCmd(connID, fmt.Sprintf("systemctl restart %s", unitName))
	return err
}

// DeleteGuardian 删除守护进程
func (s *ProcessGuardianService) DeleteGuardian(connID, name string) error {
	unitName := fmt.Sprintf("pzssh-%s.service", name)
	servicePath := fmt.Sprintf("/etc/systemd/system/%s", unitName)

	// 停止并禁用
	s.runCmd(connID, fmt.Sprintf("systemctl stop %s 2>/dev/null", unitName))
	s.runCmd(connID, fmt.Sprintf("systemctl disable %s 2>/dev/null", unitName))

	// 删除文件
	s.runCmd(connID, fmt.Sprintf("rm -f %s", servicePath))
	s.runCmd(connID, fmt.Sprintf("rm -f /var/log/pzssh-%s.log", name))

	// 重载
	s.runCmd(connID, "systemctl daemon-reload")

	return nil
}

// GetGuardianLogs 获取守护进程日志（从 systemd journal 读取）
func (s *ProcessGuardianService) GetGuardianLogs(connID, name string, lines int) string {
	if lines <= 0 {
		lines = 100
	}
	unitName := fmt.Sprintf("pzssh-%s.service", name)
	out, _ := s.runCmd(connID, fmt.Sprintf("journalctl -u %s -n %d --no-pager -o short-iso 2>/dev/null", unitName, lines))
	return out
}

// GetGuardianStats 获取守护进程统计
func (s *ProcessGuardianService) GetGuardianStats(connID, name string) map[string]interface{} {
	unitName := fmt.Sprintf("pzssh-%s.service", name)

	status := s.getSystemdStatus(connID, unitName)
	pid := s.getSystemdPID(connID, unitName)
	restarts := s.getSystemdRestarts(connID, unitName)

	// 获取运行时间
	uptime := ""
	if pid > 0 {
		out, _ := s.runCmd(connID, fmt.Sprintf("ps -o etime= -p %d 2>/dev/null", pid))
		uptime = strings.TrimSpace(out)
	}

	// 获取内存使用
	mem := ""
	if pid > 0 {
		out, _ := s.runCmd(connID, fmt.Sprintf("ps -o rss= -p %d 2>/dev/null", pid))
		mem = strings.TrimSpace(out)
	}

	return map[string]interface{}{
		"status":   status,
		"pid":      pid,
		"restarts": restarts,
		"uptime":   uptime,
		"memory":   mem,
	}
}

// ClearGuardianLogs 清空守护进程日志（清空 systemd journal 中该服务的记录）
func (s *ProcessGuardianService) ClearGuardianLogs(connID, name, logPath string) error {
	unitName := fmt.Sprintf("pzssh-%s.service", name)

	fmt.Printf("[Guardian] ClearGuardianLogs: name=%s, unit=%s\n", name, unitName)

	// 使用 journalctl 清空指定服务的日志
	out, err := s.runCmd(connID, fmt.Sprintf("journalctl --rotate --unit=%s 2>&1 && journalctl --vacuum-time=1s --unit=%s 2>&1", unitName, unitName))
	fmt.Printf("[Guardian] 清空结果: %s, err=%v\n", strings.TrimSpace(out), err)

	return err
}
