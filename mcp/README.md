# Idensyra MCP Server

Idensyra MCP Server 提供了一個模型上下文協議 (Model Context Protocol, MCP) 接口，允許 AI 代理與 Idensyra 工作區進行交互。

## 功能

### 文件操作
- `read_file` - 讀取文件內容（自動切換到該文件）
- `write_file` - 寫入或覆蓋文件（自動切換到該文件）
- `create_file` - 創建新文件（自動切換到該文件）
- `delete_file` - 刪除文件
- `rename_file` - 重命名或移動文件
- `list_files` - 列出目錄中的文件

### 代碼執行
- `execute_go_file` - 執行 Go 文件（使用 Yaegi 解釋器，自動切換到該文件）
- `execute_go_code` - 直接執行 Go 代碼
- `execute_python_file` - 執行 Python 文件（自動切換到該文件）
- `execute_python_code` - 直接執行 Python 代碼

### Notebook 操作 (igonb/ipynb)
- `modify_cell` - 修改特定儲存格（自動切換到該 notebook）
- `insert_cell` - 在指定位置插入儲存格（自動切換到該 notebook）
- `execute_cell` - 執行特定儲存格（自動切換到該 notebook）
- `execute_cell_and_after` - 執行某格及其之後的所有儲存格（自動切換到該 notebook）
- `execute_before_and_cell` - 執行某格之前及該儲存格（自動切換到該 notebook）
- `execute_all_cells` - 執行所有儲存格（自動切換到該 notebook）
- `convert_ipynb_to_igonb` - 將 ipynb 轉換為 igonb 格式

### 自動切換文件
當 AI 代理執行以下操作時，介面會自動切換到對應的文件：
- 讀取文件
- 編輯文件
- 創建文件
- 執行文件（Go 或 Python）
- 修改或執行 notebook 儲存格

這讓用戶可以即時看到 AI 代理正在操作的文件，提供更好的視覺反饋和上下文感知。

### 工作區管理
- `open_workspace` - 打開工作區目錄
- `save_temp_workspace` - 將臨時工作區保存到指定路徑
- `save_changes` - 保存所有未保存的更改
- `get_workspace_info` - 獲取當前工作區信息
- `create_workspace_directory` - 在工作區中創建新目錄
- `import_file_to_workspace` - 從電腦匯入特定檔案到工作區

## 權限配置

MCP 服務器支持三種權限級別：

- `PermissionAlways` - 無需確認即可執行操作
- `PermissionAsk` - 每次操作前需要確認（默認）
- `PermissionDeny` - 拒絕所有操作

可配置的權限項目：
- `FileEdit` - 文件編輯權限
- `FileRename` - 文件重命名權限
- `FileCreate` - 文件創建權限
- `FileDelete` - 文件刪除權限
- `ExecuteGo` - Go 代碼執行權限
- `ExecutePython` - Python 代碼執行權限
- `NotebookModify` - Notebook 修改權限
- `NotebookExecute` - Notebook 執行權限
- `WorkspaceOpen` - 工作區打開權限
- `WorkspaceSave` - 工作區保存權限
- `WorkspaceModify` - 工作區修改權限

## 使用方法

### 傳輸與連線方式（推薦）

Idensyra 支援兩種主要的 MCP 傳輸方式，請根據使用場景選擇：

- **整合於 GUI（推薦）**：當 Idensyra 以 GUI 模式執行時，會使用官方的 Model Context Protocol SDK（SSE/HTTP）來暴露 MCP 服務。此情況下，主程式會建立一個 SSE‑based HTTP handler（由 host 決定監聽的位址與埠）。在 Idensyra 中，MCP 伺服器預設監聽埠為 **14320**（例如 `http://localhost:14320`），但最終位址與埠仍以宿主應用決定。

- **獨立命令列工具（可選）**：`cmd/mcp-server` 提供一個輕量的 standalone 實作，適合在本機測試或 CLI 使用。該工具直接操作本地檔案系統，並使用 **stdin/stdout** 傳送 JSON ToolRequest/ToolResponse 物件。範例（傳送單筆請求並在 stdout 取得回應）：

```bash
printf '{"name":"read_file","arguments":{"path":"main.go"}}\n' | ./mcp-server -workspace .
```

註：某些部署可能會加入自訂的 HTTP wrapper（例如把 MCP 映射為 JSON‑RPC HTTP），但這類 wrapper 並非本專案預設提供；若您使用自建 wrapper，請確保其請求/回應格式與客戶端相符。

#### 支援的 MCP 方法

MCP 定義了一組方法（例如 `initialize`、`tools/list`、`tools/call`），具體使用方式取決於所採用的傳輸層（SDK/SSE 或 CLI stdin/stdout）。如需在 GUI 中與外部工具互動，建議使用官方 MCP SDK（參見程式碼 `mcp/mcp_server_sdk.go` 的實作範例）。

#### Python 整合範例

#### Python 整合說明

對於 GUI（SDK/SSE）整合，建議使用官方 MCP SDK 客戶端（SSE/HTTP）來呼叫工具並接收事件；請參考 MCP SDK 文件以取得客戶端示例。對於簡單的 CLI 使用，請參考上方的 stdin/stdout 範例或「Integration with AI Assistants」區段中的 subprocess 範例。

result = client.call_tool("read_file", {"path": "main.go"})
print("File content:", result["result"])
```

### 獨立命令行工具（可選）

**注意：mcp 現在要求由宿主前端提供對應的 backend callbacks；若您需要在 CLI 中運行，請使用 `cmd/mcp-server`，該工具會為本地檔案系統提供對應的 fallback callbacks（本地實作），保持行為與 GUI 一致。**

如果需要獨立的命令行工具，可以編譯：

```bash
go build -o mcp-server ./cmd/mcp-server/
```

### 運行獨立 MCP 服務器

```bash
# 使用當前目錄作為工作區
./mcp-server

# 指定工作區目錄
./mcp-server -workspace /path/to/workspace

# 使用配置文件
./mcp-server -config config.json
```

### 與 AI 助手集成

由於 MCP 服務器通過 HTTP 提供服務，可以直接從任何支持 HTTP 的 AI 助手訪問：

```python
# 範例：透過 subprocess 與 standalone mcp-server（stdin 模式）呼叫（單次請求）
import subprocess, json

req = {"name": "execute_go_code", "arguments": {"code": 'fmt.Println("Hello!")'}}
proc = subprocess.Popen(["./mcp-server", "-workspace", "."], stdin=subprocess.PIPE, stdout=subprocess.PIPE, text=True)
stdout, _ = proc.communicate(json.dumps(req) + "\n")
print("Response:", stdout)
```


### 與 Claude Desktop 集成（使用獨立工具）

如果使用獨立命令行工具，在 Claude Desktop 配置文件中添加：

```json
{
  "mcpServers": {
    "idensyra": {
      "command": "/path/to/mcp-server",
      "args": ["-workspace", "/path/to/your/workspace"]
    }
  }
}
```

### 工具調用示例

#### 讀取文件

```json
{
  "name": "read_file",
  "arguments": {
    "path": "main.go"
  }
}
```

#### 執行 Go 代碼

```json
{
  "name": "execute_go_code",
  "arguments": {
    "code": "fmt.Println(\"Hello, World!\")"
  }
}
```

#### 修改 Notebook 儲存格

```json
{
  "name": "modify_cell",
  "arguments": {
    "path": "notebook.igonb",
    "cell_index": 0,
    "new_source": "fmt.Println(\"Updated cell\")",
    "new_language": "go"
  }
}
```

#### 打開工作區

```json
{
  "name": "open_workspace",
  "arguments": {
    "path": "/path/to/workspace"
  }
}
```

## 安全注意事項

1. **權限控制**：默認情況下，所有操作都需要用戶確認（PermissionAsk）。在生產環境建議保持此設置，並謹慎修改權限設定。

2. **工作區隔離**：MCP 服務器應僅能存取指定工作區內的檔案。多數文件操作會驗證相對路徑，但 notebook 和某些工作區功能也應對路徑做清理與檢查以避免路徑穿越（directory traversal）。

3. **路徑驗證**：在讀寫檔案或 notebook 時，實作應拒絕絕對路徑與包含 `..` 的路徑段。`mcp/file_operations.go` 中的 `safeCleanRelativePath` 是一個示例。

4. **匯入檔案謹慎**：`import_file_to_workspace` 的 `source_path` 可能指向工作區外的檔案（通常為絕對路徑），在匯入前務必驗證來源並與使用者確認。

5. **代碼執行**：執行 Go 或 Python 代碼時請小心，確保來源可信，並在可能時使用受限環境執行。

## 開發

### 項目結構

```
mcp/
├── types.go                  - 類型定義和權限配置
├── file_operations.go        - 文件操作實現
├── code_execution.go         - 代碼執行實現
├── notebook_operations.go    - Notebook 操作實現
├── workspace_management.go   - 工作區管理實現
└── server.go                 - MCP 服務器主邏輯

cmd/mcp-server/
└── main.go                   - CLI 入口點
```

### 擴展 MCP 服務器

要添加新工具：

1. 在相應的操作文件中實現新方法
2. 在 `server.go` 的 `HandleRequest` 方法中添加新工具的路由
3. 在 `ListTools` 方法中添加工具信息

## 許可證

本項目採用 MIT 許可證。詳見 [LICENSE](../LICENSE) 文件。
