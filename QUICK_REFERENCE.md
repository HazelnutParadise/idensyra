# Idensyra 快速參考指南

## 新功能一覽

### 1. 工作區資料夾模式

- 可建立或開啟工作區資料夾
- Save/Save All 會將檔案寫入工作區
- 臨時工作區關閉前會提示保存

### 2. 資料夾樹與動作選單

- 支援資料夾層級
- 檔案列的 `...` 選單可 Rename / Delete
- `*` 表示已修改，`L` 表示大型檔案

### 3. 多格式預覽

- HTML/Markdown/CSV/TSV 即時預覽
- Excel 表格預覽與工作表切換
- 圖片、影片、音訊、PDF 預覽

### 4. 匯入與開啟進度提示

- 匯入與開啟工作區時顯示進度卡片

### 5. 輸出直接儲存到工作區

- Output 的 Save 會寫入 `result.txt`（或 `result_#.txt`）

---

## 鍵盤快捷鍵

| 功能         | Windows/Linux                    | macOS                          |
| ------------ | -------------------------------- | ------------------------------ |
| 執行程式碼   | `Ctrl + Enter`                   | `Cmd + Enter`                  |
| 儲存目前檔案 | `Ctrl + S`                       | `Cmd + S`                      |
| 儲存全部     | `Ctrl + Shift + S`               | `Cmd + Shift + S`              |
| 新增檔案     | `Ctrl + N`                       | `Cmd + N`                      |
| 復原         | `Ctrl + Z`                       | `Cmd + Z`                      |
| 重做         | `Ctrl + Shift + Z` 或 `Ctrl + Y` | `Cmd + Shift + Z` 或 `Cmd + Y` |
| 自動完成     | `Ctrl + Space`                   | `Ctrl + Space`                 |

---

## 工具列按鈕說明

**標題列右側（由左至右）：**

1. **Live Run** - 啟用/停用自動執行
2. **Minimap** - 切換程式碼縮略圖
3. **Word Wrap** - 切換換行模式
4. **Theme** - 切換深色/淺色主題
5. **GitHub** - 開啟 GitHub 專案頁面
6. **Insyra 官網** - 開啟 Insyra 官方網站

**工作區側邊欄：**

- **New File** - 建立新檔案
- **New Folder** - 建立資料夾
- **Import File** - 匯入外部檔案到工作區
- **Open Workspace** - 開啟現有工作區資料夾
- **Save All** - 儲存全部檔案
- **... 選單** - 重新命名/刪除
- **`*` 指示器** - 檔案已修改
- **`L` 指示器** - 大型檔案（預覽模式）

**編輯器區域：**

- **Font +/-** - 調整編輯器字體大小（8-32px）
- **Save** - 儲存目前檔案（`Ctrl/Cmd + S`）
- **Export** - 匯出目前檔案到指定位置

**輸出區域：**

- **Run** - 執行程式碼（僅 `.go`）
- **Copy** - 複製輸出內容到剪貼簿
- **Save** - 將輸出結果寫入工作區

---

## 使用提示

- Run 與 Live Run 只對 `.go` 檔案可用
- HTML/Markdown/CSV/TSV/Excel/媒體檔會進入 Preview 模式
- Save/Save All 在臨時工作區會提示建立工作區資料夾
- Output 的 Save 會建立 `result.txt` 或 `result_#.txt`
- 大型檔案與二進位檔案為預覽模式，無法直接編輯

---

## 主題說明

- 首次啟動時自動偵測系統主題
- 可手動切換深色/淺色主題
- Monaco 編輯器與介面主題同步

---

## 快速開始範例

```go
import (
    "fmt"
    "log"

    "github.com/HazelnutParadise/insyra"
    "github.com/HazelnutParadise/insyra/isr"
)

func main() {
    fmt.Println("Hello, Idensyra!")
    log.Println("this is a log message")
    dl := isr.DL.Of(1, 2, 3)
    insyra.Show("My_Data", dl)
}
```

---

## 常見問題

**Q: 工作區檔案會永久保存嗎？**  
A: 臨時工作區不會永久保存。請建立或開啟工作區資料夾後再 Save/Save All。

**Q: Save 和 Export 有什麼不同？**  
A: Save 會寫入工作區資料夾，Export 會輸出到任意位置。

**Q: Import 與 Open Workspace 有什麼不同？**  
A: Import 只加入單一檔案，Open Workspace 會載入整個資料夾並取代目前工作區。

**Q: Output Save 儲存在哪裡？**  
A: 寫入工作區的 `result.txt`（若重複則遞增編號）。

**Q: 為什麼 Run 不能點？**  
A: Run 只對 `.go` 檔案有效，且在預覽模式時會停用。

---

**版本：** Idensyra with Wails v2.11.0  
**更新日期：** 2025-12-31
