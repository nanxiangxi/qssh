# FileManager 模块重构说明

## 📁 目录结构

```
FileManager/
├── FileManagerPanel.vue          # 主容器组件（协调各子组件）
├── components/                   # 子组件目录
│   ├── FileToolbar.vue           # 工具栏组件（导航、上传等）
│   ├── FileTable.vue             # 文件列表表格组件
│   ├── ActionMenu.vue            # 三点操作菜单组件
│   ├── FilePreview.vue           # 文件预览/编辑组件
│   └── PermissionEditor.vue      # 权限编辑器组件
├── composables/                  # 组合式函数目录
│   ├── useFileOperations.js      # 文件操作逻辑（下载、上传、删除等）
│   ├── usePermissions.js         # 权限管理逻辑
│   └── useNavigation.js          # 导航和排序逻辑
└── utils/                        # 工具函数目录
    └── pathHelper.js             # 路径处理工具
```

## 🎯 设计原则

### 1. **单一职责原则**
每个组件和 composable 只负责一个功能领域：
- `FileToolbar` - 只负责工具栏 UI
- `useFileOperations` - 只负责文件操作业务逻辑
- `PermissionEditor` - 只负责权限编辑 UI

### 2. **关注点分离**
- **UI 层**：Vue 组件负责渲染和用户交互
- **逻辑层**：Composables 封装业务逻辑
- **工具层**：纯函数提供通用功能

### 3. **可维护性**
- 每个文件不超过 300 行代码
- 清晰的模块边界
- 易于单元测试

### 4. **可复用性**
- Composables 可在其他组件中复用
- 工具函数独立于 Vue
- 组件通过 props 和 events 通信

## 📦 核心模块说明

### Components（组件）

#### FileToolbar.vue
**职责**：渲染文件管理器工具栏
- 上一级按钮
- 刷新按钮
- 路径输入框
- 粘贴按钮（条件显示）
- 新建文件夹按钮
- 上传按钮

**Props**：
- `currentPath`: 当前路径
- `hasCutFile`: 是否有剪切的文件

**Events**：
- `@go-up`: 向上一级
- `@refresh`: 刷新
- `@navigate`: 导航到路径
- `@paste`: 粘贴
- `@new-folder`: 新建文件夹
- `@upload`: 触发上传

---

#### FileTable.vue
**职责**：渲染文件列表表格
- 表头（可排序）
- 文件行（双击打开）
- 三点菜单按钮

**Props**：
- `files`: 文件列表
- `loading`: 加载状态
- `sortKey`: 排序键
- `sortOrder`: 排序方向

**Events**：
- `@open-file`: 打开文件/目录
- `@sort`: 排序
- `@show-menu`: 显示操作菜单

---

#### ActionMenu.vue
**职责**：三点操作菜单
- 预览/编辑
- 下载
- 压缩为 ZIP
- 重命名
- 创建副本
- 剪切
- 修改权限
- 删除

**Props**：
- `visible`: 是否显示
- `position`: 菜单位置 {x, y}
- `file`: 目标文件

**Events**：
- `@preview`, `@download`, `@compress`, etc.

---

#### FilePreview.vue
**职责**：文件预览和编辑
- 图片预览
- 文本查看/编辑
- 保存功能

**Props**：
- `visible`: 是否显示
- `title`: 标题
- `content`: 内容
- `type`: 类型 ('image' | 'text')
- `editable`: 是否可编辑

**Events**：
- `@close`: 关闭
- `@enable-edit`: 启用编辑
- `@save`: 保存

---

#### PermissionEditor.vue
**职责**：图形化权限编辑
- 三个区域：所有者、用户组、公共
- 复选框控制读/写/执行
- 实时计算数字权限
- 预设权限按钮
- 应用到子目录选项

**Props**：
- `showChmodDialog`: 是否显示
- `chmodTarget`: 目标文件
- `permissions`: 权限状态对象
- `numericValue`: 数字权限值
- `applyToSubdirs`: 是否应用到子目录

**Events**：
- `@confirm`: 确认修改
- `@close`: 关闭
- `@set-preset`: 设置预设

---

### Composables（组合式函数）

#### useFileOperations.js
**职责**：封装所有文件操作的业务逻辑

**提供的功能**：
- `uploadFiles(files)` - 批量上传
- `downloadFile(file)` - 下载文件/目录
- `deleteFile(file)` - 删除
- `renameFile(file, newName)` - 重命名
- `duplicateFile(file)` - 创建副本
- `compressFile(file)` - 压缩为 ZIP
- `cutFile(file)` - 剪切
- `pasteFile()` - 粘贴

**返回值**：
```javascript
{
  loading,
  uploadFiles,
  downloadFile,
  deleteFile,
  renameFile,
  duplicateFile,
  compressFile,
  cutFile,
  pasteFile,
  hasCutFile  // computed
}
```

---

#### usePermissions.js
**职责**：权限管理逻辑

**提供的功能**：
- `parsePermissions(numericMode)` - 解析权限
- `calculateNumericMode()` - 计算数字权限
- `setPreset(value)` - 设置预设
- `openChmodDialog(file)` - 打开对话框
- `confirmChmod(connId)` - 确认修改
- `closeChmodDialog()` - 关闭对话框

**状态**：
```javascript
{
  showChmodDialog,
  chmodValue,
  chmodTarget,
  applyToSubdirs,
  chmodPermissions  // { owner: {r,w,x}, group: {...}, other: {...} }
}
```

---

#### useNavigation.js
**职责**：导航和排序逻辑

**提供的功能**：
- `goUp()` - 向上一级
- `navigateToPath()` - 导航到路径
- `openFile(file)` - 打开文件/目录
- `sortBy(key)` - 排序
- `sortedFiles` - 排序后的文件列表（computed）

---

### Utils（工具函数）

#### pathHelper.js
**职责**：路径处理纯函数

**提供的函数**：
- `normalizePath(path)` - 规范化路径
- `joinPath(basePath, fileName)` - 拼接路径
- `getParentPath(currentPath)` - 获取父目录
- `formatSize(bytes)` - 格式化文件大小
- `formatTime(timestamp)` - 格式化时间
- `getNumericMode(modeStr)` - 权限字符串转数字

---

## 🔄 数据流示例

### 下载文件流程

```
用户点击下载
  ↓
ActionMenu.vue 发射 @download 事件
  ↓
FileManagerPanel.vue 接收事件，调用 downloadFile(file)
  ↓
downloadFile 来自 useFileOperations composable
  ↓
useFileOperations 调用 SSHService.DownloadFile()
  ↓
创建 Blob 和下载链接
  ↓
触发浏览器下载
  ↓
显示成功消息
```

### 修改权限流程

```
用户点击"修改权限"
  ↓
ActionMenu.vue 发射 @chmod 事件
  ↓
FileManagerPanel.vue 调用 permissionState.openChmodDialog(file)
  ↓
usePermissions 解析权限并打开对话框
  ↓
PermissionEditor.vue 渲染复选框界面
  ↓
用户勾选复选框 → watch 自动计算数字权限
  ↓
用户点击"应用" → 发射 @confirm
  ↓
FileManagerPanel.vue 调用 permissionState.confirmChmod(connId)
  ↓
usePermissions 执行 chmod 命令
  ↓
重新加载文件列表
```

---

## ✅ 优势对比

### 重构前（单文件 1584 行）
❌ 所有逻辑混在一起  
❌ 难以定位问题  
❌ 修改风险高  
❌ 无法复用逻辑  
❌ 测试困难  

### 重构后（模块化）
✅ 清晰的职责划分  
✅ 问题容易定位  
✅ 修改影响范围小  
✅ Composables 可复用  
✅ 每个模块可独立测试  

---

## 🚀 后续优化建议

1. **添加更多子组件**
   - `FileToolbar.vue` - 拆分工具栏
   - `FileTable.vue` - 拆分表格
   - `ActionMenu.vue` - 拆分菜单

2. **增强 Composables**
   - 添加错误重试机制
   - 添加进度追踪
   - 添加缓存策略

3. **单元测试**
   - 测试 `pathHelper.js` 纯函数
   - 测试 Composables 逻辑
   - Mock SSHService

4. **性能优化**
   - 虚拟滚动大文件列表
   - 懒加载文件图标
   - 防抖搜索

5. **TypeScript 迁移**
   - 添加类型定义
   - 提高代码安全性
   - 更好的 IDE 支持

---

## 📝 使用示例

```vue
<template>
  <FileManagerPanel />
</template>

<script setup>
import FileManagerPanel from './FileManager/FileManagerPanel.vue'
</script>
```

就这么简单！所有复杂性都被封装在模块内部。
