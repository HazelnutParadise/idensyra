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

### Transports and how to connect (Recommended)

Idensyra supports two primary MCP transports; choose according to your environment:

- **Integrated GUI (recommended)**: When running Idensyra in GUI mode the application uses the official Model Context Protocol SDK and exposes an SSE‑based HTTP handler. Clients should use the official MCP SDK clients (for example the go-sdk) or an SSE-capable client to connect. In Idensyra the MCP server listens on port **14320** by default (e.g. `http://localhost:14320`), though the host application controls the final address and port.

- **Standalone CLI (optional)**: `cmd/mcp-server` is a lightweight standalone implementation intended for local testing / CLI use. It operates directly on the local filesystem and communicates over **stdin/stdout** using JSON ToolRequest/ToolResponse objects. Example (send a single request and receive the response on stdout):

```bash
printf '{"name":"read_file","arguments":{"path":"main.go"}}\n' | ./mcp-server -workspace .
```

Note: Some deployments may implement an HTTP wrapper that accepts JSON‑RPC or other HTTP formats; such wrappers are not provided by this project by default. If you are using a custom wrapper, ensure its request/response formats match your client implementation.

#### Supported MCP Methods

MCP defines a set of standard methods (for example `initialize`, `tools/list`, `tools/call`); how you call them depends on the transport (SDK/SSE vs CLI stdin/stdout). For GUI integrations, prefer using the official MCP SDK (see `mcp/mcp_server_sdk.go` for an example of registering tools and the SSE handler in the host app).

#### Python Integration Guidance

For GUI integrations, prefer using the official MCP SDK clients (SSE/HTTP) to call tools and receive events; see the MCP SDK documentation for client examples. For simple CLI usage, see the stdin/stdout example shown above or the subprocess example in the "Integration with AI Assistants" section.


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

The exact integration method depends on the transport you use:

- For **standalone CLI** usage, you can invoke the `mcp-server` in stdin mode (one-shot) from a script. Example using Python and subprocess:

```python
# Example: invoke the standalone mcp-server via stdin (one-shot request)
import subprocess, json

req = {"name": "execute_go_code", "arguments": {"code": 'fmt.Println("Hello!")'}}
proc = subprocess.Popen(["./mcp-server", "-workspace", "."], stdin=subprocess.PIPE, stdout=subprocess.PIPE, text=True)
stdout, _ = proc.communicate(json.dumps(req) + "\n")
print("Response:", stdout)
```

- For **GUI integrations** (SSE/HTTP), prefer using the official MCP SDK clients (SSE) to call methods and receive events; see the MCP SDK documentation for client examples.

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

1. **Permission Control**: By default, all operations require user confirmation (PermissionAsk). It's recommended to keep this setting in production and carefully review permission changes.

2. **Workspace Isolation**: The MCP server should only access files within the specified workspace. Many file operations validate relative paths, but notebook and workspace functions should also sanitize paths to avoid directory traversal.

3. **Path sanitization**: Implementations should reject absolute paths or path segments containing `..` for relative workspace operations. The standalone CLI provides `safeCleanRelativePath` in `mcp/file_operations.go` as an example.

4. **Import file caution**: `import_file_to_workspace` accepts a source path on the local machine (often absolute). Treat imports carefully, validate sources, and confirm operations with the user.

5. **Code Execution**: Be cautious when executing Go and Python code. Ensure the code source is trustworthy and run in a restricted environment where possible.

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
