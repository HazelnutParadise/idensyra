package igonb

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/HazelnutParadise/insyra/py"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

type CellResult struct {
	Index    int    `json:"index"`
	Language string `json:"language"`
	Output   string `json:"output"`
	Error    string `json:"error,omitempty"`
}

type Executor struct {
	goInterp        *interp.Interpreter
	goOutput        bytes.Buffer
	goOutputOffset  int
	pythonImports   []string
	pythonImportSet map[string]struct{}
}

type GoSetupFunc func(*interp.Interpreter) error

type goSegment struct {
	kind string
	text string
}

const (
	goSegmentImport = "import"
	goSegmentCode   = "code"
)

func NewExecutor(goSetup GoSetupFunc) (*Executor, error) {
	exec := &Executor{
		pythonImports:   make([]string, 0),
		pythonImportSet: make(map[string]struct{}),
	}

	exec.goInterp = interp.New(interp.Options{
		Stdout: &exec.goOutput,
		Stderr: &exec.goOutput,
	})

	if err := exec.goInterp.Use(stdlib.Symbols); err != nil {
		return nil, err
	}
	if goSetup != nil {
		if err := goSetup(exec.goInterp); err != nil {
			return nil, err
		}
	}

	return exec, nil
}

func NewExecutorWithSymbols(symbols map[string]map[string]reflect.Value) (*Executor, error) {
	if symbols == nil {
		return NewExecutor(nil)
	}
	return NewExecutor(func(i *interp.Interpreter) error {
		return i.Use(symbols)
	})
}

func (e *Executor) RunNotebook(nb *Notebook) ([]CellResult, error) {
	results := make([]CellResult, 0, len(nb.Cells))

	for idx, cell := range nb.Cells {
		lang := NormalizeLanguage(cell.Language)
		result := CellResult{
			Index:    idx,
			Language: lang,
		}

		switch lang {
		case "markdown":
			result.Output = ""
			results = append(results, result)
			continue
		case "go":
			output, err := e.runGoCell(cell.Source)
			result.Output = output
			if err != nil {
				result.Error = err.Error()
				results = append(results, result)
				return results, err
			}
		case "python":
			output, err := e.runPythonCell(cell.Source)
			result.Output = output
			if err != nil {
				result.Error = err.Error()
				results = append(results, result)
				return results, err
			}
		default:
			result.Error = fmt.Sprintf("unsupported language: %s", cell.Language)
			results = append(results, result)
			return results, fmt.Errorf(result.Error)
		}

		results = append(results, result)
	}

	return results, nil
}

func (e *Executor) RunNotebookUpTo(nb *Notebook, index int) ([]CellResult, error) {
	if index < 0 {
		return e.RunNotebook(nb)
	}
	if index >= len(nb.Cells) {
		return nil, fmt.Errorf("cell index out of range: %d", index)
	}

	subset := *nb
	subset.Cells = nb.Cells[:index+1]
	return e.RunNotebook(&subset)
}

func (e *Executor) runGoCell(code string) (string, error) {
	if strings.TrimSpace(code) == "" {
		return "", nil
	}

	segments := splitGoSegments(code)
	if len(segments) == 0 {
		return "", nil
	}

	var output strings.Builder
	for _, segment := range segments {
		if strings.TrimSpace(segment.text) == "" {
			continue
		}
		chunk, err := e.runGoSegment(segment.text)
		if chunk != "" {
			output.WriteString(chunk)
		}
		if err != nil {
			return output.String(), err
		}
	}

	return output.String(), nil
}

func (e *Executor) runPythonCell(code string) (string, error) {
	if strings.TrimSpace(code) == "" {
		return "", nil
	}

	script := buildPythonScript(e.pythonImports, code)
	pipeOutput, err := captureStdIO(func() error {
		return py.RunCode(nil, script)
	})
	e.mergePythonImports(code)
	return pipeOutput, err
}

func (e *Executor) runGoSegment(code string) (string, error) {
	pipeOutput, err := captureStdIO(func() error {
		_, evalErr := e.goInterp.Eval(code)
		return evalErr
	})

	all := e.goOutput.Bytes()
	var interpOutput string
	if e.goOutputOffset < len(all) {
		interpOutput = string(all[e.goOutputOffset:])
	}
	e.goOutputOffset = len(all)

	return interpOutput + pipeOutput, err
}

func buildPythonScript(imports []string, body string) string {
	if len(imports) == 0 {
		return body
	}
	return strings.Join(imports, "\n") + "\n" + body
}

func (e *Executor) mergePythonImports(code string) {
	for _, line := range strings.Split(code, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasPrefix(trimmed, "import ") || strings.HasPrefix(trimmed, "from ") {
			if _, exists := e.pythonImportSet[trimmed]; exists {
				continue
			}
			e.pythonImportSet[trimmed] = struct{}{}
			e.pythonImports = append(e.pythonImports, trimmed)
		}
	}
}

func captureStdIO(run func() error) (string, error) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w
	os.Stderr = w

	outputChan := make(chan string, 1)
	go func() {
		var output bytes.Buffer
		_, _ = io.Copy(&output, r)
		_ = r.Close()
		outputChan <- output.String()
	}()

	runErr := run()

	_ = w.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	pipeOutput := <-outputChan
	return pipeOutput, runErr
}

func splitGoSegments(code string) []goSegment {
	lines := strings.Split(code, "\n")
	segments := make([]goSegment, 0)

	var codeBuf strings.Builder
	var importBuf strings.Builder
	inImportBlock := false
	inBlockComment := false
	inRawString := false

	flushCode := func() {
		text := strings.TrimRight(codeBuf.String(), "\n")
		if strings.TrimSpace(text) != "" {
			segments = append(segments, goSegment{kind: goSegmentCode, text: text})
		}
		codeBuf.Reset()
	}
	flushImport := func() {
		text := strings.TrimRight(importBuf.String(), "\n")
		if strings.TrimSpace(text) != "" {
			segments = append(segments, goSegment{kind: goSegmentImport, text: text})
		}
		importBuf.Reset()
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if inImportBlock {
			importBuf.WriteString(line)
			importBuf.WriteString("\n")
			if trimmed == ")" {
				flushImport()
				inImportBlock = false
			}
			updateGoScanState(line, &inBlockComment, &inRawString)
			continue
		}

		if !inBlockComment && !inRawString && isImportLine(trimmed) {
			flushCode()
			if isImportBlockStart(trimmed) {
				importBuf.WriteString(line)
				importBuf.WriteString("\n")
				if isInlineImportBlock(trimmed) {
					flushImport()
				} else {
					inImportBlock = true
				}
			} else {
				segments = append(segments, goSegment{kind: goSegmentImport, text: line})
			}
			updateGoScanState(line, &inBlockComment, &inRawString)
			continue
		}

		codeBuf.WriteString(line)
		codeBuf.WriteString("\n")
		updateGoScanState(line, &inBlockComment, &inRawString)
	}

	flushCode()
	flushImport()

	return segments
}

func isImportLine(trimmed string) bool {
	if trimmed == "" {
		return false
	}
	if strings.HasPrefix(trimmed, "\"") || strings.HasPrefix(trimmed, "`") || strings.HasPrefix(trimmed, "'") {
		return false
	}
	if !strings.HasPrefix(trimmed, "import") {
		return false
	}
	if len(trimmed) == len("import") {
		return true
	}
	next := trimmed[len("import")]
	return next == ' ' || next == '\t' || next == '(' || next == '"'
}

func isImportBlockStart(trimmed string) bool {
	if !strings.HasPrefix(trimmed, "import") {
		return false
	}
	after := strings.TrimSpace(trimmed[len("import"):])
	return strings.HasPrefix(after, "(")
}

func isInlineImportBlock(trimmed string) bool {
	if !isImportBlockStart(trimmed) {
		return false
	}
	return strings.Contains(trimmed, ")")
}

func updateGoScanState(line string, inBlockComment *bool, inRawString *bool) {
	for i := 0; i < len(line); i++ {
		c := line[i]
		next := byte(0)
		if i+1 < len(line) {
			next = line[i+1]
		}

		if *inBlockComment {
			if c == '*' && next == '/' {
				*inBlockComment = false
				i++
			}
			continue
		}

		if *inRawString {
			if c == '`' {
				*inRawString = false
			}
			continue
		}

		if c == '/' && next == '*' {
			*inBlockComment = true
			i++
			continue
		}
		if c == '/' && next == '/' {
			break
		}
		if c == '`' {
			*inRawString = true
		}
	}
}
