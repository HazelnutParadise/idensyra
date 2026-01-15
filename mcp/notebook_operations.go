package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// NotebookOperations provides notebook manipulation tools for MCP
type NotebookOperations struct {
	config            *Config
	workspaceRoot     string
	confirmFunc       func(operation, details string) bool
	executeCellFunc   func(language, code string) (string, error)
	setActiveFileFunc func(path string) error
}

// NewNotebookOperations creates a new NotebookOperations instance
func NewNotebookOperations(
	config *Config,
	workspaceRoot string,
	confirmFunc func(operation, details string) bool,
	executeCellFunc func(language, code string) (string, error),
	setActiveFileFunc func(path string) error,
) *NotebookOperations {
	return &NotebookOperations{
		config:            config,
		workspaceRoot:     workspaceRoot,
		confirmFunc:       confirmFunc,
		executeCellFunc:   executeCellFunc,
		setActiveFileFunc: setActiveFileFunc,
	}
}

// Notebook represents a notebook structure
type Notebook struct {
	Version int            `json:"version"`
	Cells   []NotebookCell `json:"cells"`
}

// ReadNotebook reads and parses a notebook file
func (no *NotebookOperations) ReadNotebook(ctx context.Context, path string) (*Notebook, error) {
	fullPath := filepath.Join(no.workspaceRoot, path)

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	var notebook Notebook
	if err := json.Unmarshal(content, &notebook); err != nil {
		return nil, fmt.Errorf("error parsing notebook: %v", err)
	}

	return &notebook, nil
}

// WriteNotebook writes a notebook to file
func (no *NotebookOperations) WriteNotebook(ctx context.Context, path string, notebook *Notebook) error {
	if no.config.NotebookModify == PermissionDeny {
		return fmt.Errorf("permission denied")
	}

	if no.config.NotebookModify == PermissionAsk && no.confirmFunc != nil {
		if !no.confirmFunc("Notebook Modify", fmt.Sprintf("Save notebook: %s", path)) {
			return fmt.Errorf("cancelled by user")
		}
	}

	fullPath := filepath.Join(no.workspaceRoot, path)

	content, err := json.MarshalIndent(notebook, "", "  ")
	if err != nil {
		return fmt.Errorf("error encoding notebook: %v", err)
	}

	if err := os.WriteFile(fullPath, content, 0644); err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	return nil
}

// ModifyCell modifies a specific cell in a notebook
func (no *NotebookOperations) ModifyCell(ctx context.Context, path string, cellIndex int, newSource string, newLanguage string) (*ToolResponse, error) {
	if no.config.NotebookModify == PermissionDeny {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: "Notebook modify permission denied"}},
			IsError: true,
		}, fmt.Errorf("permission denied")
	}

	if no.config.NotebookModify == PermissionAsk && no.confirmFunc != nil {
		if !no.confirmFunc("Notebook Modify", fmt.Sprintf("Modify cell %d in: %s", cellIndex, path)) {
			return &ToolResponse{
				Content: []ContentBlock{{Type: "text", Text: "Notebook modify cancelled by user"}},
				IsError: true,
			}, fmt.Errorf("cancelled by user")
		}
	}

	notebook, err := no.ReadNotebook(ctx, path)
	if err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error reading notebook: %v", err)}},
			IsError: true,
		}, err
	}

	if cellIndex < 0 || cellIndex >= len(notebook.Cells) {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Invalid cell index: %d", cellIndex)}},
			IsError: true,
		}, fmt.Errorf("invalid cell index")
	}

	notebook.Cells[cellIndex].Source = newSource
	if newLanguage != "" {
		notebook.Cells[cellIndex].Language = newLanguage
	}

	if err := no.WriteNotebook(ctx, path, notebook); err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error writing notebook: %v", err)}},
			IsError: true,
		}, err
	}

	// Switch to the notebook being modified
	if no.setActiveFileFunc != nil {
		_ = no.setActiveFileFunc(path)
	}

	return &ToolResponse{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Cell %d modified successfully", cellIndex)}},
	}, nil
}

// InsertCell inserts a new cell at the specified position
func (no *NotebookOperations) InsertCell(ctx context.Context, path string, position int, language string, source string) (*ToolResponse, error) {
	if no.config.NotebookModify == PermissionDeny {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: "Notebook modify permission denied"}},
			IsError: true,
		}, fmt.Errorf("permission denied")
	}

	if no.config.NotebookModify == PermissionAsk && no.confirmFunc != nil {
		if !no.confirmFunc("Notebook Modify", fmt.Sprintf("Insert cell at position %d in: %s", position, path)) {
			return &ToolResponse{
				Content: []ContentBlock{{Type: "text", Text: "Notebook modify cancelled by user"}},
				IsError: true,
			}, fmt.Errorf("cancelled by user")
		}
	}

	notebook, err := no.ReadNotebook(ctx, path)
	if err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error reading notebook: %v", err)}},
			IsError: true,
		}, err
	}

	if position < 0 || position > len(notebook.Cells) {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Invalid position: %d", position)}},
			IsError: true,
		}, fmt.Errorf("invalid position")
	}

	newCell := NotebookCell{
		Language: language,
		Source:   source,
	}

	// Insert cell at position
	notebook.Cells = append(notebook.Cells[:position], append([]NotebookCell{newCell}, notebook.Cells[position:]...)...)

	if err := no.WriteNotebook(ctx, path, notebook); err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error writing notebook: %v", err)}},
			IsError: true,
		}, err
	}

	// Switch to the notebook being modified
	if no.setActiveFileFunc != nil {
		_ = no.setActiveFileFunc(path)
	}

	return &ToolResponse{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Cell inserted successfully at position %d", position)}},
	}, nil
}

// ExecuteCell executes a specific cell
func (no *NotebookOperations) ExecuteCell(ctx context.Context, path string, cellIndex int) (*ToolResponse, error) {
	if no.config.NotebookExecute == PermissionDeny {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: "Notebook execute permission denied"}},
			IsError: true,
		}, fmt.Errorf("permission denied")
	}

	if no.config.NotebookExecute == PermissionAsk && no.confirmFunc != nil {
		if !no.confirmFunc("Notebook Execute", fmt.Sprintf("Execute cell %d in: %s", cellIndex, path)) {
			return &ToolResponse{
				Content: []ContentBlock{{Type: "text", Text: "Notebook execute cancelled by user"}},
				IsError: true,
			}, fmt.Errorf("cancelled by user")
		}
	}

	notebook, err := no.ReadNotebook(ctx, path)
	if err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error reading notebook: %v", err)}},
			IsError: true,
		}, err
	}

	if cellIndex < 0 || cellIndex >= len(notebook.Cells) {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Invalid cell index: %d", cellIndex)}},
			IsError: true,
		}, fmt.Errorf("invalid cell index")
	}

	cell := notebook.Cells[cellIndex]

	if no.executeCellFunc == nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: "Cell execution function not available"}},
			IsError: true,
		}, fmt.Errorf("execution function not available")
	}

	output, err := no.executeCellFunc(cell.Language, cell.Source)
	if err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error executing cell: %v\n%s", err, output)}},
			IsError: true,
		}, err
	}

	// Switch to the notebook being executed
	if no.setActiveFileFunc != nil {
		_ = no.setActiveFileFunc(path)
	}

	return &ToolResponse{
		Content: []ContentBlock{{Type: "text", Text: output}},
	}, nil
}

// ExecuteCellAndAfter executes a cell and all subsequent cells
func (no *NotebookOperations) ExecuteCellAndAfter(ctx context.Context, path string, startIndex int) (*ToolResponse, error) {
	if no.config.NotebookExecute == PermissionDeny {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: "Notebook execute permission denied"}},
			IsError: true,
		}, fmt.Errorf("permission denied")
	}

	if no.config.NotebookExecute == PermissionAsk && no.confirmFunc != nil {
		if !no.confirmFunc("Notebook Execute", fmt.Sprintf("Execute cell %d and after in: %s", startIndex, path)) {
			return &ToolResponse{
				Content: []ContentBlock{{Type: "text", Text: "Notebook execute cancelled by user"}},
				IsError: true,
			}, fmt.Errorf("cancelled by user")
		}
	}

	notebook, err := no.ReadNotebook(ctx, path)
	if err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error reading notebook: %v", err)}},
			IsError: true,
		}, err
	}

	if startIndex < 0 || startIndex >= len(notebook.Cells) {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Invalid cell index: %d", startIndex)}},
			IsError: true,
		}, fmt.Errorf("invalid cell index")
	}

	// Switch to the notebook being executed (before starting execution)
	if no.setActiveFileFunc != nil {
		_ = no.setActiveFileFunc(path)
	}

	var outputs []string
	for i := startIndex; i < len(notebook.Cells); i++ {
		cell := notebook.Cells[i]
		output, err := no.executeCellFunc(cell.Language, cell.Source)
		if err != nil {
			outputs = append(outputs, fmt.Sprintf("Cell %d error: %v\n%s", i, err, output))
		} else {
			outputs = append(outputs, fmt.Sprintf("Cell %d output:\n%s", i, output))
		}
	}

	return &ToolResponse{
		Content: []ContentBlock{{Type: "text", Text: strings.Join(outputs, "\n\n")}},
	}, nil
}

// ExecuteBeforeAndCell executes all cells before and including the specified cell
func (no *NotebookOperations) ExecuteBeforeAndCell(ctx context.Context, path string, endIndex int) (*ToolResponse, error) {
	if no.config.NotebookExecute == PermissionDeny {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: "Notebook execute permission denied"}},
			IsError: true,
		}, fmt.Errorf("permission denied")
	}

	if no.config.NotebookExecute == PermissionAsk && no.confirmFunc != nil {
		if !no.confirmFunc("Notebook Execute", fmt.Sprintf("Execute cells up to %d in: %s", endIndex, path)) {
			return &ToolResponse{
				Content: []ContentBlock{{Type: "text", Text: "Notebook execute cancelled by user"}},
				IsError: true,
			}, fmt.Errorf("cancelled by user")
		}
	}

	notebook, err := no.ReadNotebook(ctx, path)
	if err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error reading notebook: %v", err)}},
			IsError: true,
		}, err
	}

	if endIndex < 0 || endIndex >= len(notebook.Cells) {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Invalid cell index: %d", endIndex)}},
			IsError: true,
		}, fmt.Errorf("invalid cell index")
	}

	// Switch to the notebook being executed (before starting execution)
	if no.setActiveFileFunc != nil {
		_ = no.setActiveFileFunc(path)
	}

	var outputs []string
	for i := 0; i <= endIndex; i++ {
		cell := notebook.Cells[i]
		output, err := no.executeCellFunc(cell.Language, cell.Source)
		if err != nil {
			outputs = append(outputs, fmt.Sprintf("Cell %d error: %v\n%s", i, err, output))
		} else {
			outputs = append(outputs, fmt.Sprintf("Cell %d output:\n%s", i, output))
		}
	}

	return &ToolResponse{
		Content: []ContentBlock{{Type: "text", Text: strings.Join(outputs, "\n\n")}},
	}, nil
}

// ExecuteAllCells executes all cells in a notebook
func (no *NotebookOperations) ExecuteAllCells(ctx context.Context, path string) (*ToolResponse, error) {
	if no.config.NotebookExecute == PermissionDeny {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: "Notebook execute permission denied"}},
			IsError: true,
		}, fmt.Errorf("permission denied")
	}

	if no.config.NotebookExecute == PermissionAsk && no.confirmFunc != nil {
		if !no.confirmFunc("Notebook Execute", fmt.Sprintf("Execute all cells in: %s", path)) {
			return &ToolResponse{
				Content: []ContentBlock{{Type: "text", Text: "Notebook execute cancelled by user"}},
				IsError: true,
			}, fmt.Errorf("cancelled by user")
		}
	}

	notebook, err := no.ReadNotebook(ctx, path)
	if err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error reading notebook: %v", err)}},
			IsError: true,
		}, err
	}

	// Switch to the notebook being executed (before starting execution)
	if no.setActiveFileFunc != nil {
		_ = no.setActiveFileFunc(path)
	}

	var outputs []string
	for i, cell := range notebook.Cells {
		output, err := no.executeCellFunc(cell.Language, cell.Source)
		if err != nil {
			outputs = append(outputs, fmt.Sprintf("Cell %d error: %v\n%s", i, err, output))
		} else {
			outputs = append(outputs, fmt.Sprintf("Cell %d output:\n%s", i, output))
		}
	}

	return &ToolResponse{
		Content: []ContentBlock{{Type: "text", Text: strings.Join(outputs, "\n\n")}},
	}, nil
}

// ConvertIPyNBToIgonb converts an ipynb file to igonb format
func (no *NotebookOperations) ConvertIPyNBToIgonb(ctx context.Context, ipynbPath string, igonbPath string) (*ToolResponse, error) {
	if no.config.NotebookModify == PermissionDeny {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: "Notebook convert permission denied"}},
			IsError: true,
		}, fmt.Errorf("permission denied")
	}

	if no.config.NotebookModify == PermissionAsk && no.confirmFunc != nil {
		if !no.confirmFunc("Notebook Convert", fmt.Sprintf("Convert %s to %s", ipynbPath, igonbPath)) {
			return &ToolResponse{
				Content: []ContentBlock{{Type: "text", Text: "Notebook convert cancelled by user"}},
				IsError: true,
			}, fmt.Errorf("cancelled by user")
		}
	}

	fullIPyNBPath := filepath.Join(no.workspaceRoot, ipynbPath)
	fullIgonbPath := filepath.Join(no.workspaceRoot, igonbPath)

	// Read ipynb file
	ipynbContent, err := os.ReadFile(fullIPyNBPath)
	if err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error reading ipynb file: %v", err)}},
			IsError: true,
		}, err
	}

	// Parse ipynb
	var ipynb struct {
		Cells []struct {
			CellType string   `json:"cell_type"`
			Source   []string `json:"source"`
		} `json:"cells"`
	}
	if err := json.Unmarshal(ipynbContent, &ipynb); err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error parsing ipynb: %v", err)}},
			IsError: true,
		}, err
	}

	// Convert to igonb format
	igonb := Notebook{
		Version: 1,
		Cells:   make([]NotebookCell, 0),
	}

	for _, cell := range ipynb.Cells {
		var language string
		switch cell.CellType {
		case "code":
			language = "python" // ipynb uses Python by default
		case "markdown":
			language = "markdown"
		default:
			continue // Skip unknown cell types
		}

		source := strings.Join(cell.Source, "")
		igonb.Cells = append(igonb.Cells, NotebookCell{
			Language: language,
			Source:   source,
		})
	}

	// Write igonb file
	igonbContent, err := json.MarshalIndent(igonb, "", "  ")
	if err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error encoding igonb: %v", err)}},
			IsError: true,
		}, err
	}

	if err := os.WriteFile(fullIgonbPath, igonbContent, 0644); err != nil {
		return &ToolResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error writing igonb file: %v", err)}},
			IsError: true,
		}, err
	}

	return &ToolResponse{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Successfully converted %s to %s", ipynbPath, igonbPath)}},
	}, nil
}
