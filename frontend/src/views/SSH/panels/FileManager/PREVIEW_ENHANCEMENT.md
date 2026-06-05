# 文件预览功能增强

## 📋 更新内容

### 1. 文本文件支持（60+ 种语言）

#### 编程语言
- JavaScript/TypeScript (js, jsx, ts, tsx, mjs, cjs)
- Python (py, pyw, pyi)
- Java (java)
- C/C++ (c, cpp, cc, cxx, h, hpp, hh)
- C# (cs)
- Go (go)
- Rust (rs)
- Ruby (rb, erb)
- PHP (php, phtml)
- Swift (swift)
- Kotlin (kt, kts)
- Scala (scala)
- R (r, R)
- Objective-C (m, mm)
- Lua (lua)
- Perl (pl, pm)
- Haskell (hs)
- Dart (dart)
- Elixir (ex, exs)
- Erlang (erl)
- Clojure (clj, cljs)
- F# (fs, fsx)
- VB (vb, vbs)
- Assembly (asm, s)

#### Web 技术
- HTML (html, htm, xhtml, vue, svelte, astro)
- CSS (css, scss, sass, less)

#### 数据/配置格式
- JSON (json, json5)
- XML (xml, svg, pom)
- YAML (yaml, yml)
- TOML (toml)
- INI (ini, conf, cfg, env, properties, editorconfig)
- SQL (sql)
- GraphQL (graphql, gql)
- Protocol Buffers (proto)
- Terraform (tf, tfvars)

#### 文档/标记
- Markdown (md, markdown, mdx)
- Plain Text (txt, log, csv, tsv, gitignore)

#### Shell/脚本
- Shell (sh, bash, zsh, fish)
- PowerShell (ps1)
- Batch (bat, cmd)
- Makefile (makefile)
- Dockerfile (dockerfile)
- Gradle (gradle)

### 2. 媒体文件预览

#### 图片格式
- PNG, JPG, JPEG, GIF, BMP, SVG, WebP, ICO, TIFF, TIF
- 支持缩放和全屏查看

#### 视频格式
- MP4, AVI, MOV, WMV, FLV, MKV, WebM, M4V, 3GP
- 支持播放控制（播放/暂停/进度条/音量）
- 自动播放

#### 音频格式
- MP3, WAV, OGG, FLAC, AAC, WMA, M4A
- 支持播放控制
- 自动播放

## 🔧 技术实现

### FileManagerPanel.vue
- 扩展了文件类型检测逻辑
- 添加了视频和音频预览分支
- 更新了 URL 清理逻辑（支持所有媒体类型）

### FilePreview.vue
- 添加了 `<video>` 和 `<audio>` 播放器组件
- 更新了样式以适配媒体播放器
- 扩展了语言检测映射（60+ 种）

### CodeEditor.vue
- 添加了更多语言包导入
- 使用 `StreamLanguage` 支持传统语法高亮模式
- 扩展了语言扩展映射

## 📊 对比

| 特性 | 之前 | 现在 |
|------|------|------|
| 文本语言支持 | ~20 种 | 60+ 种 |
| 图片格式 | 7 种 | 11 种 |
| 视频预览 | ❌ | ✅ 9 种格式 |
| 音频预览 | ❌ | ✅ 7 种格式 |
| 总支持格式 | ~30 种 | 90+ 种 |

## 🎯 使用示例

### 查看代码文件
```
点击任意 .go, .py, .js, .java 等文件
→ 自动识别语言
→ 显示语法高亮
→ 可编辑保存
```

### 预览图片
```
点击 .png, .jpg, .svg 等文件
→ 显示图片预览
→ 支持缩放
```

### 播放视频
```
点击 .mp4, .avi, .mov 等文件
→ 显示视频播放器
→ 支持播放控制
```

### 播放音频
```
点击 .mp3, .wav, .ogg 等文件
→ 显示音频播放器
→ 支持播放控制
```

## ⚠️ 注意事项

1. **大文件处理**：
   - 图片 > 10MB：显示警告提示
   - 音频 > 20MB：建议下载到本地收听
   - 视频 > 50MB：建议下载到本地观看
   
2. **内存管理**：关闭预览时会自动清理 Blob URL

3. **浏览器兼容性**：视频/音频播放依赖浏览器支持的编解码器

4. **性能优化**：媒体文件会完整下载到内存，大文件可能导致卡顿

## 🚀 未来改进

- [x] 添加文件大小限制提示
- [ ] 支持视频流式播放（需要后端支持 HTTP Range 请求）
- [ ] 支持 PDF 预览
- [ ] 支持 Office 文档预览（需要转换服务）
- [ ] 添加图片旋转/裁剪功能
- [ ] 视频截图功能
- [ ] 音频波形可视化
