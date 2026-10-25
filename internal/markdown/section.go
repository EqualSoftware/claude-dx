package markdown

import (
	"fmt"
	"regexp"
	"strings"
)

// headingPattern matches "## Title", "## 3. Title" and "### 7.1 Title".
var headingPattern = regexp.MustCompile(`^(#{2,3}) +(?:(\d+(?:\.\d+)*)\.? +)?(.+?) *$`)

type heading struct {
	level  int
	line   int
	number string
	title  string
}

// Section returns the content of the section that ref names, with no heading.
// ref is a number ("3", "7.1") or a title ("Goals"), with no case difference.
// The content includes the subsections.
func Section(body, ref string) (string, bool) {
	lines := strings.Split(body, "\n")
	start, end, ok := bounds(lines, ref)
	if !ok {
		return "", false
	}
	return strings.Trim(strings.Join(lines[start:end], "\n"), "\n"), true
}

// ReplaceSection replaces the content of the section that ref names. The heading stays.
func ReplaceSection(body, ref, content string) (string, error) {
	lines := strings.Split(body, "\n")
	start, end, ok := bounds(lines, ref)
	if !ok {
		return "", fmt.Errorf("section %q does not exist", ref)
	}
	middle := []string{"", strings.Trim(content, "\n")}
	if end < len(lines) {
		middle = append(middle, "")
	}
	out := append(append(append([]string{}, lines[:start]...), middle...), lines[end:]...)
	return strings.Join(out, "\n"), nil
}

// bounds returns the first and the end line of the content of the section ref.
func bounds(lines []string, ref string) (start, end int, ok bool) {
	hs := headings(lines)
	for i, h := range hs {
		if h.number != ref && !strings.EqualFold(h.title, ref) {
			continue
		}
		end = len(lines)
		for _, next := range hs[i+1:] {
			if next.level <= h.level {
				end = next.line
				break
			}
		}
		return h.line + 1, end, true
	}
	return 0, 0, false
}

// headings returns the "##" and "###" headings that are not in a code block.
func headings(lines []string) []heading {
	var hs []heading
	fenced := false
	for i, line := range lines {
		if strings.HasPrefix(line, "```") {
			fenced = !fenced
			continue
		}
		m := headingPattern.FindStringSubmatch(line)
		if fenced || m == nil {
			continue
		}
		hs = append(hs, heading{level: len(m[1]), line: i, number: m[2], title: m[3]})
	}
	return hs
}
