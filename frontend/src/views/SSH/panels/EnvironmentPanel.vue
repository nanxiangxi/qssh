<template>
  <div class="env-panel">
    <!-- 工具栏 -->
    <div class="env-toolbar">
      <button class="env-btn" @click="detect" :disabled="loading">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
        </svg>
        {{ loading ? t('env.refreshing') : t('env.refresh') }}
      </button>
      <input v-model="search" class="env-search" :placeholder="t('env.searchPlaceholder')" />
      <span v-if="!loading && envs.length > 0" class="env-count">
        {{ t('env.detected', { count: installedCount }) }}
      </span>
    </div>

    <!-- 加载中 -->
    <div v-if="loading" class="env-loading">
      <div class="spinner"></div>
      <span>{{ t('env.refreshing') }}</span>
    </div>

    <!-- 错误状态 -->
    <div v-else-if="error" class="env-empty">
      <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" opacity="0.4">
        <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
      </svg>
      <span>{{ error }}</span>
      <button class="env-btn" @click="detect" style="margin-top:0.5rem">{{ t('env.retry') }}</button>
    </div>

    <!-- 无连接 -->
    <div v-else-if="!connId" class="env-empty">
      <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" opacity="0.4">
        <path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4"/><polyline points="10 17 15 12 10 7"/><line x1="15" y1="12" x2="3" y2="12"/>
      </svg>
      <span>{{ t('env.noConnection') }}</span>
      <button class="env-btn" @click="detect" style="margin-top:0.5rem">{{ t('env.retry') }}</button>
    </div>

    <!-- 无结果 -->
    <div v-else-if="envs.length === 0" class="env-empty">
      <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" opacity="0.4">
        <rect x="2" y="2" width="20" height="8" rx="2"/><rect x="2" y="14" width="20" height="8" rx="2"/><line x1="6" y1="6" x2="6.01" y2="6"/><line x1="6" y1="18" x2="6.01" y2="18"/>
      </svg>
      <span>{{ t('env.noResults') }}</span>
    </div>

    <!-- 按分类展示 -->
    <div v-else class="env-scroll">
      <div v-for="cat in groupedEnvs" :key="cat.name" class="env-category">
        <div class="cat-header">
          <span class="cat-name">{{ cat.name }}</span>
          <span class="cat-count">{{ cat.installed }}/{{ cat.items.length }}</span>
        </div>
        <div class="env-grid">
          <div
            v-for="env in cat.items"
            :key="env.name"
            class="env-card"
            :class="{ installed: env.status === 'installed' }"
            @click="onCardClick(env)"
          >
            <div class="card-top">
              <span class="card-icon" :class="env.status">{{ env.name.charAt(0) }}</span>
              <div class="card-info">
                <div class="card-name">
                  {{ env.name }}
                  <span v-if="env.paths && env.paths.length > 1" class="multi-badge" :title="t('env.multiPath')">
                    {{ env.paths.length }}
                  </span>
                </div>
                <div v-if="env.status === 'installed' && env.version" class="card-version">v{{ env.version }}</div>
              </div>
              <span v-if="env.status === 'installed'" class="card-dot installed"></span>
              <span v-else class="card-dot notfound"></span>
            </div>
            <div v-if="env.status === 'installed'" class="card-desc">{{ t('env.desc.' + env.name) || env.description }}</div>
            <div v-if="env.status === 'installed' && env.path" class="card-path" :title="env.path">
              <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
              </svg>
              {{ env.path }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 路径选择弹窗 -->
    <div v-if="showPathDialog" class="path-dialog-overlay" @click.self="showPathDialog = false">
      <div class="path-dialog">
        <div class="path-dialog-header">
          <h3>{{ t('env.selectPath') }} - {{ selectedEnv?.name }}</h3>
          <button class="path-dialog-close" @click="showPathDialog = false">&times;</button>
        </div>
        <div class="path-dialog-body">
          <p class="path-dialog-hint">{{ t('env.multiPathHint') }}</p>
          <div
            v-for="(p, idx) in selectedEnv?.paths || []"
            :key="idx"
            class="path-option"
            @click="selectPath(p)"
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
            </svg>
            <span class="path-option-text">{{ p }}</span>
          </div>
        </div>
        <div class="path-dialog-footer">
          <button class="env-btn" @click="showPathDialog = false">{{ t('common.cancel') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { SSHService } from '@bindings/changeme/ssh/index.js'
import { useSSHLayoutStore } from '../../../stores/sshLayout'
import { Events } from '@wailsio/runtime'

const { t } = useI18n()
const sshLayoutStore = useSSHLayoutStore()

const loading = ref(false)
const envs = ref([])
const search = ref('')
const error = ref('')

// 路径选择弹窗
const showPathDialog = ref(false)
const selectedEnv = ref(null)

const installedCount = computed(() => envs.value.filter(e => e.status === 'installed').length)

const filteredEnvs = computed(() => {
  if (!search.value) return envs.value
  const q = search.value.toLowerCase()
  return envs.value.filter(e =>
    e.name.toLowerCase().includes(q) ||
    e.category.toLowerCase().includes(q) ||
    (e.description && e.description.toLowerCase().includes(q)) ||
    (e.version && e.version.toLowerCase().includes(q)) ||
    (e.path && e.path.toLowerCase().includes(q))
  )
})

const groupedEnvs = computed(() => {
  const map = new Map()
  for (const env of filteredEnvs.value) {
    const cat = env.category || '其他'
    if (!map.has(cat)) map.set(cat, { name: cat, items: [], installed: 0 })
    const entry = map.get(cat)
    entry.items.push(env)
    if (env.status === 'installed') entry.installed++
  }
  return Array.from(map.values()).sort((a, b) => b.installed - a.installed)
})

const connId = computed(() => sshLayoutStore.currentConnectionId)

const detect = async () => {
  const id = connId.value
  if (!id) {
    error.value = t('env.noConnection')
    return
  }
  loading.value = true
  error.value = ''
  try {
    const result = await SSHService.GetSystemEnvironments(id)
    envs.value = result || []
  } catch (e) {
    console.error('[EnvPanel] 检测失败:', e)
    error.value = String(e?.message || e)
    envs.value = []
  } finally {
    loading.value = false
  }
}

// 点击卡片
const onCardClick = (env) => {
  if (env.status !== 'installed') return
  if (env.paths && env.paths.length > 1) {
    // 多个路径 → 弹窗选择
    selectedEnv.value = env
    showPathDialog.value = true
  } else {
    goToPath(env.path)
  }
}

// 选择路径
const selectPath = (path) => {
  showPathDialog.value = false
  selectedEnv.value = null
  goToPath(path)
}

// 跳转文件管理器
const goToPath = (envPath) => {
  if (!envPath || !connId.value) return
  const dir = envPath.substring(0, envPath.lastIndexOf('/')) || '/'
  Events.Emit('file-manager:navigate', { connId: connId.value, path: dir })
}

watch(connId, (newId) => { if (newId) detect() })
onMounted(() => { detect() })
</script>

<style scoped>
.env-panel { display: flex; flex-direction: column; height: 100%; overflow: hidden; background: var(--bg-panel); }

.env-toolbar {
  display: flex; align-items: center; gap: 0.5rem;
  padding: 0.5rem 0.75rem; border-bottom: 1px solid var(--surface-hover); flex-shrink: 0;
}
.env-btn {
  display: flex; align-items: center; gap: 0.375rem;
  padding: 0.375rem 0.75rem; background: var(--surface-1); border: 1px solid var(--surface-hover);
  border-radius: 0.375rem; color: var(--text-secondary); font-size: 0.75rem;
  cursor: pointer; transition: all 0.15s; white-space: nowrap;
}
.env-btn:hover:not(:disabled) { background: var(--surface-hover); color: var(--text-primary); }
.env-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.env-search {
  flex: 1; padding: 0.375rem 0.625rem; background: var(--surface-3);
  border: 1px solid var(--border-strong); border-radius: 0.375rem;
  color: var(--text-primary); font-size: 0.75rem; outline: none; min-width: 0;
}
.env-search::placeholder { color: var(--text-muted); }
.env-search:focus { border-color: var(--border-accent); }
.env-count { color: var(--text-muted); font-size: 0.6875rem; white-space: nowrap; }

.env-loading, .env-empty {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  gap: 0.5rem; padding: 2rem; color: var(--text-muted); font-size: 0.8125rem;
}
.spinner {
  width: 1.5rem; height: 1.5rem; border: 0.125rem solid var(--surface-hover);
  border-top-color: var(--accent-primary); border-radius: 50%; animation: spin 0.8s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

.env-scroll { flex: 1; overflow-y: auto; padding: 0.5rem 0.75rem; }
.env-category { margin-bottom: 1rem; }
.cat-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0.375rem 0; margin-bottom: 0.375rem; border-bottom: 1px solid var(--surface-1);
}
.cat-name { font-size: 0.6875rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.5px; color: var(--text-muted); }
.cat-count { font-size: 0.625rem; color: var(--text-disabled); background: var(--surface-2); padding: 0.0625rem 0.375rem; border-radius: 0.25rem; }
.env-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(14rem, 1fr)); gap: 0.375rem; }

.env-card {
  display: flex; flex-direction: column; gap: 0.25rem;
  padding: 0.5rem 0.625rem; background: var(--surface-1);
  border: 1px solid var(--border-default); border-radius: 0.375rem; transition: all 0.15s;
}
.env-card.installed { cursor: pointer; }
.env-card.installed:hover { border-color: var(--border-accent); background: var(--surface-2); }
.env-card:not(.installed) { opacity: 0.55; }

.card-top { display: flex; align-items: center; gap: 0.5rem; }
.card-icon {
  width: 1.5rem; height: 1.5rem; display: flex; align-items: center; justify-content: center;
  border-radius: 0.25rem; font-size: 0.625rem; font-weight: 700; flex-shrink: 0;
}
.card-icon.installed { background: var(--success-bg, rgba(72,187,120,0.12)); color: var(--accent-success); }
.card-icon.notfound { background: var(--surface-2); color: var(--text-disabled); }
.card-info { flex: 1; min-width: 0; }
.card-name { font-size: 0.8125rem; font-weight: 600; color: var(--text-primary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; display: flex; align-items: center; gap: 0.375rem; }
.card-version { font-size: 0.625rem; color: var(--accent-primary); font-family: 'Cascadia Code', Consolas, monospace; }
.card-dot { width: 0.375rem; height: 0.375rem; border-radius: 50%; flex-shrink: 0; }
.card-dot.installed { background: var(--accent-success); }
.card-dot.notfound { background: var(--text-disabled); }
.card-desc { font-size: 0.625rem; color: var(--text-muted); line-height: 1.3; padding-left: 2rem; }
.card-path {
  display: flex; align-items: center; gap: 0.25rem;
  font-size: 0.625rem; color: var(--text-muted); font-family: 'Cascadia Code', Consolas, monospace;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap; padding-left: 2rem;
}
.card-path svg { flex-shrink: 0; opacity: 0.5; }

.multi-badge {
  display: inline-flex; align-items: center; justify-content: center;
  width: 1rem; height: 1rem; border-radius: 50%;
  background: var(--accent-primary); color: var(--text-on-accent, #fff);
  font-size: 0.5rem; font-weight: 700; flex-shrink: 0;
}

/* 路径选择弹窗 */
.path-dialog-overlay {
  position: fixed; inset: 0; background: var(--bg-overlay, rgba(0,0,0,0.5));
  display: flex; align-items: center; justify-content: center;
  z-index: 9999; backdrop-filter: blur(4px);
}
.path-dialog {
  background: var(--bg-panel-solid, var(--bg-panel)); border: 1px solid var(--border-default);
  border-radius: 0.75rem; width: 28rem; max-width: 90vw; max-height: 70vh;
  box-shadow: 0 16px 48px var(--shadow-lg, rgba(0,0,0,0.4));
  display: flex; flex-direction: column; overflow: hidden;
}
.path-dialog-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 1rem 1.25rem; border-bottom: 1px solid var(--surface-hover);
}
.path-dialog-header h3 { margin: 0; font-size: 0.9375rem; font-weight: 600; color: var(--text-primary); }
.path-dialog-close {
  background: none; border: none; color: var(--text-muted); font-size: 1.25rem;
  cursor: pointer; width: 1.5rem; height: 1.5rem; display: flex; align-items: center; justify-content: center;
  border-radius: 0.25rem;
}
.path-dialog-close:hover { background: var(--surface-hover); color: var(--text-primary); }
.path-dialog-body { padding: 1rem 1.25rem; overflow-y: auto; flex: 1; }
.path-dialog-hint { font-size: 0.75rem; color: var(--text-muted); margin: 0 0 0.75rem; }

.path-option {
  display: flex; align-items: center; gap: 0.625rem;
  padding: 0.625rem 0.75rem; border: 1px solid var(--border-default); border-radius: 0.375rem;
  cursor: pointer; transition: all 0.15s; margin-bottom: 0.375rem;
}
.path-option:hover { border-color: var(--border-accent); background: var(--surface-hover); }
.path-option svg { flex-shrink: 0; color: var(--text-muted); }
.path-option-text {
  font-family: 'Cascadia Code', Consolas, monospace; font-size: 0.8125rem;
  color: var(--text-primary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.path-dialog-footer {
  display: flex; justify-content: flex-end; padding: 0.75rem 1.25rem;
  border-top: 1px solid var(--surface-hover);
}
</style>
