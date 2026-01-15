package mcp

import (
	"context"
	"fmt"
	"testing"
)

func TestExecuteGoFileRequiresReadBackend(t *testing.T) {
	ce := NewCodeExecution(DefaultConfig(), ".", nil, func(c, bg string) string { return "ok" }, nil, nil, nil, nil)
	_, err := ce.ExecuteGoFile(context.Background(), "a.go")
	if err == nil {
		t.Fatalf("expected error when read backend missing")
	}
}

func TestExecuteGoFileUsesReadBackend(t *testing.T) {
	called := false
	read := func(path string) (string, error) {
		called = true
		return "package main\nfunc main(){println(\"hi\")}", nil
	}
	execGo := func(c, bg string) string { return "ran" }
	ce := NewCodeExecution(DefaultConfig(), ".", nil, execGo, nil, nil, read, nil)
	res, err := ce.ExecuteGoFile(context.Background(), "a.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("read backend not called")
	}
	if res.Content[0].Text != "ran" {
		t.Fatalf("unexpected result: %s", res.Content[0].Text)
	}
}

func TestExecutePythonFileUsesContentCallback(t *testing.T) {
	read := func(path string) (string, error) { return "print(1)", nil }
	execPyContent := func(filename, content string) (string, error) { return "ok", nil }
	ce := NewCodeExecution(DefaultConfig(), ".", nil, nil, nil, execPyContent, read, nil)
	res, err := ce.ExecutePythonFile(context.Background(), "a.py")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Content[0].Text != "ok" {
		t.Fatalf("unexpected result: %s", res.Content[0].Text)
	}
}

func TestExecutePythonCodePrefersContentCallback(t *testing.T) {
	execPyContent := func(filename, content string) (string, error) { return fmt.Sprintf("ran:%s", content), nil }
	ce := NewCodeExecution(DefaultConfig(), ".", nil, nil, nil, execPyContent, nil, nil)
	res, err := ce.ExecutePythonCode(context.Background(), "print(2)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Content[0].Text != "ran:print(2)" {
		t.Fatalf("unexpected result: %s", res.Content[0].Text)
	}
}
