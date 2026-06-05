package ssh

import (
	"fmt"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// WindowManager SSH窗口管理器
type WindowManager struct {
	app          *application.App
	windowMutex  sync.RWMutex
	windows      map[string]*application.WebviewWindow // groupID -> Window
	onGroupClose func(groupID string)                  // 分组关闭回调
}

// NewWindowManager 创建窗口管理器实例
func NewWindowManager(app *application.App, onGroupClose func(groupID string)) *WindowManager {
	return &WindowManager{
		app:          app,
		windows:      make(map[string]*application.WebviewWindow),
		onGroupClose: onGroupClose,
	}
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

	newWindow := wm.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            windowTitle,
		Width:            1600,
		MinWidth:         1600,
		Height:           1000,
		MinHeight:        1000,
		URL:              url,
		DisableResize:    false,
		Frameless:        true,
		BackgroundColour: application.NewRGB(30, 30, 30),
	})

	fmt.Printf("[WindowManager] 窗口创建成功\n")

	// 保存窗口引用
	wm.windowMutex.Lock()
	wm.windows[groupID] = newWindow
	wm.windowMutex.Unlock()

	fmt.Printf("[WindowManager] 窗口引用已保存: %s\n", groupID)

	// 监听窗口关闭事件，自动销毁分组
	newWindow.OnWindowEvent(events.Common.WindowClosing, func(e *application.WindowEvent) {
		fmt.Printf("[WindowManager] 🗑️ 窗口关闭事件触发: %s\n", groupID)

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

// CloseWindow 关闭分组窗口并清理
func (wm *WindowManager) CloseWindow(groupID string) {
	wm.windowMutex.Lock()
	defer wm.windowMutex.Unlock()

	if window, exists := wm.windows[groupID]; exists {
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
