package igonb

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/HazelnutParadise/insyra"
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
	goInterp       *interp.Interpreter
	goOutput       bytes.Buffer
	goOutputOffset int
	goImports      map[string]bool
	pythonSession  *PythonSession
	pythonMu       sync.Mutex
	pythonErr      error
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
		goImports: make(map[string]bool),
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
	return e.RunNotebookWithCallback(nb, -1, nil)
}

func (e *Executor) RunNotebookUpTo(nb *Notebook, index int) ([]CellResult, error) {
	return e.RunNotebookWithCallback(nb, index, nil)
}

func (e *Executor) RunNotebookCell(nb *Notebook, index int) ([]CellResult, error) {
	return e.RunNotebookCellWithCallback(nb, index, nil)
}

func (e *Executor) RunNotebookCellWithCallback(nb *Notebook, index int, onResult func(CellResult)) ([]CellResult, error) {
	if nb == nil {
		return nil, fmt.Errorf("notebook is nil")
	}
	if index < 0 || index >= len(nb.Cells) {
		return nil, fmt.Errorf("cell index out of range: %d", index)
	}

	emit := func(result CellResult, results []CellResult) []CellResult {
		results = append(results, result)
		if onResult != nil {
			onResult(result)
		}
		return results
	}

	cell := nb.Cells[index]
	lang := NormalizeLanguage(cell.Language)
	results := make([]CellResult, 0, 1)

	switch lang {
	case "markdown":
		results = emit(CellResult{
			Index:    index,
			Language: lang,
			Output:   "",
		}, results)
	case "go":
		output, err := e.runGoCell(cell.Source)
		result := CellResult{
			Index:    index,
			Language: lang,
			Output:   output,
		}
		if err != nil {
			result.Error = err.Error()
			results = emit(result, results)
			return results, err
		}
		results = emit(result, results)
	case "python":
		groupResults, err := e.runPythonGroup([]Cell{cell}, []int{index})
		for _, result := range groupResults {
			results = emit(result, results)
		}
		if err != nil {
			return results, err
		}
	default:
		result := CellResult{
			Index:    index,
			Language: lang,
			Error:    fmt.Sprintf("unsupported language: %s", cell.Language),
		}
		results = emit(result, results)
		return results, fmt.Errorf("%s", result.Error)
	}

	return results, nil
}

func (e *Executor) RunNotebookWithCallback(nb *Notebook, index int, onResult func(CellResult)) ([]CellResult, error) {
	if nb == nil {
		return nil, fmt.Errorf("notebook is nil")
	}
	if index >= len(nb.Cells) {
		return nil, fmt.Errorf("cell index out of range: %d", index)
	}

	maxIndex := len(nb.Cells) - 1
	if index >= 0 && index < maxIndex {
		maxIndex = index
	}

	results := make([]CellResult, 0, len(nb.Cells))
	pythonHandled := make([]bool, len(nb.Cells))

	emit := func(result CellResult) {
		results = append(results, result)
		if onResult != nil {
			onResult(result)
		}
	}

	for idx := 0; idx <= maxIndex; idx++ {
		if pythonHandled[idx] {
			continue
		}
		cell := nb.Cells[idx]
		lang := NormalizeLanguage(cell.Language)

		switch lang {
		case "markdown":
			emit(CellResult{
				Index:    idx,
				Language: lang,
				Output:   "",
			})
		case "go":
			output, err := e.runGoCell(cell.Source)
			result := CellResult{
				Index:    idx,
				Language: lang,
				Output:   output,
			}
			if err != nil {
				result.Error = err.Error()
				emit(result)
				return results, err
			}
			emit(result)
		case "python":
			group := make([]Cell, 0)
			groupIndices := make([]int, 0)
			for scan := idx; scan < len(nb.Cells); scan++ {
				if index >= 0 && scan > index {
					break
				}
				scanLang := NormalizeLanguage(nb.Cells[scan].Language)
				if scanLang == "python" {
					group = append(group, nb.Cells[scan])
					groupIndices = append(groupIndices, scan)
					pythonHandled[scan] = true
					continue
				}
				if scanLang == "markdown" {
					continue
				}
				break
			}

			if len(group) == 0 {
				break
			}

			groupResults, err := e.runPythonGroup(group, groupIndices)
			for _, result := range groupResults {
				emit(result)
			}
			if err != nil {
				return results, err
			}
		default:
			result := CellResult{
				Index:    idx,
				Language: lang,
				Error:    fmt.Sprintf("unsupported language: %s", cell.Language),
			}
			emit(result)
			return results, fmt.Errorf("%s", result.Error)
		}
	}

	return results, nil
}

func (e *Executor) runGoCell(code string) (string, error) {
	if strings.TrimSpace(code) == "" {
		return "", nil
	}

	code = normalizeGoRangeLoops(code)
	segments := expandGoSegments(splitGoSegments(code))
	if len(segments) == 0 {
		return "", nil
	}

	lastCodeIndex := -1
	for i := len(segments) - 1; i >= 0; i-- {
		if segments[i].kind == goSegmentCode && strings.TrimSpace(segments[i].text) != "" {
			lastCodeIndex = i
			break
		}
	}

	var output strings.Builder
	for idx, segment := range segments {
		if strings.TrimSpace(segment.text) == "" {
			continue
		}
		code := segment.text
		var newImports []string
		var err error
		if segment.kind == goSegmentImport {
			code, newImports, err = e.filterGoImportSegment(segment.text)
			if err != nil {
				return output.String(), err
			}
			if strings.TrimSpace(code) == "" {
				continue
			}
		}
		if segment.kind == goSegmentCode && idx == lastCodeIndex {
			prefix, expr := splitGoTrailingExpression(code)
			if expr != "" {
				if strings.TrimSpace(prefix) != "" {
					chunk, err := e.runGoSegment(prefix, false)
					if chunk != "" {
						output.WriteString(chunk)
					}
					if err != nil {
						return output.String(), err
					}
				}
				chunk, err := e.runGoSegment(expr, true)
				if chunk != "" {
					output.WriteString(chunk)
				}
				if err != nil {
					return output.String(), err
				}
				continue
			}
		}
		chunk, err := e.runGoSegment(code, false)
		if chunk != "" {
			output.WriteString(chunk)
		}
		if err != nil {
			return output.String(), err
		}
		if segment.kind == goSegmentImport && len(newImports) > 0 {
			e.trackGoImports(newImports)
		}
	}

	return output.String(), nil
}

func (e *Executor) runPythonCell(code string) (string, error) {
	if strings.TrimSpace(code) == "" {
		return "", nil
	}

	pipeOutput, err := captureStdIO(func() error {
		return py.RunCode(nil, code)
	})
	return pipeOutput, err
}

func (e *Executor) ensurePythonSession() (*PythonSession, error) {
	if e.pythonSession != nil || e.pythonErr != nil {
		return e.pythonSession, e.pythonErr
	}

	e.pythonMu.Lock()
	defer e.pythonMu.Unlock()
	if e.pythonSession != nil || e.pythonErr != nil {
		return e.pythonSession, e.pythonErr
	}

	session, err := NewPythonSession()
	if err != nil {
		e.pythonErr = err
		return nil, err
	}
	e.pythonSession = session
	return session, nil
}

func (e *Executor) Close() error {
	if e == nil {
		return nil
	}
	e.pythonMu.Lock()
	defer e.pythonMu.Unlock()
	if e.pythonSession != nil {
		_ = e.pythonSession.Close()
		e.pythonSession = nil
	}
	e.pythonErr = nil
	return nil
}

func (e *Executor) PreloadGoImports(paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	newPaths := make([]string, 0, len(paths))
	for _, path := range paths {
		trimmed := strings.TrimSpace(path)
		if trimmed == "" {
			continue
		}
		if e.goImports != nil && e.goImports[trimmed] {
			continue
		}
		newPaths = append(newPaths, trimmed)
	}
	if len(newPaths) == 0 {
		return nil
	}

	specs := make([]goImportSpec, 0, len(newPaths))
	for _, path := range newPaths {
		specs = append(specs, goImportSpec{Path: path})
	}
	code := buildGoImportSegment(specs)
	if strings.TrimSpace(code) == "" {
		return nil
	}
	if _, err := e.runGoSegment(code, false); err != nil {
		return err
	}
	e.trackGoImports(newPaths)
	return nil
}

func (e *Executor) runGoSegment(code string, allowAutoOutput bool) (string, error) {
	var evalValue reflect.Value
	pipeOutput, err := captureStdIO(func() error {
		value, evalErr := e.goInterp.Eval(code)
		evalValue = value
		return evalErr
	})

	all := e.goOutput.Bytes()
	var interpOutput string
	if e.goOutputOffset < len(all) {
		interpOutput = string(all[e.goOutputOffset:])
	}
	e.goOutputOffset = len(all)

	valueOutput := ""
	if allowAutoOutput && err == nil && interpOutput == "" && pipeOutput == "" && !isGoDeclarationChunk(code) && !isGoAssignmentChunk(code) {
		if showOutput, shown := autoShowGoValue(evalValue); shown {
			valueOutput = showOutput
		} else {
			valueOutput = formatGoEvalValue(evalValue)
			if valueOutput != "" && !strings.HasSuffix(valueOutput, "\n") {
				valueOutput += "\n"
			}
		}
	}

	return interpOutput + pipeOutput + valueOutput, err
}

func autoShowGoValue(value reflect.Value) (string, bool) {
	if !value.IsValid() || !value.CanInterface() {
		return "", false
	}

	if output, ok := showIDataList(value.Interface()); ok {
		return output, true
	}
	if output, ok := showIDataTable(value.Interface()); ok {
		return output, true
	}
	if value.Kind() == reflect.Struct && value.CanAddr() {
		if output, ok := showIDataList(value.Addr().Interface()); ok {
			return output, true
		}
		if output, ok := showIDataTable(value.Addr().Interface()); ok {
			return output, true
		}
	}
	return "", false
}

func showIDataList(value any) (string, bool) {
	list, ok := value.(insyra.IDataList)
	if !ok || list == nil {
		return "", false
	}
	output, err := captureStdIO(func() error {
		list.Show()
		return nil
	})
	if err != nil {
		return "", false
	}
	return output, true
}

func showIDataTable(value any) (string, bool) {
	table, ok := value.(insyra.IDataTable)
	if !ok || table == nil {
		return "", false
	}
	output, err := captureStdIO(func() error {
		table.Show()
		return nil
	})
	if err != nil {
		return "", false
	}
	return output, true
}

func formatGoEvalValue(value reflect.Value) string {
	if !value.IsValid() || !value.CanInterface() {
		return ""
	}
	if value.Kind() == reflect.Func {
		return ""
	}
	return fmt.Sprint(value.Interface())
}

func isGoDeclarationChunk(code string) bool {
	lines := strings.Split(code, "\n")
	inBlockComment := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if inBlockComment {
			if strings.Contains(trimmed, "*/") {
				inBlockComment = false
			}
			continue
		}
		if strings.HasPrefix(trimmed, "/*") {
			if !strings.Contains(trimmed, "*/") {
				inBlockComment = true
			}
			continue
		}
		if strings.HasPrefix(trimmed, "//") {
			continue
		}

		switch firstGoToken(trimmed) {
		case "func", "type", "var", "const", "import", "package":
			return true
		default:
			return false
		}
	}

	return false
}

func isGoAssignmentChunk(code string) bool {
	inBlockComment := false
	inRawString := false
	inString := false
	inRune := false

	for i := 0; i < len(code); i++ {
		c := code[i]
		next := byte(0)
		if i+1 < len(code) {
			next = code[i+1]
		}

		if inBlockComment {
			if c == '*' && next == '/' {
				inBlockComment = false
				i++
			}
			continue
		}
		if inRawString {
			if c == '`' {
				inRawString = false
			}
			continue
		}
		if inString {
			if c == '\\' {
				i++
				continue
			}
			if c == '"' {
				inString = false
			}
			continue
		}
		if inRune {
			if c == '\\' {
				i++
				continue
			}
			if c == '\'' {
				inRune = false
			}
			continue
		}

		if c == '/' && next == '*' {
			inBlockComment = true
			i++
			continue
		}
		if c == '/' && next == '/' {
			for i < len(code) && code[i] != '\n' {
				i++
			}
			continue
		}
		if c == '`' {
			inRawString = true
			continue
		}
		if c == '"' {
			inString = true
			continue
		}
		if c == '\'' {
			inRune = true
			continue
		}

		if c == ':' && next == '=' {
			return true
		}
		if next == '=' {
			switch c {
			case '+', '-', '*', '/', '%', '&', '|', '^':
				return true
			case '<', '>':
				if previousNonSpaceByte(code, i-1) == c {
					return true
				}
			}
		}
		if c == '=' {
			if next == '=' {
				continue
			}
			prev := previousNonSpaceByte(code, i-1)
			if prev == 0 {
				return true
			}
			if prev == '!' || prev == '<' || prev == '>' || prev == '=' || prev == ':' {
				continue
			}
			return true
		}
	}

	return false
}

func previousNonSpaceByte(code string, index int) byte {
	for idx := index; idx >= 0; idx-- {
		if code[idx] != ' ' && code[idx] != '\t' && code[idx] != '\n' && code[idx] != '\r' {
			return code[idx]
		}
	}
	return 0
}

func firstGoToken(line string) string {
	for i, r := range line {
		if r == ' ' || r == '\t' || r == '(' || r == '{' {
			return line[:i]
		}
	}
	return line
}

func buildPythonScript(imports []string, body string) string {
	if len(imports) == 0 {
		return body
	}
	return strings.Join(imports, "\n") + "\n" + body
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

type pythonCellOutput struct {
	Index  int    `json:"index"`
	Output string `json:"output"`
	Error  string `json:"error"`
}

const pythonOutputMarker = "__IGONB__"

func (e *Executor) runPythonGroup(cells []Cell, indices []int) ([]CellResult, error) {
	if len(cells) == 0 {
		return nil, nil
	}
	if len(indices) != len(cells) {
		return nil, fmt.Errorf("python group index mismatch")
	}

	session, err := e.ensurePythonSession()
	if err != nil {
		return nil, err
	}

	results := make([]CellResult, 0, len(cells))
	for i, cell := range cells {
		output, runErr := session.Run(cell.Source)
		result := CellResult{
			Index:    indices[i],
			Language: "python",
			Output:   output,
		}
		if runErr != nil {
			result.Error = runErr.Error()
			results = append(results, result)
			return results, runErr
		}
		results = append(results, result)
	}

	return results, nil
}

type goImportSpec struct {
	Name string
	Path string
}

func (e *Executor) filterGoImportSegment(code string) (string, []string, error) {
	specs, err := parseGoImportSpecs(code)
	if err != nil {
		return "", nil, err
	}
	if len(specs) == 0 {
		return "", nil, nil
	}

	newSpecs := make([]goImportSpec, 0, len(specs))
	newPaths := make([]string, 0, len(specs))
	for _, spec := range specs {
		if spec.Path == "" {
			continue
		}
		if e.goImports != nil && e.goImports[spec.Path] {
			continue
		}
		newSpecs = append(newSpecs, spec)
		newPaths = append(newPaths, spec.Path)
	}

	if len(newSpecs) == 0 {
		return "", nil, nil
	}
	return buildGoImportSegment(newSpecs), newPaths, nil
}

func (e *Executor) trackGoImports(paths []string) {
	if len(paths) == 0 {
		return
	}
	if e.goImports == nil {
		e.goImports = make(map[string]bool)
	}
	for _, path := range paths {
		if path == "" {
			continue
		}
		e.goImports[path] = true
	}
}

func parseGoImportSpecs(code string) ([]goImportSpec, error) {
	src := "package main\n" + code + "\n"
	file, err := parser.ParseFile(token.NewFileSet(), "imports.go", src, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}

	specs := make([]goImportSpec, 0, len(file.Imports))
	for _, spec := range file.Imports {
		pathValue := strings.Trim(spec.Path.Value, "`\"")
		name := ""
		if spec.Name != nil {
			name = spec.Name.Name
		}
		specs = append(specs, goImportSpec{Name: name, Path: pathValue})
	}
	return specs, nil
}

func buildGoImportSegment(specs []goImportSpec) string {
	if len(specs) == 0 {
		return ""
	}
	if len(specs) == 1 {
		spec := specs[0]
		if spec.Name == "" {
			return fmt.Sprintf("import %q", spec.Path)
		}
		return fmt.Sprintf("import %s %q", spec.Name, spec.Path)
	}

	var builder strings.Builder
	builder.WriteString("import (\n")
	for _, spec := range specs {
		if spec.Name == "" {
			builder.WriteString(fmt.Sprintf("\t%q\n", spec.Path))
		} else {
			builder.WriteString(fmt.Sprintf("\t%s %q\n", spec.Name, spec.Path))
		}
	}
	builder.WriteString(")")
	return builder.String()
}

func buildPythonGroupScript(codes []string) (string, error) {
	payload, err := json.Marshal(codes)
	if err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(payload)
	script := fmt.Sprintf(`
import base64, json, sys, io, contextlib, traceback, ast
_cells = json.loads(base64.b64decode("%s").decode("utf-8"))
_results = []
_globals = globals()
def _exec_cell(_code, _globals):
    _tree = ast.parse(_code, mode="exec")
    if _tree.body and isinstance(_tree.body[-1], ast.Expr):
        _last = _tree.body.pop()
        if _tree.body:
            _module = ast.Module(body=_tree.body, type_ignores=[])
            exec(compile(_module, "<igonb>", "exec"), _globals)
        _expr = ast.Expression(_last.value)
        return eval(compile(_expr, "<igonb>", "eval"), _globals)
    exec(_code, _globals)
    return None
for _idx, _code in enumerate(_cells):
    _buf_out = io.StringIO()
    _buf_err = io.StringIO()
    try:
        with contextlib.redirect_stdout(_buf_out), contextlib.redirect_stderr(_buf_err):
            _value = _exec_cell(_code, _globals)
            if _value is not None:
                print(_value)
        _results.append({"index": _idx, "output": _buf_out.getvalue() + _buf_err.getvalue(), "error": ""})
    except Exception:
        _results.append({"index": _idx, "output": _buf_out.getvalue() + _buf_err.getvalue(), "error": traceback.format_exc()})
        break
print("%s" + json.dumps(_results))
`, encoded, pythonOutputMarker)
	return script, nil
}

func parsePythonGroupOutput(output string) ([]pythonCellOutput, error) {
	index := strings.LastIndex(output, pythonOutputMarker)
	if index == -1 {
		return nil, fmt.Errorf("python output missing marker")
	}

	payload := strings.TrimSpace(output[index+len(pythonOutputMarker):])
	if payload == "" {
		return nil, fmt.Errorf("python output missing payload")
	}

	var parsed []pythonCellOutput
	if err := json.Unmarshal([]byte(payload), &parsed); err != nil {
		return nil, err
	}
	return parsed, nil
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

var goRangeLoopRegex = regexp.MustCompile(`(?m)^(\s*)for\s+([A-Za-z_]\w*)\s*:=\s*range\s+([0-9]+)\s*\{`)

func normalizeGoRangeLoops(code string) string {
	return goRangeLoopRegex.ReplaceAllString(code, "${1}for ${2} := 0; ${2} < ${3}; ${2}++ {")
}

func expandGoSegments(segments []goSegment) []goSegment {
	if len(segments) == 0 {
		return segments
	}

	expanded := make([]goSegment, 0, len(segments))
	for _, segment := range segments {
		if segment.kind != goSegmentCode {
			expanded = append(expanded, segment)
			continue
		}
		subSegments := splitGoCodeSegments(segment.text)
		for _, sub := range subSegments {
			expanded = append(expanded, goSegment{
				kind: goSegmentCode,
				text: sub,
			})
		}
	}
	return expanded
}

func splitGoCodeSegments(code string) []string {
	lines := strings.Split(code, "\n")
	segments := make([]string, 0)

	var buf strings.Builder
	mode := ""
	braceDepth := 0
	declParenDepth := 0
	inBlockComment := false
	inRawString := false

	flush := func() {
		text := strings.TrimRight(buf.String(), "\n")
		if strings.TrimSpace(text) != "" {
			segments = append(segments, text)
		}
		buf.Reset()
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		atTop := braceDepth == 0 && declParenDepth == 0 && !inBlockComment && !inRawString
		if atTop && trimmed != "" && !strings.HasPrefix(trimmed, "//") {
			if isDeclStart(trimmed) {
				if mode != "decl" {
					flush()
					mode = "decl"
				}
			} else {
				if mode != "stmt" {
					flush()
					mode = "stmt"
				}
			}
		}

		buf.WriteString(line)
		buf.WriteString("\n")

		braceDelta, parenDelta := scanGoLine(line, &inBlockComment, &inRawString)
		braceDepth += braceDelta

		if mode == "decl" {
			if declParenDepth > 0 {
				declParenDepth += parenDelta
				if declParenDepth < 0 {
					declParenDepth = 0
				}
			} else if isDeclBlockStart(trimmed) {
				declParenDepth = parenDelta
				if declParenDepth < 0 {
					declParenDepth = 0
				}
			}
		}
	}

	flush()
	return segments
}

func splitGoTrailingExpression(code string) (string, string) {
	if prefix, expr, ok := splitGoTrailingExpressionAST(code); ok {
		return prefix, expr
	}

	lines := strings.Split(code, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" {
			continue
		}
		trimmed = stripGoLineComment(trimmed)
		if trimmed == "" {
			continue
		}
		stmtPart, exprPart := splitGoLineLastSegment(trimmed)
		if exprPart == "" {
			return code, ""
		}
		if _, err := parser.ParseExpr(exprPart); err != nil {
			return code, ""
		}
		prefix := strings.Join(lines[:i], "\n")
		if stmtPart != "" {
			if strings.TrimSpace(prefix) != "" {
				prefix = prefix + "\n" + strings.TrimSpace(stmtPart)
			} else {
				prefix = strings.TrimSpace(stmtPart)
			}
		}
		return prefix, exprPart
	}
	return code, ""
}

func splitGoTrailingExpressionAST(code string) (string, string, bool) {
	if strings.TrimSpace(code) == "" {
		return "", "", false
	}

	const prefix = "package main\nfunc _(){\n"
	const suffix = "\n}"
	src := prefix + code + suffix

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "snippet.go", src, parser.AllErrors)
	if err != nil {
		return "", "", false
	}

	var target *ast.FuncDecl
	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Name == nil || funcDecl.Name.Name != "_" {
			continue
		}
		target = funcDecl
		break
	}
	if target == nil || target.Body == nil || len(target.Body.List) == 0 {
		return "", "", false
	}

	lastStmt := target.Body.List[len(target.Body.List)-1]
	exprStmt, ok := lastStmt.(*ast.ExprStmt)
	if !ok {
		return "", "", false
	}

	start := fset.Position(exprStmt.X.Pos()).Offset - len(prefix)
	end := fset.Position(exprStmt.X.End()).Offset - len(prefix)
	if start < 0 || end < start || end > len(code) {
		return "", "", false
	}

	exprText := strings.TrimSpace(code[start:end])
	if exprText == "" {
		return "", "", false
	}
	prefixText := strings.TrimRight(code[:start], " \t\r\n")
	return prefixText, exprText, true
}

func stripGoLineComment(line string) string {
	inString := false
	inRune := false
	inRaw := false

	for i := 0; i < len(line); i++ {
		c := line[i]
		next := byte(0)
		if i+1 < len(line) {
			next = line[i+1]
		}

		if inRaw {
			if c == '`' {
				inRaw = false
			}
			continue
		}
		if inString {
			if c == '\\' {
				i++
				continue
			}
			if c == '"' {
				inString = false
			}
			continue
		}
		if inRune {
			if c == '\\' {
				i++
				continue
			}
			if c == '\'' {
				inRune = false
			}
			continue
		}

		if c == '`' {
			inRaw = true
			continue
		}
		if c == '"' {
			inString = true
			continue
		}
		if c == '\'' {
			inRune = true
			continue
		}
		if c == '/' && next == '/' {
			return strings.TrimSpace(line[:i])
		}
	}

	return strings.TrimSpace(line)
}

func splitGoLineLastSegment(line string) (string, string) {
	inString := false
	inRune := false
	inRaw := false
	lastSemicolon := -1

	for i := 0; i < len(line); i++ {
		c := line[i]

		if inRaw {
			if c == '`' {
				inRaw = false
			}
			continue
		}
		if inString {
			if c == '\\' {
				i++
				continue
			}
			if c == '"' {
				inString = false
			}
			continue
		}
		if inRune {
			if c == '\\' {
				i++
				continue
			}
			if c == '\'' {
				inRune = false
			}
			continue
		}

		if c == '`' {
			inRaw = true
			continue
		}
		if c == '"' {
			inString = true
			continue
		}
		if c == '\'' {
			inRune = true
			continue
		}
		if c == ';' {
			lastSemicolon = i
		}
	}

	if lastSemicolon == -1 {
		return "", strings.TrimSpace(line)
	}
	stmtPart := strings.TrimSpace(line[:lastSemicolon])
	exprPart := strings.TrimSpace(line[lastSemicolon+1:])
	if exprPart == "" {
		return "", ""
	}
	return stmtPart, exprPart
}

func scanGoLine(line string, inBlockComment *bool, inRawString *bool) (int, int) {
	braceDelta := 0
	parenDelta := 0
	inString := false
	inRune := false

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
		if inString {
			if c == '\\' {
				i++
				continue
			}
			if c == '"' {
				inString = false
			}
			continue
		}
		if inRune {
			if c == '\\' {
				i++
				continue
			}
			if c == '\'' {
				inRune = false
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
			continue
		}
		if c == '"' {
			inString = true
			continue
		}
		if c == '\'' {
			inRune = true
			continue
		}

		switch c {
		case '{':
			braceDelta++
		case '}':
			braceDelta--
		case '(':
			parenDelta++
		case ')':
			parenDelta--
		}
	}

	return braceDelta, parenDelta
}

func isDeclStart(trimmed string) bool {
	return isDeclKeyword(trimmed, "func") ||
		isDeclKeyword(trimmed, "type") ||
		isDeclKeyword(trimmed, "var") ||
		isDeclKeyword(trimmed, "const") ||
		isDeclKeyword(trimmed, "import")
}

func isDeclKeyword(trimmed, keyword string) bool {
	if !strings.HasPrefix(trimmed, keyword) {
		return false
	}
	if len(trimmed) == len(keyword) {
		return true
	}
	next := trimmed[len(keyword)]
	return next == ' ' || next == '\t' || next == '('
}

func isDeclBlockStart(trimmed string) bool {
	for _, keyword := range []string{"var", "const", "type", "import"} {
		if isDeclKeyword(trimmed, keyword) {
			after := strings.TrimSpace(trimmed[len(keyword):])
			return strings.HasPrefix(after, "(")
		}
	}
	return false
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
