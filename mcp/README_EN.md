# Idensyra MCP Server

The Idensyra MCP Server provides a Model Context Protocol (MCP) interface that allows AI agents to interact with Idensyra workspaces.

## Features

### File Operations
- `read_file` - Read file contents (automatically switches to the file)
- `write_file` - Write or overwrite a file (automatically switches to the file)
- `create_file` - Create a new file (automatically switches to the file)
- `delete_file` - Delete a file
- `rename_file` - Rename or move a file
- `list_files` - List files in a directory

### Code Execution
- `execute_go_file` - Execute a Go file (using Yaegi interpreter, automatically switches to the file)
- `execute_go_code` - Execute Go code directly
- `execute_python_file` - Execute a Python file (automatically switches to the file)
- `execute_python_code` - Execute Python code directly

### Notebook Operations (igonb/ipynb)
- `modify_cell` - Modify a specific cell (automatically switches to the notebook)
- `insert_cell` - Insert a cell at a specified position (automatically switches to the notebook)
- `execute_cell` - Execute a specific cell (automatically switches to the notebook)
- `execute_cell_and_after` - Execute a cell and all subsequent cells (automatically switches to the notebook)
- `execute_before_and_cell` - Execute all cells before and including a specific cell (automatically switches to the notebook)
- `execute_all_cells` - Execute all cells (automatically switches to the notebook)
- `convert_ipynb_to_igonb` - Convert ipynb to igonb format

### Automatic File Switching
When an AI agent performs the following operations, the interface automatically switches to the corresponding file:
- Reading a file
- Editing a file
- Creating a file
- Executing a file (Go or Python)
- Modifying or executing notebook cells

This allows users to see in real-time which file the AI agent is working on, providing better visual feedback and context awareness.

### Workspace Management
- `open_workspace` - Open a workspace directory
- `save_temp_workspace` - Save temporary workspace to a specified path
- `save_changes` - Save all unsaved changes
- `get_workspace_info` - Get information about the current workspace
- `create_workspace_directory` - Create a new directory in the workspace
- `import_file_to_workspace` - Import a specific file from the computer into the workspace

## Permission Configuration

The MCP server supports three permission levels:

- `PermissionAlways` - Execute operations without confirmation
- `PermissionAsk` - Require confirmation for each operation (default)
- `PermissionDeny` - Deny all operations

Configurable permission items:
- `FileEdit` - File editing permission
- `FileRename` - File renaming permission
- `FileCreate` - File creation permission
- `FileDelete` - File deletion permission
- `ExecuteGo` - Go code execution permission
- `ExecutePython` - Python code execution permission
- `NotebookModify` - Notebook modification permission
- `NotebookExecute` - Notebook execution permission
- `WorkspaceOpen` - Workspace opening permission
- `WorkspaceSave` - Workspace saving permission
- `WorkspaceModify` - Workspace modification permission

## Usage

### Built-in HTTP Server (Recommended)

The MCP server is integrated into the main Idensyra application and automatically starts an HTTP server on `localhost:3000` when you launch Idensyra.

```bash
# Just start Idensyra - MCP server starts automatically
./idensyra
```

#### Unified HTTP API Endpoint

The MCP server now uses a single endpoint for all requests:

- `GET /mcp` - Get service status and list of available tools
- `POST /mcp` - Execute MCP tool calls

#### Usage Examples

```bash
# Get service status and tool list
curl http://localhost:3000/mcp

# Read a file
curl -X POST http://localhost:3000/mcp \
  -H "Content-Type: application/json" \
  -d '{"name": "read_file", "arguments": {"path": "main.go"}}'

# Import a file to workspace
curl -X POST http://localhost:3000/mcp \
  -H "Content-Type: application/json" \
  -d '{"name": "import_file_to_workspace", "arguments": {"source_path": "/path/to/file.txt", "target_dir": ""}}'
```

### Standalone CLI Tool (Optional)

If you need a standalone command-line tool, you can build it:

```bash
go build -o mcp-server ./cmd/mcp-server/
```

### Run Standalone MCP Server

```bash
# Use current directory as workspace
./mcp-server

# Specify workspace directory
./mcp-server -workspace /path/to/workspace

# Use configuration file
./mcp-server -config config.json
```

### Integration with AI Assistants

Since the MCP server is available via HTTP, you can access it from any AI assistant that supports HTTP:

```python
# Python example
import requests

response = requests.post('http://localhost:3000/mcp/call', json={
    "name": "execute_go_code",
    "arguments": {"code": "fmt.Println(\"Hello!\")"}
})
print(response.json())
```

### Integration with Claude Desktop (Using Standalone Tool)

If using the standalone CLI tool, add to your Claude Desktop configuration file:

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

### Tool Call Examples

#### Read a File

```json
{
  "name": "read_file",
  "arguments": {
    "path": "main.go"
  }
}
```

#### Execute Go Code

```json
{
  "name": "execute_go_code",
  "arguments": {
    "code": "fmt.Println(\"Hello, World!\")"
  }
}
```

#### Modify a Notebook Cell

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

#### Open a Workspace

```json
{
  "name": "open_workspace",
  "arguments": {
    "path": "/path/to/workspace"
  }
}
```

## Security Considerations

1. **Permission Control**: By default, all operations require user confirmation. It's recommended to keep this setting in production environments.

2. **Workspace Isolation**: The MCP server can only access files within the specified workspace and cannot access files outside the workspace.

3. **Code Execution**: Be cautious when executing Go and Python code. Ensure the code source is trustworthy.

## Development

### Project Structure

```
mcp/
├── types.go                  - Type definitions and permission configuration
├── file_operations.go        - File operations implementation
├── code_execution.go         - Code execution implementation
├── notebook_operations.go    - Notebook operations implementation
├── workspace_management.go   - Workspace management implementation
└── server.go                 - MCP server main logic

cmd/mcp-server/
└── main.go                   - CLI entry point
```

### Extending the MCP Server

To add a new tool:

1. Implement the new method in the appropriate operations file
2. Add routing for the new tool in the `HandleRequest` method in `server.go`
3. Add tool information in the `ListTools` method

## License

This project is licensed under the MIT License. See the [LICENSE](../LICENSE) file for details.
