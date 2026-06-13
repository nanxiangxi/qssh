<template>
  <Modal
    :visible="showChmodDialog"
    title="修改权限"
    width="700px"
    @confirm="$emit('confirm')"
    @close="handleClose"
    @cancel="handleCancel"
  >
    <div v-if="chmodTarget" class="chmod-dialog">
      <p class="chmod-file-name">{{ chmodTarget.name }}</p>
      
      <!-- 权限复选框区域 -->
      <div class="chmod-sections">
        <!-- 所有者 -->
        <fieldset class="chmod-section">
          <legend>所有者</legend>
          <div class="checkbox-group">
            <label class="checkbox-item">
              <input type="checkbox" :checked="getSafePermission('owner', 'r')" @change="$emit('update:permission', { group: 'owner', perm: 'r', value: $event.target.checked })">
              <span>读取</span>
            </label>
            <label class="checkbox-item">
              <input type="checkbox" :checked="getSafePermission('owner', 'w')" @change="$emit('update:permission', { group: 'owner', perm: 'w', value: $event.target.checked })">
              <span>写入</span>
            </label>
            <label class="checkbox-item">
              <input type="checkbox" :checked="getSafePermission('owner', 'x')" @change="$emit('update:permission', { group: 'owner', perm: 'x', value: $event.target.checked })">
              <span>执行</span>
            </label>
          </div>
        </fieldset>
        
        <!-- 用户组 -->
        <fieldset class="chmod-section">
          <legend>用户组</legend>
          <div class="checkbox-group">
            <label class="checkbox-item">
              <input type="checkbox" :checked="getSafePermission('group', 'r')" @change="$emit('update:permission', { group: 'group', perm: 'r', value: $event.target.checked })">
              <span>读取</span>
            </label>
            <label class="checkbox-item">
              <input type="checkbox" :checked="getSafePermission('group', 'w')" @change="$emit('update:permission', { group: 'group', perm: 'w', value: $event.target.checked })">
              <span>写入</span>
            </label>
            <label class="checkbox-item">
              <input type="checkbox" :checked="getSafePermission('group', 'x')" @change="$emit('update:permission', { group: 'group', perm: 'x', value: $event.target.checked })">
              <span>执行</span>
            </label>
          </div>
        </fieldset>
        
        <!-- 公共 -->
        <fieldset class="chmod-section">
          <legend>公共</legend>
          <div class="checkbox-group">
            <label class="checkbox-item">
              <input type="checkbox" :checked="getSafePermission('other', 'r')" @change="$emit('update:permission', { group: 'other', perm: 'r', value: $event.target.checked })">
              <span>读取</span>
            </label>
            <label class="checkbox-item">
              <input type="checkbox" :checked="getSafePermission('other', 'w')" @change="$emit('update:permission', { group: 'other', perm: 'w', value: $event.target.checked })">
              <span>写入</span>
            </label>
            <label class="checkbox-item">
              <input type="checkbox" :checked="getSafePermission('other', 'x')" @change="$emit('update:permission', { group: 'other', perm: 'x', value: $event.target.checked })">
              <span>执行</span>
            </label>
          </div>
        </fieldset>
      </div>
      
      <!-- 数字权限显示和预设按钮 -->
      <div class="chmod-footer">
        <div class="chmod-number-input">
          <label>权限值：</label>
          <input 
            :value="numericValue" 
            class="dialog-input chmod-input"
            readonly
          />
        </div>
        
        <!-- 只有在选择目录时才显示此选项 -->
        <label v-if="chmodTarget?.isDir" class="checkbox-item apply-to-subdirs">
          <input type="checkbox" :checked="applyToSubdirs" @change="$emit('update:applyToSubdirs', $event.target.checked)">
          <span>应用到子目录</span>
        </label>
      </div>
    </div>
  </Modal>
</template>

<script setup>
import { computed, watch } from 'vue'
import Modal from '@/components/Modal.vue'

const props = defineProps({
  showChmodDialog: {
    type: Boolean,
    required: true
  },
  chmodTarget: {
    type: Object,
    default: null
  },
  permissions: {
    type: Object,
    default: () => ({
      owner: { r: false, w: false, x: false },
      group: { r: false, w: false, x: false },
      other: { r: false, w: false, x: false }
    })
  },
  numericValue: {
    type: String,
    default: '755'
  },
  applyToSubdirs: {
    type: Boolean,
    default: false
  }
})

// 调试日志 - 监控 props 变化
watch(() => props.showChmodDialog, (newVal, oldVal) => {
  console.log('[PermissionEditor] showChmodDialog 变化:', { old: oldVal, new: newVal })
}, { immediate: true })

watch(() => props.chmodTarget, (newVal) => {
  console.log('[PermissionEditor] chmodTarget 变化:', newVal ? newVal.name : null)
}, { immediate: true })

const emit = defineEmits(['confirm', 'close', 'update:showChmodDialog', 'update:applyToSubdirs', 'update:permission'])

// 安全的权限访问 - 使用可选链和默认值
const getSafePermission = (group, perm) => {
  return props.permissions?.[group]?.[perm] ?? false
}

const handleClose = () => {
  emit('update:showChmodDialog', false)
}

const handleCancel = () => {
  emit('update:showChmodDialog', false)
}
</script>

<style scoped>
.chmod-dialog {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.chmod-file-name {
  color: var(--primary-light);
  font-weight: 600;
  font-size: 1rem;
  padding: 0.5rem;
  background: var(--primary-bg);
  border-radius: 0.375rem;
  text-align: center;
}

/* 权限复选框区域 */
.chmod-sections {
  display: flex;
  justify-content: space-around;
  gap: 1rem;
  padding: 1rem 0;
}

.chmod-section {
  flex: 1;
  padding: 1rem;
  background: var(--surface-2);
  border: 1px solid var(--surface-hover);
  border-radius: 0.5rem;
}

.chmod-section legend {
  color: var(--text-primary);
  font-weight: 600;
  font-size: 0.9rem;
  padding: 0 0.5rem;
}

.checkbox-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-top: 0.5rem;
}

.checkbox-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  color: var(--text-secondary);
  font-size: 0.875rem;
  transition: color 0.2s;
}

.checkbox-item:hover {
  color: var(--text-primary);
}

.checkbox-item input[type="checkbox"] {
  width: 1rem;
  height: 1rem;
  cursor: pointer;
  accent-color: var(--accent-primary);
}

/* 底部区域 */
.chmod-footer {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding-top: 1rem;
  border-top: 1px solid var(--surface-hover);
}

.chmod-number-input {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.chmod-number-input label {
  color: var(--text-secondary);
  font-size: 0.875rem;
}

.chmod-input {
  width: 6rem;
  padding: 0.5rem 0.75rem;
  background: var(--bg-input);
  border: 1px solid var(--border-strong);
  border-radius: 0.375rem;
  color: var(--text-primary);
  font-size: 1rem;
  font-family: monospace;
  text-align: center;
}

.apply-to-subdirs {
  margin-left: 1rem;
  padding-top: 0.5rem;
}
</style>
