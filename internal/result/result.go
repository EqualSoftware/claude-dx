// Package result makes compact tool results for the model (ADR-005).
package result

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Text returns a tool result with one text part.
func Text(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}}
}

// Errorf returns a tool result that tells the model that the tool failed.
// Write the message in ASD-STE100.
func Errorf(format string, args ...any) *mcp.CallToolResult {
	result := Text(fmt.Sprintf(format, args...))
	result.IsError = true
	return result
}

// Table aligns rows in columns, with two spaces between columns.
// It is compact for the model: no borders, no header, no padding after the last column.
func Table(rows [][]string) string {
	widths := map[int]int{}
	for _, row := range rows {
		for i, cell := range row {
			widths[i] = max(widths[i], len(cell))
		}
	}
	var b strings.Builder
	for _, row := range rows {
		for i, cell := range row {
			if i == len(row)-1 {
				b.WriteString(cell)
				break
			}
			b.WriteString(cell + strings.Repeat(" ", widths[i]-len(cell)+2))
		}
		b.WriteString("\n")
	}
	return b.String()
}

// Sentence makes a Go error text into a sentence for the model: a capital letter and a period.
func Sentence(err error) string {
	text := strings.TrimSuffix(err.Error(), ".")
	r, size := utf8.DecodeRuneInString(text)
	return string(unicode.ToUpper(r)) + text[size:] + "."
}

// Reply makes the tool result for text and err. An error that is normal (example: a docs
// directory that does not exist yet) gives a normal message, not a tool error.
func Reply(text string, err, normal error) (*mcp.CallToolResult, any, error) {
	if errors.Is(err, normal) {
		return Text(Sentence(err)), nil, nil
	}
	if err != nil {
		return Errorf("%s", Sentence(err)), nil, nil
	}
	return Text(text), nil, nil
}
