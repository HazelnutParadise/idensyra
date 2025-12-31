# Wails Migration Guide

## 從 Fyne UI 遷移到 Wails

本項目已從 Fyne UI 框架遷移到 Wails v2，使用原生 WebView 和現代化的 Web 技術來構建跨平台桌面應用。

## 主要變更

### 1. UI 框架變更
- **之前**: Fyne v2 (原生 Go UI)
- **之後**: Wails v2 (WebView + Go 後端)

### 2. 前端技術棧
- **HTML/CSS/JavaScript**: 使用 Vite 構建的現代 Web 前端
- **Monaco Editor**: 用於代碼編輯的專業編輯器
- **Bootstrap 5**: UI 組件庫
- **Font Awesome**: 圖標庫

### 3. 後端 API
所有的 Go 後端功能都通過 Wails 的綁定系統暴露給前端：

```go
// app.go 中的主要方法
- ExecuteCode(code string) string
- ExecuteCodeWithColorBG(code string, colorBG string) string
- GetVersion() map[string]string
- GetDefaultCode() string
- GetSymbols() []string
- SaveCode(code string) error
- SaveResult(result string) error
- OpenGitHub()
- OpenHazelnutParadise()
```

### 4. 項目結構

```
idensyra/
├── app.go                 # Wails 應用後端邏輯
├── main.go                # 應用入口點
├── wails.json             # Wails 配置文件
├── internal/              # Yaegi 提取的符號表
│   ├── ansi2html.go
│   ├── extract.go
│   └── github_com-*.go
├── frontend/              # 前端代碼
│   ├── src/
│   │   ├── main.js       # 主 JavaScript 文件
│   │   └── style.css     # 樣式表
│   ├── index.html
│   └── package.json
└── build/                 # 構建輸出目錄
    └── bin/
        └── idensyra.exe
```

### 5. 移除的文件和依賴

#### 移除的依賴
- `fyne.io/fyne/v2`
- `github.com/gorilla/websocket` (不再需要 WebSocket)

#### 保留的核心依賴
- `github.com/HazelnutParadise/insyra`
- `github.com/traefik/yaegi`
- `gorm.io/gorm`

#### 新增的依賴
- `github.com/wailsapp/wails/v2`

### 6. 備份文件
- `main.go.bak`: 原始的 Fyne UI 實現

## 開發指南

### 環境要求

1. **Go 1.23+**
2. **Node.js 16+**
3. **Wails CLI v2**

安裝 Wails CLI:
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 開發模式

啟動開發服務器（支持熱重載）：

```bash
wails dev
```

這將：
1. 啟動 Go 後端
2. 啟動前端開發服務器 (Vite)
3. 打開應用窗口
4. 監聽文件變更並自動重新加載

### 構建生產版本

構建 Windows 可執行文件：

```bash
wails build
```

構建輸出：`build/bin/idensyra.exe`

跨平台構建：

```bash
# macOS
wails build -platform darwin/amd64

# Linux
wails build -platform linux/amd64
```

### 前端開發

進入前端目錄並安裝依賴：

```bash
cd frontend
npm install
```

單獨運行前端開發服務器：

```bash
npm run dev
```

構建前端：

```bash
npm run build
```

### 重新生成 Yaegi 符號表

如果 insyra 包更新，需要重新生成符號表：

```bash
cd internal
go generate
```

## 新功能

### 1. 改進的代碼編輯器
- **Monaco Editor**: VS Code 相同的編輯器
- **語法高亮**: Go 語言語法支持
- **自動補全**: 基於符號表的智能提示
- **主題切換**: 支持明亮/暗色主題

### 2. 更好的 UI/UX
- **響應式設計**: 適配不同屏幕尺寸
- **可調整分割**: 代碼和輸出區域可調整大小
- **快捷鍵支持**:
  - `Ctrl/Cmd + Enter`: 運行代碼
  - `Ctrl/Cmd + S`: 保存代碼

### 3. Live Run 模式
- 實時執行: 編輯時自動運行代碼（帶防抖）
- 可切換: 通過復選框開關

### 4. 文件操作
- 使用原生文件對話框保存/加載代碼
- 支持保存執行結果

## 移除的功能

1. **Web UI 模式**: 不再需要獨立的 Web 服務器模式（Wails 本身就是 WebView）
2. **WebSocket 連接**: 使用 Wails 的綁定系統替代
3. **模板引擎**: 使用原生 HTML/JS 替代 Go 模板

## 性能優化

- **啟動速度**: Wails 使用原生 WebView，啟動更快
- **內存使用**: 比嵌入式瀏覽器更輕量
- **包大小**: 無需打包 Chromium，體積更小

## 已知問題

1. 首次運行需要安裝 WebView2 Runtime（Windows）
2. Monaco Editor 需要從 CDN 加載（可以考慮本地化）

## 未來計劃

- [ ] 支持更多編輯器功能（查找/替換、多光標等）
- [ ] 添加代碼片段庫
- [ ] 支持項目管理（多文件）
- [ ] 添加調試功能
- [ ] 離線 Monaco Editor 支持

## 貢獻

如果您想為項目做出貢獻，請：
1. Fork 本倉庫
2. 創建功能分支
3. 提交 Pull Request

## 許可證

本項目遵循與原項目相同的許可證。

## 鏈接

- [Wails 官方文檔](https://wails.io)
- [Monaco Editor](https://microsoft.github.io/monaco-editor/)
- [Insyra](https://insyra.hazelnut-paradise.com)