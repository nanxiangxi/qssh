<template>
  <ContextMenu
    :visible="visible"
    :position="position"
    :items="menuItems"
    @select="handleSelect"
    @close="$emit('update:visible', false)"
  />
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import ContextMenu from '../../../../../components/ContextMenu.vue'

const { t } = useI18n()

const props = defineProps({
  visible: { type: Boolean, default: false },
  position: { type: Object, default: () => ({ x: 0, y: 0 }) },
  hasSelection: { type: Boolean, default: false },
})

const emit = defineEmits(['update:visible', 'copy', 'paste', 'select-all', 'clear', 'search'])

const menuItems = computed(() => [
  {
    label: t('terminal.copy'),
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>',
    action: 'copy',
    shortcut: 'Ctrl+Shift+C',
    disabled: !props.hasSelection,
  },
  {
    label: t('terminal.paste'),
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"/><rect x="8" y="2" width="8" height="4" rx="1"/></svg>',
    action: 'paste',
    shortcut: 'Ctrl+Shift+V',
  },
  { divider: true },
  {
    label: t('terminal.selectAll'),
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 11V5a2 2 0 0 0-2-2H5a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h6"/><path d="M12 12H3M12 12V3"/></svg>',
    action: 'select-all',
    shortcut: 'Ctrl+A',
  },
  {
    label: t('terminal.clear'),
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>',
    action: 'clear',
    shortcut: 'Ctrl+L',
  },
  { divider: true },
  {
    label: t('terminal.search'),
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>',
    action: 'search',
    shortcut: 'Ctrl+F',
  },
])

function handleSelect(item) {
  if (item.action) {
    emit(item.action)
  }
}
</script>
