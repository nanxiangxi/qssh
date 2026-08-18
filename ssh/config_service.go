package ssh

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// AppConfig 应用配置
type AppConfig struct {
	Locale    string          `json:"locale"`
	Terminal  TerminalConfig  `json:"terminal"`
	UI        UIConfig        `json:"ui"`
	Cloud     CloudConfig     `json:"cloud"`
	Shortcuts ShortcutsConfig `json:"shortcuts"`
	Advanced  AdvancedConfig  `json:"advanced"`
}

type ShortcutsConfig struct {
	Enabled       bool `json:"enabled"`
	SwitchTab     bool `json:"switchTab"`
	SaveGroup     bool `json:"saveGroup"`
	CloudUpload   bool `json:"cloudUpload"`
	CloudDownload bool `json:"cloudDownload"`
}

type AdvancedConfig struct {
	GroupBehavior string `json:"groupBehavior"`
}

type CloudConfig struct {
	Enabled      bool   `json:"enabled"`
	ServerURL    string `json:"serverUrl"`
	Token        string `json:"token"`
	SyncInterval int    `json:"syncInterval"`
	AutoSyncTo   bool   `json:"autoSyncTo"`
	AutoSyncFrom bool   `json:"autoSyncFrom"`
}

type TerminalConfig struct {
	DefaultType       string `json:"defaultType"`
	AutoSwitchClassic bool   `json:"autoSwitchClassic"`
	SwitchMode        string `json:"switchMode"`
	FontSize          int    `json:"fontSize"`
	CommandSendMode   string `json:"commandSendMode"`
	CodeHighlight     bool   `json:"codeHighlight"`
}

type UIConfig struct {
	AutoTray         bool   `json:"autoTray"`
	RememberPosition bool   `json:"rememberPosition"`
	AutoShowHome     bool   `json:"autoShowHome"`
	Theme            string `json:"theme"`
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
	configDir := filepath.Join(GetDataDir(), "config")
	os.MkdirAll(configDir, 0755)
	configPath := filepath.Join(configDir, "config.json")

	svc := &ConfigService{
		filePath: configPath,
		config:   getDefaultConfig(),
	}

	if err := svc.loadAndMerge(); err != nil {
		log.Error("加载配置失败，使用默认配置", "error", err)
	}

	return svc
}

func (s *ConfigService) SetApp(app *application.App) { s.app = app }

func getDefaultConfig() *AppConfig {
	return &AppConfig{
		Terminal: TerminalConfig{
			DefaultType:       "classic",
			AutoSwitchClassic: true,
			SwitchMode:        "prompt",
			FontSize:          14,
			CommandSendMode:   "enter",
		},
		UI: UIConfig{
			RememberPosition: true,
			AutoShowHome:     true,
		},
		Cloud: CloudConfig{
			SyncInterval: 60,
		},
		Shortcuts: ShortcutsConfig{
			Enabled:       true,
			SwitchTab:     true,
			SaveGroup:     true,
			CloudUpload:   true,
			CloudDownload: true,
		},
		Advanced: AdvancedConfig{
			GroupBehavior: "prompt",
		},
	}
}

func (s *ConfigService) loadAndMerge() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dir := filepath.Dir(s.filePath)
	os.MkdirAll(dir, 0755)

	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		s.config = getDefaultConfig()
		return s.save()
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		s.config = getDefaultConfig()
		return s.save()
	}

	s.config = mergeConfig(raw, getDefaultConfig())
	s.save()
	return nil
}

// mergeConfig 以默认配置为基础，用 raw 中的值覆盖
func mergeConfig(raw map[string]interface{}, defaults *AppConfig) *AppConfig {
	result := *defaults

	if term, ok := raw["terminal"].(map[string]interface{}); ok {
		setStr(&result.Terminal.DefaultType, term, "defaultType")
		setBool(&result.Terminal.AutoSwitchClassic, term, "autoSwitchClassic")
		setStr(&result.Terminal.SwitchMode, term, "switchMode")
		setInt(&result.Terminal.FontSize, term, "fontSize")
		setStr(&result.Terminal.CommandSendMode, term, "commandSendMode")
		setBool(&result.Terminal.CodeHighlight, term, "codeHighlight")
	}
	if ui, ok := raw["ui"].(map[string]interface{}); ok {
		setBool(&result.UI.AutoTray, ui, "autoTray")
		setBool(&result.UI.RememberPosition, ui, "rememberPosition")
		setBool(&result.UI.AutoShowHome, ui, "autoShowHome")
		setStr(&result.UI.Theme, ui, "theme")
	}
	if cloud, ok := raw["cloud"].(map[string]interface{}); ok {
		setBool(&result.Cloud.Enabled, cloud, "enabled")
		setStr(&result.Cloud.ServerURL, cloud, "serverUrl")
		setStr(&result.Cloud.Token, cloud, "token")
		setInt(&result.Cloud.SyncInterval, cloud, "syncInterval")
		setBool(&result.Cloud.AutoSyncTo, cloud, "autoSyncTo")
		setBool(&result.Cloud.AutoSyncFrom, cloud, "autoSyncFrom")
	}
	if sc, ok := raw["shortcuts"].(map[string]interface{}); ok {
		setBool(&result.Shortcuts.Enabled, sc, "enabled")
		setBool(&result.Shortcuts.SwitchTab, sc, "switchTab")
		setBool(&result.Shortcuts.SaveGroup, sc, "saveGroup")
		setBool(&result.Shortcuts.CloudUpload, sc, "cloudUpload")
		setBool(&result.Shortcuts.CloudDownload, sc, "cloudDownload")
	}
	if adv, ok := raw["advanced"].(map[string]interface{}); ok {
		setStr(&result.Advanced.GroupBehavior, adv, "groupBehavior")
	}

	return &result
}

func setStr(dst *string, m map[string]interface{}, key string) {
	if v, ok := m[key].(string); ok && v != "" {
		*dst = v
	}
}
func setBool(dst *bool, m map[string]interface{}, key string) {
	if v, ok := m[key].(bool); ok {
		*dst = v
	}
}
func setInt(dst *int, m map[string]interface{}, key string) {
	if v, ok := m[key].(float64); ok {
		*dst = int(v)
	}
}

func (s *ConfigService) save() error {
	dir := filepath.Dir(s.filePath)
	os.MkdirAll(dir, 0755)

	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0644)
}

func (s *ConfigService) GetConfig() *AppConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

func (s *ConfigService) SetConfig(config *AppConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = config
	return s.save()
}

func (s *ConfigService) GetTerminalConfig() TerminalConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.Terminal
}

func (s *ConfigService) SetTerminalConfig(config TerminalConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config.Terminal = config
	return s.save()
}

func (s *ConfigService) GetUIConfig() UIConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.UI
}

func (s *ConfigService) SetUIConfig(config UIConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config.UI = config
	return s.save()
}

// configAccessor 配置字段访问器
type configAccessor struct {
	get func(c *AppConfig) interface{}
	set func(c *AppConfig, v interface{}) error
}

// buildAccessors 构建 category.key → accessor 映射
func (s *ConfigService) buildAccessors() map[string]configAccessor {
	return map[string]configAccessor{
		"locale.value":               {get: func(c *AppConfig) interface{} { return c.Locale }, set: func(c *AppConfig, v interface{}) error { s, ok := v.(string); if ok { c.Locale = s }; return typeErr(ok) }},
		"terminal.defaultType":       {get: func(c *AppConfig) interface{} { return c.Terminal.DefaultType }, set: func(c *AppConfig, v interface{}) error { s, ok := v.(string); if ok { c.Terminal.DefaultType = s }; return typeErr(ok) }},
		"terminal.autoSwitchClassic": {get: func(c *AppConfig) interface{} { return c.Terminal.AutoSwitchClassic }, set: func(c *AppConfig, v interface{}) error { s, ok := v.(bool); if ok { c.Terminal.AutoSwitchClassic = s }; return typeErr(ok) }},
		"terminal.switchMode":        {get: func(c *AppConfig) interface{} { return c.Terminal.SwitchMode }, set: func(c *AppConfig, v interface{}) error { s, ok := v.(string); if ok { c.Terminal.SwitchMode = s }; return typeErr(ok) }},
		"terminal.fontSize":          {get: func(c *AppConfig) interface{} { return c.Terminal.FontSize }, set: func(c *AppConfig, v interface{}) error { s, ok := v.(float64); if ok { c.Terminal.FontSize = int(s) }; return typeErr(ok) }},
		"terminal.commandSendMode":   {get: func(c *AppConfig) interface{} { return c.Terminal.CommandSendMode }, set: func(c *AppConfig, v interface{}) error { s, ok := v.(string); if ok { c.Terminal.CommandSendMode = s }; return typeErr(ok) }},
		"terminal.codeHighlight":     {get: func(c *AppConfig) interface{} { return c.Terminal.CodeHighlight }, set: func(c *AppConfig, v interface{}) error { s, ok := v.(bool); if ok { c.Terminal.CodeHighlight = s }; return typeErr(ok) }},
		"ui.autoTray":                {get: func(c *AppConfig) interface{} { return c.UI.AutoTray }, set: func(c *AppConfig, v interface{}) error { s, ok := v.(bool); if ok { c.UI.AutoTray = s }; return typeErr(ok) }},
		"ui.rememberPosition":        {get: func(c *AppConfig) interface{} { return c.UI.RememberPosition }, set: func(c *AppConfig, v interface{}) error { s, ok := v.(bool); if ok { c.UI.RememberPosition = s }; return typeErr(ok) }},
		"ui.autoShowHome":            {get: func(c *AppConfig) interface{} { return c.UI.AutoShowHome }, set: func(c *AppConfig, v interface{}) error { s, ok := v.(bool); if ok { c.UI.AutoShowHome = s }; return typeErr(ok) }},
		"ui.theme":                   {get: func(c *AppConfig) interface{} { return c.UI.Theme }, set: func(c *AppConfig, v interface{}) error { s, ok := v.(string); if ok { c.UI.Theme = s }; return typeErr(ok) }},
		"cloud.enabled":              {get: func(c *AppConfig) interface{} { return c.Cloud.Enabled }, set: func(c *AppConfig, v interface{}) error { s, ok := v.(bool); if ok { c.Cloud.Enabled = s }; return typeErr(ok) }},
		"cloud.serverUrl":            {get: func(c *AppConfig) interface{} { return c.Cloud.ServerURL }, set: func(c *AppConfig, v interface{}) error { s, ok := v.(string); if ok { c.Cloud.ServerURL = s }; return typeErr(ok) }},
		"cloud.token":                {get: func(c *AppConfig) interface{} { return c.Cloud.Token }, set: func(c *AppConfig, v interface{}) error { s, ok := v.(string); if ok { c.Cloud.Token = s }; return typeErr(ok) }},
		"cloud.syncInterval":         {get: func(c *AppConfig) interface{} { return c.Cloud.SyncInterval }, set: func(c *AppConfig, v interface{}) error { s, ok := v.(float64); if ok { c.Cloud.SyncInterval = int(s) }; return typeErr(ok) }},
		"cloud.autoSyncTo":           {get: func(c *AppConfig) interface{} { return c.Cloud.AutoSyncTo }, set: func(c *AppConfig, v interface{}) error { s, ok := v.(bool); if ok { c.Cloud.AutoSyncTo = s }; return typeErr(ok) }},
		"cloud.autoSyncFrom":         {get: func(c *AppConfig) interface{} { return c.Cloud.AutoSyncFrom }, set: func(c *AppConfig, v interface{}) error { s, ok := v.(bool); if ok { c.Cloud.AutoSyncFrom = s }; return typeErr(ok) }},
		"shortcuts.enabled":          {get: func(c *AppConfig) interface{} { return c.Shortcuts.Enabled }, set: func(c *AppConfig, v interface{}) error { s, ok := v.(bool); if ok { c.Shortcuts.Enabled = s }; return typeErr(ok) }},
		"shortcuts.switchTab":        {get: func(c *AppConfig) interface{} { return c.Shortcuts.SwitchTab }, set: func(c *AppConfig, v interface{}) error { s, ok := v.(bool); if ok { c.Shortcuts.SwitchTab = s }; return typeErr(ok) }},
		"shortcuts.saveGroup":        {get: func(c *AppConfig) interface{} { return c.Shortcuts.SaveGroup }, set: func(c *AppConfig, v interface{}) error { s, ok := v.(bool); if ok { c.Shortcuts.SaveGroup = s }; return typeErr(ok) }},
		"shortcuts.cloudUpload":      {get: func(c *AppConfig) interface{} { return c.Shortcuts.CloudUpload }, set: func(c *AppConfig, v interface{}) error { s, ok := v.(bool); if ok { c.Shortcuts.CloudUpload = s }; return typeErr(ok) }},
		"shortcuts.cloudDownload":    {get: func(c *AppConfig) interface{} { return c.Shortcuts.CloudDownload }, set: func(c *AppConfig, v interface{}) error { s, ok := v.(bool); if ok { c.Shortcuts.CloudDownload = s }; return typeErr(ok) }},
		"advanced.groupBehavior":     {get: func(c *AppConfig) interface{} { return c.Advanced.GroupBehavior }, set: func(c *AppConfig, v interface{}) error { s, ok := v.(string); if ok { c.Advanced.GroupBehavior = s }; return typeErr(ok) }},
	}
}

func typeErr(ok bool) error {
	if !ok {
		return fmt.Errorf("无效的值类型")
	}
	return nil
}

// Get 获取单个配置项
func (s *ConfigService) Get(category, key string) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	accessors := s.buildAccessors()
	fullKey := category + "." + key
	acc, ok := accessors[fullKey]
	if !ok {
		return nil, fmt.Errorf("未知的配置项: %s", fullKey)
	}
	return acc.get(s.config), nil
}

// Set 设置单个配置项
func (s *ConfigService) Set(category, key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	accessors := s.buildAccessors()
	fullKey := category + "." + key
	acc, ok := accessors[fullKey]
	if !ok {
		return fmt.Errorf("未知的配置项: %s", fullKey)
	}
	if err := acc.set(s.config, value); err != nil {
		return err
	}
	return s.save()
}

// ResetCategory 重置分类配置
func (s *ConfigService) ResetCategory(category string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	defaults := getDefaultConfig()

	v := reflect.ValueOf(s.config).Elem()
	dv := reflect.ValueOf(defaults).Elem()

	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == category || (len(jsonTag) > 0 && jsonTag[:len(jsonTag)-len(",omitempty")] == category) {
			v.Field(i).Set(dv.Field(i))
			return s.save()
		}
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

// ImportConfig 导入配置（合并默认值，确保新字段有默认值）
func (s *ConfigService) ImportConfig(jsonStr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return err
	}
	s.config = mergeConfig(raw, getDefaultConfig())
	return s.save()
}

// 以下方法保持向后兼容（Wails 绑定）

func (s *ConfigService) GetFileManagerConfig() FileManagerConfig {
	return FileManagerConfig{ShowHidden: false, SortBy: "name", SortOrder: "asc", ConfirmDelete: true}
}
func (s *ConfigService) SetFileManagerConfig(config FileManagerConfig) error { return nil }
func (s *ConfigService) GetAIConfig() AIConfig {
	return AIConfig{Enabled: true, AutoExecute: false, ConfirmExecution: true}
}
func (s *ConfigService) SetAIConfig(config AIConfig) error { return nil }
func (s *ConfigService) GetSSHSettings() SSHSettings {
	return SSHSettings{DefaultPort: 22, ConnectTimeout: 30, KeepAlive: true, KeepAliveInterval: 30, AutoReconnect: true, MaxReconnectAttempts: 3}
}
func (s *ConfigService) SetSSHSettings(settings SSHSettings) error { return nil }

// 预留结构体（保持 Wails 绑定兼容）
type FileManagerConfig struct {
	ShowHidden    bool   `json:"showHidden"`
	SortBy        string `json:"sortBy"`
	SortOrder     string `json:"sortOrder"`
	ConfirmDelete bool   `json:"confirmDelete"`
}

type AIConfig struct {
	Enabled          bool `json:"enabled"`
	AutoExecute      bool `json:"autoExecute"`
	ConfirmExecution bool `json:"confirmExecution"`
}

type SSHSettings struct {
	DefaultPort          int  `json:"defaultPort"`
	ConnectTimeout       int  `json:"connectTimeout"`
	KeepAlive            bool `json:"keepAlive"`
	KeepAliveInterval    int  `json:"keepAliveInterval"`
	AutoReconnect        bool `json:"autoReconnect"`
	MaxReconnectAttempts int  `json:"maxReconnectAttempts"`
}
