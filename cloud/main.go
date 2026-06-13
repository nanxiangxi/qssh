package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"changeme/cloud/server"
	"changeme/cloud/server/crypto"
)

const Version = "0.2.0"

func main() {
	port := flag.Int("p", 9527, "监听端口")
	token := flag.String("t", "", "认证令牌（必填）")
	dataDir := flag.String("data", "", "数据目录")
	flag.Parse()

	if *token == "" {
		fmt.Println("错误: 请使用 -t 指定认证令牌")
		fmt.Println("用法: cloud-server -p 9527 -t your-secret-token")
		os.Exit(1)
	}

	dir := *dataDir
	if dir == "" {
		exePath, err := os.Executable()
		if err != nil {
			dir = "data"
		} else {
			dir = filepath.Join(filepath.Dir(exePath), "data")
		}
	}
	os.MkdirAll(dir, 0755)

	configPath := filepath.Join(dir, "config.json")
	cfg := loadConfig(configPath)
	cfg.Port = *port
	cfg.Token = *token

	srv := server.New(cfg, dir)

	// 启动 WebSocket TLS 服务
	wsSrv := server.NewWSServer(cfg, srv)
	if err := wsSrv.Start(dir); err != nil {
		log.Fatalf("[WS] 启动失败: %v", err)
	}

	fmt.Println("========================================")
	fmt.Printf("  启SSH 私有云端 v%s (WSS)\n", Version)
	fmt.Println("========================================")
	fmt.Printf("  端口: %d\n", cfg.Port)
	fmt.Printf("  令牌: %s\n", crypto.MaskToken(cfg.Token))
	fmt.Println("========================================")
	fmt.Println("  输入 help 查看可用命令")
	fmt.Println("========================================")

	// 信号退出
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// CLI 交互
	go cliLoop(srv, sigCh)

	<-sigCh
	fmt.Println("\n正在关闭...")
	wsSrv.Stop()
	saveConfig(configPath, cfg)
	fmt.Println("✓ 已关闭")
}

// ==================== CLI ====================

func cliLoop(srv *server.Server, sigCh chan os.Signal) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		parts := strings.Fields(input)
		cmd := parts[0]

		switch cmd {
		case "help", "h":
			printHelp()
		case "status", "s":
			printStatus(srv)
		case "devices", "d":
			printDevices(srv)
		case "sync":
			printSync(srv)
		case "clear":
			clearSync(srv)
		case "token":
			fmt.Printf("令牌: %s\n", srv.GetConfig().Token)
		case "quit", "q", "exit":
			sigCh <- syscall.SIGINT
			return
		default:
			fmt.Printf("未知命令: %s (输入 help 查看帮助)\n", cmd)
		}
	}
}

func printHelp() {
	fmt.Println("可用命令:")
	fmt.Println("  status, s    查看服务状态")
	fmt.Println("  devices, d   查看已连接设备")
	fmt.Println("  sync         查看同步数据")
	fmt.Println("  clear        清空同步数据")
	fmt.Println("  token        显示认证令牌")
	fmt.Println("  quit, q      退出")
}

func printStatus(srv *server.Server) {
	status := srv.GetStatus()
	fmt.Printf("状态:     运行中 (TLS)\n")
	fmt.Printf("端口:     %d\n", status.Port)
	fmt.Printf("在线设备: %d\n", status.DeviceCount)
	fmt.Printf("同步连接: %d\n", status.SyncCount)
	fmt.Printf("运行时间: %s\n", formatDuration(time.Since(status.StartedAt)))
}

func printDevices(srv *server.Server) {
	devices := srv.GetDevices()
	if len(devices) == 0 {
		fmt.Println("暂无设备连接")
		return
	}
	fmt.Printf("已连接设备 (%d):\n", len(devices))
	for _, d := range devices {
		status := "离线"
		if d.Status == "online" {
			status = "在线"
		}
		fmt.Printf("  %-20s %-15s %s\n", d.Name, d.Host, status)
	}
}

func printSync(srv *server.Server) {
	data := srv.GetSyncData()
	if data == nil || len(data.Connections) == 0 {
		fmt.Println("暂无同步数据")
		return
	}
	fmt.Printf("同步连接 (%d):\n", len(data.Connections))
	for _, c := range data.Connections {
		fmt.Printf("  %-20s %s@%s:%d\n", c.Name, c.Username, c.Host, c.Port)
	}
	if !data.UpdatedAt.IsZero() {
		fmt.Printf("最后更新: %s\n", data.UpdatedAt.Format("2006-01-02 15:04:05"))
	}
}

func clearSync(srv *server.Server) {
	srv.UpdateSyncData(server.SyncData{
		Connections: []server.SyncConnection{},
		UpdatedAt:   time.Now(),
	})
	fmt.Println("✓ 同步数据已清空")
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d秒", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%d分钟", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%d小时%d分钟", int(d.Hours()), int(d.Minutes())%60)
	}
	return fmt.Sprintf("%d天%d小时", int(d.Hours()/24), int(d.Hours())%24)
}

func loadConfig(path string) *server.Config {
	cfg := server.DefaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}
	json.Unmarshal(data, cfg)
	return cfg
}

func saveConfig(path string, cfg *server.Config) {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(path, data, 0644)
}
