# Idensyra 功能總覽

## 目錄

- [核心功能](#核心功能)
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

### Live Run（即時執行）

- 程式碼變更後自動執行（防抖）
- 可隨時啟用/停用
- 狀態通知回饋

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
- 新增資料夾
- 重新命名/刪除（檔案列動作選單）
- 無法刪除最後一個檔案
- 匯入外部檔案（任意格式）
- 匯出目前檔案到指定位置

### 檔案狀態與提示

- `*` 表示已修改
- `L` 表示大型檔案
- 圖示區分資料夾、圖片、Markdown 等常見格式

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
- 多語言語法高亮（Go、JS/TS、HTML、CSS、JSON、Markdown 等）
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

---

## 使用者介面

- 標題列：Live Run、Minimap、Word Wrap、Theme、GitHub、Insyra 官網
- 工作區工具列：New File、New Folder、Import、Open Workspace、Save All
- 檔案列動作選單：Rename / Delete
- 匯入與開啟工作區顯示進度卡片
- 面板可拖曳調整寬度（20%-80%）

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

- Run 僅對 `.go` 檔案可用
- Live Run 支援自動執行
- Copy 可複製輸出內容
- Save 會將輸出寫入工作區（`result.txt` 或 `result_#.txt`）

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
- Wails: v2.11.0
- Insyra: v0.2.11
- Yaegi: v0.16.1
- Excelize: v2.10.0

### 跨平台支援

- Windows：WebView2
- macOS：WebKit
- Linux：WebKitGTK（需安裝）

---

## 使用建議

- 長檔案可開啟 Minimap 方便導航
- 長行程式碼可開啟 Word Wrap
- 測試程式碼可開啟 Live Run
- 需要永久保存請先建立工作區資料夾再 Save
- 輸出結果會存入工作區，可當作分析紀錄

---

## 功能狀態

### 已實現

- 工作區資料夾建立/開啟/保存
- 檔案與資料夾樹狀管理
- 匯入/匯出與重新命名
- 多格式預覽與媒體顯示
- Go 智慧提示與 Live Run
- Minimap 與 Word Wrap
- ANSI 彩色輸出

### 未來可能改進

- 自訂快捷鍵
- 更多編輯器主題
- 片段/snippet 管理
- 整合式偵錯工具
- 工作區設定匯出/匯入
- ZIP 格式匯出選項

---

**版本：** Idensyra with Wails v2.11.0  
**更新日期：** 2025-12-31  
**文件版本：** 1.1
