# 代码编辑器说明

## 技术选型

本项目使用 **CodeMirror 6** 作为代码编辑器。

### 为什么选择 CodeMirror 6？

1. **轻量级**：比 Monaco Editor 小很多，加载更快
2. **配置简单**：无需复杂的 Vite 插件配置
3. **性能优秀**：编译速度快，不影响开发体验
4. **功能足够**：满足文件查看和编辑的基本需求
5. **易于定制**：可以轻松实现中文界面

## 支持的语言

### 常用编程语言（15+ 种）
- **JavaScript/TypeScript**: js, jsx, ts, tsx
- **Python**: py, pyw
- **Java**: java
- **C/C++**: c, cpp, h, hpp
- **Ruby**: rb
- **HTML**: html, htm, vue, svelte
- **CSS**: css, scss, less
- **JSON**: json
- **XML**: xml, svg
- **YAML**: yaml, yml
- **SQL**: sql
- **Markdown**: md
- **Shell**: sh, bash
- **纯文本**: txt, log, ini, conf, cfg, env, toml, gitignore, dockerfile, makefile 等

### 设计原则
- ✅ **核心语言**：提供完整的语法高亮和智能提示
- ✅ **其他文本**：所有未列出的文件都作为纯文本打开，可以正常编辑和保存
- ✅ **轻量级**：只加载必要的语言包，减少编译时间和包体积

## 功能特性

✅ 语法高亮（15+ 种常用语言）
✅ 行号显示
✅ 自动换行
✅ 深色主题
✅ 只读模式
✅ 实时内容同步
✅ 简洁界面（无查找/替换等复杂功能）
✅ 所有文本文件均可编辑（未支持的语言作为纯文本）

## 媒体预览支持

### 图片格式
- PNG, JPG, JPEG, GIF, BMP, SVG, WebP, ICO, TIFF

### 视频格式
- MP4, AVI, MOV, WMV, FLV, MKV, WebM, M4V, 3GP

### 音频格式
- MP3, WAV, OGG, FLAC, AAC, WMA, M4A

## 使用方法

```vue
<CodeEditor
  :content="fileContent"
  :language="'javascript'"
  :readonly="false"
  :theme="'vs-dark'"
  :word-wrap="true"
  @update:content="handleContentChange"
/>
```

## 与 Monaco Editor 的对比

| 特性 | CodeMirror 6 | Monaco Editor |
|------|-------------|---------------|
| 体积 | ~50KB | ~2MB |
| 编译速度 | 快 | 慢（需要插件） |
| 配置复杂度 | 简单 | 复杂 |
| 中文支持 | 容易 | 困难（需要插件） |
| 功能丰富度 | 基础 | 非常强大 |
| 适用场景 | 简单编辑 | IDE 级别 |

## 总结

对于本项目的文件管理器场景，CodeMirror 6 是更合适的选择：
- ✅ 编译速度快
- ✅ 配置简单
- ✅ 足够满足需求
- ✅ 易于维护和扩展
