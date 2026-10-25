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

// SetField sets key to value (a string or a []string) and keeps a comment at the end of the line.
// If the key does not exist, SetField adds it at the end.
func SetField(front, key string, value any) string {
	line := key + ": " + format(value)
	pattern := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(key) + `:(.*)$`)
	m := pattern.FindStringSubmatchIndex(front)
	if m == nil {
		return front + line + "\n"
	}
	return front[:m[0]] + line + comment(front[m[2]:m[3]]) + front[m[1]:]
}

// format writes a string as a YAML scalar (quoted only if necessary) and a []string as a flow list.
func format(value any) string {
	if list, ok := value.([]string); ok {
		return "[" + strings.Join(list, ", ") + "]"
	}
	out, _ := yaml.Marshal(value)
	return strings.TrimSuffix(string(out), "\n")
}

// comment returns the " # ..." part at the end of a YAML value, or "".
func comment(value string) string {
	start := 0
	trimmed := strings.TrimLeft(value, " ")
	if q := trimmed[:min(1, len(trimmed))]; q == `"` || q == "'" {
		offset := len(value) - len(trimmed)
		start = offset + 1 + strings.Index(trimmed[1:], q) + 1
	}
	i := strings.Index(value[start:], " #")
	if i < 0 {
		return ""
	}
	j := start + i
	for j > 0 && value[j-1] == ' ' {
		j--
	}
	return value[j:]
}
