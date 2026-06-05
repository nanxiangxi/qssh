# Modal 组件使用指南

## 🎯 概述

Modal 是一个功能强大的对话框组件，支持高度自定义的样式、布局和行为。

---

## 📦 基础用法

### 1. 简单确认对话框

```vue
<template>
  <Modal
    v-model:visible="showDialog"
    title="确认操作"
    content="确定要删除这个文件吗？"
    @confirm="handleConfirm"
    @cancel="handleCancel"
  />
</template>

<script setup>
import { ref } from 'vue'
import Modal from '@/components/Modal.vue'

const showDialog = ref(false)

const handleConfirm = () => {
  console.log('用户确认了')
}

const handleCancel = () => {
  console.log('用户取消了')
}
</script>
```

---

## 🎨 样式自定义

### 2. 自定义圆角和颜色

```vue
<Modal
  v-model:visible="showDialog"
  title="自定义样式"
  :border-radius="20"
  background-color="rgba(30, 30, 40, 0.98)"
  border-color="rgba(66, 153, 225, 0.3)"
  overlay-opacity="0.8"
  overlay-blur="8"
/>
```

**可用样式属性：**
- `borderRadius` - 圆角大小（支持数字或字符串）
- `backgroundColor` - 背景颜色
- `borderColor` - 边框颜色
- `overlayOpacity` - 遮罩透明度 (0-1)
- `overlayBlur` - 遮罩模糊程度
- `overlayColor` - 遮罩颜色

---

## 📐 尺寸控制

### 3. 响应式尺寸

```vue
<!-- 固定尺寸 -->
<Modal
  width="600px"
  height="400px"
/>

<!-- 百分比尺寸 -->
<Modal
  width="80%"
  height="70%"
/>

<!-- 限制最大最小尺寸 -->
<Modal
  :min-width="300"
  :max-width="800"
  :min-height="200"
  :max-height="600"
/>
```

---

## 🎭 显示控制

### 4. 隐藏标题栏

```vue
<Modal
  :show-header="false"
  :show-close="false"
  :show-footer="false"
>
  <div class="custom-content">
    完全自定义的内容
  </div>
</Modal>
```

### 5. 自定义按钮

```vue
<Modal
  cancel-text="稍后提醒"
  confirm-text="立即执行"
  loading-text="执行中..."
  :danger="true"
/>
```

---

## 🎬 动画效果

### 6. 多种过渡动画

```vue
<!-- 默认滑动动画 -->
<Modal transition-name="modal" />

<!-- 淡入淡出 -->
<Modal transition-name="fade" />

<!-- 从底部滑入 -->
<Modal transition-name="slide-up" />

<!-- 缩放动画 -->
<Modal transition-name="zoom" />

<!-- 自定义动画时长 -->
<Modal :animation-duration="300" />
```

---

## 🔧 高级功能

### 7. 使用插槽自定义

```vue
<Modal v-model:visible="showDialog">
  <!-- 自定义标题 -->
  <template #header>
    <div class="custom-header">
      <svg>...</svg>
      <h3>自定义标题</h3>
    </div>
  </template>
  
  <!-- 自定义内容 -->
  <template #default>
    <div class="custom-body">
      <p>任意 HTML 内容</p>
      <input v-model="inputValue" />
    </div>
  </template>
  
  <!-- 自定义底部按钮 -->
  <template #footer>
    <div class="custom-footer">
      <button @click="handleSkip">跳过</button>
      <button @click="handleBack">上一步</button>
      <button @click="handleNext">下一步</button>
    </div>
  </template>
</Modal>
```

### 8. 按钮布局

```vue
<!-- 按钮居右（默认） -->
<Modal button-layout="right" />

<!-- 按钮居中 -->
<Modal button-layout="center" />

<!-- 按钮居左 -->
<Modal button-layout="left" />

<!-- 按钮两端对齐 -->
<Modal button-layout="space-between" />
```

---

## 💡 实际应用场景

### 9. 文件预览对话框

```vue
<Modal
  v-model:visible="showPreview"
  :title="fileName"
  width="90%"
  height="85%"
  :border-radius="8"
  :show-cancel="false"
  confirm-text="关闭"
>
  <CodeEditor
    :content="fileContent"
    :language="detectedLanguage"
    :readonly="true"
  />
</Modal>
```

### 10. 危险操作确认

```vue
<Modal
  v-model:visible="showDeleteConfirm"
  title="⚠️ 危险操作"
  content="此操作不可恢复，确定要删除吗？"
  danger
  cancel-text="取消"
  confirm-text="确认删除"
  @confirm="deleteFile"
/>
```

### 11. 加载状态

```vue
<Modal
  v-model:visible="showLoading"
  title="处理中"
  content="请稍候..."
  :show-cancel="false"
  :show-confirm="false"
  :close-on-overlay="false"
  :close-on-escape="false"
>
  <div class="loading-spinner">
    <div class="spinner"></div>
    <p>正在上传文件...</p>
  </div>
</Modal>
```

### 12. 表单对话框

```vue
<Modal
  v-model:visible="showForm"
  title="新建文件夹"
  width="400px"
  @confirm="createFolder"
>
  <div class="form-group">
    <label>文件夹名称：</label>
    <input 
      v-model="folderName" 
      placeholder="请输入文件夹名称"
      @keyup.enter="createFolder"
    />
  </div>
</Modal>
```

---

## 📋 Props 完整列表

### 基础配置
| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `visible` | Boolean | false | 是否显示 |
| `title` | String | '提示' | 标题 |
| `content` | String | '' | 内容文本 |

### 尺寸配置
| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `width` | String/Number | '400px' | 宽度 |
| `height` | String/Number | 'auto' | 高度 |
| `maxWidth` | String/Number | '90vw' | 最大宽度 |
| `maxHeight` | String/Number | '90vh' | 最大高度 |
| `minWidth` | String/Number | '300px' | 最小宽度 |
| `minHeight` | String/Number | 'auto' | 最小高度 |

### 样式配置
| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `borderRadius` | String/Number | '12px' | 圆角大小 |
| `overlayOpacity` | Number | 0.7 | 遮罩透明度 |
| `overlayBlur` | Number | 4 | 遮罩模糊 |
| `overlayColor` | String | 'rgba(0,0,0,0.7)' | 遮罩颜色 |
| `backgroundColor` | String | 'rgba(45,45,45,0.98)' | 背景色 |
| `borderColor` | String | 'rgba(255,255,255,0.15)' | 边框色 |

### 显示控制
| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `showHeader` | Boolean | true | 显示标题栏 |
| `showClose` | Boolean | true | 显示关闭按钮 |
| `showFooter` | Boolean | true | 显示底部 |
| `showCancel` | Boolean | true | 显示取消按钮 |
| `showConfirm` | Boolean | true | 显示确认按钮 |

### 按钮配置
| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `cancelText` | String | '取消' | 取消按钮文本 |
| `confirmText` | String | '确定' | 确认按钮文本 |
| `loadingText` | String | '处理中...' | 加载文本 |
| `closeTooltip` | String | '关闭' | 关闭提示 |
| `confirmDisabled` | Boolean | false | 禁用确认按钮 |
| `danger` | Boolean | false | 危险操作样式 |

### 布局配置
| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `buttonLayout` | String | 'right' | 按钮布局 |

### 交互配置
| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `closeOnOverlay` | Boolean | true | 点击遮罩关闭 |
| `closeOnEscape` | Boolean | true | ESC 键关闭 |
| `draggable` | Boolean | false | 可拖动 |
| `resizable` | Boolean | false | 可调整大小 |

### 动画配置
| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `transitionName` | String | 'modal' | 过渡动画名 |
| `animationDuration` | Number | 200 | 动画时长(ms) |

### 自定义类名
| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `overlayClass` | String/Array/Object | '' | 遮罩类名 |
| `containerClass` | String/Array/Object | '' | 容器类名 |
| `headerClass` | String/Array/Object | '' | 标题栏类名 |
| `bodyClass` | String/Array/Object | '' | 内容区类名 |
| `footerClass` | String/Array/Object | '' | 底部类名 |
| `cancelButtonClass` | String/Array/Object | '' | 取消按钮类名 |
| `confirmButtonClass` | String/Array/Object | '' | 确认按钮类名 |

---

## 🎯 Events

| 事件名 | 参数 | 说明 |
|--------|------|------|
| `close` | - | 关闭时触发 |
| `cancel` | - | 取消时触发 |
| `confirm` | - | 确认时触发 |
| `update:visible` | Boolean | visible 变化时触发 |

---

## 🎨 Slots

| 插槽名 | 说明 |
|--------|------|
| `default` | 默认内容区 |
| `header` | 自定义标题栏 |
| `close-icon` | 自定义关闭图标 |
| `footer` | 自定义底部按钮区 |

---

## ✨ 最佳实践

1. **大尺寸对话框**：使用百分比尺寸 + maxHeight/maxWidth
2. **移动端适配**：设置 maxWidth 为 90vw
3. **长内容**：确保 bodyClass 包含 overflow-y: auto
4. **自定义样式**：优先使用 props，必要时使用 class
5. **性能优化**：避免频繁创建销毁，使用 v-if 控制显示

---

## 🔗 相关组件

- [CodeEditor](../views/SSH/panels/FileManager/components/CodeEditor.vue) - 代码编辑器
- [FilePreview](../views/SSH/panels/FileManager/components/FilePreview.vue) - 文件预览
