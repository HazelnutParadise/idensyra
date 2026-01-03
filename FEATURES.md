# Idensyra 功能總覽

## 目錄

- [核心功能](#核心功能)
- [igonb Notebook](#igonb-notebook)
- [Python 支援](#python-支援)
- [工作區與檔案管理](#工作區與檔案管理)
- [檔案預覽](#檔案預覽)
- [編輯器功能](#編輯器功能)
- [主題系統](#主題系統)
- [使用者介面](#使用者介面)
- [快捷鍵](#快捷鍵)
- [輸出與執行](#輸出與執行)
- [設定持久化](#設定持久化)
- [技術特點](#技術特點)
- [使用建議](#使用建議)
- [功能狀態](#功能狀態)

---

## 核心功能

### Go 程式碼執行引擎

- 基於 Yaegi 直譯器即時執行
- 支援 Insyra 與 Go 標準庫
- 只對 `.go` 檔案提供 Run
- 錯誤訊息與 ANSI 彩色輸出即時顯示
- 支援 range over integers 語法（Go 1.22+）

### Live Run（即時執行）

- 程式碼變更後自動執行（防抖）
- 可隨時啟用/停用
- 狀態通知回饋

---

## igonb Notebook

### 概述

igonb 是 Idensyra 專屬的互動式筆記本格式，類似 Jupyter Notebook，但專為 Go 與 Insyra 生態系統設計。

### 檔案格式

- 副檔名：`.igonb`
- JSON 格式，包含版本、cells 陣列、metadata
- 每個 cell 包含：id、language、source、output、error

### 支援語言

- **Go**：完整 Insyra 與標準庫支援
- **Python**：透過 Insyra py 模組執行
- **Markdown**：即時渲染預覽

### Cell 執行模式

- **單一執行**（▶）：只執行選中的 Cell
- **與上方一起執行**（▲▶）：執行目前及所有上方的 Cell
- **向下執行**（▼▶）：從目前 Cell 向下執行全部
- **Run All**：執行所有 Cell

### Cell 管理

- 拖放排序：拖動 Cell 左側把手可重新排序
- 新增 Cell：點擊 + 按鈕
- 刪除 Cell：點擊 × 按鈕
- 摺疊/展開：點擊 Cell 標題列

### 執行控制

- **Stop**：停止目前執行（完成當前 Cell 後停止）
- **Reset**：重置執行環境，清除所有變數與狀態
- **Clear All**：清除所有 Cell 的輸出

### 輸出模式

- **Full**：完整顯示所有輸出
- **Compact**：精簡顯示，減少輸出高度

### Go-Python 互操作

- Go Cell 中定義的變數可在 Python Cell 中使用
- 支援的型別：
  - 基本型別（int、float、string、bool）
  - 切片與陣列
  - Insyra DataList 與 DataTable
- Python Cell 的結果可回傳到 Go 環境

### 自動保存

- 編輯 igonb 內容後自動保存
- 約 1 秒防抖延遲

---

## Python 支援

### Python 檔案執行

- 直接執行 `.py` 檔案
- 透過 Insyra py 模組執行
- 支援 UTF-8 編碼輸出

### Python 套件管理器

- **查看已安裝套件**：顯示套件名稱與版本
- **安裝套件**：輸入套件名稱後點擊 Install
- **解除安裝套件**：點擊套件旁的移除按鈕
- **重新安裝環境**：遇到問題時可重建 Python 環境

### IPython Notebook 支援

- 可開啟 `.ipynb` 檔案
- 顯示轉換按鈕，一鍵轉換為 `.igonb` 格式

---

## 工作區與檔案管理

### 臨時工作區與工作區資料夾

- 啟動時建立臨時工作區
- 可建立新工作區資料夾或開啟現有資料夾
- Save/Save All 會把檔案寫回工作區資料夾
- 開啟工作區前會提醒未儲存變更
- 關閉應用時會提示保存臨時工作區

### 檔案與資料夾操作

- 新增檔案（`Ctrl/Cmd + N`）
- 新增 igonb 筆記本
- 新增資料夾
- 重新命名/刪除（檔案列動作選單）
- 無法刪除最後一個檔案
- 匯入外部檔案（任意格式）
- 匯出目前檔案到指定位置
- **拖放移動**：可拖動檔案到資料夾中

### 檔案狀態與提示

- `*` 表示已修改
- `L` 表示大型檔案
- 圖示區分資料夾、圖片、Markdown、Python、igonb 等常見格式

### 自動暫存

- 編輯後約 1 秒同步到工作區暫存區
- Save/Save All 才會寫入工作區資料夾

---

## 檔案預覽

- HTML 直接在輸出面板預覽
- Markdown 即時渲染預覽
- CSV/TSV 表格預覽
- Excel（`.xlsx`/`.xlsm`/`.xltx`/`.xltm`）支援工作表切換預覽
- 圖片、影片、音訊、PDF 皆可預覽
- 二進位或大型檔案顯示預覽提示，避免誤編輯

---

## 編輯器功能

### Monaco Editor 整合

- Monaco Editor 本地化打包（無 CDN）
- 多語言語法高亮（Go、JS/TS、HTML、CSS、JSON、Markdown、Python 等）
- 括號配對、程式碼摺疊、行號顯示
- 多游標編輯與常用快捷鍵

### Go 智慧提示

- Go 關鍵字與常用型別
- 標準庫函式與 Insyra 函式
- struct 欄位與方法提示
- 觸發快捷鍵：`Ctrl + Space`

### 編輯器視圖

- Minimap（可開啟/關閉）
- Word Wrap（可開啟/關閉）
- 字體大小調整（8-32px）

---

## 主題系統

- 跟隨系統深/淺色主題
- 手動切換主題
- Monaco 編輯器與整體主題同步
- igonb Notebook 主題同步

---

## 使用者介面

### 標題列按鈕（由左至右）

1. **Live Run** - 啟用/停用自動執行
2. **Minimap** - 切換程式碼縮略圖
3. **Word Wrap** - 切換換行模式
4. **Theme** - 切換深色/淺色主題
5. **Python** - 開啟 Python 套件管理器
6. **GitHub** - 開啟 GitHub 專案頁面
7. **Insyra 官網** - 開啟 Insyra 官方網站

### 工作區工具列

- **New File** - 建立新檔案
- **New Notebook** - 建立新的 igonb 筆記本
- **New Folder** - 建立資料夾
- **Import** - 匯入外部檔案
- **Open Workspace** - 開啟現有工作區資料夾
- **Save All** - 儲存全部檔案

### 檔案列動作選單

- Rename - 重新命名
- Delete - 刪除

### 進度與通知

- 匯入與開啟工作區顯示進度卡片
- 操作結果顯示通知訊息

### 面板調整

- 可拖曳調整編輯器與輸出區寬度（20%-80%）

---

## 快捷鍵

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

## 輸出與執行

### 一般檔案

- **Run** 對 `.go` 和 `.py` 檔案可用
- **Live Run** 僅支援 `.go` 檔案
- **Copy** 可複製輸出內容
- **Save** 會將輸出寫入工作區（`result.txt` 或 `result_#.txt`）

### igonb Notebook

- **Run All** 執行所有 Cell
- **Stop** 停止執行
- **Reset** 重置環境
- **Clear All** 清除所有輸出
- **Full/Compact** 切換輸出模式

---

## 設定持久化

以下設定會保留於 localStorage：

- 主題選擇（深色/淺色）
- Minimap 狀態
- Word Wrap 狀態

字體大小與面板比例為當次啟動設定。

---

## 技術特點

### 完全本地化

- Monaco Editor、Bootstrap、Font Awesome 全部本地化
- 無需網路連線即可使用

### 相依套件版本

- Monaco Editor: v0.55.1
- Bootstrap: v5.3.8
- Font Awesome: v7.1.0
- marked.js: 用於 Markdown 渲染
- Wails: v2.11.0
- Insyra: v0.2.12
- Yaegi: v0.16.1
- Excelize: v2.10.0
- Go: 1.25

### 跨平台支援

- Windows：WebView2
- macOS：WebKit
- Linux：WebKitGTK（需安裝）

---

## 使用建議

### 一般使用

- 長檔案可開啟 Minimap 方便導航
- 長行程式碼可開啟 Word Wrap
- 測試程式碼可開啟 Live Run
- 需要永久保存請先建立工作區資料夾再 Save

### igonb Notebook

- 資料分析流程適合使用 igonb
- 使用 Markdown Cell 記錄分析過程
- 善用「與上方一起執行」確保環境正確
- 遇到問題可 Reset 重置環境

### Python 整合

- 確保已安裝所需套件
- Go-Python 互操作適合混合運算場景
- 大量 Python 運算建議直接使用 .py 檔案

---

## 功能狀態

### 已實現

- 工作區資料夾建立/開啟/保存
- 檔案與資料夾樹狀管理
- 檔案拖放移動
- 匯入/匯出與重新命名
- 多格式預覽與媒體顯示
- Go 智慧提示與 Live Run
- Minimap 與 Word Wrap
- ANSI 彩色輸出
- **igonb Notebook 完整支援**
- **Python 檔案執行**
- **Python 套件管理器**
- **Go-Python 互操作**
- **IPython Notebook 開啟與轉換**

### 未來可能改進

- 自訂快捷鍵
- 更多編輯器主題
- 片段/snippet 管理
- 整合式偵錯工具
- 工作區設定匯出/匯入
- ZIP 格式匯出選項
- igonb Cell 輸出圖表顯示
- 更多 Python-Go 型別支援

---

**版本：** Idensyra v0.2.0 (開發中)  
**更新日期：** 2025  
**文件版本：** 2.0
