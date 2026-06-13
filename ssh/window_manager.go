package ssh

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// WindowPosition 窗口位置信息
type WindowPosition struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// WindowManager SSH窗口管理器
type WindowManager struct {
	app            *application.App
	configService  *ConfigService
	windowMutex    sync.RWMutex
	windows        map[string]*application.WebviewWindow // groupID -> Window
	onGroupClose   func(groupID string)                  // 分组关闭回调
	positions      map[string]*WindowPosition            // 窗口位置缓存
	positionsFile  string                                // 位置文件路径
	positionsMutex sync.RWMutex
}

// NewWindowManager 创建窗口管理器实例
func NewWindowManager(app *application.App, onGroupClose func(groupID string)) *WindowManager {
	wm := &WindowManager{
		app:           app,
		windows:       make(map[string]*application.WebviewWindow),
		onGroupClose:  onGroupClose,
		positions:     make(map[string]*WindowPosition),
		positionsFile: getWindowPositionsFile(),
	}

	// 加载已保存的窗口位置
	wm.loadPositions()

	return wm
}

// SetConfigService 设置配置服务
func (wm *WindowManager) SetConfigService(cs *ConfigService) {
	wm.configService = cs
}

// SetOnGroupClose 设置分组关闭回调
func (wm *WindowManager) SetOnGroupClose(callback func(groupID string)) {
	wm.onGroupClose = callback
}

// getWindowPositionsFile 获取窗口位置文件路径
func getWindowPositionsFile() string {
	exePath, err := os.Executable()
	if err != nil {
		return "data/config/window_positions.json"
	}
	exeDir := filepath.Dir(exePath)
	dir := filepath.Join(exeDir, "data", "config")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "window_positions.json")
}

// loadPositions 从文件加载窗口位置
func (wm *WindowManager) loadPositions() {
	data, err := os.ReadFile(wm.positionsFile)
	if err != nil {
		return
	}
	var positions map[string]*WindowPosition
	if err := json.Unmarshal(data, &positions); err != nil {
		return
	}
	wm.positions = positions
	fmt.Printf("[WindowManager] 已加载 %d 个窗口位置记录\n", len(positions))
}

// savePositions 保存窗口位置到文件
func (wm *WindowManager) savePositions() {
	wm.positionsMutex.RLock()
	data, err := json.MarshalIndent(wm.positions, "", "  ")
	wm.positionsMutex.RUnlock()

	if err != nil {
		return
	}
	os.WriteFile(wm.positionsFile, data, 0644)
}

// isRememberPositionEnabled 检查是否启用位置记忆
func (wm *WindowManager) isRememberPositionEnabled() bool {
	if wm.configService == nil {
		return true // 默认启用
	}
	cfg := wm.configService.GetConfig()
	return cfg.UI.RememberPosition
}

// SaveMainWindowPositionIfChanged 仅在位置有变化时保存主窗口位置
func (wm *WindowManager) SaveMainWindowPositionIfChanged(window *application.WebviewWindow) {
	if !wm.isRememberPositionEnabled() {
		return
	}

	x, y := window.Position()
	w, h := window.Size()

	if x < -10000 || y < -10000 || w < 100 || h < 100 {
		return
	}

	wm.positionsMutex.RLock()
	existing := wm.positions["main"]
	wm.positionsMutex.RUnlock()

	if existing != nil && existing.X == x && existing.Y == y && existing.Width == w && existing.Height == h {
		return
	}

	wm.positionsMutex.Lock()
	wm.positions["main"] = &WindowPosition{X: x, Y: y, Width: w, Height: h}
	wm.positionsMutex.Unlock()

	wm.savePositions()
}

// SaveMainWindowPosition 保存主窗口位置
func (wm *WindowManager) SaveMainWindowPosition(window *application.WebviewWindow) {
	if !wm.isRememberPositionEnabled() {
		fmt.Println("[WindowManager] 位置记忆未启用，跳过保存主窗口位置")
		return
	}

	x, y := window.Position()
	w, h := window.Size()

	fmt.Printf("[WindowManager] 主窗口当前状态: (%d,%d %dx%d)\n", x, y, w, h)

	// 只保存有效的非最小化位置
	if x < -10000 || y < -10000 || w < 100 || h < 100 {
		fmt.Println("[WindowManager] 主窗口位置无效，跳过保存")
		return
	}

	wm.positionsMutex.Lock()
	wm.positions["main"] = &WindowPosition{X: x, Y: y, Width: w, Height: h}
	wm.positionsMutex.Unlock()

	wm.savePositions()
	fmt.Printf("[WindowManager] ✓ 已保存主窗口位置: (%d,%d %dx%d)\n", x, y, w, h)
}

// RestoreMainWindowPosition 恢复主窗口位置，返回是否恢复成功
func (wm *WindowManager) RestoreMainWindowPosition(window *application.WebviewWindow) bool {
	if !wm.isRememberPositionEnabled() {
		return false
	}

	wm.positionsMutex.RLock()
	pos, exists := wm.positions["main"]
	wm.positionsMutex.RUnlock()

	if !exists || pos == nil {
		return false
	}

	// 验证位置是否在屏幕范围内
	primary := wm.app.Screen.GetPrimary()
	if primary != nil {
		screenW := primary.Size.Width
		screenH := primary.Size.Height
		if pos.X < -100 || pos.Y < -100 || pos.X > screenW-50 || pos.Y > screenH-50 {
			fmt.Printf("[WindowManager] 主窗口位置超出屏幕范围，跳过恢复\n")
			return false
		}
	}

	window.SetPosition(pos.X, pos.Y)
	window.SetSize(pos.Width, pos.Height)
	fmt.Printf("[WindowManager] 已恢复主窗口位置: (%d,%d %dx%d)\n", pos.X, pos.Y, pos.Width, pos.Height)
	return true
}

// saveWindowPositionIfChanged 仅在位置有变化时保存
func (wm *WindowManager) saveWindowPositionIfChanged(windowID string, window *application.WebviewWindow) {
	if !wm.isRememberPositionEnabled() {
		return
	}

	x, y := window.Position()
	w, h := window.Size()

	// 无效位置跳过
	if x < -10000 || y < -10000 || w < 100 || h < 100 {
		return
	}

	wm.positionsMutex.RLock()
	existing := wm.positions[windowID]
	wm.positionsMutex.RUnlock()

	// 位置没变化，跳过
	if existing != nil && existing.X == x && existing.Y == y && existing.Width == w && existing.Height == h {
		return
	}

	wm.positionsMutex.Lock()
	wm.positions[windowID] = &WindowPosition{X: x, Y: y, Width: w, Height: h}
	wm.positionsMutex.Unlock()

	wm.savePositions()
}

// saveWindowPosition 保存 SSH 窗口位置
func (wm *WindowManager) saveWindowPosition(windowID string) {
	if !wm.isRememberPositionEnabled() {
		fmt.Println("[WindowManager] 位置记忆未启用，跳过保存窗口位置")
		return
	}

	wm.windowMutex.RLock()
	var targetWindow *application.WebviewWindow
	for gid, w := range wm.windows {
		if gid == windowID {
			targetWindow = w
			break
		}
	}
	wm.windowMutex.RUnlock()

	if targetWindow == nil {
		fmt.Printf("[WindowManager] 窗口 %s 未找到，跳过保存\n", windowID)
		return
	}

	x, y := targetWindow.Position()
	w, h := targetWindow.Size()

	fmt.Printf("[WindowManager] 窗口 %s 当前状态: (%d,%d %dx%d)\n", windowID, x, y, w, h)

	// 只保存有效的非最小化位置
	if x < -10000 || y < -10000 || w < 100 || h < 100 {
		fmt.Printf("[WindowManager] 窗口 %s 位置无效，跳过保存\n", windowID)
		return
	}

	wm.positionsMutex.Lock()
	wm.positions[windowID] = &WindowPosition{X: x, Y: y, Width: w, Height: h}
	wm.positionsMutex.Unlock()

	wm.savePositions()
	fmt.Printf("[WindowManager] ✓ 已保存窗口位置: %s (%d,%d %dx%d)\n", windowID, x, y, w, h)
}

// restoreWindowPosition 恢复窗口位置
func (wm *WindowManager) restoreWindowPosition(windowID string, window *application.WebviewWindow) {
	if !wm.isRememberPositionEnabled() {
		return
	}

	wm.positionsMutex.RLock()
	pos, exists := wm.positions[windowID]
	wm.positionsMutex.RUnlock()

	if !exists || pos == nil {
		return
	}

	// 验证位置是否在屏幕范围内
	primary := wm.app.Screen.GetPrimary()
	if primary != nil {
		screenW := primary.Size.Width
		screenH := primary.Size.Height
		// 位置超出屏幕范围，不恢复
		if pos.X < -100 || pos.Y < -100 || pos.X > screenW-50 || pos.Y > screenH-50 {
			fmt.Printf("[WindowManager] 窗口位置超出屏幕范围，跳过恢复: %s\n", windowID)
			return
		}
	}

	window.SetPosition(pos.X, pos.Y)
	window.SetSize(pos.Width, pos.Height)
	fmt.Printf("[WindowManager] 已恢复窗口位置: %s (%d,%d %dx%d)\n", windowID, pos.X, pos.Y, pos.Width, pos.Height)
}

// CreateSSHWindow 创建或聚焦SSH分组窗口
func (wm *WindowManager) CreateSSHWindow(groupID string, groupName string, activeConnID string) error {
	fmt.Printf("[WindowManager] CreateSSHWindow 被调用: groupID=%s, groupName=%s, activeConn=%s\n", groupID, groupName, activeConnID)

	wm.windowMutex.RLock()
	existingWindow, exists := wm.windows[groupID]
	wm.windowMutex.RUnlock()

	// 如果窗口已存在，聚焦现有窗口（不创建新窗口）
	if exists && existingWindow != nil {
		fmt.Printf("[WindowManager] ✓ 窗口已存在，聚焦现有窗口: %s\n", groupID)
		existingWindow.Focus()
		return nil
	}

	fmt.Printf("[WindowManager] 创建新窗口: %s\n", groupID)

	// 创建新窗口
	windowTitle := groupName
	if windowTitle == "" {
		windowTitle = "SSH 终端"
	}

	// 构建 URL，传递 groupID 和 activeConn 参数
	url := "/#/ssh?group=" + groupID
	if activeConnID != "" {
		url += "&activeConn=" + activeConnID
		fmt.Printf("[WindowManager] 窗口URL: %s (包含 activeConn)\n", url)
	} else {
		fmt.Printf("[WindowManager] 窗口URL: %s\n", url)
	}

	// 窗口名称格式：ssh-{groupID}，便于通过 GetByName 查找
	windowName := "ssh-" + groupID

	newWindow := wm.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             windowName,
		Title:            windowTitle,
		URL:              url,
		DisableResize:    false,
		Frameless:        true,
		BackgroundColour: application.NewRGB(30, 30, 30),
	})

	// 尝试恢复窗口位置，否则使用默认尺寸居中
	wm.restoreWindowPosition(groupID, newWindow)

	// 如果没有恢复到位置（没有保存过），使用默认尺寸居中
	if !wm.isRememberPositionEnabled() || wm.positions[groupID] == nil {
		w, h := wm.calculateWindowSize()
		newWindow.SetSize(w, h)
		newWindow.Center()
		fmt.Printf("[WindowManager] 使用默认窗口大小: %dx%d\n", w, h)
	}

	// 保存窗口引用
	wm.windowMutex.Lock()
	wm.windows[groupID] = newWindow
	wm.windowMutex.Unlock()

	// 显示并聚焦窗口
	newWindow.Show()
	newWindow.Focus()
	fmt.Printf("[WindowManager] 窗口引用已保存并聚焦: %s\n", groupID)

	// 定时保存窗口位置（每3秒检查一次，有变化才写盘）
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			wm.windowMutex.RLock()
			win, exists := wm.windows[groupID]
			wm.windowMutex.RUnlock()
			if !exists {
				return // 窗口已关闭，停止定时器
			}
			wm.saveWindowPositionIfChanged(groupID, win)
		}
	}()

	// 监听窗口关闭事件，保存最终位置并销毁分组
	newWindow.OnWindowEvent(events.Common.WindowClosing, func(e *application.WindowEvent) {
		fmt.Printf("[WindowManager] 🗑️ 窗口关闭事件触发: %s\n", groupID)

		// 关闭前最后保存一次
		wm.saveWindowPosition(groupID)

		// 清理窗口引用
		wm.windowMutex.Lock()
		delete(wm.windows, groupID)
		wm.windowMutex.Unlock()

		// 调用回调函数，销毁分组
		if wm.onGroupClose != nil {
			fmt.Printf("[WindowManager] 🔄 调用分组销毁回调\n")
			wm.onGroupClose(groupID)
		}
	})

	return nil
}

// calculateWindowSize 根据屏幕大小计算合适的窗口尺寸
func (wm *WindowManager) calculateWindowSize() (int, int) {
	// 档位定义
	type size struct {
		width  int
		height int
	}
	sizes := []size{
		{1920, 1080},
		{1600, 1000},
		{1400, 900},
		{1200, 800},
	}

	// 获取主屏幕
	primary := wm.app.Screen.GetPrimary()
	if primary == nil {
		return 1400, 900
	}

	screenW := primary.Size.Width
	screenH := primary.Size.Height

	// 选择不超过屏幕 85% 的最大档位
	maxW := int(float64(screenW) * 0.85)
	maxH := int(float64(screenH) * 0.85)

	fmt.Printf("[WindowManager] 屏幕大小: %dx%d, 最大窗口: %dx%d\n", screenW, screenH, maxW, maxH)

	for _, s := range sizes {
		if s.width <= maxW && s.height <= maxH {
			return s.width, s.height
		}
	}

	return 1200, 800
}

// CloseWindow 关闭分组窗口并清理
func (wm *WindowManager) CloseWindow(groupID string) {
	wm.windowMutex.Lock()
	defer wm.windowMutex.Unlock()

	if window, exists := wm.windows[groupID]; exists {
		// 保存窗口位置
		wm.saveWindowPosition(groupID)

		window.Close()
		delete(wm.windows, groupID)
		fmt.Printf("[WindowManager] 窗口已关闭并清理: %s\n", groupID)
	}
}

// CleanupWindow 清理窗口引用（不关闭窗口，仅清理引用）
// 用于前端通知后端窗口已通过系统方式关闭的情况
func (wm *WindowManager) CleanupWindow(groupID string) {
	wm.windowMutex.Lock()
	defer wm.windowMutex.Unlock()

	if _, exists := wm.windows[groupID]; exists {
		delete(wm.windows, groupID)
		fmt.Printf("[WindowManager] 窗口引用已清理: %s\n", groupID)
	}
}

// GetWindowCount 获取当前打开的窗口数量
func (wm *WindowManager) GetWindowCount() int {
	wm.windowMutex.RLock()
	defer wm.windowMutex.RUnlock()
	return len(wm.windows)
}

// HasWindow 检查指定分组的窗口是否存在
func (wm *WindowManager) HasWindow(groupID string) bool {
	wm.windowMutex.RLock()
	defer wm.windowMutex.RUnlock()
	_, exists := wm.windows[groupID]
	return exists
}

// HideAllWindows 隐藏所有窗口
func (wm *WindowManager) HideAllWindows() {
	wm.windowMutex.RLock()
	defer wm.windowMutex.RUnlock()
	for _, window := range wm.windows {
		window.Hide()
	}
	fmt.Println("[WindowManager] 已隐藏所有 SSH 窗口")
}

// ShowAllWindows 显示所有窗口
func (wm *WindowManager) ShowAllWindows() {
	wm.windowMutex.RLock()
	defer wm.windowMutex.RUnlock()
	for _, window := range wm.windows {
		window.Show()
		window.Focus()
	}
	fmt.Println("[WindowManager] 已显示所有 SSH 窗口")
}

// ClearPositions 清除所有窗口位置记忆
func (wm *WindowManager) ClearPositions() {
	wm.positionsMutex.Lock()
	wm.positions = make(map[string]*WindowPosition)
	wm.positionsMutex.Unlock()

	// 删除位置文件
	os.Remove(wm.positionsFile)
	fmt.Println("[WindowManager] 已清除所有窗口位置记忆")
}
