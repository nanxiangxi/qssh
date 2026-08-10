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

// SSHService SSH 服务（Wails 暴露给前端）。
// 职责：编排连接、Shell、文件操作、分组、窗口、事件广播。
type SSHService struct {
	mu            sync.RWMutex
	clients       map[string]*SSHClient
	storage       *StorageManager
	windowManager *WindowManager
	groupManager  *GroupManager
	auditLogger   *AuditLogger
	monitorHistory *MonitorHistory
	app           *application.App
	shellSessions map[string]bool

	reconnectMu      sync.RWMutex
	reconnectConfigs map[string]*SSHConfig

	directoryUploadCancelMu sync.Mutex
	directoryUploadCancelled bool
}

// NewSSHService 创建 SSH 服务实例
func NewSSHService() *SSHService {
	return &SSHService{
		clients:          make(map[string]*SSHClient),
		storage:          NewStorageManager(),
		groupManager:     NewGroupManager(),
		auditLogger:      NewAuditLogger(),
		monitorHistory:   NewMonitorHistory(),
		shellSessions:    make(map[string]bool),
		reconnectConfigs: make(map[string]*SSHConfig),
	}
}

// SetApp 设置 Wails 应用实例
func (s *SSHService) SetApp(app *application.App) {
	s.app = app
}

// GetClient 获取指定连接的 SSHClient
func (s *SSHService) GetClient(connID string) (*SSHClient, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	client, ok := s.clients[connID]
	if !ok {
		return nil, fmt.Errorf("连接不存在: %s", connID)
	}
	return client, nil
}

// GetApp 获取 Wails 应用实例
func (s *SSHService) GetApp() *application.App {
	return s.app
}

// CancelDirectoryUpload 取消目录上传
func (s *SSHService) CancelDirectoryUpload() {
	s.directoryUploadCancelMu.Lock()
	defer s.directoryUploadCancelMu.Unlock()
	s.directoryUploadCancelled = true
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

// formatSSHError 将 SSH 错误转换为友好提示
func formatSSHError(err error) error {
	if err == nil {
		return nil
	}
	errMsg := err.Error()

	switch {
	case contains(errMsg, "unable to authenticate", "authentication failed"):
		return &FriendlyError{Message: "认证失败：用户名或密码错误"}
	case contains(errMsg, "timeout", "i/o timeout"):
		return &FriendlyError{Message: "连接超时：请检查网络连接或服务器地址"}
	case contains(errMsg, "connection refused"):
		return &FriendlyError{Message: "连接被拒绝：请检查服务器是否运行SSH服务"}
	case contains(errMsg, "no route to host", "network is unreachable"):
		return &FriendlyError{Message: "网络不可达：请检查网络连接"}
	case contains(errMsg, "handshake failed"):
		return &FriendlyError{Message: "连接失败：无法与服务器建立安全连接"}
	}

	if idx := strings.Index(errMsg, ":"); idx > 0 {
		return &FriendlyError{Message: "连接失败：" + errMsg[:idx]}
	}
	return &FriendlyError{Message: "连接失败：" + errMsg}
}

func contains(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// FriendlyError 友好的错误类型
type FriendlyError struct{ Message string }

func (e *FriendlyError) Error() string { return e.Message }

// TestConnection 测试 SSH 连接
func (s *SSHService) TestConnection(config *SSHConfig) error {
	if config.Password == "" && config.PrivateKey == "" && config.KeyPath == "" {
		if stored := s.findStoredPassword(config.Host, config.Port, config.Username); stored != "" {
			config.Password = stored
		}
	}
	client := NewSSHClient(config)
	if err := client.Connect(); err != nil {
		return formatSSHError(err)
	}
	client.Close()
	return nil
}

// Connect 连接到 SSH 服务器
func (s *SSHService) Connect(connID string, config *SSHConfig) error {
	if config.Password == "" && config.PrivateKey == "" && config.KeyPath == "" {
		if stored := s.findStoredPassword(config.Host, config.Port, config.Username); stored != "" {
			config.Password = stored
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	client := NewSSHClient(config)
	if err := client.Connect(); err != nil {
		return formatSSHError(err)
	}
	s.clients[connID] = client
	return nil
}

// CreateAndConnect 创建新连接并建立 SSH 连接
func (s *SSHService) CreateAndConnect(config *SSHConfig) (string, error) {
	result, err := s.CreateAndConnectWithGroup(config, "group-1")
	if err != nil {
		return "", err
	}
	return result["connID"], nil
}

// findStoredPassword 从存储中查找已保存连接的密码（按 host:port:username 匹配）
func (s *SSHService) findStoredPassword(host string, port int, username string) string {
	for _, c := range s.storage.GetAllConnections() {
		if c.Host == host && c.Port == port && c.Username == username && c.Password != "" {
			return c.Password
		}
	}
	return ""
}

// CreateAndConnectWithGroup 创建连接并指定分组
func (s *SSHService) CreateAndConnectWithGroup(config *SSHConfig, groupID string) (map[string]string, error) {
	if config == nil {
		return nil, &FriendlyError{Message: "连接配置不能为空"}
	}

	// 安全设计：前端不接收密码（json:"-"），连接时后端自动从存储补全
	if config.Password == "" && config.PrivateKey == "" && config.KeyPath == "" {
		if stored := s.findStoredPassword(config.Host, config.Port, config.Username); stored != "" {
			config.Password = stored
		}
	}

	connID := generateConnectionID()

	connInfo := &ConnectionInfo{
		ID:           connID,
		Name:         config.Name,
		Host:         config.Host,
		Port:         config.Port,
		Username:     config.Username,
		Password:     config.Password,
		KeyPath:      config.KeyPath,
		PrivateKey:   config.PrivateKey,
		JumpHost:     config.JumpHost,
		JumpUsername: config.JumpUsername,
		JumpPassword: config.JumpPassword,
		JumpKeyPath:  config.JumpKeyPath,
		Status:       "disconnected",
		Saved:        true,
		GroupID:      groupID,
	}

	if err := s.storage.AddConnection(connInfo); err != nil {
		return nil, &FriendlyError{Message: "保存连接配置失败: " + err.Error()}
	}

	type connectResult struct {
		client *SSHClient
		err    error
	}
	resultChan := make(chan connectResult, 1)

	go func() {
		client := NewSSHClient(config)
		client.SetDisconnectCallback(connID, s.onConnectionDisconnected)
		err := client.Connect()
		resultChan <- connectResult{client: client, err: err}
	}()

	timeout := time.Duration(30) * time.Second
	if config.Timeout > 0 {
		timeout = time.Duration(config.Timeout) * time.Second
	}

	select {
	case result := <-resultChan:
		if result.err != nil {
			s.storage.DeleteFromCache(connID)
			return nil, formatSSHError(result.err)
		}
		s.mu.Lock()
		s.clients[connID] = result.client
		s.mu.Unlock()

	case <-time.After(timeout):
		s.storage.DeleteFromCache(connID)
		return nil, &FriendlyError{Message: fmt.Sprintf("连接超时：服务器在 %v 内未响应", timeout)}
	}

	s.storage.UpdateConnectionStatus(connID, "connected")

	group := s.groupManager.GetGroup(groupID)
	if group == nil {
		s.groupManager.CreateGroupWithName(groupID, config.Name)
	}
	if err := s.groupManager.AddConnectionToGroup(groupID, connID); err != nil {
		return nil, err
	}

	return map[string]string{"connID": connID, "groupID": groupID}, nil
}

// Disconnect 断开 SSH 连接
func (s *SSHService) Disconnect(connID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, ok := s.clients[connID]
	if !ok {
		return nil
	}

	client.Close()
	delete(s.clients, connID)
	delete(s.shellSessions, connID)
	s.storage.DeleteFromCache(connID)
	s.broadcastConnections()
	return nil
}

// GetServerKey 根据连接 ID 返回服务器标识
func (s *SSHService) GetServerKey(connID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if client, ok := s.clients[connID]; ok {
		return fmt.Sprintf("%s:%d", client.config.Host, client.config.Port)
	}
	return ""
}

// ResolveCdPath 在独立 session 中执行 cd 并返回真实路径
func (s *SSHService) ResolveCdPath(connID, currentDir, targetPath string) (string, error) {
	shellArg := func(a string) string {
		if a == "~" || strings.HasPrefix(a, "~/") {
			return a
		}
		return "'" + strings.ReplaceAll(a, "'", "'\\''") + "'"
	}
	cmd := fmt.Sprintf("cd %s && cd %s && pwd", shellArg(currentDir), shellArg(targetPath))
	output, err := s.RunCommand(connID, cmd)
	if err != nil {
		return "", err
	}
	path := strings.TrimSpace(output)
	if path == "" || !strings.HasPrefix(path, "/") {
		return "", fmt.Errorf("路径解析失败: %s", output)
	}
	return path, nil
}

// RunCommand 执行命令并返回输出字符串
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

// ExecuteCommand 执行远程命令（自动记录审计日志）
func (s *SSHService) ExecuteCommand(connID string, command string) (*CommandResult, error) {
	s.mu.RLock()
	client, ok := s.clients[connID]
	var host, username string
	if ok {
		host = client.config.Host
		username = client.config.Username
	}
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("连接不存在: %s", connID)
	}

	result, err := client.ExecuteCommand(command)

	// 记录审计日志
	success := err == nil && result != nil && result.Success
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	} else if result != nil && !result.Success {
		errMsg = result.Stderr
	}
	s.auditLogger.Log(connID, host, username, "command", command, success, errMsg)

	return result, err
}

// ExecuteCommandWithSudo 执行远程命令（权限不足时自动用 SSH 密码 sudo 提权）
func (s *SSHService) ExecuteCommandWithSudo(connID string, command string) (*CommandResult, error) {
	s.mu.RLock()
	client, ok := s.clients[connID]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("连接不存在: %s", connID)
	}

	// 先直接执行
	result, err := client.ExecuteCommand(command)
	if err != nil {
		return nil, err
	}
	if result.Success {
		return result, nil
	}

	// 检查是否是权限问题
	errOut := strings.TrimSpace(result.Stderr)
	if errOut == "" {
		errOut = strings.TrimSpace(result.Stdout)
	}
	isPermErr := strings.Contains(errOut, "Permission denied") ||
		strings.Contains(errOut, "permission denied") ||
		strings.Contains(errOut, "Cannot open")

	if !isPermErr {
		return result, nil
	}

	// 用 SSH 密码通过 sudo -S 提权
	password := client.config.Password
	if password == "" {
		return result, fmt.Errorf("权限不足且未配置密码，无法提权")
	}

	encodedPw := base64.StdEncoding.EncodeToString([]byte(password))
	sudoCmd := fmt.Sprintf("echo %s | base64 -d | sudo -S sh -c %s 2>&1",
		encodedPw, shellQuote(command))
	sudoResult, sudoErr := client.ExecuteCommand(sudoCmd)
	if sudoErr != nil {
		return result, fmt.Errorf("sudo 执行失败: %v", sudoErr)
	}
	if sudoResult.Success {
		return sudoResult, nil
	}

	errMsg := strings.TrimSpace(sudoResult.Stderr)
	if errMsg == "" {
		errMsg = strings.TrimSpace(sudoResult.Stdout)
	}
	return sudoResult, fmt.Errorf("sudo 执行失败: %s", errMsg)
}

// ReadFileForUpload 读取本地文件并返回 base64 编码（用于拖放上传）
func (s *SSHService) ReadFileForUpload(connID string, filePath string) (string, error) {
	s.mu.RLock()
	_, ok := s.clients[connID]
	s.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("连接不存在: %s", connID)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %v", err)
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

// IsLocalDirectory 检查本地路径是否是目录
func (s *SSHService) IsLocalDirectory(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// UploadFile 上传文件
func (s *SSHService) UploadFile(connID string, remotePath string, data string) error {
	s.mu.RLock()
	client, ok := s.clients[connID]
	app := s.app
	s.mu.RUnlock()

	if !ok {
		return fmt.Errorf("连接不存在: %s", connID)
	}

	decodedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return fmt.Errorf("base64 解码失败: %v", err)
	}

	progressCallback := func(percent int) {
		if app != nil {
			app.Event.Emit("ssh:upload-progress", map[string]interface{}{
				"connId":   connID,
				"path":     remotePath,
				"progress": percent,
			})
		}
	}

	return client.UploadFileFromBytes(remotePath, decodedData, progressCallback)
}

// DownloadFile 下载文件
func (s *SSHService) DownloadFile(connID string, remotePath string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.clients[connID]
	if !ok {
		return nil, fmt.Errorf("连接不存在: %s", connID)
	}
	return client.DownloadFileToBytes(remotePath)
}

// DownloadFileRange 下载文件指定范围（用于断点续传）
func (s *SSHService) DownloadFileRange(connID string, remotePath string, offset int64, limit int64) ([]byte, error) {
	s.mu.RLock()
	client, ok := s.clients[connID]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("连接不存在: %s", connID)
	}
	return client.DownloadFileRange(remotePath, offset, limit)
}

// GetRemoteFileSize 获取远程文件大小（用于前端判断是否需要续传）
func (s *SSHService) GetRemoteFileSize(connID string, remotePath string) (int64, error) {
	s.mu.RLock()
	client, ok := s.clients[connID]
	s.mu.RUnlock()

	if !ok {
		return 0, fmt.Errorf("连接不存在: %s", connID)
	}
	if client.sftpClient == nil {
		if err := client.InitSFTP(); err != nil {
			return 0, err
		}
	}
	stat, err := client.sftpClient.Stat(remotePath)
	if err != nil {
		return 0, nil // 文件不存在返回 0
	}
	return stat.Size(), nil
}

// ListDirectory 列出目录
func (s *SSHService) ListDirectory(connID string, remotePath string) ([]FileInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.clients[connID]
	if !ok {
		return nil, nil
	}
	return client.ListDirectory(remotePath)
}

// DeleteFile 删除文件
func (s *SSHService) DeleteFile(connID string, remotePath string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.clients[connID]
	if !ok {
		return fmt.Errorf("连接不存在: %s", connID)
	}
	return client.DeleteFile(remotePath)
}

// CreateDirectory 创建目录
func (s *SSHService) CreateDirectory(connID string, remotePath string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.clients[connID]
	if !ok {
		return fmt.Errorf("连接不存在")
	}
	return client.CreateDirectory(remotePath)
}

// RenameFile 重命名文件/目录
func (s *SSHService) RenameFile(connID string, oldPath, newPath string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.clients[connID]
	if !ok {
		return fmt.Errorf("连接不存在")
	}
	return client.RenameFile(oldPath, newPath)
}

// CopyFile 复制文件
func (s *SSHService) CopyFile(connID string, srcPath, destPath string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.clients[connID]
	if !ok {
		return fmt.Errorf("连接不存在")
	}
	return client.CopyFile(srcPath, destPath)
}

// UploadDirectory 上传目录
func (s *SSHService) UploadDirectory(connID string, localPath string, remotePath string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.clients[connID]
	if !ok {
		return fmt.Errorf("连接不存在")
	}
	return client.UploadDirectory(localPath, remotePath)
}

// CreateArchive 创建压缩包
func (s *SSHService) CreateArchive(connID string, files []string, archiveName string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.clients[connID]
	if !ok {
		return "", fmt.Errorf("连接不存在")
	}
	return client.CreateArchive(files, archiveName)
}

// DeleteTempFile 删除临时文件
func (s *SSHService) DeleteTempFile(connID string, filePath string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.clients[connID]
	if !ok {
		return fmt.Errorf("连接不存在")
	}
	return client.DeleteTempFile(filePath)
}

// ExtractArchive 解压压缩包
func (s *SSHService) ExtractArchive(connID string, archivePath string, targetDir string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.clients[connID]
	if !ok {
		return fmt.Errorf("连接不存在")
	}
	return client.ExtractArchive(archivePath, targetDir)
}

// SelectLocalDirectoryAndUpload 选择本地目录并上传
func (s *SSHService) SelectLocalDirectoryAndUpload(connID string, remotePath string) error {
	if s.app == nil {
		return fmt.Errorf("应用实例未初始化")
	}

	s.ResetDirectoryUploadCancel()

	localDir, err := s.app.Dialog.OpenFile().
		SetTitle("选择要上传的目录").
		CanChooseDirectories(true).
		CanChooseFiles(false).
		PromptForSingleSelection()

	if err != nil || localDir == "" {
		return fmt.Errorf("用户取消选择或发生错误")
	}

	s.emitDirectoryProgress(connID, "compressing", 10, "正在压缩目录...")

	tempArchive := filepath.Join(os.TempDir(), fmt.Sprintf("directory_upload_%d.tar.gz", time.Now().Unix()))
	if err := s.createTarGzFromDirectory(localDir, tempArchive); err != nil {
		return fmt.Errorf("压缩本地目录失败: %v", err)
	}

	archiveInfo, _ := os.Stat(tempArchive)
	if archiveInfo != nil {
		s.emitDirectoryProgress(connID, "uploading", 40, "正在上传压缩包...")
	}

	archiveData, err := os.ReadFile(tempArchive)
	os.Remove(tempArchive)
	if err != nil {
		return fmt.Errorf("读取压缩文件失败: %v", err)
	}

	s.mu.RLock()
	client, ok := s.clients[connID]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("连接不存在")
	}

	remoteArchivePath := fmt.Sprintf("%s/temp_upload_%d.tar.gz", remotePath, time.Now().Unix())
	if err := s.uploadFileWithProgress(client, remoteArchivePath, archiveData); err != nil {
		if s.IsDirectoryUploadCancelled() {
			return fmt.Errorf("用户取消了上传")
		}
		return fmt.Errorf("上传压缩文件失败: %v", err)
	}

	s.emitDirectoryProgress(connID, "extracting", 80, "正在解压...")

	if err := client.ExtractArchive(remoteArchivePath, remotePath); err != nil {
		return fmt.Errorf("解压失败: %v", err)
	}

	client.DeleteTempFile(remoteArchivePath)
	return nil
}

func (s *SSHService) emitDirectoryProgress(connID, stage string, progress int, message string) {
	if s.app != nil {
		s.app.Event.Emit("directory-upload-progress", map[string]interface{}{
			"stage":    stage,
			"progress": progress,
			"message":  message,
		})
	}
}

func (s *SSHService) uploadFileWithProgress(client *SSHClient, remotePath string, data []byte) error {
	if client.sftpClient == nil {
		if err := client.InitSFTP(); err != nil {
			return err
		}
	}

	remoteFile, err := client.sftpClient.Create(remotePath)
	if err != nil {
		// SFTP 权限不足，回退到 sudo 方式
		return client.uploadViaSudo(remotePath, data, nil)
	}
	defer remoteFile.Close()

	const chunkSize = 1024 * 1024
	totalSize := len(data)
	uploadedSize := 0

	for uploadedSize < totalSize {
		if s.IsDirectoryUploadCancelled() {
			remoteFile.Close()
			client.DeleteTempFile(remotePath)
			return fmt.Errorf("用户取消了上传")
		}

		end := uploadedSize + chunkSize
		if end > totalSize {
			end = totalSize
		}

		written, err := remoteFile.Write(data[uploadedSize:end])
		if err != nil {
			client.DeleteTempFile(remotePath)
			return fmt.Errorf("写入数据失败: %v", err)
		}

		uploadedSize += written

		if s.app != nil {
			progress := 40 + int(float64(uploadedSize)/float64(totalSize)*40)
			s.app.Event.Emit("directory-upload-progress", map[string]interface{}{
				"stage":    "uploading",
				"progress": progress,
				"message":  fmt.Sprintf("正在上传... (%.1f MB / %.1f MB)", float64(uploadedSize)/(1024*1024), float64(totalSize)/(1024*1024)),
			})
		}
	}

	return nil
}

func (s *SSHService) createTarGzFromDirectory(sourceDir string, destArchive string) error {
	file, err := os.Create(destArchive)
	if err != nil {
		return err
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	baseName := filepath.Base(sourceDir)

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}

		relPath = baseName + "/" + relPath
		relPath = filepath.ToSlash(relPath)

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			data, err := os.Open(path)
			if err != nil {
				return err
			}
			defer data.Close()
			_, err = io.Copy(tarWriter, data)
			return err
		}

		return nil
	})
}

// IsConnected 检查连接状态
func (s *SSHService) IsConnected(connID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.clients[connID]
	if !ok {
		return false
	}
	return client.IsConnected()
}

// GetConnectionList 获取所有连接 ID 列表
func (s *SSHService) GetConnectionList() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]string, 0, len(s.clients))
	for id := range s.clients {
		ids = append(ids, id)
	}
	return ids
}

// ========== 连接存储管理 API ==========

func (s *SSHService) AddConnection(conn *ConnectionInfo) error       { return s.storage.AddConnection(conn) }
func (s *SSHService) UpdateConnection(conn *ConnectionInfo) error    { return s.storage.UpdateConnection(conn) }
func (s *SSHService) DeleteConnection(id string) error               { return s.storage.DeleteConnection(id) }
func (s *SSHService) GetConnection(id string) (*ConnectionInfo, error) { return s.storage.GetConnection(id) }

func (s *SSHService) SyncImportConnection(conn *ConnectionInfo) error {
	err := s.storage.SyncImportConnection(conn)
	if err == nil {
		s.broadcastConnections()
	}
	return err
}

// ImportSSHConfig 解析 ~/.ssh/config 文件，返回可用的连接列表
func (s *SSHService) ImportSSHConfig() ([]SSHConfig, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("获取用户主目录失败: %v", err)
	}

	configPath := filepath.Join(home, ".ssh", "config")
	entries, err := ParseSSHConfig(configPath)
	if err != nil {
		return nil, err
	}

	var configs []SSHConfig
	for _, e := range entries {
		configs = append(configs, SSHConfig{
			Name:       e.Host,
			Host:       e.HostName,
			Port:       e.Port,
			Username:   e.User,
			KeyPath:    e.IdentityFile,
		})
	}

	return configs, nil
}

// GetAllConnections 获取所有连接（含分组 ID 和真实在线状态）
func (s *SSHService) GetAllConnections() []*ConnectionInfo {
	connections := s.storage.GetAllConnections()
	var result []*ConnectionInfo
	for _, conn := range connections {
		if !conn.Saved {
			continue
		}
		group := s.groupManager.GetGroupByConnID(conn.ID)
		if group != nil {
			conn.GroupID = group.ID
		}
		// 按 host:port:username 匹配在线状态（存储 ID 可能与活跃连接 ID 不同）
		s.mu.RLock()
		isOnline := false
		for _, client := range s.clients {
			if client.config.Host == conn.Host && client.config.Port == conn.Port && client.config.Username == conn.Username {
				isOnline = true
				break
			}
		}
		s.mu.RUnlock()
		if isOnline {
			conn.Status = "connected"
		} else {
			conn.Status = "disconnected"
		}
		result = append(result, conn)
	}
	return result
}

func (s *SSHService) SaveConnection(id string) error {
	err := s.storage.MarkAsSaved(id)
	if err == nil {
		s.broadcastConnections()
	}
	return err
}

func (s *SSHService) UnsaveConnection(id string) {
	s.storage.MarkAsUnsaved(id)
}

func (s *SSHService) broadcastConnections() {
	if s.app == nil {
		return
	}
	all := s.GetAllConnections()
	if all == nil {
		all = []*ConnectionInfo{}
	}
	s.app.Event.Emit("ssh:connections-updated", map[string]interface{}{
		"connections": all,
		"timestamp":   time.Now().UnixMilli(),
	})
}

// ========== 分组管理 API ==========

func (s *SSHService) CreateGroup(name string) string    { return s.groupManager.CreateGroup(name).ID }
func (s *SSHService) GetAllGroups() []*SSHGroup         { return s.groupManager.GetAllGroups() }
func (s *SSHService) GetDefaultGroupID() string         { return "group-1" }
func (s *SSHService) AddToGroup(groupID, connID string) error { return s.groupManager.AddConnectionToGroup(groupID, connID) }
func (s *SSHService) RemoveFromGroup(groupID, connID string) error { return s.groupManager.RemoveConnectionFromGroup(groupID, connID) }

// UpdateGroup 更新分组属性（名称、颜色、父分组）
func (s *SSHService) UpdateGroup(groupID, name, color, parentID string) error {
	return s.groupManager.UpdateGroup(groupID, name, color, parentID)
}

// ReorderGroups 更新分组排序
func (s *SSHService) ReorderGroups(groupIDs []string) error {
	return s.groupManager.ReorderGroups(groupIDs)
}

// GetChildGroups 获取子分组
func (s *SSHService) GetChildGroups(parentID string) []*SSHGroup {
	return s.groupManager.GetChildGroups(parentID)
}

func (s *SSHService) GetGroupConnections(groupID string) []string {
	group := s.groupManager.GetGroup(groupID)
	if group == nil {
		return []string{}
	}
	return group.ConnIDs
}

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

// CloseGroup 关闭分组
func (s *SSHService) CloseGroup(groupID string) error {
	group := s.groupManager.GetGroup(groupID)
	if group == nil {
		return nil
	}

	connIDs := make([]string, len(group.ConnIDs))
	copy(connIDs, group.ConnIDs)

	for _, connID := range connIDs {
		s.mu.Lock()
		if client, ok := s.clients[connID]; ok {
			client.Close()
			delete(s.clients, connID)
			s.storage.DeleteFromCache(connID)
		}
		s.mu.Unlock()
	}

	s.groupManager.ClearGroup(groupID)

	if s.app != nil {
		s.app.Event.Emit("ssh:group-updated", map[string]interface{}{
			"groupID":     groupID,
			"connections": []string{},
		})
	}

	s.groupManager.DeleteGroup(groupID)
	return nil
}

func (s *SSHService) OpenSSHWindow(groupID string, groupName string, activeConnID string) error {
	if s.windowManager == nil {
		return &FriendlyError{Message: "窗口管理器未初始化"}
	}
	return s.windowManager.CreateSSHWindow(groupID, groupName, activeConnID)
}

func (s *SSHService) GetGroupByConnID(connID string) *SSHGroup {
	return s.groupManager.GetGroupByConnID(connID)
}

// ========== Shell 会话 API ==========

func (s *SSHService) StartShellSession(connID string) error {
	return s.StartShellSessionWithID(connID, "default")
}

func (s *SSHService) StartShellSessionWithID(connID, sessionID string) error {
	s.mu.RLock()
	client, ok := s.clients[connID]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("连接不存在: %s", connID)
	}

	if _, err := client.CreateShell(sessionID); err != nil {
		return err
	}

	s.mu.Lock()
	s.shellSessions[connID] = true
	s.mu.Unlock()
	return nil
}

func (s *SSHService) WriteToTerminal(connID string, data string) error {
	return s.WriteToTerminalByID(connID, "default", data)
}

func (s *SSHService) WriteToTerminalByID(connID, sessionID string, data string) error {
	s.mu.RLock()
	client, ok := s.clients[connID]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("连接不存在: %s", connID)
	}
	if !client.IsShellActive(sessionID) {
		return fmt.Errorf("Shell会话未初始化: %s", sessionID)
	}
	return client.WriteToShell(sessionID, []byte(data))
}

func (s *SSHService) ReadFromShellSession(connID, sessionID string, buf []byte) (int, error) {
	s.mu.RLock()
	client, ok := s.clients[connID]
	s.mu.RUnlock()
	if !ok {
		return 0, fmt.Errorf("连接不存在")
	}
	return client.ReadFromShell(sessionID, buf)
}

func (s *SSHService) IsShellSessionActive(connID, sessionID string) bool {
	s.mu.RLock()
	client, ok := s.clients[connID]
	s.mu.RUnlock()
	if !ok {
		return false
	}
	return client.IsShellActive(sessionID)
}

func (s *SSHService) GetSessionIDs(connID string) []string {
	s.mu.RLock()
	client, ok := s.clients[connID]
	s.mu.RUnlock()
	if !ok {
		return nil
	}
	return client.GetSessionIDs()
}

func (s *SSHService) ResizeTerminal(connID string, cols, rows int) error {
	return s.ResizeTerminalByID(connID, "default", cols, rows)
}

func (s *SSHService) ResizeTerminalByID(connID, sessionID string, cols, rows int) error {
	s.mu.RLock()
	client, ok := s.clients[connID]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("连接不存在")
	}
	return client.ResizeTerminalByID(sessionID, cols, rows)
}

func (s *SSHService) IsShellActive(connID string) bool {
	return s.IsShellSessionActive(connID, "default")
}

func (s *SSHService) CloseShellSession(connID string) error {
	return s.CloseShellSessionByID(connID, "default")
}

func (s *SSHService) CloseShellSessionByID(connID, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, ok := s.clients[connID]
	if !ok {
		return fmt.Errorf("连接不存在")
	}
	client.CloseShell(sessionID)
	return nil
}

func (s *SSHService) ReadFromShell(connID string, buf []byte) (int, error) {
	return s.ReadFromShellSession(connID, "default", buf)
}

var connIDCounter int64

func generateConnectionID() string {
	connIDCounter++
	return fmt.Sprintf("conn_%d_%d", time.Now().UnixNano(), connIDCounter)
}

// ========== 系统监控 API ==========

func (s *SSHService) GetLatency(connID string) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.clients[connID]
	if !ok {
		return 0
	}
	return client.UpdateLatency()
}

func (s *SSHService) GetSystemStats(connID string) (*SystemStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.clients[connID]
	if !ok {
		return nil, fmt.Errorf("连接不存在")
	}

	stats, err := client.GetSystemStats(context.Background())
	if err != nil {
		return nil, err
	}

	// 自动记录监控快照
	if stats != nil {
		var diskPercent float64
		var diskUsed, diskTotal uint64
		if len(stats.Disk.Partitions) > 0 {
			diskPercent = stats.Disk.Partitions[0].UsedPercent
			diskUsed = stats.Disk.Partitions[0].Used
			diskTotal = stats.Disk.Partitions[0].Total
		}
		s.monitorHistory.Add(connID, MonitorSnapshot{
			Timestamp:   time.Now().UnixMilli(),
			CPUPercent:  stats.CPU.UsagePercent,
			MemPercent:  stats.Memory.UsedPercent,
			MemUsed:     stats.Memory.Used,
			MemTotal:    stats.Memory.Total,
			DiskPercent: diskPercent,
			DiskUsed:    diskUsed,
			DiskTotal:   diskTotal,
			NetRx:       stats.Network.TotalRx,
			NetTx:       stats.Network.TotalTx,
		})
	}

	return stats, nil
}

// GetMonitorHistory 获取监控历史数据
func (s *SSHService) GetMonitorHistory(connID string, count int) []MonitorSnapshot {
	return s.monitorHistory.Get(connID, count)
}

// GetMonitorHistorySince 获取指定时间以来的监控历史
func (s *SSHService) GetMonitorHistorySince(connID string, since int64) []MonitorSnapshot {
	return s.monitorHistory.GetSince(connID, since)
}

func (s *SSHService) GetProcessList(connID string) ([]ProcessInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.clients[connID]
	if !ok {
		return nil, fmt.Errorf("连接不存在")
	}
	return client.GetProcessList(context.Background())
}

func (s *SSHService) KillProcess(connID string, pid int32) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.clients[connID]
	if !ok {
		return fmt.Errorf("连接不存在")
	}
	return client.KillProcess(context.Background(), pid)
}

func (s *SSHService) SendSignal(connID string, pid int32, signal string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.clients[connID]
	if !ok {
		return fmt.Errorf("连接不存在")
	}
	return client.SendSignal(context.Background(), pid, signal)
}

func (s *SSHService) GetProcessDetail(connID string, pid int32) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.clients[connID]
	if !ok {
		return "", fmt.Errorf("连接不存在")
	}
	return client.GetProcessDetail(context.Background(), pid)
}

// ========== 文件搜索 API ==========

func (s *SSHService) ListFiles(connID string, path string) ([]FileInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.clients[connID]
	if !ok {
		return nil, fmt.Errorf("连接不存在")
	}
	return client.ListDirectory(path)
}

// GetFileOwners 按需获取文件的 owner/group（前端异步调用）
func (s *SSHService) GetFileOwners(connID string, dir string, filenames []string) map[string][2]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.clients[connID]
	if !ok {
		return nil
	}
	return client.GetFileOwners(dir, filenames)
}

func (s *SSHService) SearchFiles(connID string, basePath string, keyword string, searchID string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.clients[connID]
	if !ok {
		return fmt.Errorf("连接不存在")
	}
	return client.SearchFiles(basePath, keyword, s.app, searchID)
}

func (s *SSHService) CancelSearch(connID string, searchID string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.clients[connID]
	if !ok {
		return fmt.Errorf("连接不存在")
	}
	client.CancelSearch(searchID)
	return nil
}

// ========== 断线重连 ==========

func (s *SSHService) onConnectionDisconnected(connID string) {
	s.mu.RLock()
	client, ok := s.clients[connID]
	if !ok {
		s.mu.RUnlock()
		return
	}
	config := client.config
	s.mu.RUnlock()

	s.reconnectMu.Lock()
	s.reconnectConfigs[connID] = config
	s.reconnectMu.Unlock()

	s.storage.UpdateConnectionStatus(connID, "disconnected")

	if s.app != nil {
		s.app.Event.Emit("ssh:connection-disconnected", map[string]interface{}{
			"connID":    connID,
			"host":      config.Host,
			"timestamp": time.Now().UnixMilli(),
		})
	}

	s.broadcastConnections()

	// 自动重连（后台执行，不阻塞）
	s.AutoReconnect(connID)
}

// Reconnect 重新连接
func (s *SSHService) Reconnect(connID string) error {
	s.reconnectMu.RLock()
	config, exists := s.reconnectConfigs[connID]
	s.reconnectMu.RUnlock()

	if !exists {
		connInfo, err := s.storage.GetConnection(connID)
		if err != nil || connInfo == nil {
			return &FriendlyError{Message: "无法找到连接配置，请重新连接"}
		}
		config = &SSHConfig{
			Name:         connInfo.Name,
			Host:         connInfo.Host,
			Port:         connInfo.Port,
			Username:     connInfo.Username,
			Password:     connInfo.Password,
			KeyPath:      connInfo.KeyPath,
			PrivateKey:   connInfo.PrivateKey,
			JumpHost:     connInfo.JumpHost,
			JumpUsername: connInfo.JumpUsername,
			JumpPassword: connInfo.JumpPassword,
			JumpKeyPath:  connInfo.JumpKeyPath,
		}
	}

	s.mu.Lock()
	if oldClient, ok := s.clients[connID]; ok {
		oldClient.Close()
		delete(s.clients, connID)
	}
	delete(s.shellSessions, connID)
	s.mu.Unlock()

	client := NewSSHClient(config)
	client.SetDisconnectCallback(connID, s.onConnectionDisconnected)

	if s.app != nil {
		s.app.Event.Emit("ssh:connection-reconnecting", map[string]interface{}{"connID": connID})
	}

	if err := client.Connect(); err != nil {
		if s.app != nil {
			s.app.Event.Emit("ssh:connection-reconnect-failed", map[string]interface{}{
				"connID": connID,
				"error":  formatSSHError(err).Error(),
			})
		}
		return formatSSHError(err)
	}

	s.mu.Lock()
	s.clients[connID] = client
	s.mu.Unlock()

	s.storage.UpdateConnectionStatus(connID, "connected")

	s.reconnectMu.Lock()
	delete(s.reconnectConfigs, connID)
	s.reconnectMu.Unlock()

	if s.app != nil {
		s.app.Event.Emit("ssh:connection-reconnected", map[string]interface{}{"connID": connID})
	}

	s.broadcastConnections()
	return nil
}

// AutoReconnect 自动重连（指数退避，最多重试 10 次，约 3 分钟）
func (s *SSHService) AutoReconnect(connID string) {
	go func() {
		maxRetries := 10
		retryDelay := 2 * time.Second
		maxDelay := 30 * time.Second

		for i := 0; i < maxRetries; i++ {
			time.Sleep(retryDelay)

			if err := s.Reconnect(connID); err == nil {
				return
			}

			// 指数退避，上限 30 秒
			retryDelay *= 2
			if retryDelay > maxDelay {
				retryDelay = maxDelay
			}
		}

		if s.app != nil {
			s.app.Event.Emit("ssh:connection-reconnect-failed", map[string]interface{}{
				"connID":  connID,
				"error":   "自动重连失败，请手动重连",
				"autoMax": true,
			})
		}
	}()
}

// ========== SSH 密钥管理 API ==========

var defaultKeyManager = NewKeyManager()

// ListSSHKeys 列出所有 SSH 密钥
func (s *SSHService) ListSSHKeys() ([]SSHKeyInfo, error) {
	return defaultKeyManager.ListKeys()
}

// GenerateSSHKey 生成新密钥
func (s *SSHService) GenerateSSHKey(name, keyType, comment string, bits int) error {
	return defaultKeyManager.GenerateKey(name, keyType, comment, bits)
}

// DeploySSHPublicKey 部署公钥到远程服务器
func (s *SSHService) DeploySSHPublicKey(keyName, connID string) error {
	s.mu.RLock()
	client, ok := s.clients[connID]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("连接不存在: %s", connID)
	}
	return defaultKeyManager.DeployPublicKey(keyName, client)
}

// GetSSHPublicKey 获取公钥内容
func (s *SSHService) GetSSHPublicKey(keyName string) (string, error) {
	return defaultKeyManager.GetPublicKey(keyName)
}

// DeleteSSHKey 删除密钥
func (s *SSHService) DeleteSSHKey(keyName string) error {
	return defaultKeyManager.DeleteKey(keyName)
}

// ========== 审计日志 API ==========

// GetAuditLogs 获取审计日志（最近 N 条）
func (s *SSHService) GetAuditLogs(limit int) []AuditEntry {
	return s.auditLogger.GetLogs(limit)
}

// GetAuditLogsByDate 按日期获取审计日志
func (s *SSHService) GetAuditLogsByDate(date string) []AuditEntry {
	return s.auditLogger.GetLogsByDate(date)
}

// ExportAuditLogs 导出审计日志
func (s *SSHService) ExportAuditLogs(startDate, endDate string) (string, error) {
	return s.auditLogger.ExportLogs(startDate, endDate)
}

// CheckConnectionHealth 检查连接健康状态
func (s *SSHService) CheckConnectionHealth(connID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.clients[connID]
	if !ok {
		return false
	}
	return client.IsConnected()
}

// StartHealthCheck 启动健康检查定时器
func (s *SSHService) StartHealthCheck() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
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
					s.onConnectionDisconnected(connID)
				}
			}
		}
	}()
}

// ========== 系统环境检测 API ==========

func (s *SSHService) GetSystemEnvironments(connID string) ([]EnvInfo, error) {
	s.mu.RLock()
	client, ok := s.clients[connID]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("连接不存在: %s", connID)
	}
	return client.GetSystemEnvironments()
}
