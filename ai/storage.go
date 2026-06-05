package ai

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Project 代表一个 SSH 服务器项目
type Project struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Session 代表一个 AI 对话会话
type Session struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ChatSession 会话信息（存储用）
type ChatSession struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ChatStore 持久化存储管理器
type ChatStore struct {
	mu       sync.RWMutex
	baseDir  string
	projects map[string]*Project
	sessions map[string]*ChatSession
}

// NewChatStore 创建存储管理器
func NewChatStore(baseDir string) (*ChatStore, error) {
	store := &ChatStore{
		baseDir:  filepath.Join(baseDir, "ai"),
		projects: make(map[string]*Project),
		sessions: make(map[string]*ChatSession),
	}

	// 确保目录存在
	dirs := []string{
		store.baseDir,
		filepath.Join(store.baseDir, "projects"),
		filepath.Join(store.baseDir, "sessions"),
		filepath.Join(store.baseDir, "history"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("创建目录 %s 失败: %v", dir, err)
		}
	}

	// 加载已有数据
	store.loadProjects()
	store.loadSessions()

	return store, nil
}

// === 项目管理 ===

// CreateProject 创建项目
func (cs *ChatStore) CreateProject(name, host string, port int, username string) *Project {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	project := &Project{
		ID:        fmt.Sprintf("proj_%d", time.Now().UnixNano()),
		Name:      name,
		Host:      host,
		Port:      port,
		Username:  username,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	cs.projects[project.ID] = project
	cs.saveProject(project)
	return project
}

// GetProject 获取项目
func (cs *ChatStore) GetProject(id string) *Project {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.projects[id]
}

// GetOrCreateProjectByHost 根据主机地址获取或创建项目
func (cs *ChatStore) GetOrCreateProjectByHost(host string, port int, username string) *Project {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// 查找已有项目
	for _, p := range cs.projects {
		if p.Host == host && p.Port == port && p.Username == username {
			p.UpdatedAt = time.Now()
			cs.saveProject(p)
			return p
		}
	}

	// 创建新项目
	project := &Project{
		ID:        fmt.Sprintf("proj_%d", time.Now().UnixNano()),
		Name:      fmt.Sprintf("%s@%s", username, host),
		Host:      host,
		Port:      port,
		Username:  username,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	cs.projects[project.ID] = project
	cs.saveProject(project)
	return project
}

// ListProjects 列出所有项目
func (cs *ChatStore) ListProjects() []*Project {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	projects := make([]*Project, 0, len(cs.projects))
	for _, p := range cs.projects {
		projects = append(projects, p)
	}
	return projects
}

// === 会话管理 ===

// CreateSession 创建会话
func (cs *ChatStore) CreateSession(projectID, title string) *ChatSession {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	session := &ChatSession{
		ID:        fmt.Sprintf("sess_%d", time.Now().UnixNano()),
		ProjectID: projectID,
		Title:     title,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	cs.sessions[session.ID] = session
	cs.saveSession(session)
	return session
}

// GetSessionsByProject 获取项目的所有会话
func (cs *ChatStore) GetSessionsByProject(projectID string) []*ChatSession {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	var sessions []*ChatSession
	for _, s := range cs.sessions {
		if s.ProjectID == projectID {
			sessions = append(sessions, s)
		}
	}
	return sessions
}

// UpdateSessionTitle 更新会话标题
func (cs *ChatStore) UpdateSessionTitle(sessionID, title string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if s, ok := cs.sessions[sessionID]; ok {
		s.Title = title
		s.UpdatedAt = time.Now()
		cs.saveSession(s)
	}
}

// === 聊天历史 ===

// SaveMessages 保存完整消息历史（含 tool 上下文）
func (cs *ChatStore) SaveMessages(key string, messages []map[string]interface{}) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	dir := filepath.Join(cs.baseDir, "history")
	filePath := filepath.Join(dir, key+".json")

	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化消息历史失败: %v", err)
	}

	return os.WriteFile(filePath, data, 0644)
}

// LoadMessages 加载完整消息历史
func (cs *ChatStore) LoadMessages(key string) ([]map[string]interface{}, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	filePath := filepath.Join(cs.baseDir, "history", key+".json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("读取消息历史失败: %v", err)
	}

	var messages []map[string]interface{}
	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, fmt.Errorf("解析消息历史失败: %v", err)
	}

	return messages, nil
}

// DeleteMessages 删除消息历史
func (cs *ChatStore) DeleteMessages(key string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	filePath := filepath.Join(cs.baseDir, "history", key+".json")
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除消息历史失败: %v", err)
	}
	return nil
}

// === 内部方法 ===

func (cs *ChatStore) saveProject(p *Project) {
	dir := filepath.Join(cs.baseDir, "projects")
	filePath := filepath.Join(dir, p.ID+".json")
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		fmt.Printf("[ChatStore] 序列化项目失败: %v\n", err)
		return
	}
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		fmt.Printf("[ChatStore] 保存项目失败: %v\n", err)
	}
}

func (cs *ChatStore) saveSession(s *ChatSession) {
	dir := filepath.Join(cs.baseDir, "sessions")
	filePath := filepath.Join(dir, s.ID+".json")
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		fmt.Printf("[ChatStore] 序列化会话失败: %v\n", err)
		return
	}
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		fmt.Printf("[ChatStore] 保存会话失败: %v\n", err)
	}
}

func (cs *ChatStore) loadProjects() {
	dir := filepath.Join(cs.baseDir, "projects")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}
		var p Project
		if err := json.Unmarshal(data, &p); err != nil {
			continue
		}
		cs.projects[p.ID] = &p
	}
	fmt.Printf("[ChatStore] 加载了 %d 个项目\n", len(cs.projects))
}

func (cs *ChatStore) loadSessions() {
	dir := filepath.Join(cs.baseDir, "sessions")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}
		var s ChatSession
		if err := json.Unmarshal(data, &s); err != nil {
			continue
		}
		cs.sessions[s.ID] = &s
	}
	fmt.Printf("[ChatStore] 加载了 %d 个会话\n", len(cs.sessions))
}
