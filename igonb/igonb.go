package igonb

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const CurrentVersion = 1

type Notebook struct {
	Version  int            `json:"version"`
	Cells    []Cell         `json:"cells"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type Cell struct {
	ID       string `json:"id,omitempty"`
	Language string `json:"language"`
	Source   string `json:"source"`
}

func Parse(data []byte) (*Notebook, error) {
	if len(data) == 0 {
		return nil, errors.New("empty igonb content")
	}

	var nb Notebook
	if err := json.Unmarshal(data, &nb); err != nil {
		return nil, fmt.Errorf("invalid igonb json: %w", err)
	}

	if nb.Version == 0 {
		nb.Version = CurrentVersion
	}
	normalizeNotebook(&nb)

	if err := nb.Validate(); err != nil {
		return nil, err
	}
	return &nb, nil
}

func (n *Notebook) Validate() error {
	if n.Version <= 0 {
		return fmt.Errorf("invalid igonb version: %d", n.Version)
	}
	if len(n.Cells) == 0 {
		return errors.New("igonb must contain at least one cell")
	}

	for idx, cell := range n.Cells {
		lang := NormalizeLanguage(cell.Language)
		if lang == "" {
			return fmt.Errorf("cell %d has empty language", idx+1)
		}
		if lang != "go" && lang != "python" && lang != "markdown" {
			return fmt.Errorf("cell %d has unsupported language: %s", idx+1, cell.Language)
		}
	}
	return nil
}

func NormalizeLanguage(input string) string {
	lang := strings.ToLower(strings.TrimSpace(input))
	switch lang {
	case "py":
		return "python"
	case "md":
		return "markdown"
	default:
		return lang
	}
}

func normalizeNotebook(n *Notebook) {
	for i := range n.Cells {
		n.Cells[i].Language = NormalizeLanguage(n.Cells[i].Language)
	}
}

func NewNotebook() *Notebook {
	return &Notebook{
		Version: CurrentVersion,
		Cells: []Cell{
			{
				Language: "go",
				Source:   "",
			},
		},
	}
}

func DefaultNotebookJSON() (string, error) {
	nb := NewNotebook()
	data, err := json.MarshalIndent(nb, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
