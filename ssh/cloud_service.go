package ssh

import (
	"fmt"
	"sync"
	"time"

	"changeme/cloud/client"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// CloudService 云端服务（Wails 暴露给前端）
type CloudService struct {
	mu       sync.RWMutex
	app      *application.App
	client   *client.CloudClient
	config   *ConfigService
	running  bool
	stopCh   chan struct{}
}

// NewCloudService 创建云端服务
func NewCloudService(configService *ConfigService) *CloudService {
	return &CloudService{
		config: configService,
	}
}

// SetApp 设置应用实例
func (cs *CloudService) SetApp(app *application.App) {
	cs.app = app
}

// Connect 连接到云端（不返回 error，避免 Wails 打印 ERR 日志）
func (cs *CloudService) Connect(serverAddr, token string) bool {
	fmt.Printf("[CloudService] 连接云端: %s\n", serverAddr)

	// 断开旧连接
	cs.Disconnect()

	cs.mu.Lock()
	defer cs.mu.Unlock()

	c := client.NewCloudClient(serverAddr, token)
	fmt.Println("[CloudService] 正在建立 TLS 连接...")
	if err := c.Connect(); err != nil {
		fmt.Printf("[CloudService] ✗ 连接失败: %v\n", err)
		cs.emitStatus(false, err.Error())
		return false
	}
	fmt.Println("[CloudService] ✓ TLS 连接成功")

	// 注册设备
	fmt.Println("[CloudService] 正在注册设备...")
	if err := c.Register("启SSH客户端", "localhost", "windows", "0.3.0"); err != nil {
		fmt.Printf("[CloudService] ✗ 注册失败: %v\n", err)
		c.Disconnect()
		cs.emitStatus(false, err.Error())
		return false
	}
	fmt.Println("[CloudService] ✓ 设备注册成功")

	cs.client = c
	cs.running = true

	// 启动心跳
	cs.stopCh = make(chan struct{})
	go cs.heartbeatLoop()

	cs.emitStatus(true, "")
	fmt.Println("[CloudService] ✓ 已连接到云端")
	return true
}

// emitStatus 发送连接状态事件给前端
func (cs *CloudService) emitStatus(connected bool, errMsg string) {
	if cs.app != nil {
		cs.app.Event.Emit("cloud:status", map[string]interface{}{
			"connected": connected,
			"error":     errMsg,
		})
	}
}

// Disconnect 断开连接
func (cs *CloudService) Disconnect() {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.stopCh != nil {
		close(cs.stopCh)
		cs.stopCh = nil
	}
	if cs.client != nil {
		cs.client.Disconnect()
		cs.client = nil
		cs.running = false
		fmt.Println("[CloudService] 已断开云端连接")

		// 通知前端
		if cs.app != nil {
			cs.app.Event.Emit("cloud:status", map[string]interface{}{
				"connected": false,
			})
		}
	}
}

// IsConnected 检查是否连接
func (cs *CloudService) IsConnected() bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	connected := cs.client != nil && cs.client.IsConnected()
	return connected
}

// PullSync 拉取同步数据
func (cs *CloudService) PullSync() ([]client.SyncConnection, error) {
	cs.mu.RLock()
	c := cs.client
	cs.mu.RUnlock()

	if c == nil || !c.IsConnected() {
		return nil, fmt.Errorf("未连接到云端")
	}

	fmt.Println("[CloudService] 拉取同步数据...")
	data, err := c.PullSync()
	if err != nil {
		fmt.Printf("[CloudService] ✗ 拉取失败: %v\n", err)
		return nil, err
	}

	if data == nil {
		return []client.SyncConnection{}, nil
	}
	fmt.Printf("[CloudService] ✓ 拉取成功: %d 个连接\n", len(data.Connections))
	return data.Connections, nil
}

// PushSync 推送同步数据
func (cs *CloudService) PushSync(connections []client.SyncConnection) error {
	cs.mu.RLock()
	c := cs.client
	cs.mu.RUnlock()

	if c == nil || !c.IsConnected() {
		return fmt.Errorf("未连接到云端")
	}

	fmt.Printf("[CloudService] 推送同步数据: %d 个连接\n", len(connections))
	err := c.PushSync(client.SyncData{
		Connections: connections,
		UpdatedAt:   time.Now(),
	})
	if err != nil {
		fmt.Printf("[CloudService] ✗ 推送失败: %v\n", err)
		return err
	}
	fmt.Println("[CloudService] ✓ 推送成功")
	return nil
}

// heartbeatLoop 心跳循环
func (cs *CloudService) heartbeatLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-cs.stopCh:
			return
		case <-ticker.C:
			cs.mu.RLock()
			c := cs.client
			cs.mu.RUnlock()

			if c != nil && c.IsConnected() {
				if err := c.Heartbeat(); err != nil {
					fmt.Printf("[CloudService] 心跳失败: %v\n", err)
				}
			}
		}
	}
}
