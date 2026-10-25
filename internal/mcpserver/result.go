package mcpserver

import (
	"fmt"
	"strings"

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
