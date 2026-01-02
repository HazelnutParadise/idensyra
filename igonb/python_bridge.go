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
	bindings := e.buildPythonBindings(shared)
	wrapper, err := buildPythonWrapper(code, bindings)
	if err != nil {
		return "", err
	}

	payload, output, err := e.executePythonWrapper(wrapper, bindings)
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

func buildPythonWrapper(code string, bindings []pythonBinding) (string, error) {
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

	wrapper := fmt.Sprintf(`
import ast, traceback, types

__igonb_code = %s
__igonb_globals = globals()
__igonb_injected = %s
__igonb_reserved = {
    "insyra",
    "insyra_return",
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
try:
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

insyra.Return({"vars": __igonb_vars, "error": __igonb_error}, None)
`, string(codeLiteral), injected)

	return wrapper, nil
}

func (e *Executor) executePythonWrapper(wrapper string, bindings []pythonBinding) (pythonRunPayload, string, error) {
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
		argList = ", " + strings.Join(args, ", ")
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
	literal, ok := formatGoLiteral(value)
	if !ok {
		return fmt.Errorf("unsupported value for %s", name)
	}
	if strings.Contains(literal, "math.") {
		_ = e.PreloadGoImports([]string{"math"})
	}
	if strings.Contains(literal, "insyra.") {
		_ = e.PreloadGoImports([]string{"github.com/HazelnutParadise/insyra"})
	}
	if strings.Contains(literal, "isr.") {
		_ = e.PreloadGoImports([]string{"github.com/HazelnutParadise/insyra/isr"})
	}

	if e.goNameExists(name) {
		if _, err := e.runGoSegment(fmt.Sprintf("%s = %s", name, literal), false); err == nil {
			return nil
		} else {
			return err
		}
	}
	if _, err := e.runGoSegment(fmt.Sprintf("var %s = %s", name, literal), false); err != nil {
		return err
	}
	return nil
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

func formatDataListLiteral(list insyra.IDataList) (string, bool) {
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
	if table == nil {
		return "isr.DT.Of(nil)", true
	}
	rows := table.To2DSlice()
	rowsLiteral, ok := format2DAnySliceLiteral(rows)
	if !ok {
		return "", false
	}
	expr := fmt.Sprintf("isr.DT.Of(%s)", rowsLiteral)
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
