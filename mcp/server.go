package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
)

// Server represents the MCP server
type Server struct {
	config              *Config
	fileOps             *FileOperations
	codeExec            *CodeExecution
	notebookOps         *NotebookOperations
	workspaceManagement *WorkspaceManagement
}

// NewServer creates a new MCP server instance
func NewServer(
	config *Config,
	workspaceRoot string,
	confirmFunc func(operation, details string) bool,
	executeGoFunc func(code string, colorBG string) string,
	executePyFunc func(filePath string) (string, error),
	executeCellFunc func(language, code string) (string, error),
	openWorkspaceFunc func(path string) error,
	saveWorkspaceFunc func(path string) error,
	saveChangesFunc func() error,
	setActiveFileFunc func(path string) error,
	importFileFunc func(sourcePath, targetDir string) error,
	// Optional file backend callbacks (if provided, they will be used instead of direct FS access)
	readFileFunc func(path string) (string, error),
	writeFileFunc func(path string, content string) error,
	createFileFunc func(path string, content string) error,
	deleteFileFunc func(path string) error,
	renameFileFunc func(oldPath, newPath string) error,
	listFilesFunc func(dirPath string) (string, error),
) *Server {
	return &Server{
		config: config,
		fileOps: NewFileOperations(config, workspaceRoot, confirmFunc, setActiveFileFunc,
			readFileFunc, writeFileFunc, createFileFunc, deleteFileFunc, renameFileFunc, listFilesFunc),
		codeExec:            NewCodeExecution(config, workspaceRoot, confirmFunc, executeGoFunc, executePyFunc, setActiveFileFunc),
		notebookOps:         NewNotebookOperations(config, workspaceRoot, confirmFunc, executeCellFunc, setActiveFileFunc),
		workspaceManagement: NewWorkspaceManagement(config, confirmFunc, openWorkspaceFunc, saveWorkspaceFunc, saveChangesFunc, importFileFunc),
	}
}

// HandleRequest handles an incoming MCP tool request
func (s *Server) HandleRequest(ctx context.Context, req *ToolRequest) (*ToolResponse, error) {
	switch req.Name {
	// File operations
	case "read_file":
		path, _ := req.Arguments["path"].(string)
		return s.fileOps.ReadFile(ctx, path)
	case "write_file":
		path, _ := req.Arguments["path"].(string)
		content, _ := req.Arguments["content"].(string)
		return s.fileOps.WriteFile(ctx, path, content)
	case "create_file":
		path, _ := req.Arguments["path"].(string)
		content, _ := req.Arguments["content"].(string)
		return s.fileOps.CreateFile(ctx, path, content)
	case "delete_file":
		path, _ := req.Arguments["path"].(string)
		return s.fileOps.DeleteFile(ctx, path)
	case "rename_file":
		oldPath, _ := req.Arguments["old_path"].(string)
		newPath, _ := req.Arguments["new_path"].(string)
		return s.fileOps.RenameFile(ctx, oldPath, newPath)
	case "list_files":
		dirPath, _ := req.Arguments["dir_path"].(string)
		return s.fileOps.ListFiles(ctx, dirPath)

	// Code execution
	case "execute_go_file":
		path, _ := req.Arguments["path"].(string)
		return s.codeExec.ExecuteGoFile(ctx, path)
	case "execute_go_code":
		code, _ := req.Arguments["code"].(string)
		return s.codeExec.ExecuteGoCode(ctx, code)
	case "execute_python_file":
		path, _ := req.Arguments["path"].(string)
		return s.codeExec.ExecutePythonFile(ctx, path)
	case "execute_python_code":
		code, _ := req.Arguments["code"].(string)
		return s.codeExec.ExecutePythonCode(ctx, code)

	// Notebook operations
	case "modify_cell":
		path, _ := req.Arguments["path"].(string)
		cellIndex, _ := req.Arguments["cell_index"].(float64)
		newSource, _ := req.Arguments["new_source"].(string)
		newLanguage, _ := req.Arguments["new_language"].(string)
		return s.notebookOps.ModifyCell(ctx, path, int(cellIndex), newSource, newLanguage)
	case "insert_cell":
		path, _ := req.Arguments["path"].(string)
		position, _ := req.Arguments["position"].(float64)
		language, _ := req.Arguments["language"].(string)
		source, _ := req.Arguments["source"].(string)
		return s.notebookOps.InsertCell(ctx, path, int(position), language, source)
	case "execute_cell":
		path, _ := req.Arguments["path"].(string)
		cellIndex, _ := req.Arguments["cell_index"].(float64)
		return s.notebookOps.ExecuteCell(ctx, path, int(cellIndex))
	case "execute_cell_and_after":
		path, _ := req.Arguments["path"].(string)
		startIndex, _ := req.Arguments["start_index"].(float64)
		return s.notebookOps.ExecuteCellAndAfter(ctx, path, int(startIndex))
	case "execute_before_and_cell":
		path, _ := req.Arguments["path"].(string)
		endIndex, _ := req.Arguments["end_index"].(float64)
		return s.notebookOps.ExecuteBeforeAndCell(ctx, path, int(endIndex))
	case "execute_all_cells":
		path, _ := req.Arguments["path"].(string)
		return s.notebookOps.ExecuteAllCells(ctx, path)
	case "convert_ipynb_to_igonb":
		ipynbPath, _ := req.Arguments["ipynb_path"].(string)
		igonbPath, _ := req.Arguments["igonb_path"].(string)
		return s.notebookOps.ConvertIPyNBToIgonb(ctx, ipynbPath, igonbPath)

	// Workspace management
	case "open_workspace":
		path, _ := req.Arguments["path"].(string)
		return s.workspaceManagement.OpenWorkspace(ctx, path)
	case "save_temp_workspace":
		targetPath, _ := req.Arguments["target_path"].(string)
		return s.workspaceManagement.SaveTempWorkspace(ctx, targetPath)
	case "save_changes":
		return s.workspaceManagement.SaveChanges(ctx)
	case "get_workspace_info":
		return s.workspaceManagement.GetWorkspaceInfo(ctx)
	case "create_workspace_directory":
		relativePath, _ := req.Arguments["relative_path"].(string)
		return s.workspaceManagement.CreateWorkspaceDirectory(ctx, relativePath)
	case "import_file_to_workspace":
		sourcePath, _ := req.Arguments["source_path"].(string)
		targetDir, _ := req.Arguments["target_dir"].(string)
		return s.workspaceManagement.ImportFileToWorkspace(ctx, sourcePath, targetDir)

	default:
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Unknown tool: %s", req.Name)}},
			IsError: true,
		}, fmt.Errorf("unknown tool: %s", req.Name)
	}
}

// ListTools returns a list of available tools
func (s *Server) ListTools() []ToolInfo {
	tools := []ToolInfo{
		// File operations
		{
			Name:        "read_file",
			Description: "Operates on Idensyra workspace - Read the content of a file",
			Target:      "idensyra",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{"type": "string", "description": "Path to the file"},
				},
				"required": []string{"path"},
			},
		},
		{
			Name:        "write_file",
			Description: "Operates on Idensyra workspace - Write content to a file (creates or overwrites)",
			Target:      "idensyra",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path":    map[string]interface{}{"type": "string", "description": "Path to the file"},
					"content": map[string]interface{}{"type": "string", "description": "Content to write"},
				},
				"required": []string{"path", "content"},
			},
		},
		{
			Name:        "create_file",
			Description: "Operates on Idensyra workspace - Create a new file (fails if file already exists)",
			Target:      "idensyra",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path":    map[string]interface{}{"type": "string", "description": "Path to the file"},
					"content": map[string]interface{}{"type": "string", "description": "Initial content"},
				},
				"required": []string{"path", "content"},
			},
		},
		{
			Name:        "delete_file",
			Description: "Operates on Idensyra workspace - Delete a file",
			Target:      "idensyra",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{"type": "string", "description": "Path to the file"},
				},
				"required": []string{"path"},
			},
		},
		{
			Name:        "rename_file",
			Description: "Operates on Idensyra workspace - Rename or move a file",
			Target:      "idensyra",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"old_path": map[string]interface{}{"type": "string", "description": "Current path"},
					"new_path": map[string]interface{}{"type": "string", "description": "New path"},
				},
				"required": []string{"old_path", "new_path"},
			},
		},
		{
			Name:        "list_files",
			Description: "Operates on Idensyra workspace - List files in a directory",
			Target:      "idensyra",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"dir_path": map[string]interface{}{"type": "string", "description": "Directory path (empty for root)"},
				},
				"required": []string{},
			},
		},

		// Code execution
		{
			Name:        "execute_go_file",
			Description: "Operates on Idensyra workspace - Execute a Go file using the Yaegi interpreter",
			Target:      "idensyra",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{"type": "string", "description": "Path to the .go file"},
				},
				"required": []string{"path"},
			},
		},
		{
			Name:        "execute_go_code",
			Description: "Operates on Idensyra workspace - Execute Go code directly",
			Target:      "idensyra",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"code": map[string]interface{}{"type": "string", "description": "Go code to execute"},
				},
				"required": []string{"code"},
			},
		},
		{
			Name:        "execute_python_file",
			Description: "Operates on Idensyra workspace - Execute a Python file",
			Target:      "idensyra",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{"type": "string", "description": "Path to the .py file"},
				},
				"required": []string{"path"},
			},
		},
		{
			Name:        "execute_python_code",
			Description: "Operates on Idensyra workspace - Execute Python code directly",
			Target:      "idensyra",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"code": map[string]interface{}{"type": "string", "description": "Python code to execute"},
				},
				"required": []string{"code"},
			},
		},

		// Notebook operations
		{
			Name:        "modify_cell",
			Description: "Modify a specific cell in a notebook",
			Target:      "idensyra",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path":         map[string]interface{}{"type": "string", "description": "Path to the notebook"},
					"cell_index":   map[string]interface{}{"type": "number", "description": "Index of the cell to modify"},
					"new_source":   map[string]interface{}{"type": "string", "description": "New source code"},
					"new_language": map[string]interface{}{"type": "string", "description": "New language (optional)"},
				},
				"required": []string{"path", "cell_index", "new_source"},
			},
		},
		{
			Name:        "insert_cell",
			Description: "Insert a new cell at a specific position in a notebook",
			Target:      "idensyra",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path":     map[string]interface{}{"type": "string", "description": "Path to the notebook"},
					"position": map[string]interface{}{"type": "number", "description": "Position to insert at (0-based)"},
					"language": map[string]interface{}{"type": "string", "description": "Cell language (go, python, markdown)"},
					"source":   map[string]interface{}{"type": "string", "description": "Cell source code"},
				},
				"required": []string{"path", "position", "language", "source"},
			},
		},
		{
			Name:        "execute_cell",
			Description: "Execute a specific cell in a notebook",
			Target:      "idensyra",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path":       map[string]interface{}{"type": "string", "description": "Path to the notebook"},
					"cell_index": map[string]interface{}{"type": "number", "description": "Index of the cell to execute"},
				},
				"required": []string{"path", "cell_index"},
			},
		},
		{
			Name:        "execute_cell_and_after",
			Description: "Execute a cell and all subsequent cells",
			Target:      "idensyra",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path":        map[string]interface{}{"type": "string", "description": "Path to the notebook"},
					"start_index": map[string]interface{}{"type": "number", "description": "Index of the first cell to execute"},
				},
				"required": []string{"path", "start_index"},
			},
		},
		{
			Name:        "execute_before_and_cell",
			Description: "Execute all cells before and including a specific cell",
			Target:      "idensyra",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path":      map[string]interface{}{"type": "string", "description": "Path to the notebook"},
					"end_index": map[string]interface{}{"type": "number", "description": "Index of the last cell to execute"},
				},
				"required": []string{"path", "end_index"},
			},
		},
		{
			Name:        "execute_all_cells",
			Description: "Execute all cells in a notebook",
			Target:      "idensyra",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{"type": "string", "description": "Path to the notebook"},
				},
				"required": []string{"path"},
			},
		},
		{
			Name:        "convert_ipynb_to_igonb",
			Description: "Convert an ipynb file to igonb format",
			Target:      "idensyra",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"ipynb_path": map[string]interface{}{"type": "string", "description": "Path to the .ipynb file"},
					"igonb_path": map[string]interface{}{"type": "string", "description": "Path for the output .igonb file"},
				},
				"required": []string{"ipynb_path", "igonb_path"},
			},
		},

		// Workspace management
		{
			Name:        "open_workspace",
			Description: "Operates on Idensyra workspace - Open a workspace directory",
			Target:      "idensyra",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{"type": "string", "description": "Path to the workspace directory"},
				},
				"required": []string{"path"},
			},
		},
		{
			Name:        "save_temp_workspace",
			Description: "Operates on Idensyra workspace - Save the temporary workspace to a specified path",
			Target:      "idensyra",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"target_path": map[string]interface{}{"type": "string", "description": "Target directory path"},
				},
				"required": []string{"target_path"},
			},
		},
		{
			Name:        "save_changes",
			Description: "Operates on Idensyra workspace - Save all unsaved changes in the current workspace",
			Target:      "idensyra",
			InputSchema: map[string]interface{}{
				"type":                 "object",
				"properties":           map[string]interface{}{},
				"required":             []string{},
				"additionalProperties": true,
			},
		},
		{
			Name:        "get_workspace_info",
			Description: "Operates on Idensyra workspace - Get information about the current workspace",
			Target:      "idensyra",
			InputSchema: map[string]interface{}{
				"type":                 "object",
				"properties":           map[string]interface{}{},
				"required":             []string{},
				"additionalProperties": true,
			},
		},
		{
			Name:        "create_workspace_directory",
			Description: "Operates on Idensyra workspace - Create a new directory in the workspace",
			Target:      "idensyra",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"relative_path": map[string]interface{}{"type": "string", "description": "Relative path for the new directory"},
				},
				"required": []string{"relative_path"},
			},
		},
		{
			Name:        "import_file_to_workspace",
			Description: "Operates on Idensyra workspace - Import a file from the computer into the workspace",
			Target:      "idensyra",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"source_path": map[string]interface{}{"type": "string", "description": "Absolute path to the source file on the computer"},
					"target_dir":  map[string]interface{}{"type": "string", "description": "Target directory in workspace (relative path, empty string for root)"},
				},
				"required": []string{"source_path"},
			},
		},
	}

	// Ensure every tool has a complete inputSchema
	for i := range tools {
		if tools[i].InputSchema == nil {
			tools[i].InputSchema = map[string]interface{}{"type": "object", "properties": map[string]interface{}{}, "required": []string{}}
			continue
		}
		if _, ok := tools[i].InputSchema["type"]; !ok {
			tools[i].InputSchema["type"] = "object"
		}
		if _, ok := tools[i].InputSchema["properties"]; !ok {
			tools[i].InputSchema["properties"] = map[string]interface{}{}
		}
		if _, ok := tools[i].InputSchema["required"]; !ok {
			tools[i].InputSchema["required"] = []string{}
		}

		// If properties is empty, ensure additionalProperties is present so validators accept the schema
		if props, ok := tools[i].InputSchema["properties"].(map[string]interface{}); ok {
			if len(props) == 0 {
				if _, ok := tools[i].InputSchema["additionalProperties"]; !ok {
					tools[i].InputSchema["additionalProperties"] = true
				}
			}
		}
	}

	return tools
}

// ToolInfo represents information about a tool
type ToolInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Target      string                 `json:"target,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// Serve starts the MCP server and handles stdin/stdout communication
func (s *Server) Serve(ctx context.Context, input io.Reader, output io.Writer) error {
	decoder := json.NewDecoder(input)
	encoder := json.NewEncoder(output)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var req ToolRequest
		if err := decoder.Decode(&req); err != nil {
			if err == io.EOF {
				return nil
			}
			log.Printf("Error decoding request: %v", err)
			continue
		}

		resp, err := s.HandleRequest(ctx, &req)
		if err != nil {
			log.Printf("Error handling request: %v", err)
		}

		if err := encoder.Encode(resp); err != nil {
			log.Printf("Error encoding response: %v", err)
			return err
		}
	}
}
