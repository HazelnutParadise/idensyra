package mcp

// PermissionLevel defines the permission level for MCP operations
type PermissionLevel int

const (
	// PermissionAlways allows operations without confirmation
	PermissionAlways PermissionLevel = iota
	// PermissionAsk requires confirmation for each operation
	PermissionAsk
	// PermissionDeny denies all operations
	PermissionDeny
)

// Config holds the MCP server configuration
type Config struct {
	// File operation permissions
	FileEdit   PermissionLevel
	FileRename PermissionLevel
	FileCreate PermissionLevel
	FileDelete PermissionLevel

	// Execution permissions
	ExecuteGo     PermissionLevel
	ExecutePython PermissionLevel

	// Notebook permissions
	NotebookModify  PermissionLevel
	NotebookExecute PermissionLevel

	// Workspace permissions
	WorkspaceOpen   PermissionLevel
	WorkspaceSave   PermissionLevel
	WorkspaceModify PermissionLevel
}

// DefaultConfig returns a default configuration with all permissions set to ask
func DefaultConfig() *Config {
	return &Config{
		FileEdit:        PermissionAsk,
		FileRename:      PermissionAsk,
		FileCreate:      PermissionAsk,
		FileDelete:      PermissionAsk,
		ExecuteGo:       PermissionAsk,
		ExecutePython:   PermissionAsk,
		NotebookModify:  PermissionAsk,
		NotebookExecute: PermissionAsk,
		WorkspaceOpen:   PermissionAsk,
		WorkspaceSave:   PermissionAsk,
		WorkspaceModify: PermissionAsk,
	}
}

// ToolRequest represents a generic MCP tool request
type ToolRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolResponse represents a generic MCP tool response
type ToolResponse struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

// ContentBlock represents a content block in the response
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// NotebookCell represents a cell in a notebook
type NotebookCell struct {
	Language string `json:"language"`
	Source   string `json:"source"`
	Output   string `json:"output,omitempty"`
}
