package igonb

type OutputFormatter func(output string) string

func FormatResult(result CellResult, format OutputFormatter) CellResult {
	if format == nil {
		return result
	}
	if result.Language == "markdown" || result.Output == "" {
		return result
	}
	result.Output = format(result.Output)
	return result
}

func FormatResults(results []CellResult, format OutputFormatter) []CellResult {
	if format == nil || len(results) == 0 {
		return results
	}
	formatted := make([]CellResult, len(results))
	for i, result := range results {
		formatted[i] = FormatResult(result, format)
	}
	return formatted
}
