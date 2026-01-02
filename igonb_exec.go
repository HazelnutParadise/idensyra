package main

import (
	"fmt"
	"os"

	"github.com/HazelnutParadise/idensyra/igonb"
	"github.com/HazelnutParadise/idensyra/internal"
	"github.com/wailsapp/wails/v2/pkg/runtime"
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

	formattedResults := make([]igonb.CellResult, 0)
	results, runErr := runIgonbInWorkspace(exec, nb, cellIndex, func(result igonb.CellResult) {
		formatted := formatIgonbResult(result)
		formattedResults = append(formattedResults, formatted)
		if a != nil && a.ctx != nil {
			runtime.EventsEmit(a.ctx, "igonb:cell-result", formatted)
		}
	})
	if len(formattedResults) != len(results) {
		formattedResults = formatIgonbResults(results)
	}

	return formattedResults, runErr
}

func runIgonbInWorkspace(exec *igonb.Executor, nb *igonb.Notebook, cellIndex int, onResult func(igonb.CellResult)) ([]igonb.CellResult, error) {
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
		return exec.RunNotebookWithCallback(nb, cellIndex, onResult)
	}
	return exec.RunNotebookWithCallback(nb, -1, onResult)
}

// ExecuteIgonb runs all cells and returns formatted results.
func (a *App) ExecuteIgonb(content string) ([]igonb.CellResult, error) {
	return a.ExecuteIgonbCells(content, -1)
}

func formatIgonbResults(results []igonb.CellResult) []igonb.CellResult {
	if len(results) == 0 {
		return results
	}
	formatted := make([]igonb.CellResult, len(results))
	for i, result := range results {
		formatted[i] = formatIgonbResult(result)
	}
	return formatted
}

func formatIgonbResult(result igonb.CellResult) igonb.CellResult {
	if result.Language == "markdown" {
		return result
	}
	if result.Output == "" {
		return result
	}
	result.Output = internal.AnsiToHTMLWithBG(result.Output, "dark")
	return result
}
