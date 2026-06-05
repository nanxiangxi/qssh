package ssh

import (
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

// InitSFTP 初始化SFTP客户端
func (s *SSHClient) InitSFTP() error {
	if !s.isConnected {
		return fmt.Errorf("未连接到SSH服务器")
	}

	if s.sftpClient != nil {
		return nil // 已经初始化
	}

	client, err := sftp.NewClient(s.client)
	if err != nil {
		return fmt.Errorf("初始化SFTP失败: %v", err)
	}

	s.sftpClient = client
	return nil
}

// UploadFile 上传文件到远程服务器
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

	_, err = io.Copy(remoteFile, localFile)
	if err != nil {
		return fmt.Errorf("上传文件失败: %v", err)
	}

	return nil
}

// DownloadFile 从远程服务器下载文件
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

	_, err = io.Copy(localFile, remoteFile)
	if err != nil {
		return fmt.Errorf("下载文件失败: %v", err)
	}

	return nil
}

// ListDirectory 列出远程目录内容
func (s *SSHClient) ListDirectory(remotePath string) ([]FileInfo, error) {
	if s.sftpClient == nil {
		if err := s.InitSFTP(); err != nil {
			return nil, err
		}
	}

	// 规范化路径
	if remotePath == "" {
		remotePath = "/"
	}

	files, err := s.sftpClient.ReadDir(remotePath)
	if err != nil {
		// 检查是否是目录不存在的错误
		if strings.Contains(err.Error(), "file does not exist") || strings.Contains(err.Error(), "no such file") {
			return nil, fmt.Errorf("目录不存在: %s", remotePath)
		}
		return nil, fmt.Errorf("读取目录失败: %v", err)
	}

	var fileInfos []FileInfo
	for _, file := range files {
		// 正确拼接路径
		fullPath := remotePath
		if !strings.HasSuffix(remotePath, "/") {
			fullPath += "/"
		}
		fullPath += file.Name()
		
		// 获取所有者信息
		owner, group := s.getFileOwner(fullPath)
		
		// 使用 Lstat 获取文件信息，不跟随符号链接
		fileStat, err := s.sftpClient.Lstat(fullPath)
		isDir := false
		if err == nil && fileStat != nil {
			isDir = fileStat.IsDir()
		} else {
			// 如果 Lstat 失败，使用 ReadDir 返回的信息
			isDir = file.IsDir()
		}

		fileInfos = append(fileInfos, FileInfo{
			Name:    file.Name(),
			Path:    fullPath,
			Size:    file.Size(),
			Mode:    file.Mode().String(),
			ModTime: file.ModTime(),
			IsDir:   isDir,
			Owner:   owner,
			Group:   group,
		})
	}

	return fileInfos, nil
}

// SearchFiles 递归搜索文件（支持子目录，流式返回）
func (s *SSHClient) SearchFiles(basePath string, keyword string, app *application.App, searchID string) error {
	if s.sftpClient == nil {
		if err := s.InitSFTP(); err != nil {
			return err
		}
	}

	keyword = strings.ToLower(keyword)
	foundCount := 0
	
	// 注册搜索 ID
	s.searchMutex.Lock()
	s.searchCancelMap[searchID] = false
	s.searchMutex.Unlock()
	
	fmt.Printf("[SFTP] ========== 开始流式搜索 ==========\n")
	fmt.Printf("[SFTP] 搜索 ID: %s\n", searchID)
	fmt.Printf("[SFTP] 关键词: %s\n", keyword)
	fmt.Printf("[SFTP] 基础路径: %s\n", basePath)
	
	// 发送开始事件
	if app != nil {
		fmt.Printf("[SFTP] 发送 search-start 事件\n")
		app.Event.Emit("search-start", map[string]interface{}{
			"searchID": searchID,
			"basePath": basePath,
			"keyword":  keyword,
		})
	}
	
	// 递归搜索函数
	var searchRecursive func(path string) error
	searchRecursive = func(path string) error {
		// 检查是否被取消
		s.searchMutex.RLock()
		cancelled := s.searchCancelMap[searchID]
		s.searchMutex.RUnlock()
		
		if cancelled {
			fmt.Printf("[SFTP] 搜索已取消: %s\n", searchID)
			return fmt.Errorf("搜索已取消")
		}
		
		fmt.Printf("[SFTP] 正在读取目录: %s\n", path)
		files, err := s.sftpClient.ReadDir(path)
		if err != nil {
			fmt.Printf("[SFTP] ⚠️ 读取目录失败 %s: %v\n", path, err)
			return err
		}
		fmt.Printf("[SFTP] 目录 %s 包含 %d 个文件\n", path, len(files))

		for _, file := range files {
			// 再次检查取消标志
			s.searchMutex.RLock()
			cancelled := s.searchCancelMap[searchID]
			s.searchMutex.RUnlock()
			
			if cancelled {
				return fmt.Errorf("搜索已取消")
			}
			
			// 跳过 . 和 ..
			if file.Name() == "." || file.Name() == ".." {
				continue
			}

			fullPath := path
			if !strings.HasSuffix(path, "/") {
				fullPath += "/"
			}
			fullPath += file.Name()

			// 检查文件名是否匹配
			if strings.Contains(strings.ToLower(file.Name()), keyword) {
				// 使用 Lstat 获取准确信息
				fileStat, err := s.sftpClient.Lstat(fullPath)
				isDir := false
				if err == nil && fileStat != nil {
					isDir = fileStat.IsDir()
				} else {
					isDir = file.IsDir()
				}

				// 计算相对于 basePath 的路径
				relativePath := fullPath
				if strings.HasPrefix(fullPath, basePath) {
					relativePath = strings.TrimPrefix(fullPath, basePath)
					relativePath = strings.TrimPrefix(relativePath, "/")
				}

				fileInfo := FileInfo{
					Name:    file.Name(),
					Path:    fullPath,
					Size:    file.Size(),
					Mode:    file.Mode().String(),
					ModTime: file.ModTime(),
					IsDir:   isDir,
					Owner:   "-",
					Group:   "-",
				}

				// 发送找到的文件（包含相对路径）
				if app != nil {
					foundCount++
					fmt.Printf("[SFTP] 找到匹配文件 #%d: %s (路径: %s, 相对: %s)\n", foundCount, file.Name(), fullPath, relativePath)
					app.Event.Emit("search-result", map[string]interface{}{
						"searchID": searchID,
						"file":     fileInfo,
						"relativePath": relativePath, // 添加相对路径字段
					})
				}
			}

			// 如果是目录，递归搜索
			if file.IsDir() {
				fmt.Printf("[SFTP] 进入子目录: %s\n", fullPath)
				if err := searchRecursive(fullPath); err != nil {
					// 如果是取消导致的错误，直接返回
					if strings.Contains(err.Error(), "搜索已取消") {
						return err
					}
					// 记录错误但继续搜索其他目录
					fmt.Printf("[SFTP] ⚠️ 搜索目录失败 %s: %v\n", fullPath, err)
				}
			}
		}

		return nil
	}

	// 开始递归搜索
	fmt.Printf("[SFTP] 开始递归搜索...\n")
	err := searchRecursive(basePath)
	
	// 清理搜索 ID
	s.searchMutex.Lock()
	delete(s.searchCancelMap, searchID)
	s.searchMutex.Unlock()
	
	// 发送完成事件
	if app != nil {
		fmt.Printf("[SFTP] 搜索完成，共找到 %d 个文件\n", foundCount)
		app.Event.Emit("search-complete", map[string]interface{}{
			"searchID":   searchID,
			"totalFound": foundCount,
			"error":      err != nil,
		})
	}

	if err != nil {
		// 如果是取消导致的，不返回错误
		if strings.Contains(err.Error(), "搜索已取消") {
			fmt.Printf("[SFTP] 搜索已被用户取消\n")
			return nil
		}
		return fmt.Errorf("搜索失败: %v", err)
	}

	return nil
}

// CancelSearch 取消正在进行的搜索
func (s *SSHClient) CancelSearch(searchID string) {
	s.searchMutex.Lock()
	defer s.searchMutex.Unlock()
	
	if _, exists := s.searchCancelMap[searchID]; exists {
		s.searchCancelMap[searchID] = true
		fmt.Printf("[SFTP] 已标记搜索为取消: %s\n", searchID)
	}
}

// getFileOwner 获取文件所有者和组
func (s *SSHClient) getFileOwner(path string) (string, string) {
	// 使用 stat 命令代替 ls，更可靠
	cmd := fmt.Sprintf("stat -c '%%U %%G' '%s'", path)
	result, err := s.ExecuteCommand(cmd)
	if err != nil || !result.Success {
		fmt.Printf("[SFTP] ⚠️ 获取文件所有者失败: %v\n", err)
		return "unknown", "unknown"
	}

	fields := strings.Fields(result.Stdout)
	if len(fields) >= 2 {
		return fields[0], fields[1]
	}

	return "unknown", "unknown"
}

// DeleteFile 删除远程文件或目录
func (s *SSHClient) DeleteFile(remotePath string) error {
	fmt.Printf("[SFTP] ========== DeleteFile ==========\n")
	fmt.Printf("[SFTP] remotePath: %s\n", remotePath)
	
	if s.sftpClient == nil {
		fmt.Printf("[SFTP] 初始化 SFTP 客户端...\n")
		if err := s.InitSFTP(); err != nil {
			fmt.Printf("[SFTP] ⚠️ SFTP 初始化失败: %v\n", err)
			return err
		}
	}

	// 使用 Lstat 而不是 Stat，不跟随符号链接
	fmt.Printf("[SFTP] 获取文件信息（Lstat）...\n")
	fileInfo, err := s.sftpClient.Lstat(remotePath)
	if err != nil {
		fmt.Printf("[SFTP] ⚠️ 获取文件信息失败: %v\n", err)
		return fmt.Errorf("获取文件信息失败: %v", err)
	}

	fmt.Printf("[SFTP] 文件信息:\n")
	fmt.Printf("  - Name: %s\n", fileInfo.Name())
	fmt.Printf("  - Size: %d\n", fileInfo.Size())
	fmt.Printf("  - Mode: %s\n", fileInfo.Mode())
	fmt.Printf("  - IsDir: %v\n", fileInfo.IsDir())
	
	// 检查是否是符号链接
	isSymlink := fileInfo.Mode()&os.ModeSymlink != 0
	fmt.Printf("  - IsSymlink: %v\n", isSymlink)
	
	if isSymlink {
		// 符号链接：直接删除链接本身
		fmt.Printf("[SFTP] ✓ 这是一个符号链接，直接删除\n")
		err = s.sftpClient.Remove(remotePath)
		if err != nil {
			fmt.Printf("[SFTP] ⚠️ 删除失败: %v\n", err)
			return fmt.Errorf("删除失败: %v", err)
		}
		fmt.Printf("[SFTP] ✓ 删除成功\n")
	} else if fileInfo.IsDir() {
		// 目录：递归删除
		fmt.Printf("[SFTP] ⚠️ 这是一个目录，使用 SFTP 递归删除\n")
		err = s.removeDir(remotePath)
		if err != nil {
			fmt.Printf("[SFTP] ⚠️ 删除目录失败: %v\n", err)
			return fmt.Errorf("删除目录失败: %v", err)
		}
		fmt.Printf("[SFTP] ✓ 目录删除成功\n")
	} else {
		// 普通文件
		fmt.Printf("[SFTP] ✓ 这是一个文件\n")
		err = s.sftpClient.Remove(remotePath)
		if err != nil {
			fmt.Printf("[SFTP] ⚠️ 删除失败: %v\n", err)
			return fmt.Errorf("删除失败: %v", err)
		}
		fmt.Printf("[SFTP] ✓ 删除成功\n")
	}

	return nil
}

// removeDir 递归删除目录
func (s *SSHClient) removeDir(path string) error {
	fmt.Printf("[SFTP] removeDir 开始: path=%s\n", path)
	
	files, err := s.sftpClient.ReadDir(path)
	if err != nil {
		fmt.Printf("[SFTP] ⚠️ ReadDir 失败: %v\n", err)
		return err
	}
	
	fmt.Printf("[SFTP] 目录中有 %d 个文件/子目录\n", len(files))

	for i, file := range files {
		// 跳过 . 和 ..
		if file.Name() == "." || file.Name() == ".." {
			fmt.Printf("[SFTP] [%d/%d] 跳过: %s\n", i+1, len(files), file.Name())
			continue
		}
		
		// 正确拼接路径
		fullPath := path + "/" + file.Name()
		
		fmt.Printf("[SFTP] [%d/%d] 处理: %s (isDir=%v)\n", i+1, len(files), fullPath, file.IsDir())
		
		if file.IsDir() {
			fmt.Printf("[SFTP] 递归删除子目录: %s\n", fullPath)
			err = s.removeDir(fullPath)
		} else {
			fmt.Printf("[SFTP] 删除文件: %s\n", fullPath)
			err = s.sftpClient.Remove(fullPath)
		}
		
		if err != nil {
			fmt.Printf("[SFTP] ⚠️ 删除失败: %v\n", err)
			return err
		}
		fmt.Printf("[SFTP] ✓ 删除成功\n")
	}

	fmt.Printf("[SFTP] 所有子文件已删除，现在删除目录本身: %s\n", path)
	err = s.sftpClient.RemoveDirectory(path)
	if err != nil {
		fmt.Printf("[SFTP] ⚠️ RemoveDirectory 失败: %v\n", err)
		return err
	}
	
	fmt.Printf("[SFTP] ✓ 目录删除成功\n")
	return nil
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
		return fmt.Errorf("创建目录失败: %v", err)
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
		return fmt.Errorf("重命名失败: %v", err)
	}

	return nil
}

// UploadFileFromBytes 从字节数组上传文件（分块写入，支持进度回调）
func (s *SSHClient) UploadFileFromBytes(remotePath string, data []byte, progressCallback func(percent int)) error {
	fmt.Printf("[SFTP] UploadFileFromBytes: path=%s, size=%d bytes, hasCallback=%v\n", remotePath, len(data), progressCallback != nil)

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
	chunkSize := 32 * 1024 // 32KB per chunk
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

		// 报告进度（每 5% 报告一次，避免事件过多）
		if progressCallback != nil {
			percent := written * 100 / totalSize
			if percent >= lastReported+5 || percent >= 100 {
				lastReported = percent
				fmt.Printf("[SFTP] 上传进度: %d%%\n", percent)
				progressCallback(percent)
			}
		}
	}

	// 验证文件完整性
	fileInfo, err := s.sftpClient.Stat(remotePath)
	if err != nil {
		fmt.Printf("[SFTP] ⚠️ 验证文件失败: %v\n", err)
		return fmt.Errorf("验证文件失败: %v", err)
	}
	
	if fileInfo.Size() != int64(len(data)) {
		fmt.Printf("[SFTP] ⚠️ 文件大小不匹配: expected=%d, actual=%d\n", len(data), fileInfo.Size())
		return fmt.Errorf("文件大小不匹配: 期望 %d bytes, 实际 %d bytes", len(data), fileInfo.Size())
	}
	
	fmt.Printf("[SFTP] ✓ 上传成功并验证通过，文件大小: %d bytes\n", written)
	return nil
}

// DownloadFileToBytes 下载远程文件并返回字节数组
func (s *SSHClient) DownloadFileToBytes(remotePath string) ([]byte, error) {
	fmt.Printf("[SFTP] ========== DownloadFileToBytes ==========\n")
	fmt.Printf("[SFTP] remotePath: %s\n", remotePath)
	
	if s.sftpClient == nil {
		fmt.Printf("[SFTP] 初始化 SFTP 客户端...\n")
		if err := s.InitSFTP(); err != nil {
			fmt.Printf("[SFTP] ⚠️ SFTP 初始化失败: %v\n", err)
			return nil, err
		}
	}

	fmt.Printf("[SFTP] 打开远程文件...\n")
	remoteFile, err := s.sftpClient.Open(remotePath)
	if err != nil {
		fmt.Printf("[SFTP] ⚠️ 打开远程文件失败: %v\n", err)
		return nil, fmt.Errorf("打开远程文件失败: %v", err)
	}
	defer remoteFile.Close()

	fmt.Printf("[SFTP] 读取文件内容...\n")
	data, err := io.ReadAll(remoteFile)
	if err != nil {
		fmt.Printf("[SFTP] ⚠️ 读取远程文件失败: %v\n", err)
		return nil, fmt.Errorf("读取远程文件失败: %v", err)
	}

	fmt.Printf("[SFTP] ✓ 下载成功，文件大小: %d bytes\n", len(data))
	return data, nil
}

// UploadDirectory 上传目录（递归）
func (s *SSHClient) UploadDirectory(localPath, remotePath string) error {
	if s.sftpClient == nil {
		if err := s.InitSFTP(); err != nil {
			return err
		}
	}

	// 创建远程目录
	err := s.sftpClient.MkdirAll(remotePath)
	if err != nil {
		return fmt.Errorf("创建远程目录失败: %v", err)
	}

	// 读取本地目录
	entries, err := os.ReadDir(localPath)
	if err != nil {
		return fmt.Errorf("读取本地目录失败: %v", err)
	}

	for _, entry := range entries {
		localFilePath := localPath + "/" + entry.Name()
		remoteFilePath := remotePath + "/" + entry.Name()

		if entry.IsDir() {
			// 递归上传子目录
			err = s.UploadDirectory(localFilePath, remoteFilePath)
			if err != nil {
				return err
			}
		} else {
			// 上传文件
			err = s.UploadFile(localFilePath, remoteFilePath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// CreateArchive 创建压缩包（tar.gz）并返回临时文件路径
func (s *SSHClient) CreateArchive(files []string, archiveName string) (string, error) {
	if len(files) == 0 {
		return "", fmt.Errorf("没有文件需要压缩")
	}

	// 生成临时文件名
	tempDir := "/tmp"
	tempArchive := fmt.Sprintf("%s/%s.tar.gz", tempDir, archiveName)

	// 构建 tar 命令（允许顶级目录存在）
	cmd := fmt.Sprintf("tar czf '%s'", tempArchive)
	for _, file := range files {
		cmd += fmt.Sprintf(" '%s'", file)
	}

	fmt.Printf("[SFTP] 执行压缩命令: %s\n", cmd)
	result, err := s.ExecuteCommand(cmd)
	if err != nil {
		return "", fmt.Errorf("执行压缩命令失败: %v", err)
	}

	if !result.Success {
		return "", fmt.Errorf("压缩失败: %s", result.Stderr)
	}

	fmt.Printf("[SFTP] ✓ 压缩成功: %s\n", tempArchive)
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

	fmt.Printf("[SFTP] ✓ 临时文件已删除: %s\n", filePath)
	return nil
}

// ExtractArchive 解压压缩包到指定目录
func (s *SSHClient) ExtractArchive(archivePath string, targetDir string) error {
	// 判断压缩格式
	var cmd string
	if strings.HasSuffix(archivePath, ".zip") {
		// 解压 zip 文件
		cmd = fmt.Sprintf("cd '%s' && unzip -o '%s'", targetDir, archivePath)
	} else if strings.HasSuffix(archivePath, ".tar.gz") || strings.HasSuffix(archivePath, ".tgz") {
		// 解压 tar.gz 文件
		cmd = fmt.Sprintf("cd '%s' && tar xzf '%s'", targetDir, archivePath)
	} else if strings.HasSuffix(archivePath, ".tar") {
		// 解压 tar 文件
		cmd = fmt.Sprintf("cd '%s' && tar xf '%s'", targetDir, archivePath)
	} else {
		return fmt.Errorf("不支持的压缩格式: %s", archivePath)
	}

	fmt.Printf("[SFTP] 执行解压命令: %s\n", cmd)
	result, err := s.ExecuteCommand(cmd)
	if err != nil {
		return fmt.Errorf("执行解压命令失败: %v", err)
	}

	if !result.Success {
		return fmt.Errorf("解压失败: %s", result.Stderr)
	}

	fmt.Printf("[SFTP] ✓ 解压成功: %s -> %s\n", archivePath, targetDir)
	return nil
}

