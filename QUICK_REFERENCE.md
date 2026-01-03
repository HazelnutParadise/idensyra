# Idensyra 快速參考指南

## 新功能一覽（v0.2.0）

### 1. igonb Notebook

- 新增 `.igonb` 格式：類似 Jupyter 的互動式筆記本
- 支援 Go、Python、Markdown 三種 Cell 類型
- Cell 執行模式：單一執行、與上方一起執行、向下執行
- 拖放排序 Cell
- Markdown 即時預覽
- Go-Python 變數互操作

### 2. Python 支援

- 直接執行 `.py` 檔案
- 內建 Python 套件管理器（pip list/install/uninstall）
- 可重新安裝 Python 環境

### 3. IPython Notebook 支援

- 可開啟 `.ipynb` 檔案
- 一鍵轉換 `.ipynb` 到 `.igonb` 格式

### 4. 工作區增強

- 檔案拖放移動與排序
- 自動暫存臨時工作區內容

### 5. 工作區資料夾模式

- 可建立或開啟工作區資料夾
- Save/Save All 會將檔案寫入工作區
- 臨時工作區關閉前會提示保存

### 6. 資料夾樹與動作選單

- 支援資料夾層級
- 檔案列的 `...` 選單可 Rename / Delete
- `*` 表示已修改，`L` 表示大型檔案

### 7. 多格式預覽

- HTML/Markdown/CSV/TSV 即時預覽
- Excel 表格預覽與工作表切換
- 圖片、影片、音訊、PDF 預覽

### 8. 匯入與開啟進度提示

- 匯入與開啟工作區時顯示進度卡片

### 9. 輸出直接儲存到工作區

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

1. **Live Run** - 啟用/停用自動執行（僅 .go）
2. **Minimap** - 切換程式碼縮略圖
3. **Word Wrap** - 切換換行模式
4. **Theme** - 切換深色/淺色主題
5. **Python** - 開啟 Python 套件管理器
6. **GitHub** - 開啟 GitHub 專案頁面
7. **Insyra 官網** - 開啟 Insyra 官方網站

**工作區側邊欄：**

- **New File** - 建立新檔案
- **New Notebook** - 建立新的 igonb 筆記本
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

- **Run** - 執行程式碼（`.go` 或 `.py`）
- **Copy** - 複製輸出內容到剪貼簿
- **Save** - 將輸出結果寫入工作區

**igonb Notebook 工具列：**

- **Run All** - 執行所有 Cell
- **Stop** - 停止執行
- **Reset** - 重置執行環境（清除變數）
- **Clear All** - 清除所有輸出
- **Full/Compact** - 切換輸出顯示模式
- **Convert** - 轉換 .ipynb 到 .igonb（僅 .ipynb 檔案）
- **+ Go** - 新增 Go Cell

**igonb Cell 控制：**

- **▶** - 執行單一 Cell
- **▲▶** - 執行目前及上方所有 Cell
- **▼▶** - 從目前 Cell 向下執行
- **×** - 刪除 Cell
- **語言選擇器** - 切換 Go/Python/Markdown
- **拖動把手** - 拖放排序 Cell

---

## igonb Notebook 使用

### 建立與執行

1. 點擊 **New Notebook** 建立 `.igonb` 檔案
2. 選擇 Cell 語言（Go/Python/Markdown）
3. 輸入程式碼後點擊執行按鈕

### Cell 執行模式

- **▶ 單一執行**：只執行選中的 Cell
- **▲▶ 與上方一起**：確保所有上方 Cell 已執行
- **▼▶ 向下執行**：從目前 Cell 執行到最後
- **Run All**：執行所有 Cell

### Go-Python 互操作

```go
// Go Cell
data := isr.DL.Of(1, 2, 3, 4, 5)
result := data.Sum()
```

```python
# Python Cell - 可存取 Go 變數
print(f"Data: {data}")
print(f"Sum: {result}")
```

### 執行控制

- **Stop**：停止執行（完成當前 Cell 後停止）
- **Reset**：重置環境，清除所有變數與狀態

---

## Python 套件管理

1. 點擊標題列的 **Python** 按鈕開啟套件管理器
2. **查看套件**：顯示已安裝套件清單
3. **安裝套件**：輸入套件名稱，點擊 Install
4. **解除安裝**：點擊套件旁的移除按鈕
5. **重新安裝環境**：遇到問題時點擊 Reinstall

---

## 使用提示

### 一般使用

- Run 與 Live Run 只對 `.go` 檔案可用
- `.py` 檔案可點擊 Run 執行
- `.igonb` 檔案以 Notebook 模式顯示
- HTML/Markdown/CSV/TSV/Excel/媒體檔會進入 Preview 模式
- Save/Save All 在臨時工作區會提示建立工作區資料夾
- Output 的 Save 會建立 `result.txt` 或 `result_#.txt`
- 大型檔案與二進位檔案為預覽模式，無法直接編輯
- 拖放檔案可移動到其他資料夾

### igonb Notebook

- 適合資料分析流程與實驗
- 使用 Markdown Cell 記錄分析過程
- 善用「與上方一起執行」確保環境正確
- 遇到問題可 Reset 重置環境
- 輸出模式可切換 Full/Compact

### Python 整合

- 在 igonb 中可混用 Go 和 Python
- Go 變數會自動傳遞給 Python Cell
- 確保已安裝所需的 Python 套件

---

## 主題說明

- 首次啟動時自動偵測系統主題
- 可手動切換深色/淺色主題
- Monaco 編輯器與介面主題同步
- igonb Notebook 主題同步

---

## 快速開始範例

### Go 程式碼

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

### igonb Notebook 範例

```json
{
  "version": 1,
  "cells": [
    {
      "language": "go",
      "source": "data := isr.DL.Of(10, 20, 30)\nfmt.Println(\"Mean:\", data.Mean())"
    },
    {
      "language": "python",
      "source": "print(f\"Data from Go: {data}\")"
    },
    {
      "language": "markdown",
      "source": "## 分析完成\n以上是簡單的 Go-Python 互操作範例。"
    }
  ]
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
A: Run 只對 `.go` 和 `.py` 檔案有效，且在預覽模式時會停用。

**Q: 如何使用 igonb？**  
A: 點擊 New Notebook 建立 `.igonb` 檔案，然後新增 Cell 並選擇語言。

**Q: Go 和 Python Cell 如何共享變數？**  
A: 在 Go Cell 中定義的變數會自動傳遞給後續的 Python Cell。支援基本型別和 Insyra DataList/DataTable。

**Q: 如何安裝 Python 套件？**  
A: 點擊標題列的 Python 按鈕，在輸入框輸入套件名稱後點擊 Install。

**Q: igonb 執行卡住怎麼辦？**  
A: 點擊 Stop 停止執行，或點擊 Reset 重置環境。

**Q: .ipynb 檔案可以編輯嗎？**  
A: 可以預覽，但建議轉換為 .igonb 格式後再編輯。點擊 Convert 按鈕即可轉換。

---

**版本：** Idensyra v0.2.0 (開發中)  
**更新日期：** 2025
