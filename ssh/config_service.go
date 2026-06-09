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
}

// TerminalConfig 终端配置
type TerminalConfig struct {
	DefaultType        string `json:"defaultType"`        // structured, classic
	AutoSwitchClassic  bool   `json:"autoSwitchClassic"`  // 交互式操作自动切经典
	SwitchMode         string `json:"switchMode"`         // prompt | auto | inline
	FontSize           int    `json:"fontSize"`           // 终端字体大小
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

// UIConfig 界面配置（预留）
type UIConfig struct {
	Language         string `json:"language"`
	Theme            string `json:"theme"`
	SidebarCollapsed bool   `json:"sidebarCollapsed"`
	ShowStatusBar    bool   `json:"showStatusBar"`
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

	// 加载配置
	if err := svc.load(); err != nil {
		fmt.Printf("[ConfigService] 加载配置失败，使用默认配置: %v\n", err)
	}

	// 确保配置文件存在（如果不存在则创建）
	if err := svc.ensureConfigFile(); err != nil {
		fmt.Printf("[ConfigService] 创建配置文件失败: %v\n", err)
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

// load 加载配置
func (s *ConfigService) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查文件是否存在
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		fmt.Printf("[ConfigService] 配置文件不存在，将使用默认配置: %s\n", s.filePath)
		return nil
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, s.config)
}

// ensureConfigFile 确保配置文件存在
func (s *ConfigService) ensureConfigFile() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 检查文件是否存在
	if _, err := os.Stat(s.filePath); err == nil {
		return nil // 文件已存在
	}

	// 确保目录存在
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 保存默认配置
	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	fmt.Printf("[ConfigService] ✓ 配置文件已创建: %s\n", s.filePath)
	return nil
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

// GetUIConfig 获取界面配置（预留，供 Wails 绑定）
func (s *ConfigService) GetUIConfig() UIConfig {
	return UIConfig{
		Language:         "zh-CN",
		Theme:            "dark",
		SidebarCollapsed: false,
		ShowStatusBar:    true,
	}
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

