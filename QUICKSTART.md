# Idensyra 快速入門指南

歡迎使用 Idensyra！這是一個基於 Wails 的 Go 代碼編輯器，專為 Insyra 數據科學庫設計。

## 📋 目錄

- [安裝](#安裝)
- [首次運行](#首次運行)
- [基本使用](#基本使用)
- [功能介紹](#功能介紹)
- [快捷鍵](#快捷鍵)
- [常見問題](#常見問題)

## 🚀 安裝

### 方法 1: 下載預編譯版本

1. 前往 [Releases](https://github.com/HazelnutParadise/idensyra/releases) 頁面
2. 下載適合您系統的版本
3. 解壓縮並運行 `idensyra.exe` (Windows) 或 `idensyra` (macOS/Linux)

### 方法 2: 從源碼構建

#### 前置要求

- **Go 1.23+**: [下載安裝](https://golang.org/dl/)
- **Node.js 16+**: [下載安裝](https://nodejs.org/)
- **Wails CLI**: 
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```

#### 構建步驟

```bash
# 1. 克隆倉庫
git clone https://github.com/HazelnutParadise/idensyra.git
cd idensyra

# 2. 安裝依賴
go mod download
cd frontend && npm install && cd ..

# 3. 構建
wails build

# 4. 運行（可執行文件在 build/bin/ 目錄）
./build/bin/idensyra
```

## 🎯 首次運行

### Windows 用戶

首次運行時，如果系統沒有安裝 WebView2 Runtime，會自動提示下載安裝。這是必需的組件。

### macOS 用戶

首次運行可能會提示"無法打開"，請前往 **系統偏好設置 > 安全性與隱私** 允許運行。

### Linux 用戶

確保已安裝 WebKitGTK：

```bash
# Ubuntu/Debian
sudo apt install libwebkit2gtk-4.0-dev

# Fedora
sudo dnf install webkit2gtk3-devel

# Arch
sudo pacman -S webkit2gtk
```

## 📝 基本使用

### 1. 啟動應用

雙擊 `idensyra.exe` 或在終端運行：

```bash
./idensyra
```

### 2. 編寫代碼

應用啟動後會看到兩個面板：

- **左側**: 代碼編輯器
- **右側**: 執行結果

默認代碼示例：

```go
import (
    "fmt"
    "log"
    "github.com/HazelnutParadise/insyra/isr"
    "github.com/HazelnutParadise/insyra"
)

func main() {
    fmt.Println("Hello, World!")
    log.Println("this is a log message")
    dl := insyra.NewDataList(1, 2, 3)
    fmt.Println("This is your data list:", dl.Data())
}
```

### 3. 運行代碼

點擊 **"Run Code"** 按鈕或按 `Ctrl+Enter` (Windows/Linux) / `Cmd+Enter` (macOS)

### 4. 查看結果

執行結果會顯示在右側面板，包括：
- 標準輸出
- 錯誤信息
- 日誌輸出

## 🎨 功能介紹

### Monaco Editor

專業的代碼編輯器，提供：
- ✅ 語法高亮
- ✅ 自動縮進
- ✅ 代碼折疊
- ✅ 括號匹配
- ✅ 迷你地圖
- ✅ 行號顯示

### 主題切換

點擊工具欄的 **🌓 圖標** 在明亮和暗色主題之間切換。您的選擇會自動保存。

### Live Run 模式

啟用 **Live Run** 復選框後：
- 代碼會在您編輯時自動執行
- 有 1 秒的防抖延遲
- 適合快速測試和調試

**注意**: 如果代碼執行時間較長，建議關閉此功能。

### 保存和加載

#### 保存代碼
1. 點擊 **"Save Code"** 按鈕
2. 選擇保存位置和文件名
3. 代碼會以 `.go` 格式保存

#### 保存結果
1. 點擊 **"Save Result"** 按鈕
2. 選擇保存位置和文件名
3. 結果會以文本格式保存

### 複製結果

點擊 **"Copy"** 按鈕將執行結果複製到剪貼板。

### 外部鏈接

- **GitHub 圖標**: 打開項目 GitHub 頁面
- **鏈接圖標**: 訪問 HazelnutParadise 網站

## ⌨️ 快捷鍵

| 快捷鍵 | 功能 |
|--------|------|
| `Ctrl/Cmd + Enter` | 運行代碼 |
| `Ctrl/Cmd + S` | 保存代碼 |
| `Ctrl/Cmd + C` | 複製選中文本 |
| `Ctrl/Cmd + V` | 粘貼 |
| `Ctrl/Cmd + Z` | 撤銷 |
| `Ctrl/Cmd + Shift + Z` | 重做 |
| `Ctrl/Cmd + F` | 查找 |
| `Ctrl/Cmd + H` | 替換 |

## 📦 支持的 Insyra 包

Idensyra 完全支持以下 Insyra 子包：

```go
import (
    "github.com/HazelnutParadise/insyra"          // 核心
    "github.com/HazelnutParadise/insyra/isr"      // 數據操作
    "github.com/HazelnutParadise/insyra/stats"    // 統計分析
    "github.com/HazelnutParadise/insyra/plot"     // 繪圖
    "github.com/HazelnutParadise/insyra/gplot"    // 高級繪圖
    "github.com/HazelnutParadise/insyra/datafetch"// 數據獲取
    "github.com/HazelnutParadise/insyra/csvxl"    // CSV/Excel
    "github.com/HazelnutParadise/insyra/parallel" // 並行計算
    "github.com/HazelnutParadise/insyra/lpgen"    // 線性規劃
    "github.com/HazelnutParadise/insyra/py"       // Python 互操作
)
```

## 💡 示例代碼

### 數據列表操作

```go
import (
    "fmt"
    "github.com/HazelnutParadise/insyra"
)

func main() {
    // 創建數據列表
    dl := insyra.NewDataList(1, 2, 3, 4, 5)
    
    // 打印數據
    fmt.Println("Original:", dl.Data())
    
    // 數據轉換
    squared := dl.Map(func(x float64) float64 {
        return x * x
    })
    fmt.Println("Squared:", squared.Data())
    
    // 過濾數據
    filtered := dl.Filter(func(x float64) bool {
        return x > 2
    })
    fmt.Println("Filtered (>2):", filtered.Data())
}
```

### 統計分析

```go
import (
    "fmt"
    "github.com/HazelnutParadise/insyra"
    "github.com/HazelnutParadise/insyra/stats"
)

func main() {
    data := insyra.NewDataList(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
    
    fmt.Printf("Mean: %.2f\n", stats.Mean(data))
    fmt.Printf("Median: %.2f\n", stats.Median(data))
    fmt.Printf("StdDev: %.2f\n", stats.StdDev(data))
    fmt.Printf("Variance: %.2f\n", stats.Variance(data))
}
```

### 數據可視化

```go
import (
    "github.com/HazelnutParadise/insyra"
    "github.com/HazelnutParadise/insyra/plot"
)

func main() {
    x := insyra.NewDataList(1, 2, 3, 4, 5)
    y := insyra.NewDataList(2, 4, 6, 8, 10)
    
    // 創建折線圖
    p := plot.NewLinePlot(x, y)
    p.SetTitle("Line Plot Example")
    p.SetXLabel("X Axis")
    p.SetYLabel("Y Axis")
    
    // 保存圖表
    p.Save("line_plot.png")
}
```

## ❓ 常見問題

### Q: 為什麼代碼無法執行？

**A**: 請檢查：
1. 是否有語法錯誤
2. 導入的包是否正確
3. 是否使用了不支持的第三方包
4. 查看右側面板的錯誤信息

### Q: 支持哪些 Go 標準庫？

**A**: Idensyra 使用 Yaegi 解釋器，支持大部分 Go 標準庫。不支持的包括：
- `unsafe`
- 一些特殊的編譯器功能
- CGO 相關功能

### Q: 可以使用第三方包嗎？

**A**: 目前只支持：
- Go 標準庫
- Insyra 及其子包
- 項目 internal 目錄中預先提取的符號

不支持其他第三方包。

### Q: 如何提高執行速度？

**A**: 
1. 關閉 Live Run 模式
2. 避免大量循環和重計算
3. 使用 Insyra 的並行計算功能

### Q: 代碼會保存在哪裡？

**A**: 
- 代碼不會自動保存
- 需要手動使用 "Save Code" 功能
- 可以保存到任意位置

### Q: 如何更新 Insyra 庫？

**A**: 
如果您從源碼構建，可以：
```bash
go get -u github.com/HazelnutParadise/insyra@latest
cd internal
go generate
wails build
```

### Q: 為什麼 Windows 首次運行需要安裝 WebView2？

**A**: Wails 使用 Windows 系統的 WebView2 組件來渲染界面。這是一個輕量級的組件，安裝一次即可。

### Q: 可以修改編輯器字體大小嗎？

**A**: 目前默認字體大小為 14px。未來版本會加入設置面板。

### Q: 如何報告 Bug？

**A**: 
1. 訪問 [GitHub Issues](https://github.com/HazelnutParadise/idensyra/issues)
2. 搜索是否已有相同問題
3. 如果沒有，創建新 Issue 並提供：
   - 操作系統和版本
   - Idensyra 版本
   - 重現步驟
   - 錯誤截圖或日誌

## 🔗 相關資源

- **項目主頁**: https://github.com/HazelnutParadise/idensyra
- **Insyra 文檔**: https://insyra.hazelnut-paradise.com
- **Wails 文檔**: https://wails.io
- **HazelnutParadise**: https://hazelnut-paradise.com

## 💬 社群和支持

如果您有任何問題或建議：

1. 查看 [README.md](README.md) 和 [WAILS_MIGRATION.md](WAILS_MIGRATION.md)
2. 搜索 [GitHub Issues](https://github.com/HazelnutParadise/idensyra/issues)
3. 創建新的 Issue
4. 參與討論

## 🎓 學習資源

### Insyra 教程

訪問 [Insyra 官方文檔](https://insyra.hazelnut-paradise.com) 學習：
- 數據結構
- 統計分析
- 數據可視化
- 機器學習
- 更多進階功能

### Go 語言學習

- [Go 官方教程](https://go.dev/tour/)
- [Go by Example](https://gobyexample.com/)
- [Effective Go](https://go.dev/doc/effective_go)

---

**祝您使用愉快！** 🎉

如果 Idensyra 對您有幫助，請給我們一個 ⭐️！