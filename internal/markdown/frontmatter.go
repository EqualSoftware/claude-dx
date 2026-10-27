// Package markdown reads and changes the YAML frontmatter and the sections of dx documents.
package markdown

import (
	"regexp"
	"strings"

	"go.yaml.in/yaml/v3"
)

const delimiter = "---\n"

// Split returns the frontmatter (with no delimiters) and the body of src.
// A document with no frontmatter returns an empty frontmatter and src as the body.
func Split(src string) (front, body string) {
	src = strings.ReplaceAll(src, "\r\n", "\n")
	if !strings.HasPrefix(src, delimiter) {
		return "", src
	}
	rest := src[len(delimiter):]
	end := strings.Index(rest, "\n"+delimiter)
	if end < 0 {
		return "", src
	}
	return rest[:end+1], rest[end+1+len(delimiter):]
}

// Join puts the frontmatter and the body together again.
func Join(front, body string) string {
	if front == "" {
		return body
	}
	return delimiter + front + delimiter + body
}

// Decode reads the frontmatter into v (a pointer to a struct with yaml tags).
func Decode(front string, v any) error {
	return yaml.Unmarshal([]byte(front), v)
}

// SetField sets key to value (a string or a []string). It replaces the line of the key and its
// continuation lines (lines that start with a space or "-", as in a block list or a block scalar).
// Empty lines are replaced only when a continuation line comes after them.
// If the key does not exist, SetField adds it at the end.
func SetField(front, key string, value any) string {
	line := key + ": " + format(value)
	pattern := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(key) + `:.*(?:\n+[ -].*)*`)
	m := pattern.FindStringIndex(front)
	if m == nil {
		return front + line + "\n"
	}
	return front[:m[0]] + line + front[m[1]:]
}

// format writes a string as a YAML scalar and a []string as a flow list. Both are quoted only if necessary.
func format(value any) string {
	var node yaml.Node
	_ = node.Encode(value) // A string or a []string always encodes.
	node.Style = yaml.FlowStyle
	out, _ := yaml.Marshal(&node)
	return strings.TrimSuffix(string(out), "\n")
}
