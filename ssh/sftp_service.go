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
		return fmt.Errorf("创建远程文件失败: %v", err)
	}
	defer remoteFile.Close()

	if _, err := io.Copy(remoteFile, localFile); err != nil {
		return fmt.Errorf("上传文件失败: %v", err)
	}
	return nil
}

// DownloadFile 下载文件
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

	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("创建本地文件失败: %v", err)
	}
	defer localFile.Close()

	if _, err := io.Copy(localFile, remoteFile); err != nil {
		return fmt.Errorf("下载文件失败: %v", err)
	}
	return nil
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

	if isSymlink {
		return s.sftpClient.Remove(remotePath)
	} else if fileInfo.IsDir() {
		return s.removeDir(remotePath)
	}
	return s.sftpClient.Remove(remotePath)
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
	return s.sftpClient.MkdirAll(remotePath)
}

// RenameFile 重命名文件/目录
func (s *SSHClient) RenameFile(oldPath, newPath string) error {
	if s.sftpClient == nil {
		if err := s.InitSFTP(); err != nil {
			return err
		}
	}
	return s.sftpClient.Rename(oldPath, newPath)
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
//
// 安全修复：使用 base64 编码传递密码，避免 printf 格式化字符串漏洞。
func (s *SSHClient) copyFileViaShell(srcPath, destPath string) error {
	cmd := fmt.Sprintf("cp -p %s %s 2>&1", srcPath, destPath)
	result, err := s.ExecuteCommand(cmd)
	if err == nil && result.Success {
		return nil
	}

	password := s.config.Password
	if password == "" {
		destDir := destPath[:strings.LastIndex(destPath, "/")]
		return fmt.Errorf("权限不足且未配置密码，无法提权。请在服务器上执行：sudo chmod 777 %s", destDir)
	}

	// 安全修复：用 base64 编码密码，通过管道传给 sudo -S
	// 避免 printf '%s\n' '%s' 的格式化字符串漏洞
	encodedPw := base64.StdEncoding.EncodeToString([]byte(password))
	cmd = fmt.Sprintf("echo %s | base64 -d | sudo -S cp -p %s %s 2>/dev/null",
		encodedPw, srcPath, destPath)
	result, err = s.ExecuteCommand(cmd)
	if err == nil && result.Success {
		return nil
	}

	destDir := destPath[:strings.LastIndex(destPath, "/")]
	return fmt.Errorf("权限不足，无法写入 %s。请在服务器上执行：sudo chmod 777 %s", destDir, destDir)
}

// UploadFileFromBytes 从字节数组上传文件
func (s *SSHClient) UploadFileFromBytes(remotePath string, data []byte, progressCallback func(percent int)) error {
	if s.sftpClient == nil {
		if err := s.InitSFTP(); err != nil {
			return err
		}
	}

	remoteFile, err := s.sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("创建远程文件失败: %v", err)
	}
	defer remoteFile.Close()

	totalSize := len(data)
	chunkSize := 32 * 1024
	written := 0
	lastReported := 0

	for written < totalSize {
		end := written + chunkSize
		if end > totalSize {
			end = totalSize
		}

		n, err := remoteFile.Write(data[written:end])
		if err != nil {
			return fmt.Errorf("写入远程文件失败: %v", err)
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

	fileInfo, err := s.sftpClient.Stat(remotePath)
	if err != nil {
		return fmt.Errorf("验证文件失败: %v", err)
	}
	if fileInfo.Size() != int64(len(data)) {
		return fmt.Errorf("文件大小不匹配: 期望 %d bytes, 实际 %d bytes", len(data), fileInfo.Size())
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
		return fmt.Errorf("创建远程目录失败: %v", err)
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
