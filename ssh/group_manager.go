package ssh

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// GroupManager 分组管理器
type GroupManager struct {
	mu       sync.RWMutex
	groups   map[string]*SSHGroup
	nextID   int
	dataFile string
}

// SSHGroup SSH分组
type SSHGroup struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	ParentID  string   `json:"parent_id,omitempty"` // 父分组ID，空=顶级分组
	Color     string   `json:"color,omitempty"`     // 分组颜色标记
	Order     int      `json:"order"`               // 排序序号
	ConnIDs   []string `json:"conn_ids"`
	WindowID  string   `json:"window_id"`
	IsDefault bool     `json:"is_default"`
}

// NewGroupManager 创建分组管理器
func NewGroupManager() *GroupManager {
	dataFile := filepath.Join(GetDataDir(), "groups.json")

	gm := &GroupManager{
		groups:   make(map[string]*SSHGroup),
		nextID:   2,
		dataFile: dataFile,
	}

	// 从文件加载分组
	if err := gm.loadGroups(); err != nil {
		log.Error("加载分组失败，使用默认分组", "error", err)
	}

	// 确保默认分组存在
	if _, exists := gm.groups["group-1"]; !exists {
		gm.groups["group-1"] = &SSHGroup{
			ID:        "group-1",
			Name:      "默认分组",
			ConnIDs:   make([]string, 0),
			IsDefault: true,
			Order:     0,
		}
	}

	return gm
}

// loadGroups 从文件加载分组
func (gm *GroupManager) loadGroups() error {
	data, err := os.ReadFile(gm.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var groups []*SSHGroup
	if err := json.Unmarshal(data, &groups); err != nil {
		return err
	}

	for _, g := range groups {
		if g.ConnIDs == nil {
			g.ConnIDs = make([]string, 0)
		}
		gm.groups[g.ID] = g
	}

	return nil
}

// saveGroups 保存分组到文件
func (gm *GroupManager) saveGroups() error {
	dir := filepath.Dir(gm.dataFile)
	os.MkdirAll(dir, 0755)

	var groups []*SSHGroup
	for _, g := range gm.groups {
		groups = append(groups, g)
	}

	data, err := json.MarshalIndent(groups, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(gm.dataFile, data, 0644)
}

// CreateGroup 创建新分组
func (gm *GroupManager) CreateGroup(name string) *SSHGroup {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	
	groupID := fmt.Sprintf("group-%d", gm.nextID)
	gm.nextID++
	
	return gm.createGroupInternal(groupID, name, false)
}

// CreateGroupWithName 使用指定ID和名称创建分组（用于重新创建已删除的分组）
func (gm *GroupManager) CreateGroupWithName(groupID, name string) *SSHGroup {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	
	// 检查是否已存在
	if _, exists := gm.groups[groupID]; exists {
		log.Info("分组已存在，无需重新创建", "id", groupID)
		return gm.groups[groupID]
	}
	
	isDefault := (groupID == "group-1")
	return gm.createGroupInternal(groupID, name, isDefault)
}

// createGroupInternal 内部方法：创建分组（需要已持有锁）
func (gm *GroupManager) createGroupInternal(groupID, name string, isDefault bool) *SSHGroup {
	// 计算排序序号（追加到末尾）
	maxOrder := 0
	for _, g := range gm.groups {
		if g.Order > maxOrder {
			maxOrder = g.Order
		}
	}

	group := &SSHGroup{
		ID:        groupID,
		Name:      name,
		ConnIDs:   make([]string, 0),
		IsDefault: isDefault,
		Order:     maxOrder + 1,
	}

	gm.groups[groupID] = group
	log.Info("创建分组", "id", groupID, "name", name, "default", isDefault)

	// 更新nextID（如果是自动生成的ID）
	if !isDefault && len(groupID) > 6 {
		var num int
		fmt.Sscanf(groupID, "group-%d", &num)
		if num >= gm.nextID {
			gm.nextID = num + 1
		}
	}

	gm.saveGroups()
	return group
}

// GetGroup 获取分组
func (gm *GroupManager) GetGroup(groupID string) *SSHGroup {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	
	return gm.groups[groupID]
}

// GetAllGroups 获取所有分组
func (gm *GroupManager) GetAllGroups() []*SSHGroup {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	
	var groups []*SSHGroup
	for _, group := range gm.groups {
		groups = append(groups, group)
	}
	
	return groups
}

// AddConnectionToGroup 将连接添加到分组
func (gm *GroupManager) AddConnectionToGroup(groupID, connID string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return fmt.Errorf("分组不存在: %s", groupID)
	}

	group.ConnIDs = append(group.ConnIDs, connID)
	log.Info("添加连接到分组", "group", groupID, "conn", connID)
	gm.saveGroups()
	return nil
}

// RemoveConnectionFromGroup 从分组中移除连接
func (gm *GroupManager) RemoveConnectionFromGroup(groupID, connID string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return fmt.Errorf("分组不存在: %s", groupID)
	}

	// 移除connID
	for i, id := range group.ConnIDs {
		if id == connID {
			group.ConnIDs = append(group.ConnIDs[:i], group.ConnIDs[i+1:]...)
			break
		}
	}

	log.Info("从分组移除连接", "group", groupID, "conn", connID, "remaining", len(group.ConnIDs))

	// 如果没有连接了，删除分组（包括默认分组）
	if len(group.ConnIDs) == 0 {
		delete(gm.groups, groupID)
		log.Info("删除空分组", "id", groupID)
	}

	gm.saveGroups()
	return nil
}

// GetGroupByConnID 根据连接ID查找分组
func (gm *GroupManager) GetGroupByConnID(connID string) *SSHGroup {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	
	for _, group := range gm.groups {
		for _, id := range group.ConnIDs {
			if id == connID {
				return group
			}
		}
	}
	
	return nil
}

// GetDefaultGroup 获取默认分组
func (gm *GroupManager) GetDefaultGroup() *SSHGroup {
	return gm.GetGroup("group-1")
}

// ClearGroup 清空分组中的所有连接（但不删除分组）
func (gm *GroupManager) ClearGroup(groupID string) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	group, exists := gm.groups[groupID]
	if !exists {
		log.Warn("分组不存在，无法清空", "id", groupID)
		return
	}

	count := len(group.ConnIDs)
	group.ConnIDs = make([]string, 0)
	log.Info("已清空分组", "id", groupID, "removed", count)
	gm.saveGroups()
}

// DeleteGroup 删除分组（包括所有连接）
func (gm *GroupManager) DeleteGroup(groupID string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return fmt.Errorf("分组不存在: %s", groupID)
	}

	// 不允许删除默认分组
	if group.IsDefault {
		log.Warn("尝试删除默认分组，改为清空连接")
		group.ConnIDs = make([]string, 0)
		gm.saveGroups()
		return nil
	}

	delete(gm.groups, groupID)
	log.Info("已删除分组", "id", groupID, "connCount", len(group.ConnIDs))
	gm.saveGroups()
	return nil
}

// UpdateGroup 更新分组属性（名称、颜色、父分组）
func (gm *GroupManager) UpdateGroup(groupID, name, color, parentID string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return fmt.Errorf("分组不存在: %s", groupID)
	}

	if name != "" {
		group.Name = name
	}
	if color != "" {
		group.Color = color
	}
	if parentID != "" && parentID != groupID {
		group.ParentID = parentID
	}

	log.Info("更新分组", "id", groupID, "name", group.Name, "color", group.Color)
	gm.saveGroups()
	return nil
}

// ReorderGroups 更新分组排序
func (gm *GroupManager) ReorderGroups(groupIDs []string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	for i, id := range groupIDs {
		if group, exists := gm.groups[id]; exists {
			group.Order = i
		}
	}

	gm.saveGroups()
	return nil
}

// GetChildGroups 获取指定分组的子分组
func (gm *GroupManager) GetChildGroups(parentID string) []*SSHGroup {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	var children []*SSHGroup
	for _, g := range gm.groups {
		if g.ParentID == parentID {
			children = append(children, g)
		}
	}
	return children
}
