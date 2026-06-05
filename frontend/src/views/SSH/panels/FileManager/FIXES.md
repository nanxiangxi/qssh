# 编辑器问题修复说明

## 修复的问题

### ✅ 保存功能不正常 - 已修复

**问题原因**：
- `FilePreview.vue` 中的保存按钮直接触发 `emit('save')`
- 但没有从 CodeEditor 组件获取最新编辑的内容
- 导致保存的是旧内容，而不是用户编辑后的内容

**解决方案**：
在 `FilePreview.vue` 的 `handleSave` 函数中：
1. 先从 CodeEditor 组件获取最新内容（使用 `getValue()` 方法）
2. 通过 `emit('update:content', newContent)` 更新父组件的内容
3. 然后再触发保存事件

**修改的文件**：
- ✅ `src/views/SSH/panels/FileManager/components/FilePreview.vue`

**修复后的保存流程**：
```javascript
const handleSave = async () => {
  if (saving.value) return
  
  saving.value = true
  try {
    // ✅ 从 CodeEditor 获取最新内容
    if (codeEditorRef.value && props.type === 'text') {
      const newContent = codeEditorRef.value.getValue()
      // 先更新父组件的 content
      emit('update:content', newContent)
    }
    // 然后触发保存事件
    await emit('save')
    isModified.value = false
    originalContent = props.content
  } finally {
    saving.value = false
  }
}
```

## 测试验证

### 保存功能测试
1. 打开一个文本文件
2. 点击"编辑"按钮进入编辑模式
3. 修改文件内容
4. 点击"保存"按钮或按 Ctrl+S
5. 关闭编辑器后重新打开文件
6. 验证修改是否已保存

## 技术细节

### Vue 组件通信

文件编辑器的数据流：

```
FileManagerPanel.vue (父组件)
    ↓ props: content
FilePreview.vue (对话框)
    ↓ props: content
CodeEditor.vue (编辑器组件)
    ↑ emit: update:content (用户编辑时)
    ↑ getValue() (保存时获取最新内容)
```

**关键点**：
- 编辑时：CodeEditor 通过 `update:content` 事件实时同步内容
- 保存时：FilePreview 主动调用 `getValue()` 获取最新内容，确保数据一致

## 相关文件

- `src/views/SSH/panels/FileManager/components/CodeEditor.vue` - 编辑器组件
- `src/views/SSH/panels/FileManager/components/FilePreview.vue` - 预览/编辑对话框
- `src/views/SSH/panels/FileManager/FileManagerPanel.vue` - 文件管理器主组件

## 关于中文本地化

Monaco Editor 在 Vite + ESM 环境下的中文本地化存在技术挑战，详见 [CHINESE_LOCALIZATION.md](./CHINESE_LOCALIZATION.md)。

当前采用英文界面，但通过中文按钮和提示提供了良好的用户体验。
