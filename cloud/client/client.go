package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Message 协议消息
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// SyncConnection 同步的连接配置
type SyncConnection struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	Username  string    `json:"username"`
	Password  string    `json:"password,omitempty"`
	KeyPath   string    `json:"keyPath,omitempty"`
	Source    string    `json:"source,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

// SyncData 同步数据
type SyncData struct {
	Connections []SyncConnection `json:"connections"`
	UpdatedAt   time.Time        `json:"updatedAt"`
	DeviceID    string           `json:"deviceId"`
}

// CloudClient 云端客户端
type CloudClient struct {
	mu         sync.RWMutex
	serverAddr string
	token      string
	deviceID   string
	conn       *websocket.Conn
	connected  bool
}

// NewCloudClient 创建云端客户端
func NewCloudClient(serverAddr, token string) *CloudClient {
	return &CloudClient{
		serverAddr: serverAddr,
		token:      token,
	}
}

// Connect 连接到云端服务器
func (cc *CloudClient) Connect() error {
	u := url.URL{
		Scheme: "wss",
		Host:   cc.serverAddr,
		Path:   "/ws",
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("WebSocket 连接失败: %v", err)
	}

	cc.mu.Lock()
	cc.conn = conn
	cc.connected = true
	cc.mu.Unlock()

	return nil
}

// Disconnect 断开连接
func (cc *CloudClient) Disconnect() {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	if cc.conn != nil {
		cc.conn.Close()
		cc.conn = nil
	}
	cc.connected = false
}

// IsConnected 检查是否连接
func (cc *CloudClient) IsConnected() bool {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.connected && cc.conn != nil
}

// send 发送消息并接收响应
func (cc *CloudClient) send(msg Message) (Message, error) {
	cc.mu.RLock()
	conn := cc.conn
	connected := cc.connected
	cc.mu.RUnlock()

	if conn == nil || !connected {
		return Message{}, fmt.Errorf("未连接")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return Message{}, fmt.Errorf("序列化失败: %v", err)
	}

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		cc.Disconnect()
		return Message{}, fmt.Errorf("发送失败: %v", err)
	}

	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	_, respData, err := conn.ReadMessage()
	if err != nil {
		cc.Disconnect()
		return Message{}, fmt.Errorf("接收失败: %v", err)
	}

	var resp Message
	if err := json.Unmarshal(respData, &resp); err != nil {
		return Message{}, fmt.Errorf("解析失败: %v", err)
	}

	return resp, nil
}

// Register 注册设备
func (cc *CloudClient) Register(deviceName, host, osInfo, version string) error {
	payload, _ := json.Marshal(map[string]interface{}{
		"name":      deviceName,
		"host":      host,
		"port":      0,
		"os":        osInfo,
		"version":   version,
		"token":     cc.token,
		"timestamp": time.Now().Format(time.RFC3339),
		"nonce":     generateNonce(),
	})

	resp, err := cc.send(Message{Type: "register", Payload: payload})
	if err != nil {
		return err
	}

	if resp.Type == "error" {
		return fmt.Errorf("注册失败: %s", string(resp.Payload))
	}

	var result map[string]interface{}
	json.Unmarshal(resp.Payload, &result)
	if id, ok := result["id"].(string); ok {
		cc.deviceID = id
	}

	return nil
}

// Heartbeat 发送心跳
func (cc *CloudClient) Heartbeat() error {
	payload, _ := json.Marshal(map[string]string{"deviceId": cc.deviceID})
	_, err := cc.send(Message{Type: "heartbeat", Payload: payload})
	return err
}

// PullSync 拉取同步数据
func (cc *CloudClient) PullSync() (*SyncData, error) {
	resp, err := cc.send(Message{Type: "sync-pull", Payload: json.RawMessage("{}")})
	if err != nil {
		return nil, err
	}

	if resp.Type == "error" {
		return nil, fmt.Errorf("拉取失败: %s", string(resp.Payload))
	}

	var data SyncData
	json.Unmarshal(resp.Payload, &data)
	return &data, nil
}

// PushSync 推送同步数据
func (cc *CloudClient) PushSync(data SyncData) error {
	data.DeviceID = cc.deviceID
	payload, _ := json.Marshal(data)

	resp, err := cc.send(Message{Type: "sync-push", Payload: payload})
	if err != nil {
		return err
	}

	if resp.Type == "error" {
		return fmt.Errorf("推送失败: %s", string(resp.Payload))
	}

	return nil
}

func generateNonce() string {
	b := make([]byte, 16)
	for i := range b {
		b[i] = byte(time.Now().UnixNano() >> (8 * (i % 8)))
	}
	return fmt.Sprintf("%x", b)
}
