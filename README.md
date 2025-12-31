# Idensyra

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Wails](https://img.shields.io/badge/Wails-v2.11.0-DF5320?style=flat)](https://wails.io)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Idensyra** 是一個基於 Wails v2 的跨平台 Go 代碼編輯器和運行環境，專為 [Insyra](https://insyra.hazelnut-paradise.com) 數據科學庫設計。

![Idensyra Screenshot](gui_example.png)

## 特性

- 🚀 **即時執行**: 使用 Yaegi 解釋器即時運行 Go 代碼，無需編譯
- 💻 **Monaco Editor**: 集成 VS Code 同款編輯器，提供語法高亮和智能提示
- 🎨 **主題切換**: 支持明亮和暗色主題
- 📊 **Insyra 集成**: 完整支持 Insyra 數據科學庫的所有功能
- 🔄 **Live Run 模式**: 編輯時自動執行代碼
- 💾 **文件操作**: 使用原生對話框保存和加載代碼
- 🌐 **跨平台**: 支持 Windows、macOS 和 Linux
- ⚡ **輕量快速**: 使用系統原生 WebView，體積小啟動快

## 系統要求

### 開發環境
- Go 1.23 或更高版本
- Node.js 16 或更高版本
- Wails CLI v2.11.0 或更高版本

### 運行環境
- **Windows**: Windows 10/11，需要 WebView2 Runtime
- **macOS**: macOS 10.13 或更高版本
- **Linux**: 需要 WebKitGTK

## 快速開始

### 安裝 Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 克隆項目

```bash
git clone https://github.com/HazelnutParadise/idensyra.git
cd idensyra
```

### 安裝依賴

```bash
# 安裝 Go 依賴
go mod download

# 安裝前端依賴
cd frontend
npm install
cd ..
```

### 開發模式

啟動開發服務器（支持熱重載）：

```bash
wails dev
```

這將啟動應用並自動監聽文件變更。

### 構建生產版本

```bash
wails build
```

構建完成後，可執行文件位於 `build/bin/` 目錄。

## 使用方法

### 基本操作

1. **編寫代碼**: 在左側編輯器中輸入 Go 代碼
2. **運行代碼**: 點擊 "Run Code" 按鈕或按 `Ctrl/Cmd + Enter`
3. **查看結果**: 右側面板顯示執行結果

### 快捷鍵

- `Ctrl/Cmd + Enter`: 運行代碼
- `Ctrl/Cmd + S`: 保存代碼

### Live Run 模式

啟用 "Live Run" 復選框後，代碼將在您編輯時自動執行（帶有 1 秒防抖）。

### 主題切換

點擊工具欄的主題切換按鈕在明亮和暗色主題之間切換。

## 示例代碼

```go
import (
    "fmt"
    "log"
    "github.com/HazelnutParadise/insyra/isr"
    "github.com/HazelnutParadise/insyra"
    "github.com/HazelnutParadise/insyra/stats"
)

func main() {
    // 創建數據列表
    dl := insyra.NewDataList(1, 2, 3, 4, 5)
    fmt.Println("Data:", dl.Data())
    
    // 計算統計量
    mean := stats.Mean(dl)
    fmt.Printf("Mean: %.2f\n", mean)
    
    // 數據轉換
    squared := dl.Map(func(x float64) float64 {
        return x * x
    })
    fmt.Println("Squared:", squared.Data())
}
```

## 支持的包

Idensyra 支持以下 Insyra 子包：

- `insyra`: 核心數據結構
- `insyra/isr`: 數據列表和數據表操作
- `insyra/stats`: 統計分析
- `insyra/plot`: 數據可視化
- `insyra/gplot`: 高級繪圖
- `insyra/datafetch`: 數據獲取
- `insyra/csvxl`: CSV/Excel 處理
- `insyra/parallel`: 並行計算
- `insyra/lpgen`: 線性規劃
- `insyra/py`: Python 互操作

## 項目結構

```
idensyra/
├── app.go                 # Wails 應用後端邏輯
├── main.go                # 應用入口點
├── wails.json             # Wails 配置文件
├── go.mod                 # Go 模塊定義
├── internal/              # Yaegi 符號表
│   ├── ansi2html.go      # ANSI 轉 HTML
│   ├── extract.go        # 符號提取
│   └── github_com-*.go   # 提取的符號表
├── frontend/              # 前端代碼
│   ├── src/
│   │   ├── main.js       # 主 JavaScript 文件
│   │   └── style.css     # 樣式表
│   ├── index.html        # HTML 入口
│   └── package.json      # 前端依賴
└── build/                 # 構建輸出
    └── bin/
        └── idensyra.exe
```

## 開發指南

### 重新生成 Yaegi 符號表

如果 Insyra 包更新，需要重新生成符號表：

```bash
cd internal
go generate
```

### 前端開發

```bash
cd frontend
npm run dev      # 開發服務器
npm run build    # 構建生產版本
```

### 構建選項

```bash
# 默認構建（當前平台）
wails build

# 跨平台構建
wails build -platform darwin/amd64   # macOS
wails build -platform linux/amd64    # Linux
wails build -platform windows/amd64  # Windows

# 壓縮構建
wails build -upx

# 調試構建
wails build -debug
```

## 從 Fyne UI 遷移

本項目已從 Fyne UI 遷移到 Wails。詳細的遷移指南請參閱 [WAILS_MIGRATION.md](WAILS_MIGRATION.md)。

原始的 Fyne 實現已備份為 `main.go.bak`。

## 貢獻

歡迎提交 Issue 和 Pull Request！

在提交 PR 之前，請確保：

1. 代碼遵循 Go 標準格式（`go fmt`）
2. 所有測試通過
3. 添加適當的註釋
4. 更新相關文檔

詳細貢獻指南請參閱 [CONTRIBUTING.md](CONTRIBUTING.md)。

## 版本歷史

- **v0.1.0** (2024-12-31): 遷移到 Wails v2
- **v0.0.6** (2024): Fyne UI 版本

## 許可證

本項目採用 MIT 許可證。詳見 [LICENSE](LICENSE) 文件。

## 致謝

- [Wails](https://wails.io) - 跨平台桌面應用框架
- [Insyra](https://insyra.hazelnut-paradise.com) - Go 數據科學庫
- [Yaegi](https://github.com/traefik/yaegi) - Go 解釋器
- [Monaco Editor](https://microsoft.github.io/monaco-editor/) - 代碼編輯器

## 鏈接

- 官方網站: [HazelnutParadise](https://hazelnut-paradise.com)
- GitHub: [https://github.com/HazelnutParadise/idensyra](https://github.com/HazelnutParadise/idensyra)
- Insyra 文檔: [https://insyra.hazelnut-paradise.com](https://insyra.hazelnut-paradise.com)

## 支持

如果您遇到問題或有建議，請：

1. 查看 [Issue](https://github.com/HazelnutParadise/idensyra/issues) 列表
2. 提交新的 Issue
3. 參與討論

---

<div align="center">

Made with ❤️ by [HazelnutParadise](https://hazelnut-paradise.com)

如果這個項目對您有幫助，請給它一個 ⭐️！

</div>