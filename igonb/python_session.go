package igonb

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/HazelnutParadise/insyra/py"
)

type PythonSession struct {
	mu     sync.Mutex
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	closed bool
}

type pythonSessionRequest struct {
	Code string `json:"code"`
}

type pythonSessionResponse struct {
	Output string `json:"output"`
	Error  string `json:"error"`
}

var pythonBaseDir = func() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return wd
}()

var pythonDefaultImports = []string{
	"import requests",
	"import json",
	"import numpy as np",
	"import pandas as pd",
	"import polars as pl",
	"import matplotlib.pyplot as plt",
	"import seaborn as sns",
	"import scipy",
	"import sklearn",
	"import statsmodels.api as sm",
	"import plotly.graph_objects as go",
	"import spacy",
	"import bs4",
}

func NewPythonSession() (*PythonSession, error) {
	pythonPath, err := pythonExecutablePath()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(pythonPath, "-u", "-c", pythonSessionScript())
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &PythonSession{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReader(stdout),
	}, nil
}

func (s *PythonSession) Run(code string) (string, error) {
	if s == nil {
		return "", fmt.Errorf("python session not initialized")
	}
	if strings.TrimSpace(code) == "" {
		return "", nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return "", fmt.Errorf("python session closed")
	}

	payload, err := json.Marshal(pythonSessionRequest{Code: code})
	if err != nil {
		return "", err
	}
	if _, err := s.stdin.Write(append(payload, '\n')); err != nil {
		return "", err
	}

	line, err := s.stdout.ReadBytes('\n')
	if err != nil {
		return "", err
	}
	trimmed := strings.TrimSpace(string(line))
	if trimmed == "" {
		return "", fmt.Errorf("empty python response")
	}

	var resp pythonSessionResponse
	if err := json.Unmarshal([]byte(trimmed), &resp); err != nil {
		return "", err
	}
	if resp.Error != "" {
		return resp.Output, fmt.Errorf(resp.Error)
	}
	return resp.Output, nil
}

func (s *PythonSession) Close() error {
	if s == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	if s.stdin != nil {
		_ = s.stdin.Close()
	}
	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
		_, _ = s.cmd.Process.Wait()
	}
	return nil
}

func pythonExecutablePath() (string, error) {
	baseDir := pythonBaseDir
	if baseDir == "" {
		if wd, err := os.Getwd(); err == nil {
			baseDir = wd
		}
	}
	if baseDir == "" {
		return "", fmt.Errorf("unable to determine working directory")
	}

	installDir := filepath.Join(
		baseDir,
		".insyra_env",
		fmt.Sprintf("py25c_%s_%s", runtime.GOOS, runtime.GOARCH),
	)
	var pythonPath string
	if runtime.GOOS == "windows" {
		pythonPath = filepath.Join(installDir, ".venv", "Scripts", "python.exe")
	} else {
		pythonPath = filepath.Join(installDir, ".venv", "bin", "python")
	}

	if _, err := os.Stat(pythonPath); err == nil {
		return pythonPath, nil
	}

	if err := py.RunCode(nil, "pass"); err != nil {
		return "", err
	}
	if _, err := os.Stat(pythonPath); err == nil {
		return pythonPath, nil
	}
	return "", fmt.Errorf("python executable not found: %s", pythonPath)
}

func pythonSessionScript() string {
	imports := strings.Join(pythonDefaultImports, "\n")
	return fmt.Sprintf(`%s
import sys, json, traceback, ast, io, contextlib

_globals = globals()

def _exec_cell(_code):
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

def _run(code):
    _buf_out = io.StringIO()
    _buf_err = io.StringIO()
    try:
        with contextlib.redirect_stdout(_buf_out), contextlib.redirect_stderr(_buf_err):
            _value = _exec_cell(code)
            if _value is not None:
                print(_value)
        return {"output": _buf_out.getvalue() + _buf_err.getvalue(), "error": ""}
    except Exception:
        return {"output": _buf_out.getvalue() + _buf_err.getvalue(), "error": traceback.format_exc()}

for _line in sys.stdin:
    if not _line:
        break
    try:
        _req = json.loads(_line)
    except Exception:
        continue
    _resp = _run(_req.get("code", ""))
    sys.stdout.write(json.dumps(_resp) + "\n")
    sys.stdout.flush()
`, imports)
}
