package ssh

import (
	"fmt"
	"sync"
)

// GroupManager 分组管理器
type GroupManager struct {
	mu     sync.RWMutex
	groups map[string]*SSHGroup
	nextID int
}

// SSHGroup SSH分组
type SSHGroup struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	ConnIDs   []string `json:"conn_ids"`
	WindowID  string   `json:"window_id"`
	IsDefault bool     `json:"is_default"`
}

// NewGroupManager 创建分组管理器
func NewGroupManager() *GroupManager {
	gm := &GroupManager{
		groups: make(map[string]*SSHGroup),
		nextID: 2, // 从2开始，因为id-1是默认分组
	}
	
	// 创建默认分组
	gm.groups["group-1"] = &SSHGroup{
		ID:        "group-1",
		Name:      "默认分组",
		ConnIDs:   make([]string, 0),
		IsDefault: true,
	}
	
	return gm
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
		fmt.Printf("[GroupManager] 分组 %s 已存在，无需重新创建\n", groupID)
		return gm.groups[groupID]
	}
	
	isDefault := (groupID == "group-1")
	return gm.createGroupInternal(groupID, name, isDefault)
}

// createGroupInternal 内部方法：创建分组（需要已持有锁）
func (gm *GroupManager) createGroupInternal(groupID, name string, isDefault bool) *SSHGroup {
	group := &SSHGroup{
		ID:        groupID,
		Name:      name,
		ConnIDs:   make([]string, 0),
		IsDefault: isDefault,
	}
	
	gm.groups[groupID] = group
	fmt.Printf("[GroupManager] 创建分组: %s, 名称: %s, 默认: %v\n", groupID, name, isDefault)
	
	// 更新nextID（如果是自动生成的ID）
	if !isDefault && len(groupID) > 6 {
		// 提取ID中的数字部分
		var num int
		fmt.Sscanf(groupID, "group-%d", &num)
		if num >= gm.nextID {
			gm.nextID = num + 1
		}
	}
	
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
	fmt.Printf("[GroupManager] 添加连接到分组: group=%s, conn=%s\n", groupID, connID)
	
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
	
	fmt.Printf("[GroupManager] 从分组移除连接: group=%s, conn=%s, 剩余连接数: %d\n", 
		groupID, connID, len(group.ConnIDs))
	
	// 如果没有连接了，删除分组（包括默认分组）
	if len(group.ConnIDs) == 0 {
		delete(gm.groups, groupID)
		fmt.Printf("[GroupManager] 删除空分组: %s\n", groupID)
	}
	
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
		fmt.Printf("[GroupManager] ⚠️ 分组不存在，无法清空: %s\n", groupID)
		return
	}
	
	count := len(group.ConnIDs)
	group.ConnIDs = make([]string, 0)
	fmt.Printf("[GroupManager] 🧹 已清空分组: %s (移除 %d 个连接)\n", groupID, count)
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
		fmt.Printf("[GroupManager] ⚠️ 尝试删除默认分组，改为清空连接\n")
		group.ConnIDs = make([]string, 0)
		return nil
	}
	
	delete(gm.groups, groupID)
	fmt.Printf("[GroupManager] 🗑️ 已删除分组: %s (连接数: %d)\n", groupID, len(group.ConnIDs))
	
	return nil
}
