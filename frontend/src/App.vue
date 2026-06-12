<template>
  <router-view />
  <Message ref="messageRef" />
  <ConfirmDialog />
</template>

<script setup>
import { ref, onMounted } from 'vue'
import Message from './components/Message.vue'
import ConfirmDialog from './components/ConfirmDialog.vue'
import { setMessageInstance } from './utils/message'

const messageRef = ref(null)

onMounted(() => {
  // 设置全局消息实例
  if (messageRef.value) {
    setMessageInstance(messageRef.value)
  }
})
</script>

<style>
/* 全局滚动条样式 */
*::-webkit-scrollbar {
  width: 3px;
  height: 3px;
}

*::-webkit-scrollbar-track {
  background: transparent;
}

*::-webkit-scrollbar-thumb {
  background: linear-gradient(180deg,
    color-mix(in srgb, var(--accent-primary) 15%, transparent),
    color-mix(in srgb, var(--accent-primary) 30%, transparent));
  border-radius: 999px;
}

*::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(180deg,
    color-mix(in srgb, var(--accent-primary) 30%, transparent),
    color-mix(in srgb, var(--accent-primary) 50%, transparent));
}

*::-webkit-scrollbar-button {
  display: none;
}
</style>
