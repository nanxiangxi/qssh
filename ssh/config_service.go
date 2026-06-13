package ssh

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// AppConfig 应用配置（仅保留实际使用的字段）
type AppConfig struct {
	Terminal TerminalConfig `json:"terminal"`
	UI       UIConfig       `json:"ui"`
	Cloud    CloudConfig    `json:"cloud"`
}

// CloudConfig 云端配置
type CloudConfig struct {
	Enabled      bool   `json:"enabled"`
	ServerURL    string `json:"serverUrl"`
	Token        string `json:"token"`
	SyncInterval int    `json:"syncInterval"` // 同步间隔（秒）
	AutoSyncTo   bool   `json:"autoSyncTo"`   // 自动同步到云端
	AutoSyncFrom bool   `json:"autoSyncFrom"` // 自动从云端同步
}

// TerminalConfig 终端配置
type TerminalConfig struct {
	DefaultType        string `json:"defaultType"`        // structured, classic
	AutoSwitchClassic  bool   `json:"autoSwitchClassic"`  // 交互式操作自动切经典
	SwitchMode         string `json:"switchMode"`         // prompt | auto | inline
	FontSize           int    `json:"fontSize"`           // 终端字体大小
	CommandSendMode    string `json:"commandSendMode"`    // enter | button
	CodeHighlight      bool   `json:"codeHighlight"`      // 经典终端代码高亮
}

// FileManagerConfig 文件管理配置（预留，供 Wails 绑定使用）
type FileManagerConfig struct {
	ShowHidden    bool   `json:"showHidden"`
	SortBy        string `json:"sortBy"`
	SortOrder     string `json:"sortOrder"`
	ConfirmDelete bool   `json:"confirmDelete"`
}

// AIConfig AI 配置（预留）
type AIConfig struct {
	Enabled          bool `json:"enabled"`
	AutoExecute      bool `json:"autoExecute"`
	ConfirmExecution bool `json:"confirmExecution"`
}

// UIConfig 界面配置
type UIConfig struct {
	AutoTray         bool   `json:"autoTray"`         // SSH 连接成功后自动最小化到托盘
	RememberPosition bool   `json:"rememberPosition"` // 记忆窗口位置
	AutoShowHome     bool   `json:"autoShowHome"`     // SSH 窗口全部关闭后自动显示首页
	Theme            string `json:"theme"`            // 主题 dark/light
}

// SSHSettings SSH 设置（预留）
type SSHSettings struct {
	DefaultPort          int  `json:"defaultPort"`
	ConnectTimeout       int  `json:"connectTimeout"`
	KeepAlive            bool `json:"keepAlive"`
	KeepAliveInterval    int  `json:"keepAliveInterval"`
	AutoReconnect        bool `json:"autoReconnect"`
	MaxReconnectAttempts int  `json:"maxReconnectAttempts"`
}

// ConfigService 配置服务
type ConfigService struct {
	mu       sync.RWMutex
	config   *AppConfig
	filePath string
	app      *application.App
}

// NewConfigService 创建配置服务
func NewConfigService() *ConfigService {
	dataDir := getDataDir()
	configPath := filepath.Join(dataDir, "config.json")

	fmt.Printf("[ConfigService] 配置文件路径: %s\n", configPath)

	svc := &ConfigService{
		filePath: configPath,
		config:   getDefaultConfig(),
	}

	// 加载配置并补全缺失字段
	if err := svc.loadAndMerge(); err != nil {
		fmt.Printf("[ConfigService] 加载配置失败，使用默认配置: %v\n", err)
	}

	return svc
}

// SetApp 设置应用实例
func (s *ConfigService) SetApp(app *application.App) {
	s.app = app
}

// 获取默认配置
func getDefaultConfig() *AppConfig {
	return &AppConfig{
		Terminal: TerminalConfig{
			DefaultType:       "classic",
			AutoSwitchClassic: true,
			SwitchMode:        "prompt",
			FontSize:          14,
			CommandSendMode:   "enter",
			CodeHighlight:     false,
		},
		UI: UIConfig{
			AutoTray:         false,
			RememberPosition: true,
			AutoShowHome:     true,
		},
		Cloud: CloudConfig{
			Enabled:      false,
			ServerURL:    "",
			Token:        "",
			SyncInterval: 60,
		},
	}
}

// 获取数据目录（与 storage.go 保持一致）
func getDataDir() string {
	// 获取可执行文件目录
	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("[ConfigService] 获取可执行文件路径失败: %v，使用当前目录\n", err)
		// 降级到当前工作目录
		cwd, _ := os.Getwd()
		configDir := filepath.Join(cwd, "data", "config")
		os.MkdirAll(configDir, 0755)
		return configDir
	}
	exeDir := filepath.Dir(exePath)
	fmt.Printf("[ConfigService] 可执行文件目录: %s\n", exeDir)

	// 配置文件目录：exeDir/data/config
	configDir := filepath.Join(exeDir, "data", "config")
	fmt.Printf("[ConfigService] 尝试创建配置目录: %s\n", configDir)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Printf("[ConfigService] 创建配置目录失败: %v\n", err)
		// 降级到 data 目录
		configDir = filepath.Join(exeDir, "data")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			// 降级到程序目录
			configDir = exeDir
		}
	}

	fmt.Printf("[ConfigService] ✓ 配置目录: %s\n", configDir)
	return configDir
}

// loadAndMerge 加载配置文件并补全缺失字段
// 文件不存在 → 使用默认配置并保存
// 文件存在 → 读取 JSON，与默认配置合并（缺失字段自动补充），合并后保存
func (s *ConfigService) loadAndMerge() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 确保目录存在
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	defaults := getDefaultConfig()

	// 检查文件是否存在
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		fmt.Printf("[ConfigService] 配置文件不存在，创建默认配置: %s\n", s.filePath)
		s.config = defaults
		return s.save()
	}

	// 文件存在，读取内容
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	// 解析为 map，保留已有字段的值
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		fmt.Printf("[ConfigService] 配置文件解析失败，使用默认配置: %v\n", err)
		s.config = defaults
		return s.save()
	}

	// 以默认配置为基础，用文件中的值覆盖
	merged := mergeConfig(raw, defaults)
	s.config = merged

	// 保存合并后的配置（确保文件包含所有字段）
	if err := s.save(); err != nil {
		fmt.Printf("[ConfigService] 保存合并配置失败: %v\n", err)
	} else {
		fmt.Printf("[ConfigService] ✓ 配置已加载并补全缺失字段\n")
	}

	return nil
}

// mergeConfig 以默认配置为基础，用 raw 中的值覆盖
func mergeConfig(raw map[string]interface{}, defaults *AppConfig) *AppConfig {
	result := *defaults // 复制默认值

	// 合并 terminal 配置
	if termRaw, ok := raw["terminal"].(map[string]interface{}); ok {
		if v, ok := termRaw["defaultType"].(string); ok && v != "" {
			result.Terminal.DefaultType = v
		}
		if v, ok := termRaw["autoSwitchClassic"].(bool); ok {
			result.Terminal.AutoSwitchClassic = v
		}
		if v, ok := termRaw["switchMode"].(string); ok && v != "" {
			result.Terminal.SwitchMode = v
		}
		if v, ok := termRaw["fontSize"].(float64); ok && v > 0 {
			result.Terminal.FontSize = int(v)
		}
		if v, ok := termRaw["commandSendMode"].(string); ok && v != "" {
			result.Terminal.CommandSendMode = v
		}
		if v, ok := termRaw["codeHighlight"].(bool); ok {
			result.Terminal.CodeHighlight = v
		}
	}

	// 合并 ui 配置
	if uiRaw, ok := raw["ui"].(map[string]interface{}); ok {
		if v, ok := uiRaw["autoTray"].(bool); ok {
			result.UI.AutoTray = v
		}
		if v, ok := uiRaw["rememberPosition"].(bool); ok {
			result.UI.RememberPosition = v
		}
		if v, ok := uiRaw["autoShowHome"].(bool); ok {
			result.UI.AutoShowHome = v
		}
		if v, ok := uiRaw["theme"].(string); ok && v != "" {
			result.UI.Theme = v
		}
	}

	// 合并 cloud 配置
	if cloudRaw, ok := raw["cloud"].(map[string]interface{}); ok {
		if v, ok := cloudRaw["enabled"].(bool); ok {
			result.Cloud.Enabled = v
		}
		if v, ok := cloudRaw["serverUrl"].(string); ok {
			result.Cloud.ServerURL = v
		}
		if v, ok := cloudRaw["token"].(string); ok {
			result.Cloud.Token = v
		}
		if v, ok := cloudRaw["syncInterval"].(float64); ok && v > 0 {
			result.Cloud.SyncInterval = int(v)
		}
		if v, ok := cloudRaw["autoSyncTo"].(bool); ok {
			result.Cloud.AutoSyncTo = v
		}
		if v, ok := cloudRaw["autoSyncFrom"].(bool); ok {
			result.Cloud.AutoSyncFrom = v
		}
	}

	return &result
}

// save 保存配置（调用者需持有锁）
func (s *ConfigService) save() error {
	// 确保目录存在
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("[ConfigService] 创建目录失败: %v\n", err)
		return err
	}

	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return err
	}

	fmt.Printf("[ConfigService] 保存配置到: %s\n", s.filePath)
	return os.WriteFile(s.filePath, data, 0644)
}

// GetConfig 获取完整配置
func (s *ConfigService) GetConfig() *AppConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// SetConfig 设置完整配置
func (s *ConfigService) SetConfig(config *AppConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config = config
	return s.save()
}

// GetTerminalConfig 获取终端配置
func (s *ConfigService) GetTerminalConfig() TerminalConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.Terminal
}

// SetTerminalConfig 设置终端配置
func (s *ConfigService) SetTerminalConfig(config TerminalConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config.Terminal = config
	return s.save()
}

// GetFileManagerConfig 获取文件管理配置（预留，供 Wails 绑定）
func (s *ConfigService) GetFileManagerConfig() FileManagerConfig {
	return FileManagerConfig{
		ShowHidden:    false,
		SortBy:        "name",
		SortOrder:     "asc",
		ConfirmDelete: true,
	}
}

// SetFileManagerConfig 设置文件管理配置（预留）
func (s *ConfigService) SetFileManagerConfig(config FileManagerConfig) error {
	return nil
}

// GetAIConfig 获取 AI 配置（预留，供 Wails 绑定）
func (s *ConfigService) GetAIConfig() AIConfig {
	return AIConfig{
		Enabled:          true,
		AutoExecute:      false,
		ConfirmExecution: true,
	}
}

// SetAIConfig 设置 AI 配置（预留）
func (s *ConfigService) SetAIConfig(config AIConfig) error {
	return nil
}

// GetUIConfig 获取界面配置
func (s *ConfigService) GetUIConfig() UIConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.UI
}

// SetUIConfig 设置界面配置（预留）
func (s *ConfigService) SetUIConfig(config UIConfig) error {
	return nil
}

// GetSSHSettings 获取 SSH 设置（预留，供 Wails 绑定）
func (s *ConfigService) GetSSHSettings() SSHSettings {
	return SSHSettings{
		DefaultPort:          22,
		ConnectTimeout:       30,
		KeepAlive:            true,
		KeepAliveInterval:    30,
		AutoReconnect:        true,
		MaxReconnectAttempts: 3,
	}
}

// SetSSHSettings 设置 SSH 设置（预留）
func (s *ConfigService) SetSSHSettings(settings SSHSettings) error {
	return nil
}

// Get 获取单个配置项
func (s *ConfigService) Get(category, key string) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if category == "terminal" {
		if key == "defaultType" { return s.config.Terminal.DefaultType, nil }
		if key == "autoSwitchClassic" { return s.config.Terminal.AutoSwitchClassic, nil }
		if key == "switchMode" { return s.config.Terminal.SwitchMode, nil }
		if key == "fontSize" { return s.config.Terminal.FontSize, nil }
		if key == "commandSendMode" { return s.config.Terminal.CommandSendMode, nil }
		if key == "codeHighlight" { return s.config.Terminal.CodeHighlight, nil }
	}

	if category == "ui" {
		if key == "autoTray" { return s.config.UI.AutoTray, nil }
		if key == "rememberPosition" { return s.config.UI.RememberPosition, nil }
		if key == "autoShowHome" { return s.config.UI.AutoShowHome, nil }
		if key == "theme" { return s.config.UI.Theme, nil }
	}

	if category == "cloud" {
		if key == "enabled" { return s.config.Cloud.Enabled, nil }
		if key == "serverUrl" { return s.config.Cloud.ServerURL, nil }
		if key == "token" { return s.config.Cloud.Token, nil }
		if key == "syncInterval" { return s.config.Cloud.SyncInterval, nil }
		if key == "autoSyncTo" { return s.config.Cloud.AutoSyncTo, nil }
		if key == "autoSyncFrom" { return s.config.Cloud.AutoSyncFrom, nil }
	}

	return nil, fmt.Errorf("未知的配置项: %s.%s", category, key)
}

// Set 设置单个配置项
func (s *ConfigService) Set(category, key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if category == "terminal" {
		if key == "defaultType" {
			if v, ok := value.(string); ok { s.config.Terminal.DefaultType = v; return s.save() }
			return fmt.Errorf("无效的值类型")
		}
		if key == "autoSwitchClassic" {
			if v, ok := value.(bool); ok { s.config.Terminal.AutoSwitchClassic = v; return s.save() }
			return fmt.Errorf("无效的值类型")
		}
		if key == "switchMode" {
			if v, ok := value.(string); ok { s.config.Terminal.SwitchMode = v; return s.save() }
			return fmt.Errorf("无效的值类型")
		}
		if key == "fontSize" {
			if v, ok := value.(float64); ok { s.config.Terminal.FontSize = int(v); return s.save() }
			return fmt.Errorf("无效的值类型")
		}
		if key == "commandSendMode" {
			if v, ok := value.(string); ok { s.config.Terminal.CommandSendMode = v; return s.save() }
			return fmt.Errorf("无效的值类型")
		}
		if key == "codeHighlight" {
			if v, ok := value.(bool); ok { s.config.Terminal.CodeHighlight = v; return s.save() }
			return fmt.Errorf("无效的值类型")
		}
	}

	if category == "ui" {
		if key == "autoTray" {
			if v, ok := value.(bool); ok { s.config.UI.AutoTray = v; return s.save() }
			return fmt.Errorf("无效的值类型")
		}
		if key == "rememberPosition" {
			if v, ok := value.(bool); ok { s.config.UI.RememberPosition = v; return s.save() }
			return fmt.Errorf("无效的值类型")
		}
		if key == "autoShowHome" {
			if v, ok := value.(bool); ok { s.config.UI.AutoShowHome = v; return s.save() }
			return fmt.Errorf("无效的值类型")
		}
		if key == "theme" {
			if v, ok := value.(string); ok { s.config.UI.Theme = v; return s.save() }
			return fmt.Errorf("无效的值类型")
		}
	}

	if category == "cloud" {
		if key == "enabled" {
			if v, ok := value.(bool); ok { s.config.Cloud.Enabled = v; return s.save() }
			return fmt.Errorf("无效的值类型")
		}
		if key == "serverUrl" {
			if v, ok := value.(string); ok { s.config.Cloud.ServerURL = v; return s.save() }
			return fmt.Errorf("无效的值类型")
		}
		if key == "token" {
			if v, ok := value.(string); ok { s.config.Cloud.Token = v; return s.save() }
			return fmt.Errorf("无效的值类型")
		}
		if key == "syncInterval" {
			if v, ok := value.(float64); ok { s.config.Cloud.SyncInterval = int(v); return s.save() }
			return fmt.Errorf("无效的值类型")
		}
		if key == "autoSyncTo" {
			if v, ok := value.(bool); ok { s.config.Cloud.AutoSyncTo = v; return s.save() }
			return fmt.Errorf("无效的值类型")
		}
		if key == "autoSyncFrom" {
			if v, ok := value.(bool); ok { s.config.Cloud.AutoSyncFrom = v; return s.save() }
			return fmt.Errorf("无效的值类型")
		}
	}

	return fmt.Errorf("未知的配置项: %s.%s", category, key)
}

// ResetCategory 重置分类配置
func (s *ConfigService) ResetCategory(category string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	defaults := getDefaultConfig()

	if category == "terminal" {
		s.config.Terminal = defaults.Terminal
		return s.save()
	}
	if category == "ui" {
		s.config.UI = defaults.UI
		return s.save()
	}

	return fmt.Errorf("未知的配置分类: %s", category)
}

// ResetAll 重置所有配置
func (s *ConfigService) ResetAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config = getDefaultConfig()
	return s.save()
}

// ExportConfig 导出配置
func (s *ConfigService) ExportConfig() (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// ImportConfig 导入配置
func (s *ConfigService) ImportConfig(jsonStr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var config AppConfig
	if err := json.Unmarshal([]byte(jsonStr), &config); err != nil {
		return err
	}

	s.config = &config
	return s.save()
}

