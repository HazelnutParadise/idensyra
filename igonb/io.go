package igonb

import (
	"encoding/json"
	"fmt"
	"os"
)

func ReadFile(path string) (*Notebook, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read igonb file: %w", err)
	}
	return Parse(data)
}

func WriteFile(path string, nb *Notebook) error {
	if nb == nil {
		return fmt.Errorf("igonb notebook is nil")
	}
	if err := nb.Validate(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(nb, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal igonb: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write igonb file: %w", err)
	}
	return nil
}
