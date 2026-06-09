<template>
  <Transition name="fade">
    <div v-if="visible" class="dialog-mask" @mousedown="close">
      <div class="dialog" @mousedown.stop>
        <!-- 标题栏 -->
        <div class="dialog-header">
          <h3 class="dialog-title">命令历史</h3>
          <div class="header-actions">
            <button class="header-btn" :class="{ active: showFavorites }" @click="showFavorites = !showFavorites" title="收藏">
              <svg width="14" height="14" viewBox="0 0 24 24" :fill="showFavorites ? '#e0af68' : 'none'" stroke="currentColor" stroke-width="2">
                <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
              </svg>
            </button>
            <button class="header-btn" @click="close">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="18" y1="6" x2="6" y2="18"/>
                <line x1="6" y1="6" x2="18" y2="18"/>
              </svg>
            </button>
          </div>
        </div>

        <!-- 搜索栏 -->
        <div class="search-bar">
          <svg class="search-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8"/>
            <line x1="21" y1="21" x2="16.65" y2="16.65"/>
          </svg>
          <input v-model="searchKeyword" class="search-input" placeholder="搜索命令..." />
        </div>

        <!-- 命令列表 -->
        <div class="command-list">
          <template v-if="showFavorites">
            <div v-for="fav in favorites" :key="fav.id"
              class="command-item"
              @click="selectCommand(fav.command)">
              <span class="cmd-star">★</span>
              <span class="cmd-text">{{ fav.command }}</span>
              <button class="cmd-remove" @click.stop="removeFavorite(fav.id)" title="取消收藏">×</button>
            </div>
            <div v-if="favorites.length === 0" class="empty-state">暂无收藏的命令</div>
          </template>

          <template v-else>
            <div v-for="entry in filteredHistory" :key="entry.id"
              class="command-item"
              @click="selectCommand(entry.command)">
              <span class="cmd-text">{{ entry.command }}</span>
              <span class="cmd-time">{{ formatTime(entry.timestamp) }}</span>
              <button
                class="cmd-fav"
                :class="{ active: isFavorite(entry.command) }"
                @click.stop="toggleFavorite(entry.command)"
                :title="isFavorite(entry.command) ? '取消收藏' : '收藏'"
              >★</button>
            </div>
            <div v-if="filteredHistory.length === 0" class="empty-state">
              {{ searchKeyword ? '未找到匹配的命令' : '暂无命令历史' }}
            </div>
          </template>
        </div>

        <!-- 底部 -->
        <div class="dialog-footer">
          <span class="footer-hint">点击填入</span>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  visible: { type: Boolean, default: false },
  history: { type: Array, default: () => [] },
  favorites: { type: Array, default: () => [] }
})

const emit = defineEmits(['update:visible', 'select-command', 'add-favorite', 'remove-favorite'])

const searchKeyword = ref('')
const showFavorites = ref(false)

const filteredHistory = computed(() => {
  const list = props.history
  if (!searchKeyword.value) return list.slice(-100).reverse()
  const kw = searchKeyword.value.toLowerCase()
  return list.filter(h => h.command.toLowerCase().includes(kw)).slice(-100).reverse()
})

function selectCommand(command) {
  emit('select-command', command)
}

function isFavorite(command) {
  return props.favorites.some(f => f.command === command.trim())
}

function toggleFavorite(command) {
  if (isFavorite(command)) {
    const fav = props.favorites.find(f => f.command === command.trim())
    if (fav) emit('remove-favorite', fav.id)
  } else {
    emit('add-favorite', command)
  }
}

function removeFavorite(id) {
  emit('remove-favorite', id)
}

function close() {
  emit('update:visible', false)
}

function formatTime(timestamp) {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}
</script>

<style scoped>
.dialog-mask {
  position: fixed; inset: 0; background: rgba(0,0,0,.6);
  display: flex; align-items: center; justify-content: center;
  z-index: 99999; backdrop-filter: blur(2px);
}

.dialog {
  width: 480px; max-width: 90vw; max-height: 60vh;
  background: #1a1a1a; border: 1px solid #333; border-radius: 10px;
  box-shadow: 0 16px 48px rgba(0,0,0,.5);
  display: flex; flex-direction: column; overflow: hidden;
}

.dialog-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 12px 16px; border-bottom: 1px solid #2a2a2a;
}

.dialog-title { margin: 0; font-size: 14px; color: #d4d4d4; }

.header-actions { display: flex; gap: 4px; }

.header-btn {
  display: flex; align-items: center; justify-content: center;
  width: 26px; height: 26px; background: transparent; border: none;
  border-radius: 4px; color: #666; cursor: pointer;
}
.header-btn:hover { background: #333; color: #aaa; }
.header-btn.active { color: #e0af68; }

.search-bar {
  display: flex; align-items: center; padding: 8px 12px;
  border-bottom: 1px solid #2a2a2a;
}

.search-icon { color: #666; margin-right: 8px; flex-shrink: 0; }

.search-input {
  flex: 1; background: transparent; border: none; outline: none;
  color: #d4d4d4; font-size: 13px;
}
.search-input::placeholder { color: #555; }

.command-list {
  flex: 1; min-height: 0; overflow-y: auto; padding: 4px;
}

.command-item {
  display: flex; align-items: center; gap: 8px;
  padding: 7px 10px; border-radius: 4px; cursor: pointer;
}

.command-item:hover { background: #222; }

.cmd-star { color: #e0af68; font-size: 11px; flex-shrink: 0; }

.cmd-text {
  flex: 1; font-size: 13px; color: #d4d4d4;
  font-family: 'Cascadia Code', monospace;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}

.cmd-time { font-size: 11px; color: #555; flex-shrink: 0; }

.cmd-remove {
  display: flex; align-items: center; justify-content: center;
  width: 18px; height: 18px; background: transparent; border: none;
  border-radius: 3px; color: #666; cursor: pointer; font-size: 14px;
}
.cmd-remove:hover { background: #333; color: #f44336; }

.cmd-fav {
  display: flex; align-items: center; justify-content: center;
  width: 20px; height: 20px; background: transparent; border: none;
  border-radius: 3px; color: #555; cursor: pointer; font-size: 11px;
}
.cmd-fav:hover { background: #333; color: #e0af68; }
.cmd-fav.active { color: #e0af68; }

.empty-state { padding: 24px; text-align: center; color: #555; font-size: 13px; }

.dialog-footer {
  padding: 8px 16px; border-top: 1px solid #2a2a2a;
}

.footer-hint { font-size: 11px; color: #555; }

.command-list::-webkit-scrollbar { width: 4px; }
.command-list::-webkit-scrollbar-track { background: transparent; }
.command-list::-webkit-scrollbar-thumb { background: #333; border-radius: 2px; }

.fade-enter-active { transition: all .2s ease; }
.fade-leave-active { transition: all .15s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
