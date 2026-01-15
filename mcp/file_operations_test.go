package mcp

import (
	"context"
	"fmt"
	"testing"
)

func TestCreateFileRejectsInvalidPath(t *testing.T) {
	fo := NewFileOperations(DefaultConfig(), ".", nil, nil, nil, nil, nil, nil, nil, nil)
	resp, err := fo.CreateFile(context.Background(), "../outside.txt", "x")
	if err == nil {
		t.Fatalf("expected error for invalid path")
	}
	if !resp.IsError {
		t.Fatalf("expected ToolResponse.IsError true")
	}
}

func TestReadRequiresBackend(t *testing.T) {
	fo := NewFileOperations(DefaultConfig(), ".", nil, nil, nil, nil, nil, nil, nil, nil)
	resp, err := fo.ReadFile(context.Background(), "sub/ok.txt")
	if err == nil {
		t.Fatalf("expected error when read backend missing")
	}
	if !resp.IsError {
		t.Fatalf("expected ToolResponse.IsError true")
	}
}

func TestWriteRequiresBackend(t *testing.T) {
	fo := NewFileOperations(DefaultConfig(), ".", nil, nil, nil, nil, nil, nil, nil, nil)
	resp, err := fo.WriteFile(context.Background(), "sub/ok.txt", "x")
	if err == nil {
		t.Fatalf("expected error when write backend missing")
	}
	if !resp.IsError {
		t.Fatalf("expected ToolResponse.IsError true")
	}
}

func TestDeleteRequiresBackend(t *testing.T) {
	fo := NewFileOperations(DefaultConfig(), ".", nil, nil, nil, nil, nil, nil, nil, nil)
	resp, err := fo.DeleteFile(context.Background(), "sub/ok.txt")
	if err == nil {
		t.Fatalf("expected error when delete backend missing")
	}
	if !resp.IsError {
		t.Fatalf("expected ToolResponse.IsError true")
	}
}

func TestRenameRequiresBackend(t *testing.T) {
	fo := NewFileOperations(DefaultConfig(), ".", nil, nil, nil, nil, nil, nil, nil, nil)
	resp, err := fo.RenameFile(context.Background(), "a.txt", "b.txt")
	if err == nil {
		t.Fatalf("expected error when rename backend missing")
	}
	if !resp.IsError {
		t.Fatalf("expected ToolResponse.IsError true")
	}
}

func TestListRequiresBackend(t *testing.T) {
	fo := NewFileOperations(DefaultConfig(), ".", nil, nil, nil, nil, nil, nil, nil, nil)
	resp, err := fo.ListFiles(context.Background(), "")
	if err == nil {
		t.Fatalf("expected error when list backend missing")
	}
	if !resp.IsError {
		t.Fatalf("expected ToolResponse.IsError true")
	}
}

func TestCreateFileUsesBackend(t *testing.T) {
	called := false
	createFunc := func(path string, content string) error {
		called = true
		if path != "sub/ok.txt" {
			return fmt.Errorf("unexpected path: %s", path)
		}
		return nil
	}
	fo := NewFileOperations(DefaultConfig(), ".", nil, nil, nil, nil, createFunc, nil, nil, nil)
	resp, err := fo.CreateFile(context.Background(), "sub/ok.txt", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.IsError {
		t.Fatalf("unexpected ToolResponse.IsError")
	}
	if !called {
		t.Fatalf("backend createFunc was not called")
	}
}

func TestCreateFileRequiresBackend(t *testing.T) {
	fo := NewFileOperations(DefaultConfig(), ".", nil, nil, nil, nil, nil, nil, nil, nil)
	resp, err := fo.CreateFile(context.Background(), "sub/ok.txt", "x")
	if err == nil {
		t.Fatalf("expected error when create backend missing")
	}
	if !resp.IsError {
		t.Fatalf("expected ToolResponse.IsError true")
	}
}
