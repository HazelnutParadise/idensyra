#!/bin/bash
# Example script to demonstrate MCP server usage

echo "Starting Idensyra MCP Server Example"
echo "======================================"
echo

# Build the MCP server
echo "Building MCP server..."
cd "$(dirname "$0")/.."
go build -o /tmp/idensyra-mcp-server ./cmd/mcp-server/
echo "✓ Build complete"
echo

# Create test workspace
echo "Creating test workspace..."
mkdir -p /tmp/example-workspace
cat > /tmp/example-workspace/hello.go << 'EOF'
package main

import "fmt"

func main() {
    fmt.Println("Hello from MCP Server!")
}
EOF
echo "✓ Test workspace created"
echo

# Example 1: List files
echo "Example 1: List files in workspace"
echo '{"name": "list_files", "arguments": {"dir_path": ""}}' | \
  /tmp/idensyra-mcp-server -workspace /tmp/example-workspace 2>/dev/null
echo

# Example 2: Read file
echo "Example 2: Read file content"
echo '{"name": "read_file", "arguments": {"path": "hello.go"}}' | \
  /tmp/idensyra-mcp-server -workspace /tmp/example-workspace 2>/dev/null
echo

# Example 3: Create a new file
echo "Example 3: Create a new file"
echo '{"name": "create_file", "arguments": {"path": "data.txt", "content": "Sample data"}}' | \
  /tmp/idensyra-mcp-server -workspace /tmp/example-workspace 2>/dev/null
echo

# Example 4: Execute Go code
echo "Example 4: Execute Go code"
echo '{"name": "execute_go_code", "arguments": {"code": "fmt.Println(\"Hello from Go!\")"}}' | \
  /tmp/idensyra-mcp-server -workspace /tmp/example-workspace 2>/dev/null
echo

echo "======================================"
echo "Examples complete. Cleaning up..."
rm -rf /tmp/example-workspace
rm -f /tmp/idensyra-mcp-server
echo "Done!"
