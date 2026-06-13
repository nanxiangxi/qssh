package ssh

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// SSHService SSH服务结构体，供前端调用
type SSHService struct {
	mu            sync.RWMutex
	clients       map[string]*SSHClient  // 支持多个SSH连接
	storage       *StorageManager        // 连接存储管理器
	windowManager *WindowManager         // 窗口管理器
	groupManager  *GroupManager          // 分组管理器
	app           *application.App       // Wails应用实例
	shellSessions map[string]bool        // 跟踪哪些连接已启动Shell会话

	// 目录上传取消标志
	directoryUploadCancelMu sync.Mutex
	directoryUploadCancelled bool
}

// NewSSHService 创建SSH服务实例
func NewSSHService() *SSHService {
	return &SSHService{
		clients: make(map[string]*SSHClient),
		storage: NewStorageManager(),
		groupManager: NewGroupManager(),
		shellSessions: make(map[string]bool),
		// layoutService 已废弃，不再使用主题功能
	}
}

// SetApp 设置Wails应用实例
func (s *SSHService) SetApp(app *application.App) {
	s.app = app
}

// GetClient 获取指定连接的 SSHClient（供其他服务使用）
func (s *SSHService) GetClient(connID string) (*SSHClient, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	client, exists := s.clients[connID]
	if !exists {
		return nil, fmt.Errorf("连接不存在: %s", connID)
	}
	return client, nil
}

// GetApp 获取 Wails 应用实例（供其他服务使用）
func (s *SSHService) GetApp() *application.App {
	return s.app
}

// CancelDirectoryUpload 取消目录上传
func (s *SSHService) CancelDirectoryUpload() {
	s.directoryUploadCancelMu.Lock()
	defer s.directoryUploadCancelMu.Unlock()
	s.directoryUploadCancelled = true
	fmt.Println("[SSHService] ✓ 目录上传已标记为取消")
}

// ResetDirectoryUploadCancel 重置取消标志
func (s *SSHService) ResetDirectoryUploadCancel() {
	s.directoryUploadCancelMu.Lock()
	defer s.directoryUploadCancelMu.Unlock()
	s.directoryUploadCancelled = false
}

// IsDirectoryUploadCancelled 检查是否被取消
func (s *SSHService) IsDirectoryUploadCancelled() bool {
	s.directoryUploadCancelMu.Lock()
	defer s.directoryUploadCancelMu.Unlock()
	return s.directoryUploadCancelled
}

// SetWindowManager 设置窗口管理器
func (s *SSHService) SetWindowManager(wm *WindowManager) {
	s.windowManager = wm
}

// ClearWindowPositions 清除所有窗口位置记忆
func (s *SSHService) ClearWindowPositions() {
	if s.windowManager != nil {
		s.windowManager.ClearPositions()
	}
}

// formatSSHError 将SSH错误消息转换为友好的中文提示
func formatSSHError(err error) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	// 认证失败
	if strings.Contains(errMsg, "unable to authenticate") ||
		strings.Contains(errMsg, "authentication failed") {
		return &FriendlyError{Message: "认证失败：用户名或密码错误"}
	}

	// 连接超时
	if strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "i/o timeout") {
		return &FriendlyError{Message: "连接超时：请检查网络连接或服务器地址"}
	}

	// 连接拒绝
	if strings.Contains(errMsg, "connection refused") {
		return &FriendlyError{Message: "连接被拒绝：请检查服务器是否运行SSH服务"}
	}

	// 主机不可达
	if strings.Contains(errMsg, "no route to host") ||
		strings.Contains(errMsg, "network is unreachable") {
		return &FriendlyError{Message: "网络不可达：请检查网络连接"}
	}

	// 端口错误
	if strings.Contains(errMsg, "port") && strings.Contains(errMsg, "invalid") {
		return &FriendlyError{Message: "端口错误：请检查端口号是否正确"}
	}

	// 密钥文件问题
	if strings.Contains(errMsg, "key") || strings.Contains(errMsg, "private key") {
		return &FriendlyError{Message: "密钥错误：请检查密钥文件路径或格式"}
	}

	// 握手失败
	if strings.Contains(errMsg, "handshake failed") {
		return &FriendlyError{Message: "连接失败：无法与服务器建立安全连接"}
	}

	// 默认返回简化后的错误消息
	if idx := strings.Index(errMsg, ":"); idx > 0 {
		return &FriendlyError{Message: "连接失败：" + errMsg[:idx]}
	}

	return &FriendlyError{Message: "连接失败：" + errMsg}
}

// FriendlyError 友好的错误类型
type FriendlyError struct {
	Message string
}

func (e *FriendlyError) Error() string {
	return e.Message
}

// TestConnection 测试SSH连接（连接成功后立即断开，不保存）
func (s *SSHService) TestConnection(config *SSHConfig) error {
	// 创建临时客户端
	client := NewSSHClient(config)

	// 尝试连接
	if err := client.Connect(); err != nil {
		// 转换为友好的错误消息
		return formatSSHError(err)
	}

	// 连接成功，立即关闭
	client.Close()

	return nil
}

// Connect 连接到SSH服务器
func (s *SSHService) Connect(connID string, config *SSHConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client := NewSSHClient(config)
	if err := client.Connect(); err != nil {
		return formatSSHError(err)
	}

	s.clients[connID] = client
	return nil
}

// CreateAndConnect 创建新连接并建立SSH连接（后端生成ID）
// 流程：
// 1. 生成唯一 connID
// 2. 保存连接配置到本地存储
// 3. 建立 SSH 连接
// 4. 添加到默认分组
// 5. 返回 connID 供前端打开窗口
func (s *SSHService) CreateAndConnect(config *SSHConfig) (string, error) {
	fmt.Printf("[SSHService] CreateAndConnect 被调用: %+v\n", config)
	
	// 使用默认分组创建连接
	result, err := s.CreateAndConnectWithGroup(config, "group-1")
	if err != nil {
		return "", err
	}
	
	// 只返回 connID
	return result["connID"], nil
}

// CreateAndConnectWithGroup 创建连接并指定分组
func (s *SSHService) CreateAndConnectWithGroup(config *SSHConfig, groupID string) (map[string]string, error) {
	fmt.Printf("[SSHService] CreateAndConnectWithGroup 被调用: %+v, groupID=%s\n", config, groupID)
	
	// 检查config是否为nil
	if config == nil {
		return nil, &FriendlyError{Message: "连接配置不能为空"}
	}
	
	// 1. 生成唯一连接ID
	connID := generateConnectionID()
	fmt.Printf("[SSHService] 生成 connID: %s\n", connID)

	// 2. 构建连接信息
	connInfo := &ConnectionInfo{
		ID:         connID,
		Name:       config.Name,
		Host:       config.Host,
		Port:       config.Port,
		Username:   config.Username,
		Password:   config.Password,
		KeyPath:    config.KeyPath,
		PrivateKey: config.PrivateKey,
		Status:     "disconnected",
		Saved:      true,
		GroupID:    groupID,
	}

	// 3. 保存到本地存储
	fmt.Printf("[SSHService] 保存连接配置...\n")
	if err := s.storage.AddConnection(connInfo); err != nil {
		fmt.Printf("[SSHService] 保存失败: %v\n", err)
		return nil, &FriendlyError{Message: "保存连接配置失败: " + err.Error()}
	}
	fmt.Printf("[SSHService] 保存成功\n")

	// 4. 建立 SSH 连接（带超时控制）
	fmt.Printf("[SSHService] 开始建立 SSH 连接...\n")
	
	// 使用 channel 和 goroutine 实现超时控制
	type connectResult struct {
		client *SSHClient
		err    error
	}
	
	resultChan := make(chan connectResult, 1)
	
	go func() {
		client := NewSSHClient(config)
		// 设置断线回调
		client.SetDisconnectCallback(connID, s.onConnectionDisconnected)
		err := client.Connect()
		resultChan <- connectResult{client: client, err: err}
	}()
	
	// 设置超时时间（默认 30 秒，可配置）
	timeout := time.Duration(30) * time.Second
	if config.Timeout > 0 {
		timeout = time.Duration(config.Timeout) * time.Second
	}
	
	select {
	case result := <-resultChan:
		if result.err != nil {
			fmt.Printf("[SSHService] 连接失败: %v\n", result.err)
			// 连接失败，从缓存中删除（保留永久记录）
			s.storage.DeleteFromCache(connID)
			return nil, formatSSHError(result.err)
		}
		
		// 连接成功
		s.mu.Lock()
		s.clients[connID] = result.client
		s.mu.Unlock()
		fmt.Printf("[SSHService] SSH 连接成功\n")
		
	case <-time.After(timeout):
		fmt.Printf("[SSHService] ⏱️ 连接超时 (%v)\n", timeout)
		// 超时，从缓存中删除（保留永久记录）
		s.storage.DeleteFromCache(connID)
		return nil, &FriendlyError{Message: fmt.Sprintf("连接超时：服务器在 %v 内未响应，请检查网络连接", timeout)}
	}

	// 更新连接状态为已连接
	s.storage.UpdateConnectionStatus(connID, "connected")
	fmt.Printf("[SSHService] 更新状态为 connected\n")

	// 5. 添加到指定分组
	// 检查分组是否存在，如果不存在则重新创建
	group := s.groupManager.GetGroup(groupID)
	if group == nil {
		fmt.Printf("[SSHService] 分组 %s 不存在，重新创建\n", groupID)
		// 重新创建分组（使用连接名称作为分组名称）
		s.groupManager.CreateGroupWithName(groupID, config.Name)
		fmt.Printf("[SSHService] 已重新创建分组: %s\n", groupID)
	}
	
	if err := s.groupManager.AddConnectionToGroup(groupID, connID); err != nil {
		fmt.Printf("[SSHService] 添加到分组失败: %v\n", err)
		return nil, err
	}
	fmt.Printf("[SSHService] 已添加到分组: %s\n", groupID)

	// 6. 返回 connID 和 groupID
	result := map[string]string{
		"connID":  connID,
		"groupID": groupID,
	}
	fmt.Printf("[SSHService] 返回结果: %+v\n", result)
	return result, nil
}

// Disconnect 断开SSH连接（删除缓存记录，保留永久记录）
func (s *SSHService) Disconnect(connID string) error {
	fmt.Printf("[SSHService] ========== 开始断开连接: %s ==========\n", connID)

	s.mu.Lock()
	defer s.mu.Unlock()

	if client, exists := s.clients[connID]; exists {
		// 1. 关闭 SSH 连接
		client.Close()
		delete(s.clients, connID)
		fmt.Printf("[SSHService] ✓ SSH 连接已关闭\n")

		// 2. 清理 Shell 会话标记
		delete(s.shellSessions, connID)
		fmt.Printf("[SSHService] ✓ Shell 会话标记已清理\n")

		// 3. 从缓存中删除（保留永久记录）
		if err := s.storage.DeleteFromCache(connID); err != nil {
			fmt.Printf("[SSHService] ⚠️ 删除缓存记录失败: %v\n", err)
		} else {
			fmt.Printf("[SSHService] ✓ 缓存记录已删除\n")
		}

		// 4. 广播连接状态更新
		s.broadcastConnections()

		fmt.Printf("[SSHService] ========== 连接已断开: %s ==========\n", connID)
		return nil
	}

	fmt.Printf("[SSHService] ⚠️ 连接不存在: %s\n", connID)
	return nil
}

// GetServerKey 根据连接 ID 返回服务器标识（host:port），供 AI 历史存储使用
func (s *SSHService) GetServerKey(connID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if client, exists := s.clients[connID]; exists {
		return fmt.Sprintf("%s:%d", client.config.Host, client.config.Port)
	}
	return ""
}

// RunCommand 执行命令并返回输出字符串（供 AI 工具调用）
func (s *SSHService) RunCommand(connID, command string) (string, error) {
	result, err := s.ExecuteCommand(connID, command)
	if err != nil {
		return "", err
	}
	if !result.Success {
		return fmt.Sprintf("命令执行失败 (exit code %d):\n%s\n%s", result.ExitCode, result.Stdout, result.Stderr), nil
	}
	output := result.Stdout
	if result.Stderr != "" {
		output += "\n[stderr] " + result.Stderr
	}
	return output, nil
}

// ExecuteCommand 执行远程命令
func (s *SSHService) ExecuteCommand(connID string, command string) (*CommandResult, error) {
	fmt.Printf("[SSHService] ========== ExecuteCommand 被调用 ==========\n")
	fmt.Printf("[SSHService] connID: %s\n", connID)
	fmt.Printf("[SSHService] command: %s\n", command)
	
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.clients[connID]
	if !exists {
		fmt.Printf("[SSHService] ⚠️ 连接不存在: %s\n", connID)
		return nil, fmt.Errorf("连接不存在: %s", connID)
	}

	fmt.Printf("[SSHService] 开始执行命令...\n")
	result, err := client.ExecuteCommand(command)
	if err != nil {
		fmt.Printf("[SSHService] ⚠️ 命令执行失败: %v\n", err)
	} else if result != nil {
		fmt.Printf("[SSHService] ✓ 命令执行成功, Success=%v\n", result.Success)
		if !result.Success {
			fmt.Printf("[SSHService] Stderr: %s\n", result.Stderr)
		}
	}
	
	return result, err
}

// UploadFile 上传文件（从字节数组）
func (s *SSHService) UploadFile(connID string, remotePath string, data string) error {
	s.mu.RLock()
	client, exists := s.clients[connID]
	app := s.app
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("连接不存在: %s", connID)
	}

	decodedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return fmt.Errorf("base64 解码失败: %v", err)
	}

	// 进度回调：通过事件通知前端
	progressCallback := func(percent int) {
		if app != nil {
			fmt.Printf("[SSHService] 发送上传进度事件: %d%%\n", percent)
			app.Event.Emit("ssh:upload-progress", map[string]interface{}{
				"connId":   connID,
				"path":     remotePath,
				"progress": percent,
			})
		} else {
			fmt.Printf("[SSHService] app 为 nil，无法发送进度事件\n")
		}
	}

	return client.UploadFileFromBytes(remotePath, decodedData, progressCallback)
}

// DownloadFile 下载文件（返回字节数组）
func (s *SSHService) DownloadFile(connID string, remotePath string) ([]byte, error) {
	fmt.Printf("[SSHService] ========== DownloadFile 被调用 ==========\n")
	fmt.Printf("[SSHService] connID: %s\n", connID)
	fmt.Printf("[SSHService] remotePath: %s\n", remotePath)
	
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.clients[connID]
	if !exists {
		fmt.Printf("[SSHService] ⚠️ 连接不存在: %s\n", connID)
		return nil, fmt.Errorf("连接不存在: %s", connID)
	}

	fmt.Printf("[SSHService] 开始下载文件...\n")
	data, err := client.DownloadFileToBytes(remotePath)
	if err != nil {
		fmt.Printf("[SSHService] ⚠️ 下载失败: %v\n", err)
	} else {
		fmt.Printf("[SSHService] ✓ 下载成功，文件大小: %d bytes\n", len(data))
	}
	
	return data, err
}

// ListDirectory 列出目录
func (s *SSHService) ListDirectory(connID string, remotePath string) ([]FileInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.clients[connID]
	if !exists {
		return nil, nil // 连接不存在
	}

	return client.ListDirectory(remotePath)
}

// DeleteFile 删除文件
func (s *SSHService) DeleteFile(connID string, remotePath string) error {
	fmt.Printf("[SSHService] ========== DeleteFile 被调用 ==========\n")
	fmt.Printf("[SSHService] connID: %s\n", connID)
	fmt.Printf("[SSHService] remotePath: %s\n", remotePath)
	
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.clients[connID]
	if !exists {
		fmt.Printf("[SSHService] ⚠️ 连接不存在: %s\n", connID)
		return fmt.Errorf("连接不存在: %s", connID)
	}

	fmt.Printf("[SSHService] 开始删除文件...\n")
	err := client.DeleteFile(remotePath)
	if err != nil {
		fmt.Printf("[SSHService] ⚠️ 删除失败: %v\n", err)
	} else {
		fmt.Printf("[SSHService] ✓ 删除成功\n")
	}
	
	return err
}

// CreateDirectory 创建目录
func (s *SSHService) CreateDirectory(connID string, remotePath string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.clients[connID]
	if !exists {
		return fmt.Errorf("连接不存在")
	}

	return client.CreateDirectory(remotePath)
}

// RenameFile 重命名文件/目录
func (s *SSHService) RenameFile(connID string, oldPath, newPath string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.clients[connID]
	if !exists {
		return fmt.Errorf("连接不存在")
	}

	return client.RenameFile(oldPath, newPath)
}

// UploadDirectory 上传目录
func (s *SSHService) UploadDirectory(connID string, localPath string, remotePath string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.clients[connID]
	if !exists {
		return fmt.Errorf("连接不存在")
	}

	return client.UploadDirectory(localPath, remotePath)
}

// CreateArchive 创建压缩包并返回临时文件路径
func (s *SSHService) CreateArchive(connID string, files []string, archiveName string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.clients[connID]
	if !exists {
		return "", fmt.Errorf("连接不存在")
	}

	return client.CreateArchive(files, archiveName)
}

// DeleteTempFile 删除临时文件
func (s *SSHService) DeleteTempFile(connID string, filePath string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.clients[connID]
	if !exists {
		return fmt.Errorf("连接不存在")
	}

	return client.DeleteTempFile(filePath)
}

// ExtractArchive 解压压缩包到指定目录
func (s *SSHService) ExtractArchive(connID string, archivePath string, targetDir string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.clients[connID]
	if !exists {
		return fmt.Errorf("连接不存在")
	}

	return client.ExtractArchive(archivePath, targetDir)
}

// SelectLocalDirectoryAndUpload 选择本地目录并上传到远程服务器
func (s *SSHService) SelectLocalDirectoryAndUpload(connID string, remotePath string) error {
	if s.app == nil {
		return fmt.Errorf("应用实例未初始化")
	}

	// 重置取消标志
	s.ResetDirectoryUploadCancel()

	// 1. 使用原生对话框选择本地目录
	localDir, err := s.app.Dialog.OpenFile().
		SetTitle("选择要上传的目录").
		CanChooseDirectories(true).
		CanChooseFiles(false).
		PromptForSingleSelection()

	if err != nil || localDir == "" {
		return fmt.Errorf("用户取消选择或发生错误")
	}

	fmt.Printf("[SSHService] 选择的本地目录: %s\n", localDir)

	// 发送进度事件：开始压缩
	if s.app != nil {
		fmt.Println("[SSHService] 发送进度事件: compressing")
		s.app.Event.Emit("directory-upload-progress", map[string]interface{}{
			"stage":    "compressing",
			"progress": 10,
			"message":  "正在压缩目录...",
		})
	}

	// 2. 在本地创建临时 tar.gz 文件（使用 Go 原生 tar.gz）
	tempArchive := filepath.Join(os.TempDir(), fmt.Sprintf("directory_upload_%d.tar.gz", time.Now().Unix()))
	
	err = s.createTarGzFromDirectory(localDir, tempArchive)
	if err != nil {
		return fmt.Errorf("压缩本地目录失败: %v", err)
	}

	fmt.Printf("[SSHService] 本地压缩完成: %s\n", tempArchive)

	// 获取压缩文件大小
	archiveInfo, err := os.Stat(tempArchive)
	if err == nil {
		fmt.Printf("[SSHService] 压缩包大小: %d bytes (%.2f MB)\n", archiveInfo.Size(), float64(archiveInfo.Size())/(1024*1024))
	}

	// 发送进度事件：压缩完成，开始上传
	if s.app != nil {
		s.app.Event.Emit("directory-upload-progress", map[string]interface{}{
			"stage":    "uploading",
			"progress": 40,
			"message":  "正在上传压缩包...",
			"fileSize": archiveInfo.Size(), // 发送文件大小
		})
	}

	// 3. 读取压缩文件内容
	archiveData, err := os.ReadFile(tempArchive)
	if err != nil {
		os.Remove(tempArchive)
		return fmt.Errorf("读取压缩文件失败: %v", err)
	}

	// 4. 删除本地临时文件
	os.Remove(tempArchive)

	// 5. 上传到远程服务器
	s.mu.RLock()
	client, exists := s.clients[connID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("连接不存在")
	}

	// 6. 分块上传 tar.gz 文件到远程服务器（支持进度和取消）
	remoteArchivePath := fmt.Sprintf("%s/temp_upload_%d.tar.gz", remotePath, time.Now().Unix())
	err = s.uploadFileWithProgress(client, remoteArchivePath, archiveData, s.app)
	if err != nil {
		// 检查是否是被取消
		if s.IsDirectoryUploadCancelled() {
			fmt.Println("[SSHService] 目录上传已被用户取消")
			return fmt.Errorf("用户取消了上传")
		}
		return fmt.Errorf("上传压缩文件失败: %v", err)
	}

	// 发送进度事件：上传完成，开始解压
	if s.app != nil {
		s.app.Event.Emit("directory-upload-progress", map[string]interface{}{
			"stage":    "extracting",
			"progress": 80,
			"message":  "正在解压...",
		})
	}

	// 7. 在远程服务器上解压
	err = client.ExtractArchive(remoteArchivePath, remotePath)
	if err != nil {
		return fmt.Errorf("解压失败: %v", err)
	}

	// 8. 删除远程临时文件
	err = client.DeleteTempFile(remoteArchivePath)
	if err != nil {
		fmt.Printf("[SSHService] 警告: 删除临时文件失败: %v\n", err)
	}

	fmt.Printf("[SSHService] ✓ 目录上传成功\n")
	return nil
}

// uploadFileWithProgress 分块上传文件，支持进度更新和取消检查
func (s *SSHService) uploadFileWithProgress(client *SSHClient, remotePath string, data []byte, app *application.App) error {
	if client.sftpClient == nil {
		if err := client.InitSFTP(); err != nil {
			return err
		}
	}

	// 创建远程文件
	fmt.Printf("[SSHService] 创建远程文件: %s\n", remotePath)
	remoteFile, err := client.sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("创建远程文件失败: %v", err)
	}
	defer remoteFile.Close()

	// 分块上传（每块 1MB）
	const chunkSize = 1024 * 1024 // 1MB
	totalSize := len(data)
	uploadedSize := 0

	fmt.Printf("[SSHService] 开始分块上传，总大小: %d bytes, 块大小: %d bytes\n", totalSize, chunkSize)

	for uploadedSize < totalSize {
		// 检查是否被取消
		if s.IsDirectoryUploadCancelled() {
			fmt.Println("[SSHService] 检测到取消请求，停止上传")
			remoteFile.Close()
			// 删除已上传的部分文件
			client.DeleteTempFile(remotePath)
			return fmt.Errorf("用户取消了上传")
		}
		
		// 计算当前块的大小
		end := uploadedSize + chunkSize
		if end > totalSize {
			end = totalSize
		}
		chunk := data[uploadedSize:end]

		// 写入当前块
		written, err := remoteFile.Write(chunk)
		if err != nil {
			// 上传失败，删除临时文件
			client.DeleteTempFile(remotePath)
			return fmt.Errorf("写入数据失败: %v", err)
		}

		uploadedSize += written

		// 发送进度事件
		if app != nil {
			progress := 40 + int(float64(uploadedSize)/float64(totalSize)*40) // 40%-80%
			app.Event.Emit("directory-upload-progress", map[string]interface{}{
				"stage":    "uploading",
				"progress": progress,
				"message":  fmt.Sprintf("正在上传... (%.1f MB / %.1f MB)", 
					float64(uploadedSize)/(1024*1024), 
					float64(totalSize)/(1024*1024)),
			})
		}

		fmt.Printf("[SSHService] 上传进度: %d/%d bytes (%.1f%%)\n", 
			uploadedSize, totalSize, float64(uploadedSize)/float64(totalSize)*100)
	}

	fmt.Printf("[SSHService] ✓ 上传完成: %d bytes\n", uploadedSize)
	return nil
}

// createTarGzFromDirectory 将目录压缩为 tar.gz 文件（跨平台）
func (s *SSHService) createTarGzFromDirectory(sourceDir string, destArchive string) error {
	// 创建 tar.gz 文件
	file, err := os.Create(destArchive)
	if err != nil {
		return err
	}
	defer file.Close()

	// 创建 gzip writer
	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	// 创建 tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// 获取源目录的基本名称（作为顶层目录）
	baseName := filepath.Base(sourceDir)
	fmt.Printf("[SSHService] 顶层目录名: %s\n", baseName)

	// 遍历目录
	fileCount := 0
	dirCount := 0
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 获取相对路径
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// 跳过根目录本身
		if relPath == "." {
			return nil
		}

		// 添加顶层目录名前缀（保留目录结构）
		relPath = baseName + "/" + relPath

		// 转换为 Unix 风格路径
		relPath = filepath.ToSlash(relPath)

		// 创建 tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		// 写入 header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// 统计
		if info.IsDir() {
			dirCount++
			fmt.Printf("[SSHService] 添加目录: %s\n", relPath)
		} else {
			fileCount++
			// 如果是文件，写入内容
			data, err := os.Open(path)
			if err != nil {
				return err
			}
			defer data.Close()

			if _, err := io.Copy(tarWriter, data); err != nil {
				return err
			}
		}

		return nil
	})

	fmt.Printf("[SSHService] 压缩完成: %d 个文件, %d 个目录\n", fileCount, dirCount)
	return err
}

// IsConnected 检查连接状态
func (s *SSHService) IsConnected(connID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.clients[connID]
	if !exists {
		return false
	}

	return client.IsConnected()
}

// GetConnectionList 获取所有连接ID列表
func (s *SSHService) GetConnectionList() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var ids []string
	for id := range s.clients {
		ids = append(ids, id)
	}
	return ids
}

// ========== 连接存储管理 API ==========

// AddConnection 添加新连接（保存到本地）
func (s *SSHService) AddConnection(conn *ConnectionInfo) error {
	return s.storage.AddConnection(conn)
}

// UpdateConnection 更新连接信息
func (s *SSHService) UpdateConnection(conn *ConnectionInfo) error {
	return s.storage.UpdateConnection(conn)
}

// SyncImportConnection 云端同步导入（按 host:port 去重：存在则更新，不存在则新增，永久保存）
func (s *SSHService) SyncImportConnection(conn *ConnectionInfo) error {
	err := s.storage.SyncImportConnection(conn)
	if err == nil {
		s.broadcastConnections()
	}
	return err
}

// DeleteConnection 删除连接
func (s *SSHService) DeleteConnection(id string) error {
	return s.storage.DeleteConnection(id)
}

// GetConnection 获取单个连接
func (s *SSHService) GetConnection(id string) (*ConnectionInfo, error) {
	return s.storage.GetConnection(id)
}

// GetAllConnections 获取所有保存的连接（包含分组ID和真实在线状态）
func (s *SSHService) GetAllConnections() []*ConnectionInfo {
	connections := s.storage.GetAllConnections()
	
	// 为每个连接填充GroupID和真实在线状态
	for _, conn := range connections {
		// 从分组管理器获取GroupID
		group := s.groupManager.GetGroupByConnID(conn.ID)
		if group != nil {
			conn.GroupID = group.ID
		}
		
		// 根据 s.clients 判断真实在线状态
		s.mu.RLock()
		_, isOnline := s.clients[conn.ID]
		s.mu.RUnlock()
		
		if isOnline {
			conn.Status = "connected"
		} else {
			conn.Status = "disconnected"
		}
	}
	
	return connections
}

// SaveConnection 标记连接为已保存
func (s *SSHService) SaveConnection(id string) error {
	err := s.storage.MarkAsSaved(id)
	if err != nil {
		return err
	}
	// 立即广播更新，让前端侧栏马上刷新
	s.broadcastConnections()
	return nil
}

// broadcastConnections 立即广播所有连接状态
func (s *SSHService) broadcastConnections() {
	if s.app == nil {
		return
	}
	allConnections := s.GetAllConnections()
	if allConnections == nil {
		allConnections = []*ConnectionInfo{}
	}
	s.app.Event.Emit("ssh:connections-updated", map[string]interface{}{
		"connections": allConnections,
		"timestamp":   time.Now().UnixMilli(),
	})
}

// UnsaveConnection 标记连接为未保存
func (s *SSHService) UnsaveConnection(id string) {
	s.storage.MarkAsUnsaved(id)
}

// ========== 分组管理 API ==========

// CreateGroup 创建新分组
func (s *SSHService) CreateGroup(name string) string {
	group := s.groupManager.CreateGroup(name)
	return group.ID
}

// GetAllGroups 获取所有分组
func (s *SSHService) GetAllGroups() []*SSHGroup {
	return s.groupManager.GetAllGroups()
}

// GetGroupConnections 获取指定分组的所有连接
func (s *SSHService) GetGroupConnections(groupID string) []string {
	group := s.groupManager.GetGroup(groupID)
	if group == nil {
		return []string{}
	}
	return group.ConnIDs
}

// GetGroupConnectionInfos 获取指定分组的完整连接信息
func (s *SSHService) GetGroupConnectionInfos(groupID string) []*ConnectionInfo {
	group := s.groupManager.GetGroup(groupID)
	if group == nil {
		return []*ConnectionInfo{}
	}
	
	var infos []*ConnectionInfo
	for _, connID := range group.ConnIDs {
		conn, err := s.storage.GetConnection(connID)
		if err == nil && conn != nil {
			infos = append(infos, conn)
		}
	}
	
	return infos
}

// AddToGroup 将连接添加到分组
func (s *SSHService) AddToGroup(groupID, connID string) error {
	return s.groupManager.AddConnectionToGroup(groupID, connID)
}

// RemoveFromGroup 从分组中移除连接
func (s *SSHService) RemoveFromGroup(groupID, connID string) error {
	return s.groupManager.RemoveConnectionFromGroup(groupID, connID)
}

// GetDefaultGroupID 获取默认分组ID
func (s *SSHService) GetDefaultGroupID() string {
	return "group-1"
}

// CloseGroup 关闭分组（断开所有连接并删除分组）
func (s *SSHService) CloseGroup(groupID string) error {
	fmt.Printf("[SSHService] ========== 开始关闭分组: %s ==========\n", groupID)
	
	// 1. 获取分组
	group := s.groupManager.GetGroup(groupID)
	if group == nil {
		fmt.Printf("[SSHService] 分组不存在，可能已被删除: %s\n", groupID)
		return nil // 分组不存在，直接返回成功（幂等性）
	}
	
	// 2. 直接关闭所有连接（不从分组中移除，避免触发自动删除逻辑）
	connIDs := make([]string, len(group.ConnIDs))
	copy(connIDs, group.ConnIDs)
	
	for _, connID := range connIDs {
		fmt.Printf("[SSHService] 关闭连接: %s\n", connID)
		
		s.mu.Lock()
		if client, exists := s.clients[connID]; exists {
			client.Close()
			delete(s.clients, connID)
			fmt.Printf("[SSHService] SSH 连接已关闭\n")
			
			// 从缓存中删除（保留永久记录）
			if err := s.storage.DeleteFromCache(connID); err != nil {
				fmt.Printf("[SSHService] ⚠️ 删除缓存记录失败: %v\n", err)
			} else {
				fmt.Printf("[SSHService] ✓ 缓存记录已删除\n")
			}
		}
		s.mu.Unlock()
	}
	
	// 3. 清空分组中的连接列表（但不删除分组）
	s.groupManager.ClearGroup(groupID)
	
	// 4. 广播分组更新事件（通知前端刷新）
	fmt.Printf("[SSHService] 📡 准备广播分组更新事件, s.app=%v\n", s.app != nil)
	if s.app != nil {
		eventData := map[string]interface{}{
			"groupID":     groupID,
			"connections": []string{}, // 空列表，表示所有连接已关闭
		}
		s.app.Event.Emit("ssh:group-updated", eventData)
		fmt.Printf("[SSHService] ✓ 已广播分组更新事件\n")
	} else {
		fmt.Printf("[SSHService] ⚠️⚠️⚠️ s.app 为 nil，无法广播事件！\n")
	}
	
	// 5. 删除分组
	if err := s.groupManager.DeleteGroup(groupID); err != nil {
		fmt.Printf("[SSHService] ⚠️ 删除分组失败: %v\n", err)
		// 不返回错误，因为连接已经关闭
	}
	
	fmt.Printf("[SSHService] ========== 分组已关闭: %s ==========\n", groupID)
	return nil
}

// OpenSSHWindow 打开SSH分组窗口（如果已存在则聚焦，否则创建新窗口）
func (s *SSHService) OpenSSHWindow(groupID string, groupName string, activeConnID string) error {
	if s.windowManager == nil {
		return &FriendlyError{Message: "窗口管理器未初始化"}
	}
	return s.windowManager.CreateSSHWindow(groupID, groupName, activeConnID)
}

// GetGroupByConnID 根据连接ID获取所属分组
func (s *SSHService) GetGroupByConnID(connID string) *SSHGroup {
	fmt.Printf("[SSHService] GetGroupByConnID 被调用: connID=%s\n", connID)
	group := s.groupManager.GetGroupByConnID(connID)
	if group != nil {
		fmt.Printf("[SSHService] 找到分组: ID=%s, Name=%s\n", group.ID, group.Name)
	} else {
		fmt.Printf("[SSHService] 未找到分组\n")
	}
	return group
}

// StartShellSession 启动Shell会话（兼容旧调用，默认 sessionID = "default"）
func (s *SSHService) StartShellSession(connID string) error {
	return s.StartShellSessionWithID(connID, "default")
}

// StartShellSessionWithID 启动指定 ID 的 Shell 会话
func (s *SSHService) StartShellSessionWithID(connID, sessionID string) error {
	fmt.Printf("[SSHService] StartShellSession: connID=%s, sessionID=%s\n", connID, sessionID)

	s.mu.RLock()
	client, exists := s.clients[connID]
	s.mu.RUnlock()
	if !exists {
		return fmt.Errorf("连接不存在: %s", connID)
	}

	_, err := client.CreateShell(sessionID)
	if err != nil {
		return err
	}

	// 标记该连接已有 Shell 会话
	s.mu.Lock()
	s.shellSessions[connID] = true
	s.mu.Unlock()

	fmt.Printf("[SSHService] ✓ Shell会话已启动: %s/%s\n", connID, sessionID)
	return nil
}

// WriteToTerminal 向终端写入数据（兼容旧调用）
func (s *SSHService) WriteToTerminal(connID string, data string) error {
	return s.WriteToTerminalByID(connID, "default", data)
}

// WriteToTerminalByID 向指定 session 写入数据
func (s *SSHService) WriteToTerminalByID(connID, sessionID string, data string) error {
	s.mu.RLock()
	client, exists := s.clients[connID]
	s.mu.RUnlock()
	if !exists {
		return fmt.Errorf("连接不存在: %s", connID)
	}
	// 检查 session 是否存在
	if !client.IsShellActive(sessionID) {
		available := client.GetSessionIDs()
		return fmt.Errorf("Shell会话未初始化: %s (可用: %v)", sessionID, available)
	}
	return client.WriteToShell(sessionID, []byte(data))
}

// ReadFromShellSession 从指定 session 读取数据
func (s *SSHService) ReadFromShellSession(connID, sessionID string, buf []byte) (int, error) {
	s.mu.RLock()
	client, exists := s.clients[connID]
	s.mu.RUnlock()
	if !exists {
		return 0, fmt.Errorf("连接不存在")
	}
	return client.ReadFromShell(sessionID, buf)
}

// IsShellSessionActive 检查指定 session 是否活跃
func (s *SSHService) IsShellSessionActive(connID, sessionID string) bool {
	s.mu.RLock()
	client, exists := s.clients[connID]
	s.mu.RUnlock()
	if !exists {
		return false
	}
	active := client.IsShellActive(sessionID)
	if !active {
		available := client.GetSessionIDs()
		fmt.Printf("[SSHService] Session %s 不活跃, 可用: %v\n", sessionID, available)
	}
	return active
}

// GetSessionIDs 获取连接的所有活跃 session ID
func (s *SSHService) GetSessionIDs(connID string) []string {
	s.mu.RLock()
	client, exists := s.clients[connID]
	s.mu.RUnlock()
	if !exists {
		return nil
	}
	return client.GetSessionIDs()
}

// ResizeTerminal 调整终端大小（兼容旧调用）
func (s *SSHService) ResizeTerminal(connID string, cols, rows int) error {
	return s.ResizeTerminalByID(connID, "default", cols, rows)
}

// ResizeTerminalByID 调整指定 session 的终端大小
func (s *SSHService) ResizeTerminalByID(connID, sessionID string, cols, rows int) error {
	s.mu.RLock()
	client, exists := s.clients[connID]
	s.mu.RUnlock()
	if !exists {
		return fmt.Errorf("连接不存在")
	}
	return client.ResizeTerminalByID(sessionID, cols, rows)
}

// IsShellActive 检查Shell会话是否活跃（兼容旧调用，默认 session）
func (s *SSHService) IsShellActive(connID string) bool {
	return s.IsShellSessionActive(connID, "default")
}

// CloseShellSession 关闭Shell会话（兼容旧调用，默认 session）
func (s *SSHService) CloseShellSession(connID string) error {
	return s.CloseShellSessionByID(connID, "default")
}

// CloseShellSessionByID 关闭指定 session
func (s *SSHService) CloseShellSessionByID(connID, sessionID string) error {
	fmt.Printf("[SSHService] 关闭Shell会话: connID=%s, sessionID=%s\n", connID, sessionID)
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	client, exists := s.clients[connID]
	if !exists {
		fmt.Printf("[SSHService] ⚠️ 连接不存在: %s\n", connID)
		return fmt.Errorf("连接不存在")
	}
	
	// 关闭指定 session
	client.CloseShell(sessionID)
	fmt.Printf("[SSHService] ✓ Shell会话已关闭: %s/%s\n", connID, sessionID)
	return nil
}

// ReadFromShell 从Shell读取数据（兼容旧调用，默认 session）
func (s *SSHService) ReadFromShell(connID string, buf []byte) (int, error) {
	return s.ReadFromShellSession(connID, "default", buf)
}

// generateConnectionID 生成唯一的连接ID
var connIDCounter int64

func generateConnectionID() string {
	connIDCounter++
	return fmt.Sprintf("conn_%d_%d", time.Now().UnixNano(), connIDCounter)
}

// GetLatency 获取连接延迟（实时测量）
func (s *SSHService) GetLatency(connID string) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.clients[connID]
	if !exists {
		return 0
	}

	// 实时更新延迟
	return client.UpdateLatency()
}

// GetSystemStats 获取系统资源统计
func (s *SSHService) GetSystemStats(connID string) (*SystemStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.clients[connID]
	if !exists {
		return nil, fmt.Errorf("连接不存在")
	}

	ctx := context.Background()
	return client.GetSystemStats(ctx)
}

// GetProcessList 获取进程列表
func (s *SSHService) GetProcessList(connID string) ([]ProcessInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.clients[connID]
	if !exists {
		return nil, fmt.Errorf("连接不存在")
	}

	ctx := context.Background()
	return client.GetProcessList(ctx)
}

// KillProcess 终止进程
func (s *SSHService) KillProcess(connID string, pid int32) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.clients[connID]
	if !exists {
		return fmt.Errorf("连接不存在")
	}

	ctx := context.Background()
	return client.KillProcess(ctx, pid)
}

// SendSignal 向进程发送信号
func (s *SSHService) SendSignal(connID string, pid int32, signal string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.clients[connID]
	if !exists {
		return fmt.Errorf("连接不存在")
	}

	ctx := context.Background()
	return client.SendSignal(ctx, pid, signal)
}

// GetProcessDetail 获取进程详细信息
func (s *SSHService) GetProcessDetail(connID string, pid int32) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.clients[connID]
	if !exists {
		return "", fmt.Errorf("连接不存在")
	}

	ctx := context.Background()
	return client.GetProcessDetail(ctx, pid)
}

// ListFiles 列出目录内容
func (s *SSHService) ListFiles(connID string, path string) ([]FileInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.clients[connID]
	if !exists {
		return nil, fmt.Errorf("连接不存在")
	}

	return client.ListDirectory(path)
}

// SearchFiles 递归搜索文件（支持子目录，流式返回）
func (s *SSHService) SearchFiles(connID string, basePath string, keyword string, searchID string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.clients[connID]
	if !exists {
		return fmt.Errorf("连接不存在")
	}

	return client.SearchFiles(basePath, keyword, s.app, searchID)
}

// CancelSearch 取消正在进行的搜索
func (s *SSHService) CancelSearch(connID string, searchID string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.clients[connID]
	if !exists {
		return fmt.Errorf("连接不存在")
	}

	client.CancelSearch(searchID)
	return nil
}

// ========== 断线重连功能 ==========

// reconnectConfigs 保存已断开连接的配置（用于重连）
var reconnectConfigs = struct {
	mu      sync.RWMutex
	configs map[string]*SSHConfig
}{configs: make(map[string]*SSHConfig)}

// onConnectionDisconnected 连接断开回调
func (s *SSHService) onConnectionDisconnected(connID string) {
	fmt.Printf("[SSHService] 📡 连接断开回调: %s\n", connID)

	// 保存配置用于重连
	s.mu.RLock()
	client, exists := s.clients[connID]
	if !exists {
		s.mu.RUnlock()
		return
	}
	config := client.config
	s.mu.RUnlock()

	// 保存配置
	reconnectConfigs.mu.Lock()
	reconnectConfigs.configs[connID] = config
	reconnectConfigs.mu.Unlock()

	// 更新存储状态
	s.storage.UpdateConnectionStatus(connID, "disconnected")

	// 广播断线事件
	if s.app != nil {
		s.app.Event.Emit("ssh:connection-disconnected", map[string]interface{}{
			"connID":    connID,
			"host":      config.Host,
			"timestamp": time.Now().UnixMilli(),
		})
	}

	// 广播连接状态更新
	s.broadcastConnections()
}

// Reconnect 重新连接（前端调用）
func (s *SSHService) Reconnect(connID string) error {
	fmt.Printf("[SSHService] 🔄 尝试重连: %s\n", connID)

	// 1. 获取配置
	reconnectConfigs.mu.RLock()
	config, exists := reconnectConfigs.configs[connID]
	reconnectConfigs.mu.RUnlock()

	if !exists {
		// 尝试从存储获取配置
		connInfo, err := s.storage.GetConnection(connID)
		if err != nil || connInfo == nil {
			return &FriendlyError{Message: "无法找到连接配置，请重新连接"}
		}
		config = &SSHConfig{
			Name:       connInfo.Name,
			Host:       connInfo.Host,
			Port:       connInfo.Port,
			Username:   connInfo.Username,
			Password:   connInfo.Password,
			KeyPath:    connInfo.KeyPath,
			PrivateKey: connInfo.PrivateKey,
		}
	}

	// 2. 清理旧连接
	s.mu.Lock()
	if oldClient, exists := s.clients[connID]; exists {
		oldClient.Close()
		delete(s.clients, connID)
	}
	delete(s.shellSessions, connID)
	s.mu.Unlock()

	// 3. 创建新连接
	client := NewSSHClient(config)
	client.SetDisconnectCallback(connID, s.onConnectionDisconnected)

	// 广播重连中状态
	if s.app != nil {
		s.app.Event.Emit("ssh:connection-reconnecting", map[string]interface{}{
			"connID": connID,
		})
	}

	// 4. 尝试连接
	if err := client.Connect(); err != nil {
		// 连接失败
		if s.app != nil {
			s.app.Event.Emit("ssh:connection-reconnect-failed", map[string]interface{}{
				"connID": connID,
				"error":  formatSSHError(err).Error(),
			})
		}
		return formatSSHError(err)
	}

	// 5. 连接成功
	s.mu.Lock()
	s.clients[connID] = client
	s.mu.Unlock()

	// 更新状态
	s.storage.UpdateConnectionStatus(connID, "connected")

	// 清理重连配置
	reconnectConfigs.mu.Lock()
	delete(reconnectConfigs.configs, connID)
	reconnectConfigs.mu.Unlock()

	// 广播重连成功
	if s.app != nil {
		s.app.Event.Emit("ssh:connection-reconnected", map[string]interface{}{
			"connID": connID,
		})
	}

	// 广播连接状态更新
	s.broadcastConnections()

	fmt.Printf("[SSHService] ✅ 重连成功: %s\n", connID)
	return nil
}

// AutoReconnect 自动重连（后台执行，不阻塞前端）
func (s *SSHService) AutoReconnect(connID string) {
	go func() {
		fmt.Printf("[SSHService] 🔄 开始自动重连: %s\n", connID)

		maxRetries := 3
		retryDelay := 2 * time.Second

		for i := 0; i < maxRetries; i++ {
			if i > 0 {
				fmt.Printf("[SSHService] ⏳ 等待 %v 后重试 (%d/%d)...\n", retryDelay, i+1, maxRetries)
				time.Sleep(retryDelay)
				retryDelay *= 2 // 指数退避
			}

			err := s.Reconnect(connID)
			if err == nil {
				fmt.Printf("[SSHService] ✅ 自动重连成功: %s\n", connID)
				return
			}

			fmt.Printf("[SSHService] ⚠️ 自动重连失败 (%d/%d): %v\n", i+1, maxRetries, err)
		}

		fmt.Printf("[SSHService] ❌ 自动重连失败，已达到最大重试次数: %s\n", connID)

		// 通知前端自动重连失败
		if s.app != nil {
			s.app.Event.Emit("ssh:connection-reconnect-failed", map[string]interface{}{
				"connID":  connID,
				"error":   "自动重连失败，请手动重连",
				"autoMax": true,
			})
		}
	}()
}

// CheckConnectionHealth 检查连接健康状态（心跳检测）
func (s *SSHService) CheckConnectionHealth(connID string) bool {
	s.mu.RLock()
	client, exists := s.clients[connID]
	s.mu.RUnlock()

	if !exists {
		return false
	}

	return client.IsConnected()
}

// StartHealthCheck 启动健康检查定时器
func (s *SSHService) StartHealthCheck() {
	go func() {
		ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
		defer ticker.Stop()

		for range ticker.C {
			s.mu.RLock()
			connIDs := make([]string, 0, len(s.clients))
			for id := range s.clients {
				connIDs = append(connIDs, id)
			}
			s.mu.RUnlock()

			for _, connID := range connIDs {
				if !s.CheckConnectionHealth(connID) {
					fmt.Printf("[HealthCheck] ⚠️ 连接不健康: %s\n", connID)
					s.onConnectionDisconnected(connID)
				}
			}
		}
	}()
}

// DownloadFile 下载文件

