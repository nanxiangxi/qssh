package ai

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"changeme/ssh"
)

// AIConfig AI 配置（OpenAI 兼容格式）
type AIConfig struct {
	APIEndpoint  string  `json:"api_endpoint"`
	APIKey       string  `json:"api_key"`
	Model        string  `json:"model"`
	Timeout      int     `json:"timeout"`
	Temperature  float64 `json:"temperature"`
	MaxTokens    int     `json:"max_tokens"`
	TopP         float64 `json:"top_p,omitempty"`
	SystemPrompt string  `json:"system_prompt"`
}

// ConfigManager 配置管理器
type ConfigManager struct {
	mu       sync.RWMutex
	config   *AIConfig
	dataPath string
}

// NewConfigManager 创建配置管理器
func NewConfigManager() *ConfigManager {
	configPath := filepath.Join(ssh.GetDataDir(), "ai_config.json")

	cm := &ConfigManager{
		config:   getDefaultConfig(),
		dataPath: configPath,
	}

	cm.loadConfig()
	return cm
}

func getDefaultConfig() *AIConfig {
	return &AIConfig{
		Timeout:     120,
		Temperature: 1.0,
		MaxTokens:   4096,
		TopP:        0.95,
		SystemPrompt: `你是"启SSH AI"，一个专业的 SSH 服务器管理和系统运维 AI 助手。

【核心规则】
1. 始终使用中文回复。
2. 使用 Markdown 格式，代码块标注语言类型。

【工具调用规则 - 非常严格】
⚠️ 以下情况绝对不能调用工具：
- 任何形式的问候：你好、hello、hi、嗨、在吗 等
- 概念性问题：什么是 SSH、如何学习 Linux、Docker 是什么 等
- 闲聊：今天天气怎么样、你叫什么名字 等
- 请求解释：解释某个命令的含义、某个概念 等
- 已经有足够信息可以回答的问题

✅ 只有以下情况才调用工具：
- 用户明确说"查看"、"检查"、"执行"、"运行"某个具体命令
- 用户明确要求获取服务器当前状态
- 用户要求分析当前服务器的安全状况

⚠️ 重要限制：
- 每次回复尽量只调用 1-2 个工具，获取到数据后立即分析回复
- 不要反复调用相同或类似的工具，已获取的数据直接使用
- 工具调用会弹出确认框让用户审批，滥用严重影响体验
- 如果不确定是否需要工具，就不要调用
- 获取到工具结果后，优先基于已有数据直接回复，不要继续调用更多工具

【可用工具】
- execute_command: 执行 shell 命令。参数: command（完整命令字符串）
- get_server_info: 获取服务器基本信息
- get_system_status: 获取服务器实时状态
- open_terminal: 打开一个新的终端面板
- close_terminal: 关闭 AI 自己创建的终端

【终端管理规则】
- execute_command 会自动复用已有的 AI 终端，不需要每次开新终端
- 只有需要同时监控多个独立任务时，才用 open_terminal 开新终端
- 任务完成后，必须调用 close_terminal 释放不需要的终端
- 绝对不要关闭用户手动创建的终端`,
	}
}

func (cm *ConfigManager) loadConfig() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	data, err := os.ReadFile(cm.dataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config AIConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	cm.config = &config
	return nil
}

func (cm *ConfigManager) SaveConfig(config *AIConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if err := validateConfig(config); err != nil {
		return err
	}

	dir := filepath.Dir(cm.dataPath)
	os.MkdirAll(dir, 0755)

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	if err := os.WriteFile(cm.dataPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	cm.config = config
	return nil
}

func (cm *ConfigManager) GetConfig() *AIConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cp := *cm.config
	return &cp
}

func (cm *ConfigManager) IsConfigured() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config.APIEndpoint != "" && cm.config.APIKey != ""
}

func validateConfig(config *AIConfig) error {
	if config.APIEndpoint == "" {
		return fmt.Errorf("API端点不能为空")
	}
	if config.APIKey == "" {
		return fmt.Errorf("API密钥不能为空")
	}
	if config.Model == "" {
		return fmt.Errorf("模型名称不能为空")
	}
	if config.Timeout <= 0 {
		config.Timeout = 30
	}
	return nil
}
