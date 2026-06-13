<template>
  <div class="code-editor-container">
    <!-- 加载状态 -->
    <div v-if="loading" class="editor-loading">
      <div class="spinner"></div>
      <p>加载编辑器...</p>
    </div>
    
    <!-- CodeMirror Editor 容器 -->
    <div ref="editorContainer" class="codemirror-wrapper" v-show="!loading"></div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { EditorView, basicSetup } from 'codemirror'
import { EditorState } from '@codemirror/state'
import { javascript } from '@codemirror/lang-javascript'
import { python } from '@codemirror/lang-python'
import { html } from '@codemirror/lang-html'
import { css } from '@codemirror/lang-css'
import { json } from '@codemirror/lang-json'
import { markdown } from '@codemirror/lang-markdown'
import { xml } from '@codemirror/lang-xml'
import { java } from '@codemirror/lang-java'
import { cpp } from '@codemirror/lang-cpp'
import { StreamLanguage } from '@codemirror/language'
import { ruby } from '@codemirror/legacy-modes/mode/ruby'
import { shell } from '@codemirror/legacy-modes/mode/shell'
import { yaml } from '@codemirror/legacy-modes/mode/yaml'
import { sql } from '@codemirror/legacy-modes/mode/sql'
import { oneDark } from '@codemirror/theme-one-dark'

// ✅ 使用 basicSetup，它已经包含了行号等基础功能
// basicSetup 包含：行号、高亮活动行、括号匹配、自动闭合等

// 语言扩展映射（常用语言）
const languageExtensions = {
  // JavaScript/TypeScript
  javascript: javascript(),
  js: javascript(),
  jsx: javascript(),
  typescript: javascript(),
  ts: javascript(),
  tsx: javascript(),
  // Python
  python: python(),
  py: python(),
  // HTML
  html: html(),
  htm: html(),
  vue: html(),
  svelte: html(),
  // CSS
  css: css(),
  scss: css(),
  less: css(),
  // JSON
  json: json(),
  // Markdown
  markdown: markdown(),
  md: markdown(),
  // XML
  xml: xml(),
  svg: xml(),
  // Java
  java: java(),
  // C/C++
  c: cpp(),
  cpp: cpp(),
  h: cpp(),
  hpp: cpp(),
  // Legacy modes
  ruby: StreamLanguage.define(ruby),
  rb: StreamLanguage.define(ruby),
  shell: StreamLanguage.define(shell),
  sh: StreamLanguage.define(shell),
  bash: StreamLanguage.define(shell),
  yaml: StreamLanguage.define(yaml),
  yml: StreamLanguage.define(yaml),
  sql: StreamLanguage.define(sql),
}

const props = defineProps({
  content: {
    type: String,
    required: true
  },
  language: {
    type: String,
    default: 'plaintext'
  },
  readonly: {
    type: Boolean,
    default: false
  },
  theme: {
    type: String,
    default: 'vs-dark'
  },
  wordWrap: {
    type: Boolean,
    default: true
  }
})

const emit = defineEmits(['update:content'])

const editorContainer = ref(null)
const loading = ref(true)
let editorView = null

function isDarkTheme() {
  return document.documentElement.dataset.theme !== 'light'
}

// 获取语言扩展
function getLanguageExtension(lang) {
  return languageExtensions[lang.toLowerCase()] || null
}

// 初始化编辑器
onMounted(async () => {
  await nextTick()
  
  if (!editorContainer.value) return
  
  const langExtension = getLanguageExtension(props.language)
  
  // 创建 CodeMirror 编辑器
  const extensions = [
    basicSetup,
    (props.theme === 'vs-dark' && isDarkTheme()) ? oneDark : [],
    props.readonly ? EditorView.editable.of(false) : [],
    props.wordWrap ? EditorView.lineWrapping : [],
    EditorView.updateListener.of((update) => {
      if (update.docChanged) {
        const newValue = update.state.doc.toString()
        emit('update:content', newValue)
      }
    }),
  ]
  
  // 只添加支持的语言扩展
  if (langExtension) {
    extensions.push(langExtension)
  }
  
  const startState = EditorState.create({
    doc: props.content,
    extensions,
  })
  
  editorView = new EditorView({
    state: startState,
    parent: editorContainer.value,
  })
  
  loading.value = false
})

// 监听内容变化（外部更新）
watch(() => props.content, (newContent) => {
  if (editorView && newContent !== editorView.state.doc.toString()) {
    editorView.dispatch({
      changes: {
        from: 0,
        to: editorView.state.doc.length,
        insert: newContent,
      },
    })
  }
})

// 监听只读状态变化 - CodeMirror 6 不支持动态切换 readonly，需要重建
watch(() => props.readonly, () => {
  // 当 readonly 变化时，销毁并重建编辑器
  if (editorView) {
    const currentValue = editorView.state.doc.toString()
    editorView.destroy()
    
    const langExtension = getLanguageExtension(props.language)
    const extensions = [
      basicSetup,
      (props.theme === 'vs-dark' && isDarkTheme()) ? oneDark : [],
      props.readonly ? EditorView.editable.of(false) : [],
      props.wordWrap ? EditorView.lineWrapping : [],
      EditorView.updateListener.of((update) => {
        if (update.docChanged) {
          const newValue = update.state.doc.toString()
          emit('update:content', newValue)
        }
      }),
    ]
    
    if (langExtension) {
      extensions.push(langExtension)
    }
    
    const startState = EditorState.create({
      doc: currentValue,
      extensions,
    })
    
    editorView = new EditorView({
      state: startState,
      parent: editorContainer.value,
    })
  }
})

// 获取编辑器内容
const getValue = () => {
  return editorView ? editorView.state.doc.toString() : ''
}

// 设置编辑器内容
const setValue = (value) => {
  if (editorView) {
    editorView.dispatch({
      changes: {
        from: 0,
        to: editorView.state.doc.length,
        insert: value,
      },
    })
  }
}

// 暴露方法给父组件
defineExpose({
  getValue,
  setValue
})

// 清理
onUnmounted(() => {
  if (editorView) {
    editorView.destroy()
    editorView = null
  }
})
</script>

<style scoped>
.code-editor-container {
  width: 100%;
  height: 100%;
  position: relative;
}

.editor-loading {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: var(--bg-toolbar);
  color: var(--text-primary);
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid var(--surface-hover);
  border-top-color: var(--accent-primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-bottom: 16px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.codemirror-wrapper {
  width: 100%;
  height: 100%;
  overflow: auto;
}

/* CodeMirror 样式定制 */
:deep(.cm-editor) {
  height: 100%;
  font-size: 14px;
  font-family: 'Cascadia Code', 'Fira Code', Consolas, 'Courier New', monospace;
}

:deep(.cm-scroller) {
  overflow: auto;
}

:deep(.cm-gutters) {
  background-color: var(--bg-toolbar);
  border-right: 1px solid var(--border-default);
}

:deep(.cm-activeLineGutter) {
  background-color: var(--bg-panel-solid);
}

:deep(.cm-cursor) {
  border-left-color: var(--text-primary);
}

:deep(.cm-content) {
  caret-color: var(--text-primary);
}

:deep(.cm-selectionBackground) {
  background: var(--bg-selected) !important;
}

:deep(.cm-activeLine) {
  background-color: var(--surface-1) !important;
}
</style>
