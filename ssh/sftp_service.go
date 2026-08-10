package ssh

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// FileInfo 文件信息结构体
type FileInfo struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	Mode    string    `json:"mode"`
	ModTime time.Time `json:"modTime"`
	IsDir   bool      `json:"isDir"`
	Owner   string    `json:"owner"`
	Group   string    `json:"group"`
}

// InitSFTP 初始化 SFTP 客户端
func (s *SSHClient) InitSFTP() error {
	if s.isConnected.Load() == 0 {
		return fmt.Errorf("未连接到SSH服务器")
	}
	if s.sftpClient != nil {
		return nil
	}

	client, err := sftp.NewClient(s.client)
	if err != nil {
		return fmt.Errorf("初始化SFTP失败: %v", err)
	}
	s.sftpClient = client
	return nil
}

// UploadFile 上传文件
func (s *SSHClient) UploadFile(localPath, remotePath string) error {
	if s.sftpClient == nil {
		if err := s.InitSFTP(); err != nil {
			return err
		}
	}

	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("打开本地文件失败: %v", err)
	}
	defer localFile.Close()

	remoteFile, err := s.sftpClient.Create(remotePath)
	if err != nil {
		// SFTP 权限不足，回退到 sudo 方式
		localFile.Close()
		data, readErr := os.ReadFile(localPath)
		if readErr != nil {
			return fmt.Errorf("读取本地文件失败: %v", readErr)
		}
		return s.uploadViaSudo(remotePath, data, nil)
	}
	defer remoteFile.Close()

	if _, err := io.Copy(remoteFile, localFile); err != nil {
		return fmt.Errorf("上传文件失败: %v", err)
	}
	return nil
}

// DownloadFile 下载文件（支持断点续传）
// 如果本地文件已存在且小于远程文件大小，自动从断点继续下载
func (s *SSHClient) DownloadFile(remotePath, localPath string) error {
	if s.sftpClient == nil {
		if err := s.InitSFTP(); err != nil {
			return err
		}
	}

	remoteFile, err := s.sftpClient.Open(remotePath)
	if err != nil {
		return fmt.Errorf("打开远程文件失败: %v", err)
	}
	defer remoteFile.Close()

	// 获取远程文件大小
	remoteStat, err := remoteFile.Stat()
	if err != nil {
		return fmt.Errorf("获取远程文件信息失败: %v", err)
	}
	remoteSize := remoteStat.Size()

	// 检查本地文件是否已存在（断点续传）
	var localFile *os.File
	offset := int64(0)
	if info, err := os.Stat(localPath); err == nil {
		existingSize := info.Size()
		if existingSize >= remoteSize {
			// 本地文件已完整
			return nil
		}
		offset = existingSize
		localFile, err = os.OpenFile(localPath, os.O_WRONLY|os.O_APPEND, 0644)
	} else {
		localFile, err = os.Create(localPath)
	}
	if err != nil {
		return fmt.Errorf("创建/打开本地文件失败: %v", err)
	}
	defer localFile.Close()

	// 从断点位置开始读取
	if offset > 0 {
		if _, err := remoteFile.Seek(offset, io.SeekStart); err != nil {
			return fmt.Errorf("seek 到断点位置失败: %v", err)
		}
	}

	if _, err := io.Copy(localFile, remoteFile); err != nil {
		return fmt.Errorf("下载文件失败 (已下载 %d/%d bytes): %v", offset, remoteSize, err)
	}
	return nil
}

// DownloadFileRange 下载文件指定范围（用于续传或分块下载）
func (s *SSHClient) DownloadFileRange(remotePath string, offset int64, limit int64) ([]byte, error) {
	if s.sftpClient == nil {
		if err := s.InitSFTP(); err != nil {
			return nil, err
		}
	}

	remoteFile, err := s.sftpClient.Open(remotePath)
	if err != nil {
		return nil, fmt.Errorf("打开远程文件失败: %v", err)
	}
	defer remoteFile.Close()

	// seek 到指定位置
	if offset > 0 {
		if _, err := remoteFile.Seek(offset, io.SeekStart); err != nil {
			return nil, fmt.Errorf("seek 失败: %v", err)
		}
	}

	// 读取指定范围
	if limit > 0 {
		data := make([]byte, limit)
		n, err := io.ReadFull(remoteFile, data)
		if err != nil && err != io.ErrUnexpectedEOF {
			return nil, fmt.Errorf("读取失败: %v", err)
		}
		return data[:n], nil
	}

	// 无限制，读取全部
	data, err := io.ReadAll(remoteFile)
	if err != nil {
		return nil, fmt.Errorf("读取失败: %v", err)
	}
	return data, nil
}

// ListDirectory 列出远程目录内容（性能优化：移除逐文件 Lstat，异步获取 owner/group）
func (s *SSHClient) ListDirectory(remotePath string) ([]FileInfo, error) {
	if s.sftpClient == nil {
		if err := s.InitSFTP(); err != nil {
			return nil, err
		}
	}

	if remotePath == "" {
		remotePath = "/"
	}

	files, err := s.sftpClient.ReadDir(remotePath)
	if err != nil {
		if strings.Contains(err.Error(), "file does not exist") || strings.Contains(err.Error(), "no such file") {
			return nil, fmt.Errorf("目录不存在: %s", remotePath)
		}
		return nil, fmt.Errorf("读取目录失败: %v", err)
	}

	var fileInfos []FileInfo
	for _, file := range files {
		fullPath := remotePath
		if !strings.HasSuffix(remotePath, "/") {
			fullPath += "/"
		}
		fullPath += file.Name()

		fileInfos = append(fileInfos, FileInfo{
			Name:    file.Name(),
			Path:    fullPath,
			Size:    file.Size(),
			Mode:    file.Mode().String(),
			ModTime: file.ModTime(),
			IsDir:   file.IsDir(),
			Owner:   "-",
			Group:   "-",
		})
	}

	return fileInfos, nil
}

// batchGetFileOwners 一次 SSH 调用获取目录下所有文件的 owner 和 group
func (s *SSHClient) batchGetFileOwners(dir string, files []os.FileInfo) []string {
	n := len(files)
	result := make([]string, n*2)
	for i := range result {
		result[i] = "unknown"
	}
	if n == 0 {
		return result
	}

	args := make([]string, n)
	for i, f := range files {
		args[i] = fmt.Sprintf("'%s/%s'", strings.TrimSuffix(dir, "/"), f.Name())
	}
	cmd := fmt.Sprintf("stat -c '%%U %%G' %s", strings.Join(args, " "))

	res, err := s.ExecuteCommand(cmd)
	if err != nil || !res.Success {
		return result
	}

	lines := strings.Split(strings.TrimSpace(res.Stdout), "\n")
	for i, line := range lines {
		if i >= n {
			break
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			result[i] = fields[0]
			result[n+i] = fields[1]
		}
	}
	return result
}

// GetFileOwners 按需获取文件的 owner/group（前端可异步调用）
func (s *SSHClient) GetFileOwners(dir string, filenames []string) map[string][2]string {
	result := make(map[string][2]string, len(filenames))
	for _, name := range filenames {
		result[name] = [2]string{"unknown", "unknown"}
	}
	if len(filenames) == 0 {
		return result
	}

	args := make([]string, len(filenames))
	for i, name := range filenames {
		args[i] = fmt.Sprintf("'%s/%s'", strings.TrimSuffix(dir, "/"), name)
	}
	cmd := fmt.Sprintf("stat -c '%%U %%G' %s", strings.Join(args, " "))

	res, err := s.ExecuteCommand(cmd)
	if err != nil || !res.Success {
		return result
	}

	lines := strings.Split(strings.TrimSpace(res.Stdout), "\n")
	for i, line := range lines {
		if i >= len(filenames) {
			break
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			result[filenames[i]] = [2]string{fields[0], fields[1]}
		}
	}
	return result
}

// SearchFiles 递归搜索文件
func (s *SSHClient) SearchFiles(basePath string, keyword string, app *application.App, searchID string) error {
	if s.sftpClient == nil {
		if err := s.InitSFTP(); err != nil {
			return err
		}
	}

	keyword = strings.ToLower(keyword)
	foundCount := 0

	s.searchMutex.Lock()
	s.searchCancelMap[searchID] = false
	s.searchMutex.Unlock()

	if app != nil {
		app.Event.Emit("search-start", map[string]interface{}{
			"searchID": searchID,
			"basePath": basePath,
			"keyword":  keyword,
		})
	}

	var searchRecursive func(path string) error
	searchRecursive = func(path string) error {
		s.searchMutex.RLock()
		cancelled := s.searchCancelMap[searchID]
		s.searchMutex.RUnlock()
		if cancelled {
			return fmt.Errorf("搜索已取消")
		}

		files, err := s.sftpClient.ReadDir(path)
		if err != nil {
			return err
		}

		for _, file := range files {
			s.searchMutex.RLock()
			cancelled := s.searchCancelMap[searchID]
			s.searchMutex.RUnlock()
			if cancelled {
				return fmt.Errorf("搜索已取消")
			}

			if file.Name() == "." || file.Name() == ".." {
				continue
			}

			fullPath := path
			if !strings.HasSuffix(path, "/") {
				fullPath += "/"
			}
			fullPath += file.Name()

			if strings.Contains(strings.ToLower(file.Name()), keyword) {
				fileStat, err := s.sftpClient.Lstat(fullPath)
				isDir := false
				if err == nil && fileStat != nil {
					isDir = fileStat.IsDir()
				} else {
					isDir = file.IsDir()
				}

				relativePath := fullPath
				if strings.HasPrefix(fullPath, basePath) {
					relativePath = strings.TrimPrefix(fullPath, basePath)
					relativePath = strings.TrimPrefix(relativePath, "/")
				}

				if app != nil {
					foundCount++
					app.Event.Emit("search-result", map[string]interface{}{
						"searchID":     searchID,
						"file":         FileInfo{Name: file.Name(), Path: fullPath, Size: file.Size(), Mode: file.Mode().String(), ModTime: file.ModTime(), IsDir: isDir, Owner: "-", Group: "-"},
						"relativePath": relativePath,
					})
				}
			}

			if file.IsDir() {
				if err := searchRecursive(fullPath); err != nil {
					if strings.Contains(err.Error(), "搜索已取消") {
						return err
					}
				}
			}
		}
		return nil
	}

	err := searchRecursive(basePath)

	s.searchMutex.Lock()
	delete(s.searchCancelMap, searchID)
	s.searchMutex.Unlock()

	if app != nil {
		app.Event.Emit("search-complete", map[string]interface{}{
			"searchID":   searchID,
			"totalFound": foundCount,
			"error":      err != nil,
		})
	}

	if err != nil && strings.Contains(err.Error(), "搜索已取消") {
		return nil
	}
	return err
}

// CancelSearch 取消搜索
func (s *SSHClient) CancelSearch(searchID string) {
	s.searchMutex.Lock()
	defer s.searchMutex.Unlock()
	if _, ok := s.searchCancelMap[searchID]; ok {
		s.searchCancelMap[searchID] = true
	}
}

// DeleteFile 删除远程文件或目录
func (s *SSHClient) DeleteFile(remotePath string) error {
	if s.sftpClient == nil {
		if err := s.InitSFTP(); err != nil {
			return err
		}
	}

	fileInfo, err := s.sftpClient.Lstat(remotePath)
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %v", err)
	}

	isSymlink := fileInfo.Mode()&os.ModeSymlink != 0

	var deleteErr error
	if isSymlink {
		deleteErr = s.sftpClient.Remove(remotePath)
	} else if fileInfo.IsDir() {
		deleteErr = s.removeDir(remotePath)
	} else {
		deleteErr = s.sftpClient.Remove(remotePath)
	}

	// SFTP 删除失败，尝试 sudo
	if deleteErr != nil {
		return s.deleteViaSudo(remotePath, fileInfo.IsDir())
	}
	return nil
}

// removeDir 递归删除目录
func (s *SSHClient) removeDir(path string) error {
	files, err := s.sftpClient.ReadDir(path)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.Name() == "." || file.Name() == ".." {
			continue
		}

		fullPath := path + "/" + file.Name()

		if file.IsDir() {
			err = s.removeDir(fullPath)
		} else {
			err = s.sftpClient.Remove(fullPath)
		}
		if err != nil {
			return err
		}
	}

	return s.sftpClient.RemoveDirectory(path)
}

// CreateDirectory 创建目录
func (s *SSHClient) CreateDirectory(remotePath string) error {
	if s.sftpClient == nil {
		if err := s.InitSFTP(); err != nil {
			return err
		}
	}

	err := s.sftpClient.MkdirAll(remotePath)
	if err != nil {
		// SFTP 失败，尝试 sudo
		return s.mkdirViaSudo(remotePath)
	}
	return nil
}

// mkdirViaSudo 通过 sudo 创建目录
func (s *SSHClient) mkdirViaSudo(remotePath string) error {
	password := s.config.Password
	if password == "" {
		return fmt.Errorf("权限不足且未配置密码，无法创建目录 %s", remotePath)
	}

	encodedPw := base64.StdEncoding.EncodeToString([]byte(password))
	cmd := fmt.Sprintf("echo %s | base64 -d | sudo -S mkdir -p %s 2>&1",
		encodedPw, remotePath)
	result, err := s.ExecuteCommand(cmd)
	if err != nil || !result.Success {
		return fmt.Errorf("权限不足，无法创建目录 %s", remotePath)
	}
	return nil
}

// RenameFile 重命名文件/目录
func (s *SSHClient) RenameFile(oldPath, newPath string) error {
	if s.sftpClient == nil {
		if err := s.InitSFTP(); err != nil {
			return err
		}
	}

	err := s.sftpClient.Rename(oldPath, newPath)
	if err != nil {
		// SFTP 失败，尝试 sudo
		return s.renameViaSudo(oldPath, newPath)
	}
	return nil
}

// renameViaSudo 通过 sudo 重命名文件
func (s *SSHClient) renameViaSudo(oldPath, newPath string) error {
	password := s.config.Password
	if password == "" {
		return fmt.Errorf("权限不足且未配置密码，无法重命名 %s", oldPath)
	}

	encodedPw := base64.StdEncoding.EncodeToString([]byte(password))
	cmd := fmt.Sprintf("echo %s | base64 -d | sudo -S mv %s %s 2>&1",
		encodedPw, oldPath, newPath)
	result, err := s.ExecuteCommand(cmd)
	if err != nil || !result.Success {
		return fmt.Errorf("权限不足，无法重命名 %s", oldPath)
	}
	return nil
}

// CopyFile 复制文件
func (s *SSHClient) CopyFile(srcPath, destPath string) error {
	if s.sftpClient == nil {
		if err := s.InitSFTP(); err != nil {
			return err
		}
	}

	srcFile, err := s.sftpClient.Open(srcPath)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %v", err)
	}

	srcInfo, err := srcFile.Stat()
	if err != nil {
		srcFile.Close()
		return fmt.Errorf("获取源文件信息失败: %v", err)
	}

	destFile, err := s.sftpClient.Create(destPath)
	if err != nil {
		srcFile.Close()
		return s.copyFileViaShell(srcPath, destPath)
	}

	if _, err := io.Copy(destFile, srcFile); err != nil {
		srcFile.Close()
		destFile.Close()
		return s.copyFileViaShell(srcPath, destPath)
	}
	srcFile.Close()
	destFile.Close()

	_ = s.sftpClient.Chmod(destPath, srcInfo.Mode())
	return nil
}

// copyFileViaShell 用 shell 命令复制（SFTP 权限不足时的 fallback）。
func (s *SSHClient) copyFileViaShell(srcPath, destPath string) error {
	cmd := fmt.Sprintf("cp -p %s %s 2>&1", srcPath, destPath)
	result, err := s.ExecuteCommand(cmd)
	if err == nil && result.Success {
		return nil
	}

	password := s.config.Password
	if password == "" {
		return fmt.Errorf("权限不足且未配置密码，无法复制文件")
	}

	encodedPw := base64.StdEncoding.EncodeToString([]byte(password))
	cmd = fmt.Sprintf("echo %s | base64 -d | sudo -S cp -p %s %s 2>&1",
		encodedPw, srcPath, destPath)
	result, err = s.ExecuteCommand(cmd)
	if err != nil {
		return fmt.Errorf("sudo cp 执行失败: %v", err)
	}
	if !result.Success {
		errMsg := strings.TrimSpace(result.Stderr)
		if errMsg == "" {
			errMsg = strings.TrimSpace(result.Stdout)
		}
		return fmt.Errorf("sudo cp 失败: %s", errMsg)
	}
	return nil
}

// UploadFileFromBytes 从字节数组上传文件（支持断点续传）
// 如果远程文件已存在且小于本地数据大小，自动从断点继续写入
func (s *SSHClient) UploadFileFromBytes(remotePath string, data []byte, progressCallback func(percent int)) error {
	if s.sftpClient == nil {
		if err := s.InitSFTP(); err != nil {
			return err
		}
	}

	totalSize := len(data)
	offset := int64(0)

	// 检查远程文件是否已存在（断点续传）
	if stat, err := s.sftpClient.Stat(remotePath); err == nil {
		existingSize := stat.Size()
		if existingSize >= int64(totalSize) {
			// 远程文件已完整，跳过上传
			if progressCallback != nil {
				progressCallback(100)
			}
			return nil
		}
		offset = existingSize
	}

	var remoteFile *sftp.File
	var err error

	if offset > 0 {
		// 追加模式打开
		remoteFile, err = s.sftpClient.OpenFile(remotePath, os.O_WRONLY|os.O_APPEND)
	} else {
		// 新建文件
		remoteFile, err = s.sftpClient.Create(remotePath)
	}

	if err != nil {
		return s.uploadViaSudo(remotePath, data, progressCallback)
	}
	defer remoteFile.Close()

	// 从断点位置开始写入
	startPos := int(offset)
	chunkSize := 32 * 1024
	written := startPos
	lastReported := 0

	// 计算初始进度
	if progressCallback != nil && startPos > 0 {
		lastReported = startPos * 100 / totalSize
		progressCallback(lastReported)
	}

	for written < totalSize {
		end := written + chunkSize
		if end > totalSize {
			end = totalSize
		}

		n, err := remoteFile.Write(data[written:end])
		if err != nil {
			return fmt.Errorf("写入远程文件失败 (已写入 %d/%d bytes): %v", written, totalSize, err)
		}
		written += n

		if progressCallback != nil {
			percent := written * 100 / totalSize
			if percent >= lastReported+5 || percent >= 100 {
				lastReported = percent
				progressCallback(percent)
			}
		}
	}

	// 验证最终文件大小
	fileInfo, err := s.sftpClient.Stat(remotePath)
	if err != nil {
		return fmt.Errorf("验证文件失败: %v", err)
	}
	if fileInfo.Size() != int64(totalSize) {
		return fmt.Errorf("文件大小不匹配: 期望 %d bytes, 实际 %d bytes", totalSize, fileInfo.Size())
	}

	return nil
}

// DownloadFileToBytes 下载远程文件并返回字节数组
func (s *SSHClient) DownloadFileToBytes(remotePath string) ([]byte, error) {
	if s.sftpClient == nil {
		if err := s.InitSFTP(); err != nil {
			return nil, err
		}
	}

	remoteFile, err := s.sftpClient.Open(remotePath)
	if err != nil {
		return nil, fmt.Errorf("打开远程文件失败: %v", err)
	}
	defer remoteFile.Close()

	data, err := io.ReadAll(remoteFile)
	if err != nil {
		return nil, fmt.Errorf("读取远程文件失败: %v", err)
	}
	return data, nil
}

// UploadDirectory 上传目录（递归）
func (s *SSHClient) UploadDirectory(localPath, remotePath string) error {
	if s.sftpClient == nil {
		if err := s.InitSFTP(); err != nil {
			return err
		}
	}

	if err := s.sftpClient.MkdirAll(remotePath); err != nil {
		// SFTP 权限不足，尝试 sudo
		if mkdirErr := s.mkdirViaSudo(remotePath); mkdirErr != nil {
			return fmt.Errorf("创建远程目录失败: %v", mkdirErr)
		}
	}

	entries, err := os.ReadDir(localPath)
	if err != nil {
		return fmt.Errorf("读取本地目录失败: %v", err)
	}

	for _, entry := range entries {
		localFilePath := localPath + "/" + entry.Name()
		remoteFilePath := remotePath + "/" + entry.Name()

		if entry.IsDir() {
			if err := s.UploadDirectory(localFilePath, remoteFilePath); err != nil {
				return err
			}
		} else {
			if err := s.UploadFile(localFilePath, remoteFilePath); err != nil {
				return err
			}
		}
	}
	return nil
}

// CreateArchive 创建压缩包
func (s *SSHClient) CreateArchive(files []string, archiveName string) (string, error) {
	if len(files) == 0 {
		return "", fmt.Errorf("没有文件需要压缩")
	}

	tempArchive := fmt.Sprintf("/tmp/%s.tar.gz", archiveName)

	cmd := fmt.Sprintf("tar czf '%s'", tempArchive)
	for _, file := range files {
		cmd += fmt.Sprintf(" '%s'", file)
	}

	result, err := s.ExecuteCommand(cmd)
	if err != nil {
		return "", fmt.Errorf("执行压缩命令失败: %v", err)
	}
	if !result.Success {
		return "", fmt.Errorf("压缩失败: %s", result.Stderr)
	}

	return tempArchive, nil
}

// DeleteTempFile 删除临时文件
func (s *SSHClient) DeleteTempFile(filePath string) error {
	cmd := fmt.Sprintf("rm -f '%s'", filePath)
	result, err := s.ExecuteCommand(cmd)
	if err != nil {
		return fmt.Errorf("删除临时文件失败: %v", err)
	}
	if !result.Success {
		return fmt.Errorf("删除失败: %s", result.Stderr)
	}
	return nil
}

// ExtractArchive 解压压缩包
func (s *SSHClient) ExtractArchive(archivePath string, targetDir string) error {
	var cmd string
	switch {
	case strings.HasSuffix(archivePath, ".zip"):
		cmd = fmt.Sprintf("cd '%s' && unzip -o '%s'", targetDir, archivePath)
	case strings.HasSuffix(archivePath, ".tar.gz") || strings.HasSuffix(archivePath, ".tgz"):
		cmd = fmt.Sprintf("cd '%s' && tar xzf '%s'", targetDir, archivePath)
	case strings.HasSuffix(archivePath, ".tar"):
		cmd = fmt.Sprintf("cd '%s' && tar xf '%s'", targetDir, archivePath)
	default:
		return fmt.Errorf("不支持的压缩格式: %s", archivePath)
	}

	result, err := s.ExecuteCommand(cmd)
	if err != nil {
		return fmt.Errorf("执行解压命令失败: %v", err)
	}
	if !result.Success {
		return fmt.Errorf("解压失败: %s", result.Stderr)
	}
	return nil
}

// deleteViaSudo 通过 sudo 删除文件（SFTP 权限不足时的回退方案）
func (s *SSHClient) deleteViaSudo(remotePath string, isDir bool) error {
	password := s.config.Password
	if password == "" {
		return fmt.Errorf("权限不足且未配置密码，无法删除 %s", remotePath)
	}

	rmFlag := "-f"
	if isDir {
		rmFlag = "-rf"
	}

	encodedPw := base64.StdEncoding.EncodeToString([]byte(password))
	cmd := fmt.Sprintf("echo %s | base64 -d | sudo -S rm %s %s 2>&1",
		encodedPw, rmFlag, remotePath)
	result, err := s.ExecuteCommand(cmd)
	if err != nil {
		return fmt.Errorf("sudo rm 执行失败: %v", err)
	}
	if !result.Success {
		errMsg := strings.TrimSpace(result.Stderr)
		if errMsg == "" {
			errMsg = strings.TrimSpace(result.Stdout)
		}
		return fmt.Errorf("权限不足，无法删除 %s: %s", remotePath, errMsg)
	}
	return nil
}

// uploadViaSudo 通过 sudo 上传文件（SFTP 权限不足时的回退方案）
// 先上传到 /tmp，再用 sudo mv 到目标位置
func (s *SSHClient) uploadViaSudo(remotePath string, data []byte, progressCallback func(percent int)) error {
	// 1. 上传到 /tmp
	tmpPath := fmt.Sprintf("/tmp/pzssh_upload_%d", time.Now().UnixNano())
	tmpFile, err := s.sftpClient.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %v", err)
	}

	// 写入数据
	written := 0
	totalSize := len(data)
	for written < totalSize {
		end := written + 32*1024
		if end > totalSize {
			end = totalSize
		}
		n, err := tmpFile.Write(data[written:end])
		if err != nil {
			tmpFile.Close()
			s.sftpClient.Remove(tmpPath)
			return fmt.Errorf("写入临时文件失败: %v", err)
		}
		written += n
		if progressCallback != nil {
			progressCallback(written * 100 / totalSize)
		}
	}
	tmpFile.Close()

	// 2. 用 sudo mv 移动到目标位置
	password := s.config.Password
	if password == "" {
		s.sftpClient.Remove(tmpPath)
		return fmt.Errorf("权限不足且未配置密码，无法上传到 %s", remotePath)
	}

	encodedPw := base64.StdEncoding.EncodeToString([]byte(password))
	cmd := fmt.Sprintf("echo %s | base64 -d | sudo -S mv %s %s 2>&1",
		encodedPw, tmpPath, remotePath)
	result, err := s.ExecuteCommand(cmd)
	if err != nil {
		s.sftpClient.Remove(tmpPath)
		return fmt.Errorf("sudo mv 执行失败: %v", err)
	}
	if !result.Success {
		s.sftpClient.Remove(tmpPath)
		errMsg := strings.TrimSpace(result.Stderr)
		if errMsg == "" {
			errMsg = strings.TrimSpace(result.Stdout)
		}
		return fmt.Errorf("sudo mv 失败: %s", errMsg)
	}

	return nil
}
