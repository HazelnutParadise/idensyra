# Idensyra 快速參考指南

## 🎯 新功能一覽

### 1. 系統主題跟隨
程式啟動時會自動偵測並跟隨系統的深色/淺色主題設定。

### 2. Minimap 開關
點擊工具列的 **地圖圖示 (📍)** 開啟/關閉編輯器右側的 Minimap。

### 3. 自動換行開關
點擊工具列的 **文字寬度圖示 (↔️)** 切換長行文字的顯示方式：
- **關閉**：顯示水平捲軸
- **開啟**：自動換行顯示

### 4. 復原/重做
- **復原**：`Ctrl + Z` (Win/Linux) 或 `Cmd + Z` (Mac)
- **重做**：`Ctrl + Shift + Z` 或 `Ctrl + Y` (Win/Linux)，`Cmd + Shift + Z` 或 `Cmd + Y` (Mac)

### 5. Live Run 狀態通知
啟用或停用 Live Run 時會顯示通知訊息，清楚顯示目前狀態。

---

## ⌨️ 鍵盤快捷鍵

| 功能 | Windows/Linux | macOS |
|------|--------------|-------|
| 執行程式碼 | `Ctrl + Enter` | `Cmd + Enter` |
| 儲存程式碼 | `Ctrl + S` | `Cmd + S` |
| 復原 | `Ctrl + Z` | `Cmd + Z` |
| 重做 | `Ctrl + Shift + Z` 或 `Ctrl + Y` | `Cmd + Shift + Z` 或 `Cmd + Y` |
| 自動完成 | `Ctrl + Space` | `Ctrl + Space` |

---

## 🎨 工具列按鈕說明

**標題列右側（由左至右）：**

1. **Live Run 核取方塊** - 啟用/停用自動執行
2. **地圖圖示 (📍)** - 切換 Minimap
3. **文字寬度圖示 (↔️)** - 切換自動換行
4. **調色圖示 (🎨)** - 切換深色/淺色主題
5. **GitHub 圖示** - 開啟 GitHub 專案頁面
6. **連結圖示 (🔗)** - 開啟 HazelnutParadise 網站

**編輯器區域：**
- **Font +/-** - 調整編輯器字體大小（8-32px）
- **Load** - 從檔案載入程式碼
- **Save** - 儲存程式碼到檔案

**輸出區域：**
- **Font +/-** - 調整輸出字體大小（8-32px）
- **Run Code** - 執行程式碼（或按 `Ctrl + Enter`）
- **Copy** - 複製輸出內容到剪貼簿
- **Save** - 儲存輸出內容到檔案

---

## 💡 使用提示

### 啟用的功能會高亮顯示
當 Minimap 或自動換行啟用時，對應的按鈕會變成藍色，方便識別目前狀態。

### 所有設定都會自動儲存
以下設定會在關閉程式後保留：
- 主題選擇（深色/淺色）
- Minimap 開啟/關閉
- 自動換行開啟/關閉
- 編輯器字體大小
- 輸出區字體大小
- 面板分割比例

### 自動完成功能
在編輯器中輸入套件名稱後按 `Ctrl + Space` 可以看到：
- Go 語言關鍵字
- 標準庫函式（fmt, strings, time 等）
- Insyra 庫函式
- 常用型別

### 面板調整
拖曳編輯器與輸出區之間的分隔線可以調整兩邊的大小。

---

## 🎨 主題說明

### 自動跟隨系統
- 首次啟動時自動偵測系統主題
- Windows 10/11：跟隨「設定 > 個人化 > 色彩」
- macOS：跟隨「系統偏好設定 > 一般 > 外觀」

### 手動切換
- 點擊工具列的調色圖示切換主題
- 手動設定會覆蓋系統設定
- 下次啟動時會記住手動選擇

---

## 🚀 快速開始範例

### 1. 基本 Hello World
```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, Idensyra!")
}
```

### 2. 使用 Insyra 彩色輸出
```go
package main

import "insyra"

func main() {
    insyra.Println(insyra.Green("成功訊息"))
    insyra.Println(insyra.Yellow("警告訊息"))
    insyra.Println(insyra.Red("錯誤訊息"))
}
```

### 3. 啟用 Live Run 體驗
1. 勾選工具列的「Live Run」
2. 在編輯器中修改程式碼
3. 稍等片刻，程式會自動執行並顯示結果

---

## ❓ 常見問題

**Q: 如何恢復預設設定？**  
A: 清除瀏覽器的 localStorage 或刪除設定檔（未來版本會提供重設按鈕）。

**Q: Minimap 和自動換行哪個比較好？**  
A: 依個人喜好：
- Minimap：適合長檔案，方便快速導航
- 自動換行：適合編寫長行程式碼，避免水平捲動

**Q: 復原/重做有步數限制嗎？**  
A: 使用 Monaco Editor 的內建功能，通常可以復原數百步。

---

## 📦 無 CDN 依賴

所有資源（Monaco Editor、Bootstrap、Font Awesome）都已本地化打包，無需網路連線即可使用。

---

**版本：** Idensyra with Wails v2.11.0  
**更新日期：** 2025-12-31