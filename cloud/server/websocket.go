package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WSServer WebSocket TLS 服务器
type WSServer struct {
	mu      sync.RWMutex
	server  *http.Server
	config  *Config
	srv     *Server
	clients map[string]*websocket.Conn
	running bool
}

// Message 协议消息
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// NewWSServer 创建 WebSocket 服务器
func NewWSServer(config *Config, server *Server) *WSServer {
	return &WSServer{
		config:  config,
		srv:     server,
		clients: make(map[string]*websocket.Conn),
	}
}

// Start 启动 WebSocket 服务器
func (ws *WSServer) Start(dataDir string) error {
	certFile := filepath.Join(dataDir, "cert.pem")
	keyFile := filepath.Join(dataDir, "key.pem")

	if err := generateCertIfNeeded(certFile, keyFile); err != nil {
		return fmt.Errorf("生成证书失败: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", ws.handleWS)

	addr := fmt.Sprintf(":%d", ws.config.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	ws.server = srv
	ws.running = true

	go func() {
		fmt.Printf("[WS] ✓ 服务启动: wss://0.0.0.0%s/ws\n", addr)
		if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			fmt.Printf("[WS] 服务异常: %v\n", err)
		}
	}()

	return nil
}

// Stop 停止服务器
func (ws *WSServer) Stop() {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.running = false
	for _, conn := range ws.clients {
		conn.Close()
	}
	if ws.server != nil {
		ws.server.Close()
	}
}

// handleWS 处理 WebSocket 连接
func (ws *WSServer) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("[WS] 升级失败: %v\n", err)
		return
	}

	clientAddr := conn.RemoteAddr().String()
	fmt.Printf("[WS] 新连接: %s\n", clientAddr)

	var deviceID string

	defer func() {
		conn.Close()
		fmt.Printf("[WS] 连接断开: %s\n", clientAddr)
		if deviceID != "" {
			ws.srv.MarkDeviceOffline(deviceID)
		}
	}()

	// 设置读超时
	conn.SetReadDeadline(time.Now().Add(90 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(90 * time.Second))
		return nil
	})

	// 启动 ping 活跃检测
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			time.Sleep(30 * time.Second)
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		fmt.Printf("[WS] 收到消息: type=%s from=%s\n", msg.Type, clientAddr)

		resp := ws.processMessage(msg)

		// 记录设备ID
		if msg.Type == "register" && resp.Type == "ok" {
			var result map[string]interface{}
			json.Unmarshal(resp.Payload, &result)
			if id, ok := result["id"].(string); ok {
				deviceID = id
			}
		}

		respData, _ := json.Marshal(resp)
		conn.WriteMessage(websocket.TextMessage, respData)
	}
}

// processMessage 处理消息
func (ws *WSServer) processMessage(msg Message) Message {
	switch msg.Type {
	case "register":
		var req RegisterRequest
		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			return errorResponse("invalid_payload")
		}
		device, err := ws.srv.RegisterDevice(req)
		if err != nil {
			return errorResponse(err.Error())
		}
		return okResponse(device)

	case "heartbeat":
		var req struct {
			DeviceID string `json:"deviceId"`
		}
		json.Unmarshal(msg.Payload, &req)
		ws.srv.UpdateDeviceHeartbeat(req.DeviceID)
		return okResponse(nil)

	case "sync-pull":
		data := ws.srv.GetSyncData()
		return okResponse(data)

	case "sync-push":
		var data SyncData
		if err := json.Unmarshal(msg.Payload, &data); err != nil {
			return errorResponse("invalid_payload")
		}
		ws.srv.UpdateSyncData(data)
		return okResponse(nil)

	default:
		return errorResponse("unknown_type")
	}
}

func okResponse(payload interface{}) Message {
	data, _ := json.Marshal(payload)
	return Message{Type: "ok", Payload: data}
}

func errorResponse(err string) Message {
	data, _ := json.Marshal(map[string]string{"error": err})
	return Message{Type: "error", Payload: data}
}

// ==================== 证书生成 ====================

func generateCertIfNeeded(certFile, keyFile string) error {
	if _, err := os.Stat(certFile); err == nil {
		if _, err := os.Stat(keyFile); err == nil {
			return nil
		}
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"启SSH Cloud"},
			CommonName:   "启SSH Private Cloud",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return err
	}

	certOut, _ := os.Create(certFile)
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	certOut.Close()

	keyDER, _ := x509.MarshalECPrivateKey(key)
	keyOut, _ := os.Create(keyFile)
	pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	keyOut.Close()

	fmt.Printf("[WS] ✓ 证书已生成: %s\n", certFile)
	return nil
}
