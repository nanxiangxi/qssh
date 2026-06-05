package ssh

import (
	"fmt"
	"strings"
	"sync"
)

// FirewallRule 防火墙规则
type FirewallRule struct {
	Index    int    `json:"index"`    // 规则序号
	Chain    string `json:"chain"`    // INPUT / OUTPUT / FORWARD / zone
	Target   string `json:"target"`   // ACCEPT / DROP / REJECT / allow / deny
	Protocol string `json:"protocol"` // tcp / udp / icmp / all
	Source   string `json:"source"`   // 源地址
	Dest     string `json:"dest"`     // 目标地址
	Port     string `json:"port"`     // 端口
	Comment  string `json:"comment"`  // 备注
	Raw      string `json:"raw"`      // 原始行
}

// FirewallInfo 防火墙信息
type FirewallInfo struct {
	Type       string         `json:"type"`       // iptables / firewalld / ufw / unknown
	Status     string         `json:"status"`     // active / inactive / unknown
	Rules      []FirewallRule `json:"rules"`
	RawOutput  string         `json:"rawOutput"`  // 原始输出
	Chains     []string       `json:"chains"`     // 可用的链/区域
}

// FirewallService 防火墙管理服务
type FirewallService struct {
	mu     sync.RWMutex
	sshSvc *SSHService
}

// NewFirewallService 创建防火墙服务
func NewFirewallService(sshSvc *SSHService) *FirewallService {
	return &FirewallService{
		sshSvc: sshSvc,
	}
}

// runCmd 执行远程命令
func (s *FirewallService) runCmd(connID, cmd string) (string, error) {
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

// detectType 检测防火墙类型（优先级：ufw > firewalld > iptables）
func (s *FirewallService) detectType(connID string) string {
	// ufw（检查是否安装，不管是否启用）
	if out, err := s.runCmd(connID, "which ufw 2>/dev/null || command -v ufw 2>/dev/null"); err == nil && strings.TrimSpace(out) != "" {
		return "ufw"
	}
	// firewalld（检查是否安装，不管是否运行）
	if out, err := s.runCmd(connID, "which firewall-cmd 2>/dev/null || command -v firewall-cmd 2>/dev/null"); err == nil && strings.TrimSpace(out) != "" {
		return "firewalld"
	}
	// iptables
	if out, err := s.runCmd(connID, "which iptables 2>/dev/null || command -v iptables 2>/dev/null"); err == nil && strings.TrimSpace(out) != "" {
		return "iptables"
	}
	return "unknown"
}

// GetFirewallInfo 获取防火墙信息
func (s *FirewallService) GetFirewallInfo(connID string) *FirewallInfo {
	fwType := s.detectType(connID)
	info := &FirewallInfo{Type: fwType, Rules: []FirewallRule{}, Chains: []string{}}

	switch fwType {
	case "iptables":
		s.loadIptables(connID, info)
	case "firewalld":
		s.loadFirewalld(connID, info)
	case "ufw":
		s.loadUfw(connID, info)
	default:
		info.Status = "unknown"
		info.RawOutput = "未检测到支持的防火墙（iptables / firewalld / ufw）"
	}
	return info
}

// ==================== iptables ====================

func (s *FirewallService) loadIptables(connID string, info *FirewallInfo) {
	info.Chains = []string{"INPUT", "OUTPUT", "FORWARD"}

	hasRules := false
	for _, chain := range info.Chains {
		chainOut, _ := s.runCmd(connID, fmt.Sprintf("iptables -L %s -n -v --line-numbers 2>/dev/null", chain))
		info.RawOutput += chainOut + "\n\n"
		s.parseIptablesChain(chain, chainOut, info)
		// 检查是否有实际规则（不只是标题和策略行）
		lines := strings.Split(chainOut, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "Chain") || strings.HasPrefix(line, "num") || strings.HasPrefix(line, "target") {
				continue
			}
			hasRules = true
		}
	}

	if hasRules {
		info.Status = "active"
	} else {
		info.Status = "inactive"
	}
}

func (s *FirewallService) parseIptablesChain(chain, output string, info *FirewallInfo) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Chain") || strings.HasPrefix(line, "num") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		// num pkts bytes target prot opt source dest [extra...]
		rule := FirewallRule{
			Raw:      line,
			Chain:    chain,
			Protocol: fields[3],
			Target:   fields[2],
			Source:   fields[7],
			Dest:     fields[8],
		}
		// 解析序号
		fmt.Sscanf(fields[0], "%d", &rule.Index)
		// 解析端口
		for i := 9; i < len(fields); i++ {
			if fields[i] == "dpt:" && i+1 < len(fields) {
				rule.Port = fields[i+1]
			}
			if fields[i] == "spt:" && i+1 < len(fields) {
				// 源端口暂不处理
			}
		}
		// 解析注释
		if idx := strings.Index(line, "/* "); idx != -1 {
			endIdx := strings.Index(line[idx:], " */")
			if endIdx != -1 {
				rule.Comment = line[idx+3 : idx+endIdx]
			}
		}
		info.Rules = append(info.Rules, rule)
	}
}

// AddIptablesRule 添加 iptables 规则
func (s *FirewallService) AddIptablesRule(connID, chain, target, protocol, port, source, comment string) error {
	if chain == "" {
		chain = "INPUT"
	}
	if target == "" {
		target = "ACCEPT"
	}
	if protocol == "" {
		protocol = "tcp"
	}
	if source == "" {
		source = "0.0.0.0/0"
	}

	cmd := fmt.Sprintf("iptables -A %s -p %s -s %s -j %s", chain, protocol, source, target)
	if port != "" {
		cmd += fmt.Sprintf(" --dport %s", port)
	}
	if comment != "" {
		cmd += fmt.Sprintf(" -m comment --comment '%s'", comment)
	}

	_, err := s.runCmd(connID, cmd)
	if err != nil {
		return fmt.Errorf("添加规则失败: %v", err)
	}
	// 保存规则
	s.runCmd(connID, "iptables-save > /etc/iptables/rules.v4 2>/dev/null || true")
	return nil
}

// DeleteIptablesRule 删除 iptables 规则
func (s *FirewallService) DeleteIptablesRule(connID, chain string, index int) error {
	if chain == "" {
		chain = "INPUT"
	}
	cmd := fmt.Sprintf("iptables -D %s %d", chain, index)
	_, err := s.runCmd(connID, cmd)
	if err != nil {
		return fmt.Errorf("删除规则失败: %v", err)
	}
	s.runCmd(connID, "iptables-save > /etc/iptables/rules.v4 2>/dev/null || true")
	return nil
}

// ==================== firewalld ====================

func (s *FirewallService) loadFirewalld(connID string, info *FirewallInfo) {
	// 检查是否运行
	state, _ := s.runCmd(connID, "firewall-cmd --state 2>/dev/null")
	if strings.TrimSpace(state) == "running" {
		info.Status = "active"
	} else {
		info.Status = "inactive"
	}

	info.Chains = []string{"public", "trusted", "drop"}

	// 获取默认区域
	defaultZone, _ := s.runCmd(connID, "firewall-cmd --get-default-zone 2>/dev/null")
	if defaultZone != "" {
		info.Chains = []string{defaultZone}
	}

	// 获取所有区域
	zones, _ := s.runCmd(connID, "firewall-cmd --get-zones 2>/dev/null")
	if zones != "" {
		info.Chains = strings.Fields(zones)
	}

	// 获取当前区域的规则
	zone := defaultZone
	if zone == "" {
		zone = "public"
	}

	out, _ := s.runCmd(connID, fmt.Sprintf("firewall-cmd --zone=%s --list-all 2>/dev/null", zone))
	info.RawOutput = out

	// 解析端口
	portsOut, _ := s.runCmd(connID, fmt.Sprintf("firewall-cmd --zone=%s --list-ports 2>/dev/null", zone))
	for i, port := range strings.Fields(portsOut) {
		info.Rules = append(info.Rules, FirewallRule{
			Index:    i + 1,
			Chain:    zone,
			Target:   "allow",
			Protocol: "tcp",
			Port:     port,
			Raw:      port,
		})
	}

	// 解析服务
	servicesOut, _ := s.runCmd(connID, fmt.Sprintf("firewall-cmd --zone=%s --list-services 2>/dev/null", zone))
	for _, svc := range strings.Fields(servicesOut) {
		info.Rules = append(info.Rules, FirewallRule{
			Index:    len(info.Rules) + 1,
			Chain:    zone,
			Target:   "allow",
			Protocol: "service",
			Port:     svc,
			Raw:      svc,
		})
	}
}

// AddFirewalldRule 添加 firewalld 规则
func (s *FirewallService) AddFirewalldRule(connID, zone, port, protocol string) error {
	if zone == "" {
		zone = "public"
	}
	if protocol == "" {
		protocol = "tcp"
	}

	// 判断是端口还是服务
	if strings.Contains(port, "/") || strings.Contains(port, ":") {
		// 端口范围
		cmd := fmt.Sprintf("firewall-cmd --zone=%s --add-port=%s/%s --permanent 2>/dev/null", zone, port, protocol)
		_, err := s.runCmd(connID, cmd)
		if err != nil {
			return fmt.Errorf("添加端口规则失败: %v", err)
		}
	} else {
		// 尝试作为服务添加
		cmd := fmt.Sprintf("firewall-cmd --zone=%s --add-service=%s --permanent 2>/dev/null", zone, port)
		_, err := s.runCmd(connID, cmd)
		if err != nil {
			// 作为端口添加
			cmd = fmt.Sprintf("firewall-cmd --zone=%s --add-port=%s/%s --permanent 2>/dev/null", zone, port, protocol)
			_, err = s.runCmd(connID, cmd)
			if err != nil {
				return fmt.Errorf("添加规则失败: %v", err)
			}
		}
	}

	// 重载
	s.runCmd(connID, "firewall-cmd --reload 2>/dev/null")
	return nil
}

// DeleteFirewalldRule 删除 firewalld 规则
func (s *FirewallService) DeleteFirewalldRule(connID, zone, port, protocol string) error {
	if zone == "" {
		zone = "public"
	}
	if protocol == "" {
		protocol = "tcp"
	}

	cmd := fmt.Sprintf("firewall-cmd --zone=%s --remove-port=%s/%s --permanent 2>/dev/null", zone, port, protocol)
	_, err := s.runCmd(connID, cmd)
	if err != nil {
		// 尝试作为服务删除
		cmd = fmt.Sprintf("firewall-cmd --zone=%s --remove-service=%s --permanent 2>/dev/null", zone, port)
		_, err = s.runCmd(connID, cmd)
		if err != nil {
			return fmt.Errorf("删除规则失败: %v", err)
		}
	}
	s.runCmd(connID, "firewall-cmd --reload 2>/dev/null")
	return nil
}

// ==================== ufw ====================

func (s *FirewallService) loadUfw(connID string, info *FirewallInfo) {
	out, _ := s.runCmd(connID, "ufw status numbered 2>/dev/null")
	info.RawOutput = out

	if strings.Contains(out, "Status: active") {
		info.Status = "active"
	} else {
		info.Status = "inactive"
	}

	lines := strings.Split(out, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Status:") || strings.HasPrefix(line, "To") || strings.HasPrefix(line, "--") {
			continue
		}

		// 格式: [ N] to                  action      FROM
		// 或: [ N] 22/tcp                ALLOW IN    Anywhere
		if !strings.HasPrefix(line, "[") {
			continue
		}

		rule := FirewallRule{Raw: line}

		// 解析序号
		if idxEnd := strings.Index(line, "]"); idxEnd > 1 {
			fmt.Sscanf(line[1:idxEnd], "%d", &rule.Index)
		}

		// 解析剩余部分
		rest := strings.TrimSpace(line[strings.Index(line, "]")+1:])
		fields := strings.Fields(rest)

		if len(fields) >= 3 {
			rule.Port = fields[0]
			rule.Target = fields[1]
			rule.Chain = fields[2] // IN / OUT
			if len(fields) >= 4 {
				rule.Source = fields[3]
			}
		}

		// 解析协议
		if parts := strings.Split(rule.Port, "/"); len(parts) == 2 {
			rule.Protocol = parts[1]
			rule.Port = parts[0]
		}

		info.Rules = append(info.Rules, rule)
	}

	info.Chains = []string{"IN", "OUT"}
}

// AddUfwRule 添加 ufw 规则
func (s *FirewallService) AddUfwRule(connID, action, port, protocol, source string) error {
	if action == "" {
		action = "allow"
	}
	action = strings.ToLower(action)

	cmd := fmt.Sprintf("ufw %s", action)
	if port != "" {
		cmd += " " + port
		if protocol != "" && protocol != "all" {
			cmd += "/" + protocol
		}
	}
	if source != "" && source != "anywhere" && source != "0.0.0.0/0" {
		cmd += " from " + source
	}

	_, err := s.runCmd(connID, cmd)
	if err != nil {
		return fmt.Errorf("添加规则失败: %v", err)
	}
	return nil
}

// DeleteUfwRule 删除 ufw 规则
func (s *FirewallService) DeleteUfwRule(connID string, index int) error {
	// ufw delete 需要确认，使用 --force
	cmd := fmt.Sprintf("echo y | ufw delete %d 2>/dev/null", index)
	_, err := s.runCmd(connID, cmd)
	if err != nil {
		return fmt.Errorf("删除规则失败: %v", err)
	}
	return nil
}

// ToggleUfw 启用/禁用 ufw
func (s *FirewallService) ToggleUfw(connID string, enable bool) error {
	cmd := "ufw --force disable"
	if enable {
		cmd = "ufw --force enable"
	}
	_, err := s.runCmd(connID, cmd)
	return err
}

// ==================== 通用接口 ====================

// AddRule 添加规则（根据防火墙类型自动适配）
func (s *FirewallService) AddRule(connID, chain, target, protocol, port, source, comment string) error {
	fwType := s.detectType(connID)
	switch fwType {
	case "iptables":
		return s.AddIptablesRule(connID, chain, target, protocol, port, source, comment)
	case "firewalld":
		return s.AddFirewalldRule(connID, chain, port, protocol)
	case "ufw":
		action := "allow"
		if target == "DROP" || target == "REJECT" {
			action = "deny"
		}
		return s.AddUfwRule(connID, action, port, protocol, source)
	default:
		return fmt.Errorf("不支持的防火墙类型")
	}
}

// DeleteRule 删除规则（根据防火墙类型自动适配）
func (s *FirewallService) DeleteRule(connID, chain string, index int, port, protocol string) error {
	fwType := s.detectType(connID)
	switch fwType {
	case "iptables":
		return s.DeleteIptablesRule(connID, chain, index)
	case "firewalld":
		return s.DeleteFirewalldRule(connID, chain, port, protocol)
	case "ufw":
		return s.DeleteUfwRule(connID, index)
	default:
		return fmt.Errorf("不支持的防火墙类型")
	}
}

// ToggleFirewall 启用/禁用防火墙
func (s *FirewallService) ToggleFirewall(connID string, enable bool) error {
	fwType := s.detectType(connID)
	switch fwType {
	case "ufw":
		return s.ToggleUfw(connID, enable)
	case "firewalld":
		cmd := "systemctl stop firewalld 2>/dev/null"
		if enable {
			cmd = "systemctl start firewalld 2>/dev/null"
		}
		_, err := s.runCmd(connID, cmd)
		return err
	case "iptables":
		if enable {
			// 恢复已保存的规则
			_, err := s.runCmd(connID, "iptables-restore < /etc/iptables/rules.v4 2>/dev/null")
			if err != nil {
				// 没有保存的规则，设置默认放行策略
				s.runCmd(connID, "iptables -P INPUT ACCEPT")
				s.runCmd(connID, "iptables -P OUTPUT ACCEPT")
				s.runCmd(connID, "iptables -P FORWARD ACCEPT")
			}
			return nil
		}
		// 先保存当前规则，再清空
		s.runCmd(connID, "mkdir -p /etc/iptables && iptables-save > /etc/iptables/rules.v4 2>/dev/null")
		// 清空所有规则，设置默认放行
		_, err := s.runCmd(connID, "iptables -F && iptables -X && iptables -P INPUT ACCEPT && iptables -P OUTPUT ACCEPT && iptables -P FORWARD ACCEPT")
		return err
	default:
		return fmt.Errorf("不支持的防火墙类型")
	}
}

// RunCustomCommand 执行自定义防火墙命令
func (s *FirewallService) RunCustomCommand(connID, command string) (string, error) {
	// 安全检查：只允许防火墙相关命令
	allowed := false
	prefixes := []string{"iptables", "ufw", "firewall-cmd", "nft", "ip6tables"}
	for _, p := range prefixes {
		if strings.HasPrefix(strings.TrimSpace(command), p) {
			allowed = true
			break
		}
	}
	if !allowed {
		return "", fmt.Errorf("只允许执行防火墙相关命令 (iptables/ufw/firewall-cmd/nft)")
	}

	return s.runCmd(connID, command)
}
