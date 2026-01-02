package main

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/HazelnutParadise/idensyra/igonb"
	"github.com/HazelnutParadise/idensyra/internal"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var igonbRunner = igonb.NewRunner(
	internal.Symbols,
	igonb.WithDefaultGoImports(igonb.DefaultGoImports),
)

func getIgonbExecutorKey() string {
	if globalWorkspace == nil {
		return "default"
	}
	globalWorkspace.mu.RLock()
	activeFile := globalWorkspace.activeFile
	workDir := globalWorkspace.workDir
	globalWorkspace.mu.RUnlock()

	if activeFile == "" {
		return "default"
	}
	if workDir == "" {
		return activeFile
	}
	return filepath.Join(workDir, filepath.FromSlash(activeFile))
}

// ExecuteIgonbCells executes an igonb notebook up to cellIndex.
// Use -1 to run all cells, and use <= -2 to run a single cell (index = -cellIndex - 2).
func (a *App) ExecuteIgonbCells(content string, cellIndex int) ([]igonb.CellResult, error) {
	mode := igonb.RunUpTo
	targetIndex := cellIndex
	if cellIndex == -1 {
		mode = igonb.RunAll
		targetIndex = -1
	} else if cellIndex <= -2 {
		mode = igonb.RunSingle
		targetIndex = -cellIndex - 2
	}

	formatOutput := func(output string) string {
		return internal.AnsiToHTMLWithBG(output, "dark")
	}

	run := func() ([]igonb.CellResult, error) {
		return igonbRunner.Execute(content, igonb.RunOptions{
			Key:       getIgonbExecutorKey(),
			Mode:      mode,
			Index:     targetIndex,
			Formatter: formatOutput,
			OnResult: func(result igonb.CellResult) {
				if a != nil && a.ctx != nil {
					runtime.EventsEmit(a.ctx, "igonb:cell-result", result)
				}
			},
		})
	}

	results, runErr := runIgonbInWorkspace(run)
	if runErr != nil && errors.Is(runErr, igonb.ErrExecutionStopped) {
		if results == nil {
			results = []igonb.CellResult{}
		}
		return results, nil
	}
	if runErr != nil && len(results) > 0 {
		return results, nil
	}
	return results, runErr
}

// ExecuteIgonb runs all cells and returns formatted results.
func (a *App) ExecuteIgonb(content string) ([]igonb.CellResult, error) {
	return a.ExecuteIgonbCells(content, -1)
}

func runIgonbInWorkspace(run func() ([]igonb.CellResult, error)) ([]igonb.CellResult, error) {
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

	return run()
}

// ResetIgonbEnvironment clears the Go/Python execution environment for the active notebook.
func (a *App) ResetIgonbEnvironment() error {
	key := getIgonbExecutorKey()
	return igonbRunner.Reset(key)
}

// StopIgonbExecution requests the current notebook execution to stop after the active cell.
func (a *App) StopIgonbExecution() error {
	key := getIgonbExecutorKey()
	igonbRunner.Cancel(key)
	return nil
}
