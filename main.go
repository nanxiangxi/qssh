package main

import (
	"changeme/ai"
	"changeme/ssh"
	"embed"
	_ "embed"
	"fmt"
	"log"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// Wails 使用 Go 的 `embed` 包将前端文件嵌入到二进制文件中。
// frontend/dist 文件夹中的所有文件将被嵌入到二进制文件中，
// 并提供给前端使用。
// 更多信息请参见 https://pkg.go.dev/embed

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var trayIcon []byte

// 版本信息
const (
	AppVersion = "0.4.0"
	AppName    = "启SSH"
	AppNameEN  = "Qssh"
)

// getAppName 根据语言设置返回应用名称
func getAppName(locale string) string {
	if locale == "en-US" {
		return AppNameEN
	}
	return AppName
}

func init() {
	application.RegisterEvent[string]("time")
}

// appServices 应用服务集合
type appServices struct {
	sshService    *ssh.SSHService
	configService *ssh.ConfigService
	windowManager *ssh.WindowManager
	services      []application.Service
}

// initServices 初始化所有服务并处理依赖注入
func initServices(app *application.App) *appServices {
	// 核心服务
	sshService := ssh.NewSSHService()
	aiService := ai.NewAIService()
	configService := ssh.NewConfigService()

	// 配置依赖
	configService.SetApp(app)
	sshService.SetApp(app)
	aiService.SetApp(app)

	// 窗口管理器
	windowManager := ssh.NewWindowManager(app, nil)
	windowManager.SetConfigService(configService)
	sshService.SetWindowManager(windowManager)

	// AI 依赖注入
	ai.InitDeps(aiService, sshService, sshService)

	// 启动健康检查
	sshService.StartHealthCheck()

	// 依赖 sshService 的服务
	portForwardService := ssh.NewPortForwardService(sshService)
	firewallService := ssh.NewFirewallService(sshService)
	guardianService := ssh.NewProcessGuardianService(sshService)
	databaseService := ssh.NewDatabaseService(sshService)

	// 依赖 configService 的服务
	cloudService := ssh.NewCloudService(configService)
	cloudService.SetApp(app)

	return &appServices{
		sshService:    sshService,
		configService: configService,
		windowManager: windowManager,
		services: []application.Service{
			application.NewService(ssh.NewGreetService(AppVersion)),
			application.NewService(sshService),
			application.NewService(aiService),
			application.NewService(configService),
			application.NewService(portForwardService),
			application.NewService(firewallService),
			application.NewService(guardianService),
			application.NewService(databaseService),
			application.NewService(cloudService),
		},
	}
}

func main() {
	// 创建应用实例
	app := application.New(application.Options{
		Name:        AppName,
		Description: fmt.Sprintf("一个中文的SSH工具，便携，简单，开源 (v%s)", AppVersion),
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
	})

	// 初始化所有服务（处理依赖注入）
	as := initServices(app)
	sshService := as.sshService
	configService := as.configService
	windowManager := as.windowManager

	// 注册服务
	for _, svc := range as.services {
		app.RegisterService(svc)
	}

	// 根据语言设置确定应用名称
	locale := configService.GetConfig().Locale
	appDisplayName := getAppName(locale)

	// 使用必要的选项创建一个新窗口。
	mainWindow := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "main",
		Title:            fmt.Sprintf("%s - SSH工具 (v%s)", appDisplayName, AppVersion),
		URL:              "/",
		DisableResize:    false,
		Frameless:        true,
		EnableFileDrop:   true,
		BackgroundColour: application.NewRGB(255, 255, 255),
	})

	// 文件拖放处理
	mainWindow.OnWindowEvent(events.Common.WindowFilesDropped, func(event *application.WindowEvent) {
		files := event.Context().DroppedFiles()
		if len(files) > 0 {
			app.Event.Emit("file-manager:files-dropped", map[string]interface{}{
				"files": files,
			})
		}
	})

	// 尝试恢复主窗口位置，否则使用默认尺寸居中
	mainWindowRegistered := windowManager.RestoreMainWindowPosition(mainWindow)
	if !mainWindowRegistered {
		w, h := calculateWindowSize(app)
		mainWindow.SetSize(w, h)
		mainWindow.Center()
		fmt.Printf("[Main] 主窗口大小: %dx%d\n", w, h)
	}

	// 设置分组关闭回调（所有 SSH 窗口关闭后自动显示主窗口）
	windowManager.SetOnGroupClose(func(groupID string) {
		fmt.Printf("[Main] 🗑️ 窗口销毁，关闭分组: %s\n", groupID)
		if err := sshService.CloseGroup(groupID); err != nil {
			fmt.Printf("[Main] ⚠️ 关闭分组失败: %v\n", err)
		}
		// 检查是否启用了自动显示首页
		cfg := configService.GetConfig()
		if windowManager.GetWindowCount() == 0 && cfg.UI.AutoShowHome {
			fmt.Println("[Main] 所有 SSH 窗口已关闭，显示主窗口")
			mainWindow.Show()
			mainWindow.Focus()
		}
	})

	// 定时保存主窗口位置（每3秒检查一次）
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			windowManager.SaveMainWindowPositionIfChanged(mainWindow)
		}
	}()

	// 监听主窗口关闭事件，保存位置
	mainWindow.OnWindowEvent(events.Common.WindowClosing, func(e *application.WindowEvent) {
		fmt.Println("[Main] 主窗口关闭，保存位置")
		windowManager.SaveMainWindowPosition(mainWindow)
	})

	// ========== 系统托盘（延迟初始化，避免 GetCursorPos 崩溃） ==========
	go func() {
		// 等待 WebView2 环境就绪
		time.Sleep(2 * time.Second)

		systray := app.SystemTray.New()
		systray.SetIcon(trayIcon)
		systray.SetLabel(AppName)

		// 托盘菜单
		trayMenu := app.NewMenu()
		trayMenu.Add("显示主窗口").OnClick(func(ctx *application.Context) {
			mainWindow.Show()
			mainWindow.Focus()
		})
		trayMenu.Add("显示全部窗口").OnClick(func(ctx *application.Context) {
			mainWindow.Show()
			mainWindow.Focus()
			windowManager.ShowAllWindows()
		})
		trayMenu.AddSeparator()
		trayMenu.Add("隐藏全部窗口").OnClick(func(ctx *application.Context) {
			windowManager.HideAllWindows()
			windowManager.SaveMainWindowPosition(mainWindow)
			mainWindow.Hide()
		})
		trayMenu.AddSeparator()
		trayMenu.Add("退出").OnClick(func(ctx *application.Context) {
			mainWindow.Show() // 退出前显示窗口，确保正常关闭
			windowManager.ShowAllWindows()
			app.Quit()
		})
		systray.SetMenu(trayMenu)

		// 左键点击托盘图标切换主窗口显示/隐藏
		systray.OnClick(func() {
			if mainWindow.IsVisible() {
				windowManager.SaveMainWindowPosition(mainWindow)
				mainWindow.Hide()
			} else {
				mainWindow.Show()
				mainWindow.Focus()
			}
		})

		fmt.Println("[Main] 系统托盘已初始化")
	}()

	// 监听前端请求隐藏主窗口的事件（SSH连接成功后自动隐藏主窗口，不影响SSH窗口）
	app.Event.On("ssh:tray-hide", func(event *application.CustomEvent) {
		fmt.Println("[Main] 收到托盘隐藏请求，隐藏主窗口")
		windowManager.SaveMainWindowPosition(mainWindow)
		mainWindow.Hide()
	})

	// 创建一个 goroutine，每秒发射一个包含当前时间的事件。
	// 前端可以监听此事件并相应地更新 UI。
	go func() {
		for {
			now := time.Now().Format(time.RFC1123)
			app.Event.Emit("time", now)
			time.Sleep(time.Second)
		}
	}()

	// 创建一个 goroutine，每500ms广播SSH分组状态
	go func() {
		for {
			time.Sleep(500 * time.Millisecond)

			// 获取所有分组
			groups := sshService.GetAllGroups()

			// 为每个分组发送状态
			for _, group := range groups {
				// 获取完整的连接信息（包含名称）
				connInfos := sshService.GetGroupConnectionInfos(group.ID)

				// 使用 map[string]interface{} 作为事件数据
				app.Event.Emit("ssh:group-updated", map[string]interface{}{
					"groupID":     group.ID,
					"action":      "state-sync",
					"connections": connInfos, // 发送完整连接信息
				})
			}

			// 广播全局连接状态（用于首页侧边栏）
			allConnections := sshService.GetAllConnections()

			// 确保connections不是nil，而是空数组
			if allConnections == nil {
				allConnections = []*ssh.ConnectionInfo{}
			}

			app.Event.Emit("ssh:connections-updated", map[string]interface{}{
				"connections": allConnections,
				"timestamp":   time.Now().UnixMilli(),
			})
		}
	}()

	// 创建一个 goroutine，定期读取所有 Shell session 的输出并发送到前端
	go func() {
		buf := make([]byte, 4096)
		for {
			time.Sleep(20 * time.Millisecond)

			connIDs := sshService.GetConnectionList()
			for _, connID := range connIDs {
				sessionIDs := sshService.GetSessionIDs(connID)
				for _, sessionID := range sessionIDs {
					if !sshService.IsShellSessionActive(connID, sessionID) {
						continue
					}
					n, err := sshService.ReadFromShellSession(connID, sessionID, buf)
					if err != nil {
						continue
					}
					if n > 0 {
						app.Event.Emit("ssh:terminal-output", map[string]interface{}{
							"connID":    connID,
							"sessionID": sessionID,
							"data":      string(buf[:n]),
						})
					}
				}
			}
		}
	}()

	// 运行应用程序。这将阻塞直到应用程序退出。
	appErr := app.Run()

	// 如果运行应用程序时发生错误，记录错误并退出。
	if appErr != nil {
		log.Fatal(appErr)
	}
}

// calculateWindowSize 根据屏幕大小计算合适的窗口尺寸
func calculateWindowSize(app *application.App) (int, int) {
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

	primary := app.Screen.GetPrimary()
	if primary == nil {
		return 1400, 900
	}

	screenW := primary.Size.Width
	screenH := primary.Size.Height

	maxW := int(float64(screenW) * 0.85)
	maxH := int(float64(screenH) * 0.85)

	fmt.Printf("[Main] 屏幕大小: %dx%d, 最大窗口: %dx%d\n", screenW, screenH, maxW, maxH)

	for _, s := range sizes {
		if s.width <= maxW && s.height <= maxH {
			return s.width, s.height
		}
	}

	return 1200, 800
}

// readFromShell 从Shell读取数据（非阻塞）
func readFromShell(sshService *ssh.SSHService, connID string, buf []byte) (int, error) {
	return sshService.ReadFromShell(connID, buf)
}
