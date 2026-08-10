package ssh

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// SSHConfigEntry SSH config 中的一个 Host 块
type SSHConfigEntry struct {
	Host       string // 原始 Host 名称
	HostName   string // 实际主机地址
	Port       int
	User       string
	IdentityFile string
}

// ParseSSHConfig 解析 SSH config 文件，返回连接列表
func ParseSSHConfig(path string) ([]SSHConfigEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开 SSH config 失败: %w", err)
	}
	defer file.Close()

	var entries []SSHConfigEntry
	var current *SSHConfigEntry

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value := parseConfigLine(line)
		if key == "" {
			continue
		}

		keyLower := strings.ToLower(key)

		if keyLower == "host" {
			// 新的 Host 块
			if current != nil && current.Host != "" {
				entries = append(entries, *current)
			}
			current = &SSHConfigEntry{
				Host: value,
				Port: 22,
			}
			continue
		}

		if current == nil {
			continue
		}

		switch keyLower {
		case "hostname":
			current.HostName = value
		case "port":
			if p, err := strconv.Atoi(value); err == nil && p > 0 {
				current.Port = p
			}
		case "user":
			current.User = value
		case "identityfile":
			// 展开 ~ 为用户主目录
			if strings.HasPrefix(value, "~") {
				if home, err := os.UserHomeDir(); err == nil {
					value = filepath.Join(home, value[1:])
				}
			}
			current.IdentityFile = value
		}
	}

	// 保存最后一个 Host 块
	if current != nil && current.Host != "" {
		entries = append(entries, *current)
	}

	// 过滤掉通配符 Host（如 *）
	var result []SSHConfigEntry
	for _, e := range entries {
		if e.Host == "*" || strings.Contains(e.Host, "*") || strings.Contains(e.Host, "?") {
			continue
		}
		// 如果没有 HostName，使用 Host 作为主机地址
		if e.HostName == "" {
			e.HostName = e.Host
		}
		// 如果没有 User，使用默认值
		if e.User == "" {
			e.User = "root"
		}
		result = append(result, e)
	}

	return result, nil
}

// parseConfigLine 解析一行 SSH config，返回 key 和 value
func parseConfigLine(line string) (string, string) {
	// 跳过注释
	if idx := strings.Index(line, "#"); idx == 0 {
		return "", ""
	}

	// 分割 key 和 value（第一个空格）
	parts := strings.SplitN(line, " ", 2)
	if len(parts) < 2 {
		return strings.TrimSpace(parts[0]), ""
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	// 去掉引号
	if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
		value = value[1 : len(value)-1]
	}

	return key, value
}
