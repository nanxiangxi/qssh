# Monaco Editor 中文本地化实现

## ✅ 已完成

本项目已成功实现 Monaco Editor 的完整中文本地化！

## 实现方案

使用自定义 Vite 插件在构建时注入中文语言包，替换 Monaco Editor 的 NLS（National Language Support）系统。

### 技术原理

1. **开发环境**：使用 esbuild 插件在预编译时替换 `nls.js` 文件
2. **生产环境**：使用 Rollup 插件在打包时注入中文翻译
3. **语言包来源**：VS Code 官方中文语言包 (`main.i18n.json`)

### 核心文件

- `vite-plugin-monaco-nls.js` - 自定义 Vite 插件（184 行）
- `src/main.i18n.json` - 中文语言包（22668 行完整翻译）
- `vite.config.js` - Vite 配置，集成插件

## 汉化效果

✅ **完全中文化**：
- 所有菜单项（文件、编辑、选择等）
- 右键上下文菜单（复制、粘贴、剪切等）
- 查找/替换对话框（Ctrl+F / Ctrl+H）
- 命令面板（Ctrl+Shift+P）
- 状态栏信息
- 错误提示和警告
- 所有 UI 交互文本

## 配置说明

### vite.config.js 配置

```javascript
import monacoEditorNlsPlugin, { esbuildPluginMonacoEditorNls } from "./vite-plugin-monaco-nls.js";
import zhHans from "./src/main.i18n.json";

export default defineConfig({
  plugins: [
    // 生产环境使用 Rollup 插件
    process.env.NODE_ENV !== "development" && monacoEditorNlsPlugin({
      locale: "zh-hans",
      localeData: zhHans,
    }),
  ].filter(Boolean),
  optimizeDeps: {
    esbuildOptions: {
      plugins: [
        // 开发环境使用 esbuild 插件
        esbuildPluginMonacoEditorNls({
          locale: "zh-hans",
          localeData: zhHans,
        }),
      ],
    },
  },
});
```

### 工作原理

插件会拦截 Monaco Editor 的模块加载：

1. **拦截 `nls.js`**：替换为包含中文翻译的版本
2. **修改 `localize()` 函数**：从中文语言包中查找翻译
3. **自动应用**：所有 Monaco Editor 实例自动使用中文

## 使用方法

只需正常启动开发服务器即可，中文会自动生效：

```bash
npm run dev
```

无需任何额外配置或代码修改！

## 技术细节

### 插件工作流程

```
用户访问页面
    ↓
Vite 加载 Monaco Editor 模块
    ↓
esbuild/Rollup 插件拦截
    ↓
替换 nls.js 中的 localize 函数
    ↓
注入中文语言包数据
    ↓
Monaco Editor 显示中文界面
```

### 语言包结构

```json
{
  "contents": {
    "vs/editor/editor.main": {
      "actions.find": "查找",
      "actions.replace": "替换",
      ...
    },
    "vs/platform/actions/common/actions": {
      "miCopy": "复制",
      "miPaste": "粘贴",
      ...
    }
  }
}
```

## 优势

✅ **完整汉化**：所有 UI 元素都是中文
✅ **无缝集成**：无需修改业务代码
✅ **开发友好**：开发和生产环境都支持
✅ **性能优秀**：只在构建时处理，运行时零开销
✅ **易于维护**：语言包来自 VS Code 官方，定期更新

## 参考资源

- [掘金文章：发布一个monaco-editor 汉化包](https://juejin.cn/post/7524175079073546275)
- [VS Code 语言包仓库](https://github.com/microsoft/vscode-loc)
- [Monaco Editor 官方文档](https://microsoft.github.io/monaco-editor/)
