package igonb

import (
	"fmt"
	"reflect"
	"sync"
)

type RunMode int

const (
	RunAll RunMode = iota
	RunUpTo
	RunSingle
)

type RunOptions struct {
	Key       string
	Mode      RunMode
	Index     int
	OnResult  func(CellResult)
	Formatter OutputFormatter
}

type RunnerOption func(*Runner)

type Runner struct {
	mu             sync.Mutex
	executors      map[string]*Executor
	symbols        map[string]map[string]reflect.Value
	defaultImports []string
}

func NewRunner(symbols map[string]map[string]reflect.Value, options ...RunnerOption) *Runner {
	runner := &Runner{
		executors: make(map[string]*Executor),
		symbols:   symbols,
	}
	for _, option := range options {
		if option != nil {
			option(runner)
		}
	}
	return runner
}

func WithDefaultGoImports(imports []string) RunnerOption {
	return func(r *Runner) {
		r.defaultImports = append([]string(nil), imports...)
	}
}

func (r *Runner) Execute(content string, options RunOptions) ([]CellResult, error) {
	nb, err := Parse([]byte(content))
	if err != nil {
		return nil, err
	}
	return r.ExecuteNotebook(nb, options)
}

func (r *Runner) ExecuteNotebook(nb *Notebook, options RunOptions) ([]CellResult, error) {
	if nb == nil {
		return nil, fmt.Errorf("notebook is nil")
	}
	key := options.Key
	if key == "" {
		key = "default"
	}

	exec, err := r.getExecutor(key)
	if err != nil {
		return nil, err
	}
	exec.ClearStop()

	formattedResults := make([]CellResult, 0)
	callback := func(result CellResult) {
		formatted := FormatResult(result, options.Formatter)
		formattedResults = append(formattedResults, formatted)
		if options.OnResult != nil {
			options.OnResult(formatted)
		}
	}

	var results []CellResult
	var runErr error
	switch options.Mode {
	case RunSingle:
		results, runErr = exec.RunNotebookCellWithCallback(nb, options.Index, callback)
	case RunAll:
		results, runErr = exec.RunNotebookWithCallback(nb, -1, callback)
	default:
		results, runErr = exec.RunNotebookWithCallback(nb, options.Index, callback)
	}

	if len(formattedResults) != len(results) {
		formattedResults = FormatResults(results, options.Formatter)
	}
	return formattedResults, runErr
}

func (r *Runner) Cancel(key string) {
	if key == "" {
		key = "default"
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if exec, ok := r.executors[key]; ok {
		exec.RequestStop()
	}
}

func (r *Runner) Reset(key string) error {
	if key == "" {
		key = "default"
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if exec, ok := r.executors[key]; ok {
		_ = exec.Close()
		delete(r.executors, key)
	}
	return nil
}

func (r *Runner) getExecutor(key string) (*Executor, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if exec, ok := r.executors[key]; ok {
		return exec, nil
	}

	exec, err := NewExecutorWithSymbols(r.symbols)
	if err != nil {
		return nil, err
	}
	if len(r.defaultImports) > 0 {
		if err := exec.PreloadGoImports(r.defaultImports); err != nil {
			return nil, err
		}
	}
	r.executors[key] = exec
	return exec, nil
}
