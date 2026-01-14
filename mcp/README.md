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

### 內建 HTTP 服務器（推薦）

MCP 服務器已整合到 Idensyra 主程式中，啟動 Idensyra 時會自動在 `localhost:3000` 啟動符合 MCP 標準的 HTTP 服務器。

```bash
# 直接啟動 Idensyra，MCP 服務器會自動啟動
./idensyra
```

#### MCP 協議端點

服務器完全符合 Model Context Protocol (MCP) 標準，使用 JSON-RPC 2.0 格式通信：

- `POST /` - MCP 協議端點（支持所有 MCP 方法）

#### 支持的 MCP 方法

1. **initialize** - 初始化 MCP 連接
2. **tools/list** - 列出所有可用工具
3. **tools/call** - 執行特定工具

#### 使用範例

```bash
# 初始化連接
curl -X POST http://localhost:3000/ \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2024-11-05",
      "capabilities": {},
      "clientInfo": {"name": "test-client", "version": "1.0.0"}
    }
  }'

# 列出所有工具
curl -X POST http://localhost:3000/ \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/list"
  }'

# 執行工具 - 讀取文件
curl -X POST http://localhost:3000/ \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "read_file",
      "arguments": {"path": "main.go"}
    }
  }'

# 匯入檔案到工作區
curl -X POST http://localhost:3000/ \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 4,
    "method": "tools/call",
    "params": {
      "name": "import_file_to_workspace",
      "arguments": {
        "source_path": "/path/to/file.txt",
        "target_dir": ""
      }
    }
  }'
```

#### Python 整合範例

```python
import requests
import json

class IdensyraMCPClient:
    def __init__(self, url="http://localhost:3000/"):
        self.url = url
        self.request_id = 0
    
    def _call(self, method, params=None):
        self.request_id += 1
        payload = {
            "jsonrpc": "2.0",
            "id": self.request_id,
            "method": method
        }
        if params:
            payload["params"] = params
        
        response = requests.post(self.url, json=payload)
        return response.json()
    
    def initialize(self):
        return self._call("initialize", {
            "protocolVersion": "2024-11-05",
            "capabilities": {},
            "clientInfo": {"name": "python-client", "version": "1.0.0"}
        })
    
    def list_tools(self):
        return self._call("tools/list")
    
    def call_tool(self, name, arguments=None):
        return self._call("tools/call", {
            "name": name,
            "arguments": arguments or {}
        })

# 使用範例
client = IdensyraMCPClient()

# 初始化
init_result = client.initialize()
print("Server:", init_result["result"]["serverInfo"])

# 列出工具
tools = client.list_tools()
print(f"Available tools: {len(tools['result']['tools'])}")

# 讀取文件
result = client.call_tool("read_file", {"path": "main.go"})
print("File content:", result["result"])
```

### 獨立命令行工具（可選）

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
# Python 示例
import requests

response = requests.post('http://localhost:3000/mcp/call', json={
    "name": "execute_go_code",
    "arguments": {"code": "fmt.Println(\"Hello!\")"}
})
print(response.json())
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

1. **權限控制**：默認情況下，所有操作都需要用戶確認。在生產環境中建議保持此設置。

2. **工作區隔離**：MCP 服務器只能訪問指定工作區內的文件，無法訪問工作區外的文件系統。

3. **代碼執行**：執行 Go 和 Python 代碼時需要謹慎，確保代碼來源可信。

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
