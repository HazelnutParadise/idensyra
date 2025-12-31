# Idensyra

**Idensyra** 是一個基於 Wails v2 的跨平台 Go 代碼編輯器和運行環境，專為 [Insyra](https://insyra.hazelnut-paradise.com) 數據科學庫設計。

![Idensyra Screenshot](gui_example.png)

## ✨ 特性

### 核心功能
- 🚀 **即時執行**: 使用 Yaegi 解釋器即時運行 Go 代碼，無需編譯
- 💻 **Monaco Editor**: 集成 VS Code 同款編輯器，提供完整的語法高亮和智能提示
- 📊 **Insyra 集成**: 完整支持 Insyra 數據科學庫的所有功能
- 🔄 **Live Run 模式**: 編輯時自動執行代碼（可切換）
- 💾 **文件操作**: 使用原生對話框保存和加載代碼
- 🌐 **跨平台**: 支持 Windows、macOS 和 Linux
- ⚡ **輕量快速**: 使用系統原生 WebView，體積小啟動快
- 📦 **完全本地化**: 無 CDN 依賴，離線可用

### 編輯器功能
- ↩️ **復原/重做**: 完整的 Undo/Redo 支援，快捷鍵 `Ctrl+Z` / `Ctrl+Shift+Z`
- 🗺️ **Minimap 開關**: 可選擇顯示或隱藏程式碼縮略圖
- 🔄 **自動換行**: 可選擇長行自動換行或水平捲動
- 🔍 **智能提示**: Go 關鍵字、標準庫和 Insyra 函式自動完成
- 🎯 **多游標編輯**: 支援多行同時編輯
- 📏 **程式碼摺疊**: 折疊/展開程式碼區塊

### 主題系統
- 🎨 **跟隨系統主題**: 自動偵測並跟隨作業系統的深色/淺色設定
- 🌓 **手動切換主題**: 可隨時在深色與淺色主題間切換
- 🎭 **主題一致性**: Monaco 編輯器與整個程式完全同步
- 💾 **設定持久化**: 所有主題和編輯器設定自動儲存

### 使用者體驗
- 🔔 **智慧通知**: 操作回饋清晰，新通知自動隱藏舊通知
- 🎯 **視覺指示**: 啟用的功能按鈕顯示高亮狀態
- ⌨️ **完整快捷鍵**: 所有主要功能都支援鍵盤快捷鍵
- 🖱️ **可調整面板**: 拖曳調整編輯器與輸出區大小
- 🎨 **ANSI 顏色**: 支援彩色輸出，深色/淺色主題皆清晰可讀

## 系統要求

### 開發環境
- Go 1.25 或更高版本
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

| 功能 | Windows/Linux | macOS |
|------|--------------|-------|
| 執行程式碼 | `Ctrl + Enter` | `Cmd + Enter` |
| 儲存程式碼 | `Ctrl + S` | `Cmd + S` |
| 復原 | `Ctrl + Z` | `Cmd + Z` |
| 重做 | `Ctrl + Shift + Z` 或 `Ctrl + Y` | `Cmd + Shift + Z` 或 `Cmd + Y` |
| 自動完成 | `Ctrl + Space` | `Ctrl + Space` |

### 工具列功能

**標題列右側按鈕（由左至右）：**
1. **Live Run 核取方塊** - 啟用/停用自動執行
2. **Minimap 按鈕** (📍) - 切換程式碼縮略圖
3. **自動換行按鈕** (↔️) - 切換換行模式
4. **主題切換按鈕** (🎨) - 切換深色/淺色主題
5. **GitHub 按鈕** - 開啟專案 GitHub 頁面
6. **HazelnutParadise 按鈕** (🔗) - 開啟官方網站

### Live Run 模式

啟用 "Live Run" 復選框後，代碼將在您編輯時自動執行（帶有防抖機制）。

### 編輯器設定

- **字體大小**: 使用 +/- 按鈕調整（8-32px）
- **Minimap**: 點擊工具列按鈕開啟/關閉
- **自動換行**: 點擊工具列按鈕切換換行/捲動模式
- 所有設定會自動儲存並在下次啟動時恢復

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
    
    // 彩色輸出
    insyra.Println(insyra.Green("成功！"))
    insyra.Println(insyra.Yellow("警告"))
    insyra.Println(insyra.Red("錯誤"))
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

以及完整的 Go 標準庫支援。

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
│   ├── package.json      # 前端依賴
│   └── vite.config.js    # Vite 配置
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

## 技術架構

### 前端
- **框架**: Vanilla JavaScript + Vite
- **編輯器**: Monaco Editor v0.55.1（本地化）
- **UI 框架**: Bootstrap v5.3.8（本地化）
- **圖標**: Font Awesome v7.1.0（本地化）

### 後端
- **框架**: Wails v2.11.0
- **解釋器**: Yaegi v0.16.1
- **核心庫**: Insyra v0.2.10

## 更新日誌

### v0.1.0 (2025-12-31)
- 🎉 從 Fyne UI 遷移到 Wails v2
- 💻 整合 Monaco Editor
- 🎨 新增主題切換功能與跟隨系統主題
- 🔄 新增 Live Run 模式
- 📊 整合 ANSI 顏色輸出
- ✨ 新增復原/重做功能（Undo/Redo）
- ✨ 新增 Minimap 開關功能
- ✨ 新增自動換行切換功能
- 💡 智慧通知管理
- 📦 完全本地化所有前端資源
- 🐛 修正自動完成問題

## 貢獻

歡迎提交 Issue 和 Pull Request！

在提交 PR 之前，請確保：

1. 代碼遵循 Go 標準格式（`go fmt`）
2. 所有測試通過
3. 添加適當的註釋
4. 更新相關文檔

## 📚 文檔

- **[README.md](README.md)** - 主要說明文件（本文件）
- **[CHANGELOG.md](CHANGELOG.md)** - 版本更新日誌
- **[FEATURES.md](FEATURES.md)** - 完整功能總覽（中文）
- **[QUICK_REFERENCE.md](QUICK_REFERENCE.md)** - 快速參考指南（中文）
- **[CONTRIBUTING.md](CONTRIBUTING.md)** - 貢獻指南
- **[LICENSE](LICENSE)** - MIT 授權條款

## 許可證

本項目採用 MIT 許可證。詳見 [LICENSE](LICENSE) 文件。

## 致謝

- [Wails](https://wails.io) - 跨平台桌面應用框架
- [Insyra](https://insyra.hazelnut-paradise.com) - Go 數據科學庫
- [Yaegi](https://github.com/traefik/yaegi) - Go 解釋器
- [Monaco Editor](https://microsoft.github.io/monaco-editor/) - 代碼編輯器
- [Bootstrap](https://getbootstrap.com/) - UI 框架
- [Font Awesome](https://fontawesome.com/) - 圖標庫

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
