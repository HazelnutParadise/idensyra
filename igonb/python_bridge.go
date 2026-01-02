package igonb

import (
	"encoding/json"
	"fmt"
	"go/token"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/HazelnutParadise/insyra"
)

type pythonRunPayload struct {
	Vars  map[string]any `json:"vars"`
	Error string         `json:"error"`
	State string         `json:"state"`
	Defs  []pythonDef    `json:"defs"`
}

type pythonDef struct {
	Name   string `json:"name"`
	Source string `json:"source"`
}

type pythonBinding struct {
	name        string
	placeholder string
	argExpr     string
}

func (e *Executor) runPythonCell(code string) (string, error) {
	if strings.TrimSpace(code) == "" {
		return "", nil
	}
	if err := e.PreloadGoImports([]string{
		"github.com/HazelnutParadise/insyra/py",
		"github.com/HazelnutParadise/insyra",
		"github.com/HazelnutParadise/insyra/isr",
		"math",
	}); err != nil {
		return "", err
	}

	shared := e.snapshotSharedVars()
	state := e.snapshotPythonState()
	defs := e.snapshotPythonDefs()
	bindings := e.buildPythonBindings(shared)
	defsJSON, err := json.Marshal(defs)
	if err != nil {
		return "", err
	}
	defsJSONStr := string(defsJSON)
	if defsJSONStr == "null" || defsJSONStr == "" {
		defsJSONStr = "[]"
	}
	wrapper, err := buildPythonWrapper(code, bindings, len(bindings))
	if err != nil {
		return "", err
	}

	payload, output, err := e.executePythonWrapper(wrapper, bindings, defsJSONStr, state)
	if err != nil {
		return output, err
	}
	if err := e.applyPythonPayload(payload); err != nil {
		return output, err
	}
	if payload.Error != "" {
		return output, fmt.Errorf("%s", payload.Error)
	}
	return output, nil
}

func (e *Executor) snapshotPythonState() string {
	if e == nil {
		return ""
	}
	e.sharedMu.Lock()
	defer e.sharedMu.Unlock()
	return e.pythonState
}

func (e *Executor) setPythonState(state string) {
	if e == nil {
		return
	}
	e.sharedMu.Lock()
	e.pythonState = state
	e.sharedMu.Unlock()
}

func (e *Executor) snapshotPythonDefs() []pythonDef {
	if e == nil {
		return nil
	}
	e.sharedMu.Lock()
	defer e.sharedMu.Unlock()
	if len(e.pythonDefs) == 0 {
		return nil
	}
	copied := make([]pythonDef, len(e.pythonDefs))
	copy(copied, e.pythonDefs)
	return copied
}

func (e *Executor) updatePythonDefs(defs []pythonDef) {
	if e == nil || len(defs) == 0 {
		return
	}
	e.sharedMu.Lock()
	defer e.sharedMu.Unlock()
	index := make(map[string]int, len(e.pythonDefs))
	for i, def := range e.pythonDefs {
		if def.Name != "" {
			index[def.Name] = i
		}
	}
	for _, def := range defs {
		if def.Name == "" || strings.TrimSpace(def.Source) == "" {
			continue
		}
		if idx, ok := index[def.Name]; ok {
			e.pythonDefs[idx] = def
			continue
		}
		e.pythonDefs = append(e.pythonDefs, def)
		index[def.Name] = len(e.pythonDefs) - 1
	}
}

func (e *Executor) buildPythonBindings(shared map[string]any) []pythonBinding {
	if len(shared) == 0 {
		return nil
	}
	names := make([]string, 0, len(shared))
	for name := range shared {
		names = append(names, name)
	}
	sort.Strings(names)

	bindings := make([]pythonBinding, 0, len(names))
	for _, name := range names {
		value := shared[name]
		if !canPassToPython(value) {
			continue
		}
		argExpr, ok := e.pythonArgExpr(name, value)
		if !ok {
			continue
		}
		bindings = append(bindings, pythonBinding{
			name:        name,
			placeholder: fmt.Sprintf("$v%d", len(bindings)+1),
			argExpr:     argExpr,
		})
	}
	return bindings
}

func (e *Executor) pythonArgExpr(name string, value any) (string, bool) {
	if isGoIdentifier(name) && e.goNameExists(name) {
		return name, true
	}
	literal, ok := formatGoLiteral(value)
	if !ok {
		return "", false
	}
	return literal, true
}

func (e *Executor) goNameExists(name string) bool {
	if e == nil || e.goInterp == nil || !isGoIdentifier(name) {
		return false
	}
	_, err := e.goInterp.Eval(name)
	return err == nil
}

func buildPythonWrapper(code string, bindings []pythonBinding, bindingCount int) (string, error) {
	codeLiteral, err := json.Marshal(code)
	if err != nil {
		return "", err
	}

	injected := "{}"
	if len(bindings) > 0 {
		var builder strings.Builder
		builder.WriteString("{\n")
		for i, binding := range bindings {
			builder.WriteString("    ")
			builder.WriteString(strconv.Quote(binding.name))
			builder.WriteString(": ")
			builder.WriteString(binding.placeholder)
			if i < len(bindings)-1 {
				builder.WriteString(",")
			}
			builder.WriteString("\n")
		}
		builder.WriteString("}")
		injected = builder.String()
	}

	defsPlaceholder := fmt.Sprintf("$v%d", bindingCount+1)
	statePlaceholder := fmt.Sprintf("$v%d", bindingCount+2)

	wrapper := fmt.Sprintf(`
import ast, traceback, types, base64, json, pickle

__igonb_code = %s
__igonb_defs_json = %s
__igonb_state_b64 = %s
__igonb_globals = globals()
__igonb_injected = %s
__igonb_reserved = {
    "insyra",
    "insyra_return",
    "pickle",
    "base64",
    "np",
    "pd",
    "pl",
    "plt",
    "sns",
    "scipy",
    "sklearn",
    "sm",
    "go",
    "spacy",
    "bs4",
    "requests",
    "json",
}

def __igonb_exec(code, globs):
    tree = ast.parse(code, mode="exec")
    if tree.body and isinstance(tree.body[-1], ast.Expr):
        last = tree.body.pop()
        if tree.body:
            module = ast.Module(body=tree.body, type_ignores=[])
            exec(compile(module, "<igonb>", "exec"), globs)
        expr = ast.Expression(last.value)
        return eval(compile(expr, "<igonb>", "eval"), globs)
    exec(code, globs)
    return None

def __igonb_collect_defs(code):
    try:
        tree = ast.parse(code, mode="exec")
    except Exception:
        return []
    defs = []
    for node in tree.body:
        if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef, ast.ClassDef)):
            try:
                src = ast.get_source_segment(code, node)
            except Exception:
                src = None
            if src:
                defs.append({"name": node.name, "source": src})
    return defs

def __igonb_load_state(state_b64, globs):
    if not state_b64:
        return
    try:
        data = base64.b64decode(state_b64.encode("utf-8"))
        state = pickle.loads(data)
    except Exception:
        return
    if isinstance(state, dict):
        for key, value in state.items():
            if key.startswith("_") or key in __igonb_reserved:
                continue
            globs[key] = value

def __igonb_dump_state(globs):
    state = {}
    for key, value in globs.items():
        if key.startswith("_") or key in __igonb_reserved:
            continue
        if isinstance(value, (types.ModuleType, types.FunctionType, type)):
            continue
        try:
            pickle.dumps(value)
        except Exception:
            continue
        state[key] = value
    try:
        data = pickle.dumps(state)
        return base64.b64encode(data).decode("utf-8")
    except Exception:
        return ""

def __igonb_export_value(value):
    try:
        import pandas as pd
        if isinstance(value, pd.Series):
            return True, {
                "__igonb_type__": "datalist",
                "data": value.tolist(),
                "name": value.name if value.name is not None else "",
            }
        if isinstance(value, pd.DataFrame):
            return True, {
                "__igonb_type__": "datatable",
                "data": value.to_numpy().tolist(),
                "columns": list(value.columns),
                "index": list(value.index),
            }
    except Exception:
        pass
    try:
        import numpy as np
        if isinstance(value, np.ndarray):
            return True, value.tolist()
        if isinstance(value, np.generic):
            return True, value.item()
    except Exception:
        pass
    if value is None:
        return True, None
    if isinstance(value, (str, int, float, bool)):
        return True, value
    try:
        import json
        if isinstance(value, (list, tuple)):
            json.dumps(value)
            return True, list(value)
        if isinstance(value, dict):
            json.dumps(value)
            return True, value
    except Exception:
        return False, None
    return False, None

def __igonb_export(globs):
    exported = {}
    for key, value in globs.items():
        if key.startswith("_") or key in __igonb_reserved:
            continue
        if isinstance(value, (types.ModuleType, types.FunctionType)):
            continue
        ok, converted = __igonb_export_value(value)
        if ok:
            exported[key] = converted
    return exported

__igonb_error = ""
__igonb_state = ""
__igonb_new_defs = []
try:
    __igonb_defs = []
    if __igonb_defs_json:
        try:
            __igonb_defs = json.loads(__igonb_defs_json)
        except Exception:
            __igonb_defs = []
    for _def in __igonb_defs:
        if isinstance(_def, dict) and _def.get("source"):
            try:
                exec(_def["source"], __igonb_globals)
            except Exception:
                pass
    __igonb_load_state(__igonb_state_b64, __igonb_globals)
    for _k, _v in __igonb_injected.items():
        __igonb_globals[_k] = _v
    _value = __igonb_exec(__igonb_code, __igonb_globals)
    if _value is not None:
        print(_value)
except Exception:
    __igonb_error = traceback.format_exc()

try:
    __igonb_vars = __igonb_export(__igonb_globals)
except Exception:
    if not __igonb_error:
        __igonb_error = traceback.format_exc()
    __igonb_vars = {}

try:
    __igonb_state = __igonb_dump_state(__igonb_globals)
except Exception:
    __igonb_state = ""

try:
    __igonb_new_defs = __igonb_collect_defs(__igonb_code)
except Exception:
    __igonb_new_defs = []

insyra.Return({"vars": __igonb_vars, "error": __igonb_error, "state": __igonb_state, "defs": __igonb_new_defs}, None)
`, string(codeLiteral), defsPlaceholder, statePlaceholder, injected)

	return wrapper, nil
}

func (e *Executor) executePythonWrapper(wrapper string, bindings []pythonBinding, defsJSON string, state string) (pythonRunPayload, string, error) {
	var payload pythonRunPayload
	if e == nil || e.goInterp == nil {
		return payload, "", fmt.Errorf("executor not initialized")
	}

	runID := e.nextPythonRunID()
	resultVar := fmt.Sprintf("__igonb_py_result_%d", runID)
	errVar := fmt.Sprintf("__igonb_py_err_%d", runID)

	if _, err := e.runGoSegment(fmt.Sprintf("var %s map[string]any", resultVar), false); err != nil {
		return payload, "", err
	}
	if _, err := e.runGoSegment(fmt.Sprintf("var %s error", errVar), false); err != nil {
		return payload, "", err
	}

	codeLiteral := strconv.Quote(wrapper)
	argList := ""
	if len(bindings) > 0 {
		args := make([]string, len(bindings))
		for i, binding := range bindings {
			args[i] = binding.argExpr
		}
		args = append(args, strconv.Quote(defsJSON), strconv.Quote(state))
		argList = ", " + strings.Join(args, ", ")
	} else {
		argList = ", " + strings.Join([]string{strconv.Quote(defsJSON), strconv.Quote(state)}, ", ")
	}

	call := fmt.Sprintf("%s = py.RunCodef(&%s, %s%s)", errVar, resultVar, codeLiteral, argList)
	output, err := e.runGoSegment(call, false)
	if err != nil {
		return payload, output, err
	}

	errValue, err := e.goInterp.Eval(errVar)
	if err != nil {
		return payload, output, err
	}
	if errValue.IsValid() && !errValue.IsNil() {
		if errInterface, ok := errValue.Interface().(error); ok && errInterface != nil {
			return payload, output, errInterface
		}
		return payload, output, fmt.Errorf("python execution failed")
	}

	resultValue, err := e.goInterp.Eval(resultVar)
	if err != nil {
		return payload, output, err
	}
	if resultValue.IsValid() && resultValue.CanInterface() {
		if err := decodePythonPayload(resultValue.Interface(), &payload); err != nil {
			return payload, output, err
		}
	}

	return payload, output, nil
}

func decodePythonPayload(value any, payload *pythonRunPayload) error {
	if payload == nil || value == nil {
		return nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, payload)
}

func (e *Executor) applyPythonPayload(payload pythonRunPayload) error {
	if payload.State != "" {
		e.setPythonState(payload.State)
	}
	if len(payload.Defs) > 0 {
		e.updatePythonDefs(payload.Defs)
	}
	if len(payload.Vars) == 0 {
		return nil
	}
	for name, value := range payload.Vars {
		if name == "" || strings.HasPrefix(name, "__igonb") {
			continue
		}
		goValue, ok := convertPythonValue(value)
		if !ok {
			continue
		}
		e.setSharedVar(name, goValue)
		if !isGoIdentifier(name) {
			continue
		}
		if err := e.setGoVariable(name, goValue); err != nil {
			return err
		}
	}
	return nil
}

func convertPythonValue(value any) (any, bool) {
	if value == nil {
		return nil, true
	}
	rawMap, ok := value.(map[string]any)
	if !ok {
		return value, true
	}
	typeTag, ok := rawMap["__igonb_type__"].(string)
	if !ok {
		return value, true
	}
	switch typeTag {
	case "datalist":
		data, ok := coerceAnySlice(rawMap["data"])
		if !ok {
			return nil, false
		}
		list := insyra.NewDataList(data...)
		if name, ok := rawMap["name"].(string); ok && name != "" {
			list.SetName(name)
		}
		return list, true
	case "datatable":
		rows, ok := coerce2DAnySlice(rawMap["data"])
		if !ok {
			return nil, false
		}
		var table *insyra.DataTable
		if len(rows) == 0 {
			table = insyra.NewDataTable()
		} else {
			var err error
			table, err = insyra.Slice2DToDataTable(rows)
			if err != nil {
				return nil, false
			}
		}
		cols := coerceStringSlice(rawMap["columns"])
		if hasNonEmptyStrings(cols) {
			table.SetColNames(cols)
		}
		index := coerceStringSlice(rawMap["index"])
		if hasNonEmptyStrings(index) {
			table.SetRowNames(index)
		}
		return table, true
	default:
		return value, true
	}
}

func (e *Executor) setGoVariable(name string, value any) error {
	targetType := e.goVarType(name)
	if targetType != nil {
		if list, ok := value.(insyra.IDataList); ok {
			return e.assignDataList(name, list, targetType, false)
		}
		if table, ok := value.(insyra.IDataTable); ok {
			return e.assignDataTable(name, table, targetType, false)
		}
		literal, ok := formatGoLiteralForType(value, targetType)
		if !ok {
			return fmt.Errorf("unsupported value for %s", name)
		}
		e.ensureLiteralImports(literal)
		if _, err := e.runGoSegment(fmt.Sprintf("%s = %s", name, literal), false); err != nil {
			return err
		}
		return nil
	}

	if list, ok := value.(insyra.IDataList); ok {
		return e.assignDataList(name, list, nil, true)
	}
	if table, ok := value.(insyra.IDataTable); ok {
		return e.assignDataTable(name, table, nil, true)
	}

	literal, ok := formatGoLiteral(value)
	if !ok {
		return fmt.Errorf("unsupported value for %s", name)
	}
	e.ensureLiteralImports(literal)
	if _, err := e.runGoSegment(fmt.Sprintf("var %s = %s", name, literal), false); err != nil {
		return err
	}
	return nil
}

func (e *Executor) goVarType(name string) reflect.Type {
	if e == nil || e.goInterp == nil || !isGoIdentifier(name) {
		return nil
	}
	value, err := e.goInterp.Eval(name)
	if err != nil || !value.IsValid() {
		return nil
	}
	return value.Type()
}

func (e *Executor) assignDataList(name string, list insyra.IDataList, targetType reflect.Type, declareInterface bool) error {
	if targetType != nil && isIsrDlType(targetType) {
		code, ok := buildIsrDataListAssignment(name, list)
		if !ok {
			return fmt.Errorf("unsupported value for %s", name)
		}
		e.ensureLiteralImports(code)
		if _, err := e.runGoSegment(code, false); err != nil {
			return err
		}
		return nil
	}

	literal, ok := formatDataListLiteralInsyra(list)
	if !ok {
		return fmt.Errorf("unsupported value for %s", name)
	}
	e.ensureLiteralImports(literal)

	if declareInterface {
		if _, err := e.runGoSegment(fmt.Sprintf("var %s insyra.IDataList = %s", name, literal), false); err != nil {
			return err
		}
		return nil
	}
	if _, err := e.runGoSegment(fmt.Sprintf("%s = %s", name, literal), false); err != nil {
		return err
	}
	return nil
}

func (e *Executor) assignDataTable(name string, table insyra.IDataTable, targetType reflect.Type, declareInterface bool) error {
	if targetType != nil && isIsrDtType(targetType) {
		code, ok := buildIsrDataTableAssignment(name, table)
		if !ok {
			return fmt.Errorf("unsupported value for %s", name)
		}
		e.ensureLiteralImports(code)
		if _, err := e.runGoSegment(code, false); err != nil {
			return err
		}
		return nil
	}

	var literal string
	var ok bool
	if targetType != nil && (isInsyraDataTableType(targetType) || isInsyraIDataTableType(targetType)) {
		literal, ok = formatDataTableLiteralInsyra(table)
	} else {
		literal, ok = formatDataTableLiteralIsr(table)
	}
	if !ok {
		return fmt.Errorf("unsupported value for %s", name)
	}
	e.ensureLiteralImports(literal)

	if declareInterface {
		if _, err := e.runGoSegment(fmt.Sprintf("var %s insyra.IDataTable = %s", name, literal), false); err != nil {
			return err
		}
		return nil
	}
	if _, err := e.runGoSegment(fmt.Sprintf("%s = %s", name, literal), false); err != nil {
		return err
	}
	return nil
}

func (e *Executor) ensureLiteralImports(literal string) {
	if strings.Contains(literal, "math.") {
		_ = e.PreloadGoImports([]string{"math"})
	}
	if strings.Contains(literal, "insyra.") {
		_ = e.PreloadGoImports([]string{"github.com/HazelnutParadise/insyra"})
	}
	if strings.Contains(literal, "isr.") {
		_ = e.PreloadGoImports([]string{"github.com/HazelnutParadise/insyra/isr"})
	}
}

func canPassToPython(value any) bool {
	if value == nil {
		return true
	}
	switch value.(type) {
	case insyra.IDataList, insyra.IDataTable:
		return true
	case string, bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return true
	case []any, []string, []float64, []int, []bool, map[string]any:
		return true
	}
	if _, ok := value.(map[any]any); ok {
		return false
	}
	if _, err := json.Marshal(value); err == nil {
		return true
	}
	return false
}

func isGoIdentifier(name string) bool {
	return token.IsIdentifier(name)
}

func coerceAnySlice(value any) ([]any, bool) {
	if value == nil {
		return nil, true
	}
	switch v := value.(type) {
	case []any:
		return v, true
	}
	return nil, false
}

func coerce2DAnySlice(value any) ([][]any, bool) {
	if value == nil {
		return nil, true
	}
	switch v := value.(type) {
	case [][]any:
		return v, true
	case []any:
		rows := make([][]any, len(v))
		for i, row := range v {
			converted, ok := coerceAnySlice(row)
			if !ok {
				return nil, false
			}
			rows[i] = converted
		}
		return rows, true
	}
	return nil, false
}

func coerceStringSlice(value any) []string {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case []string:
		return v
	case []any:
		out := make([]string, len(v))
		for i, item := range v {
			out[i] = fmt.Sprint(item)
		}
		return out
	}
	return nil
}

func hasNonEmptyStrings(values []string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return true
		}
	}
	return false
}

func isPkgType(t reflect.Type, pkgPath string, name string) bool {
	if t == nil {
		return false
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t.PkgPath() == pkgPath && t.Name() == name
}

func isPkgInterfaceType(t reflect.Type, pkgPath string, name string) bool {
	if t == nil {
		return false
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Interface {
		return false
	}
	return t.PkgPath() == pkgPath && t.Name() == name
}

func isIsrDlType(t reflect.Type) bool {
	return isPkgType(t, "github.com/HazelnutParadise/insyra/isr", "dl")
}

func isIsrDtType(t reflect.Type) bool {
	return isPkgType(t, "github.com/HazelnutParadise/insyra/isr", "dt")
}

func isInsyraDataListType(t reflect.Type) bool {
	return isPkgType(t, "github.com/HazelnutParadise/insyra", "DataList")
}

func isInsyraDataTableType(t reflect.Type) bool {
	return isPkgType(t, "github.com/HazelnutParadise/insyra", "DataTable")
}

func isInsyraIDataListType(t reflect.Type) bool {
	return isPkgInterfaceType(t, "github.com/HazelnutParadise/insyra", "IDataList")
}

func isInsyraIDataTableType(t reflect.Type) bool {
	return isPkgInterfaceType(t, "github.com/HazelnutParadise/insyra", "IDataTable")
}

func formatGoLiteral(value any) (string, bool) {
	if value == nil {
		return "nil", true
	}
	switch v := value.(type) {
	case bool:
		return strconv.FormatBool(v), true
	case string:
		return strconv.Quote(v), true
	case int:
		return strconv.Itoa(v), true
	case int8:
		return strconv.FormatInt(int64(v), 10), true
	case int16:
		return strconv.FormatInt(int64(v), 10), true
	case int32:
		return strconv.FormatInt(int64(v), 10), true
	case int64:
		return strconv.FormatInt(v, 10), true
	case uint:
		return strconv.FormatUint(uint64(v), 10), true
	case uint8:
		return strconv.FormatUint(uint64(v), 10), true
	case uint16:
		return strconv.FormatUint(uint64(v), 10), true
	case uint32:
		return strconv.FormatUint(uint64(v), 10), true
	case uint64:
		return strconv.FormatUint(v, 10), true
	case float32:
		return formatFloatLiteral(float64(v)), true
	case float64:
		return formatFloatLiteral(v), true
	case []any:
		return formatAnySliceLiteral(v)
	case []string:
		return formatStringSliceLiteral(v), true
	case [][]any:
		return format2DAnySliceLiteral(v)
	case map[string]any:
		return formatAnyMapLiteral(v)
	case insyra.IDataList:
		return formatDataListLiteral(v)
	case insyra.IDataTable:
		return formatDataTableLiteral(v)
	default:
		rv := reflect.ValueOf(value)
		if rv.IsValid() {
			switch rv.Kind() {
			case reflect.Slice:
				length := rv.Len()
				items := make([]any, length)
				for i := 0; i < length; i++ {
					items[i] = rv.Index(i).Interface()
				}
				return formatAnySliceLiteral(items)
			case reflect.Map:
				if rv.Type().Key().Kind() == reflect.String {
					keys := rv.MapKeys()
					out := make(map[string]any, len(keys))
					for _, key := range keys {
						out[key.String()] = rv.MapIndex(key).Interface()
					}
					return formatAnyMapLiteral(out)
				}
			}
		}
	}
	return "", false
}

func formatGoLiteralForType(value any, targetType reflect.Type) (string, bool) {
	if targetType == nil {
		return formatGoLiteral(value)
	}
	if value == nil {
		if isNilAssignable(targetType) {
			return "nil", true
		}
		return "", false
	}
	return formatValueForType(value, targetType)
}

func isNilAssignable(targetType reflect.Type) bool {
	if targetType == nil {
		return false
	}
	switch targetType.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Map, reflect.Interface, reflect.Func, reflect.Chan:
		return true
	default:
		return false
	}
}

func formatValueForType(value any, targetType reflect.Type) (string, bool) {
	if targetType == nil {
		return formatGoLiteral(value)
	}
	if value == nil {
		if isNilAssignable(targetType) {
			return "nil", true
		}
		return "", false
	}
	switch targetType.Kind() {
	case reflect.Interface:
		if targetType.NumMethod() == 0 {
			return formatGoLiteral(value)
		}
		valueType := reflect.TypeOf(value)
		if valueType != nil && valueType.Implements(targetType) {
			return formatGoLiteral(value)
		}
		return "", false
	case reflect.Bool, reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		return formatScalarLiteralForType(value, targetType)
	case reflect.Slice:
		return formatSliceLiteralForType(value, targetType)
	case reflect.Array:
		return formatArrayLiteralForType(value, targetType)
	case reflect.Map:
		return formatMapLiteralForType(value, targetType)
	default:
		valueType := reflect.TypeOf(value)
		if valueType != nil && valueType.AssignableTo(targetType) {
			return formatGoLiteral(value)
		}
		return "", false
	}
}

func formatScalarLiteralForType(value any, targetType reflect.Type) (string, bool) {
	switch targetType.Kind() {
	case reflect.Bool:
		v, ok := toBool(value)
		if !ok {
			return "", false
		}
		return strconv.FormatBool(v), true
	case reflect.String:
		v, ok := toString(value)
		if !ok {
			return "", false
		}
		return strconv.Quote(v), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, ok := toInt64(value)
		if !ok {
			return "", false
		}
		return strconv.FormatInt(v, 10), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, ok := toUint64(value)
		if !ok {
			return "", false
		}
		return strconv.FormatUint(v, 10), true
	case reflect.Float32, reflect.Float64:
		v, ok := toFloat64(value)
		if !ok {
			return "", false
		}
		return formatFloatLiteral(v), true
	default:
		return "", false
	}
}

func formatSliceLiteralForType(value any, targetType reflect.Type) (string, bool) {
	if targetType == nil || targetType.Kind() != reflect.Slice {
		return "", false
	}
	if value == nil {
		return "nil", true
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return "", false
	}
	var builder strings.Builder
	builder.WriteString(targetType.String())
	builder.WriteString("{")
	for i := 0; i < rv.Len(); i++ {
		item := rv.Index(i).Interface()
		literal, ok := formatValueForType(item, targetType.Elem())
		if !ok {
			return "", false
		}
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(literal)
	}
	builder.WriteString("}")
	return builder.String(), true
}

func formatArrayLiteralForType(value any, targetType reflect.Type) (string, bool) {
	if targetType == nil || targetType.Kind() != reflect.Array {
		return "", false
	}
	if value == nil {
		return "", false
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return "", false
	}
	if rv.Len() != targetType.Len() {
		return "", false
	}
	var builder strings.Builder
	builder.WriteString(targetType.String())
	builder.WriteString("{")
	for i := 0; i < rv.Len(); i++ {
		item := rv.Index(i).Interface()
		literal, ok := formatValueForType(item, targetType.Elem())
		if !ok {
			return "", false
		}
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(literal)
	}
	builder.WriteString("}")
	return builder.String(), true
}

func formatMapLiteralForType(value any, targetType reflect.Type) (string, bool) {
	if targetType == nil || targetType.Kind() != reflect.Map {
		return "", false
	}
	if value == nil {
		return "nil", true
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Map {
		return "", false
	}
	if targetType.Key().Kind() != reflect.String || rv.Type().Key().Kind() != reflect.String {
		return "", false
	}
	keys := rv.MapKeys()
	names := make([]string, 0, len(keys))
	for _, key := range keys {
		names = append(names, key.String())
	}
	sort.Strings(names)

	var builder strings.Builder
	builder.WriteString(targetType.String())
	builder.WriteString("{")
	for i, key := range names {
		valueLiteral, ok := formatValueForType(rv.MapIndex(reflect.ValueOf(key)).Interface(), targetType.Elem())
		if !ok {
			return "", false
		}
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(strconv.Quote(key))
		builder.WriteString(": ")
		builder.WriteString(valueLiteral)
	}
	builder.WriteString("}")
	return builder.String(), true
}

func toBool(value any) (bool, bool) {
	v, ok := value.(bool)
	return v, ok
}

func toString(value any) (string, bool) {
	v, ok := value.(string)
	return v, ok
}

func toInt64(value any) (int64, bool) {
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case uint:
		if uint64(v) > math.MaxInt64 {
			return 0, false
		}
		return int64(v), true
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint64:
		if v > math.MaxInt64 {
			return 0, false
		}
		return int64(v), true
	case float32:
		return floatToInt64(float64(v))
	case float64:
		return floatToInt64(v)
	default:
		return 0, false
	}
}

func floatToInt64(value float64) (int64, bool) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, false
	}
	if math.Trunc(value) != value {
		return 0, false
	}
	if value > math.MaxInt64 || value < math.MinInt64 {
		return 0, false
	}
	return int64(value), true
}

func toUint64(value any) (uint64, bool) {
	switch v := value.(type) {
	case int:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case int8:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case int16:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case int32:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case int64:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case uint:
		return uint64(v), true
	case uint8:
		return uint64(v), true
	case uint16:
		return uint64(v), true
	case uint32:
		return uint64(v), true
	case uint64:
		return v, true
	case float32:
		return floatToUint64(float64(v))
	case float64:
		return floatToUint64(v)
	default:
		return 0, false
	}
}

func floatToUint64(value float64) (uint64, bool) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, false
	}
	if value < 0 {
		return 0, false
	}
	if math.Trunc(value) != value {
		return 0, false
	}
	if value > math.MaxUint64 {
		return 0, false
	}
	return uint64(value), true
}

func toFloat64(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	default:
		return 0, false
	}
}

func formatFloatLiteral(value float64) string {
	if math.IsNaN(value) {
		return "math.NaN()"
	}
	if math.IsInf(value, 1) {
		return "math.Inf(1)"
	}
	if math.IsInf(value, -1) {
		return "math.Inf(-1)"
	}
	text := strconv.FormatFloat(value, 'f', -1, 64)
	if !strings.ContainsAny(text, ".eE") {
		text += ".0"
	}
	return text
}

func formatAnySliceLiteral(items []any) (string, bool) {
	if items == nil {
		return "[]any{}", true
	}
	var builder strings.Builder
	builder.WriteString("[]any{")
	for i, item := range items {
		literal, ok := formatGoLiteral(item)
		if !ok {
			return "", false
		}
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(literal)
	}
	builder.WriteString("}")
	return builder.String(), true
}

func formatStringSliceLiteral(items []string) string {
	if items == nil {
		return "[]string{}"
	}
	var builder strings.Builder
	builder.WriteString("[]string{")
	for i, item := range items {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(strconv.Quote(item))
	}
	builder.WriteString("}")
	return builder.String()
}

func formatAnyMapLiteral(items map[string]any) (string, bool) {
	if items == nil {
		return "map[string]any{}", true
	}
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var builder strings.Builder
	builder.WriteString("map[string]any{")
	for i, key := range keys {
		value := items[key]
		literal, ok := formatGoLiteral(value)
		if !ok {
			return "", false
		}
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(strconv.Quote(key))
		builder.WriteString(": ")
		builder.WriteString(literal)
	}
	builder.WriteString("}")
	return builder.String(), true
}

func buildIsrDataListAssignment(name string, list insyra.IDataList) (string, bool) {
	if name == "" {
		return "", false
	}
	if list == nil {
		return fmt.Sprintf("%s = isr.DL.Of()", name), true
	}
	dataLiteral, ok := formatAnySliceLiteral(list.Data())
	if !ok {
		return "", false
	}
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%s = isr.DL.Of(%s...)", name, dataLiteral))
	if listName := list.GetName(); listName != "" {
		builder.WriteString("\n")
		builder.WriteString(fmt.Sprintf("%s.SetName(%s)", name, strconv.Quote(listName)))
	}
	return builder.String(), true
}

func buildIsrDataTableAssignment(name string, table insyra.IDataTable) (string, bool) {
	if name == "" {
		return "", false
	}
	if table == nil {
		return fmt.Sprintf("%s = isr.DT.Of(nil)", name), true
	}
	expr, ok := formatDataTableLiteralIsr(table)
	if !ok {
		return "", false
	}
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%s = %s", name, expr))
	cols := table.ColNames()
	if hasNonEmptyStrings(cols) {
		builder.WriteString("\n")
		builder.WriteString(fmt.Sprintf("%s.SetColNames(%s)", name, formatStringSliceLiteral(cols)))
	}
	index := table.RowNames()
	if hasNonEmptyStrings(index) {
		builder.WriteString("\n")
		builder.WriteString(fmt.Sprintf("%s.SetRowNames(%s)", name, formatStringSliceLiteral(index)))
	}
	return builder.String(), true
}

func formatDataListLiteral(list insyra.IDataList) (string, bool) {
	return formatDataListLiteralInsyra(list)
}

func formatDataListLiteralInsyra(list insyra.IDataList) (string, bool) {
	if list == nil {
		return "insyra.NewDataList()", true
	}
	dataLiteral, ok := formatAnySliceLiteral(list.Data())
	if !ok {
		return "", false
	}
	expr := fmt.Sprintf("insyra.NewDataList(%s...)", dataLiteral)
	if name := list.GetName(); name != "" {
		expr = fmt.Sprintf("%s.SetName(%s)", expr, strconv.Quote(name))
	}
	return expr, true
}

func formatDataTableLiteral(table insyra.IDataTable) (string, bool) {
	return formatDataTableLiteralInsyra(table)
}

func formatDataTableLiteralInsyra(table insyra.IDataTable) (string, bool) {
	if table == nil {
		return "insyra.NewDataTable()", true
	}
	rows := table.To2DSlice()
	var expr string
	if len(rows) == 0 || len(rows[0]) == 0 {
		expr = "insyra.NewDataTable()"
	} else {
		rowsLiteral, ok := format2DAnySliceLiteral(rows)
		if !ok {
			return "", false
		}
		expr = fmt.Sprintf("isr.DT.Of(%s).DataTable", rowsLiteral)
	}
	cols := table.ColNames()
	if hasNonEmptyStrings(cols) {
		expr = fmt.Sprintf("%s.SetColNames(%s)", expr, formatStringSliceLiteral(cols))
	}
	index := table.RowNames()
	if hasNonEmptyStrings(index) {
		expr = fmt.Sprintf("%s.SetRowNames(%s)", expr, formatStringSliceLiteral(index))
	}
	return expr, true
}

func formatDataTableLiteralIsr(table insyra.IDataTable) (string, bool) {
	if table == nil {
		return "isr.DT.Of(nil)", true
	}
	rows := table.To2DSlice()
	if len(rows) == 0 || len(rows[0]) == 0 {
		return "isr.DT.Of(nil)", true
	}
	rowsLiteral, ok := format2DAnySliceLiteral(rows)
	if !ok {
		return "", false
	}
	return fmt.Sprintf("isr.DT.Of(%s)", rowsLiteral), true
}

func format2DAnySliceLiteral(rows [][]any) (string, bool) {
	if rows == nil {
		return "[][]any{}", true
	}
	var builder strings.Builder
	builder.WriteString("[][]any{")
	for i, row := range rows {
		rowLiteral, ok := formatAnySliceLiteral(row)
		if !ok {
			return "", false
		}
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(rowLiteral)
	}
	builder.WriteString("}")
	return builder.String(), true
}
