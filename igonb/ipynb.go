package igonb

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// IPyNBNotebook represents a Jupyter Notebook structure
type IPyNBNotebook struct {
	Cells    []IPyNBCell   `json:"cells"`
	Metadata IPyNBMetadata `json:"metadata"`
	NBFormat int           `json:"nbformat"`
	Minor    int           `json:"nbformat_minor"`
}

// IPyNBCell represents a cell in Jupyter Notebook
type IPyNBCell struct {
	CellType       string        `json:"cell_type"`
	Source         interface{}   `json:"source"` // Can be string or []string
	Outputs        []IPyNBOutput `json:"outputs,omitempty"`
	ExecutionCount *int          `json:"execution_count,omitempty"`
	Metadata       interface{}   `json:"metadata,omitempty"`
}

// IPyNBOutput represents output from a Jupyter cell
type IPyNBOutput struct {
	OutputType string      `json:"output_type"`
	Text       interface{} `json:"text,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Name       string      `json:"name,omitempty"`
	EName      string      `json:"ename,omitempty"`
	EValue     string      `json:"evalue,omitempty"`
	Traceback  []string    `json:"traceback,omitempty"`
}

// IPyNBMetadata contains Jupyter notebook metadata
type IPyNBMetadata struct {
	KernelSpec   *IPyNBKernelSpec `json:"kernelspec,omitempty"`
	LanguageInfo *IPyNBLangInfo   `json:"language_info,omitempty"`
}

// IPyNBKernelSpec contains kernel specification
type IPyNBKernelSpec struct {
	DisplayName string `json:"display_name"`
	Language    string `json:"language"`
	Name        string `json:"name"`
}

// IPyNBLangInfo contains language information
type IPyNBLangInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// ReadIPyNBFile reads a .ipynb file and converts it to igonb Notebook format
func ReadIPyNBFile(path string) (*Notebook, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read ipynb file: %w", err)
	}
	return ParseIPyNB(data)
}

// ParseIPyNB converts Jupyter notebook JSON to igonb Notebook format
func ParseIPyNB(data []byte) (*Notebook, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty ipynb content")
	}

	var ipynb IPyNBNotebook
	if err := json.Unmarshal(data, &ipynb); err != nil {
		return nil, fmt.Errorf("invalid ipynb json: %w", err)
	}

	// Detect the language from kernel spec
	defaultLang := detectIPyNBLanguage(&ipynb)

	nb := &Notebook{
		Version:  CurrentVersion,
		Cells:    make([]Cell, 0, len(ipynb.Cells)),
		Metadata: make(map[string]any),
	}

	// Store original ipynb metadata
	nb.Metadata["ipynb_source"] = true
	if ipynb.NBFormat > 0 {
		nb.Metadata["ipynb_format"] = ipynb.NBFormat
	}

	for _, ipyCell := range ipynb.Cells {
		cell := convertIPyNBCell(ipyCell, defaultLang)
		if cell != nil {
			nb.Cells = append(nb.Cells, *cell)
		}
	}

	// Ensure at least one cell
	if len(nb.Cells) == 0 {
		nb.Cells = append(nb.Cells, Cell{
			Language: defaultLang,
			Source:   "",
		})
	}

	return nb, nil
}

// detectIPyNBLanguage determines the notebook language from metadata
func detectIPyNBLanguage(ipynb *IPyNBNotebook) string {
	if ipynb.Metadata.KernelSpec != nil {
		lang := strings.ToLower(ipynb.Metadata.KernelSpec.Language)
		if lang == "python" || strings.HasPrefix(lang, "python") {
			return "python"
		}
		if lang == "go" || lang == "golang" {
			return "go"
		}
	}

	if ipynb.Metadata.LanguageInfo != nil {
		lang := strings.ToLower(ipynb.Metadata.LanguageInfo.Name)
		if lang == "python" || strings.HasPrefix(lang, "python") {
			return "python"
		}
		if lang == "go" || lang == "golang" {
			return "go"
		}
	}

	// Default to Python as most .ipynb files are Python notebooks
	return "python"
}

// convertIPyNBCell converts a Jupyter cell to igonb Cell
func convertIPyNBCell(ipyCell IPyNBCell, defaultLang string) *Cell {
	source := extractIPyNBSource(ipyCell.Source)

	var cell Cell

	switch ipyCell.CellType {
	case "code":
		cell = Cell{
			Language: defaultLang,
			Source:   source,
			Output:   extractIPyNBOutputs(ipyCell.Outputs),
			Error:    extractIPyNBError(ipyCell.Outputs),
		}
	case "markdown":
		cell = Cell{
			Language: "markdown",
			Source:   source,
		}
	case "raw":
		// Convert raw cells to markdown
		cell = Cell{
			Language: "markdown",
			Source:   source,
		}
	default:
		// Skip unknown cell types
		return nil
	}

	return &cell
}

// extractIPyNBSource extracts source text from cell
func extractIPyNBSource(source interface{}) string {
	switch s := source.(type) {
	case string:
		return s
	case []interface{}:
		var lines []string
		for _, line := range s {
			if str, ok := line.(string); ok {
				lines = append(lines, str)
			}
		}
		return strings.Join(lines, "")
	default:
		return ""
	}
}

// extractIPyNBOutputs extracts output text from cell outputs
func extractIPyNBOutputs(outputs []IPyNBOutput) string {
	var parts []string

	for _, out := range outputs {
		switch out.OutputType {
		case "stream":
			text := extractOutputText(out.Text)
			if text != "" {
				parts = append(parts, text)
			}
		case "execute_result", "display_data":
			// Try to get text/plain from data
			if data, ok := out.Data.(map[string]interface{}); ok {
				if textPlain, ok := data["text/plain"]; ok {
					text := extractOutputText(textPlain)
					if text != "" {
						parts = append(parts, text)
					}
				}
			}
		}
	}

	return strings.TrimRight(strings.Join(parts, ""), "\n")
}

// extractIPyNBError extracts error information from cell outputs
func extractIPyNBError(outputs []IPyNBOutput) string {
	for _, out := range outputs {
		if out.OutputType == "error" {
			if len(out.Traceback) > 0 {
				// Clean ANSI escape codes from traceback
				var cleanLines []string
				for _, line := range out.Traceback {
					cleanLines = append(cleanLines, stripANSI(line))
				}
				return strings.Join(cleanLines, "\n")
			}
			if out.EName != "" {
				return fmt.Sprintf("%s: %s", out.EName, out.EValue)
			}
		}
	}
	return ""
}

// extractOutputText handles both string and []string output formats
func extractOutputText(text interface{}) string {
	switch t := text.(type) {
	case string:
		return t
	case []interface{}:
		var lines []string
		for _, line := range t {
			if str, ok := line.(string); ok {
				lines = append(lines, str)
			}
		}
		return strings.Join(lines, "")
	default:
		return ""
	}
}

// stripANSI removes ANSI escape codes from a string
func stripANSI(s string) string {
	var result strings.Builder
	inEscape := false

	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (s[i] >= 'a' && s[i] <= 'z') || (s[i] >= 'A' && s[i] <= 'Z') {
				inEscape = false
			}
			continue
		}
		result.WriteByte(s[i])
	}

	return result.String()
}

// ConvertIPyNBToIgonb reads an .ipynb file and saves it as .igonb
func ConvertIPyNBToIgonb(ipynbPath string) (string, error) {
	nb, err := ReadIPyNBFile(ipynbPath)
	if err != nil {
		return "", err
	}

	// Generate output path by replacing extension
	igonbPath := strings.TrimSuffix(ipynbPath, filepath.Ext(ipynbPath)) + ".igonb"

	if err := WriteFile(igonbPath, nb); err != nil {
		return "", err
	}

	return igonbPath, nil
}

// IPyNBToIgonbJSON converts ipynb content to igonb JSON string
func IPyNBToIgonbJSON(ipynbData []byte) (string, error) {
	nb, err := ParseIPyNB(ipynbData)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(nb, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal igonb: %w", err)
	}

	return string(data), nil
}

// IgonbToIPyNBJSON converts igonb JSON content back to ipynb JSON format
func IgonbToIPyNBJSON(igonbData []byte) (string, error) {
	nb, err := Parse(igonbData)
	if err != nil {
		return "", fmt.Errorf("failed to parse igonb: %w", err)
	}

	// Detect language from cells (default to Python for ipynb)
	language := "python"
	for _, cell := range nb.Cells {
		if cell.Language == "go" {
			language = "go"
			break
		}
		if cell.Language == "python" {
			language = "python"
			break
		}
	}

	// Create ipynb notebook structure
	ipynb := IPyNBNotebook{
		Cells:    make([]IPyNBCell, 0, len(nb.Cells)),
		NBFormat: 4,
		Minor:    5,
		Metadata: IPyNBMetadata{
			KernelSpec: &IPyNBKernelSpec{
				DisplayName: strings.Title(language),
				Language:    language,
				Name:        language,
			},
			LanguageInfo: &IPyNBLangInfo{
				Name: language,
			},
		},
	}

	// Convert cells
	execCount := 1
	for _, cell := range nb.Cells {
		ipyCell := convertCellToIPyNB(cell, &execCount)
		ipynb.Cells = append(ipynb.Cells, ipyCell)
	}

	data, err := json.MarshalIndent(ipynb, "", " ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal ipynb: %w", err)
	}

	return string(data), nil
}

// convertCellToIPyNB converts an igonb Cell to an IPyNB cell
func convertCellToIPyNB(cell Cell, execCount *int) IPyNBCell {
	// Split source into lines for ipynb format
	sourceLines := splitSourceToLines(cell.Source)

	if cell.Language == "markdown" {
		return IPyNBCell{
			CellType: "markdown",
			Source:   sourceLines,
			Metadata: map[string]interface{}{},
		}
	}

	// Code cell
	ipyCell := IPyNBCell{
		CellType:       "code",
		Source:         sourceLines,
		Metadata:       map[string]interface{}{},
		ExecutionCount: execCount,
		Outputs:        make([]IPyNBOutput, 0),
	}
	*execCount++

	// Convert output
	if cell.Output != "" {
		outputLines := splitSourceToLines(cell.Output)
		ipyCell.Outputs = append(ipyCell.Outputs, IPyNBOutput{
			OutputType: "stream",
			Name:       "stdout",
			Text:       outputLines,
		})
	}

	// Convert error
	if cell.Error != "" {
		errorLines := strings.Split(cell.Error, "\n")
		ipyCell.Outputs = append(ipyCell.Outputs, IPyNBOutput{
			OutputType: "error",
			EName:      "Error",
			EValue:     cell.Error,
			Traceback:  errorLines,
		})
	}

	return ipyCell
}

// splitSourceToLines splits source text into lines array for ipynb format
func splitSourceToLines(source string) []string {
	if source == "" {
		return []string{}
	}

	lines := strings.Split(source, "\n")
	result := make([]string, len(lines))

	for i, line := range lines {
		// Add newline back except for the last line
		if i < len(lines)-1 {
			result[i] = line + "\n"
		} else {
			result[i] = line
		}
	}

	return result
}
