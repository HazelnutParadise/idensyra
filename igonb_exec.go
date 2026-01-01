package main

import (
	"fmt"
	"os"

	"github.com/HazelnutParadise/idensyra/igonb"
	"github.com/HazelnutParadise/idensyra/internal"
)

// ExecuteIgonbCells executes an igonb notebook up to cellIndex.
// Use cellIndex < 0 to run all cells.
func (a *App) ExecuteIgonbCells(content string, cellIndex int) ([]igonb.CellResult, error) {
	nb, err := igonb.Parse([]byte(content))
	if err != nil {
		return nil, err
	}
	if cellIndex >= 0 && cellIndex >= len(nb.Cells) {
		return nil, fmt.Errorf("cell index out of range: %d", cellIndex)
	}

	exec, err := igonb.NewExecutorWithSymbols(internal.Symbols)
	if err != nil {
		return nil, err
	}

	results, _ := runIgonbInWorkspace(exec, nb, cellIndex)
	for i := range results {
		if results[i].Language == "markdown" {
			continue
		}
		if results[i].Output == "" {
			continue
		}
		results[i].Output = internal.AnsiToHTMLWithBG(results[i].Output, "dark")
	}

	return results, nil
}

func runIgonbInWorkspace(exec *igonb.Executor, nb *igonb.Notebook, cellIndex int) ([]igonb.CellResult, error) {
	var oldWD string
	var restoreWD bool
	if globalWorkspace != nil {
		globalWorkspace.mu.RLock()
		workspaceDir := globalWorkspace.workDir
		globalWorkspace.mu.RUnlock()
		if workspaceDir != "" {
			if wd, err := os.Getwd(); err == nil {
				if err := os.Chdir(workspaceDir); err == nil {
					oldWD = wd
					restoreWD = true
				}
			}
		}
	}
	if restoreWD {
		defer os.Chdir(oldWD)
	}

	if cellIndex >= 0 {
		return exec.RunNotebookUpTo(nb, cellIndex)
	}
	return exec.RunNotebook(nb)
}

// ExecuteIgonb runs all cells and returns formatted results.
func (a *App) ExecuteIgonb(content string) ([]igonb.CellResult, error) {
	return a.ExecuteIgonbCells(content, -1)
}
