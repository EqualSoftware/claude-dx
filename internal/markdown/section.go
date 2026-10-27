package markdown

import (
	"fmt"
	"regexp"
	"strings"
)

// headingPattern matches "## Title", "## 3. Title" and "### 7.1 Title".
var headingPattern = regexp.MustCompile(`^(#{1,6}) +(?:(\d+(?:\.\d+)*)\.? +)?(.+?) *$`)

// fencePattern matches the start of a code fence: up to 3 spaces, then 3 or more backticks or tildes.
var fencePattern = regexp.MustCompile("^ {0,3}(`{3,}|~{3,})(.*)$")

type heading struct {
	level  int
	line   int
	number string
	title  string
}

// Section returns the content of the section that ref names, with no heading.
// ref is a number ("3", "7.1") or a title ("Goals"), with no case difference.
// Only a heading of level 2 matches, and a dotted number also matches a heading of level 3.
// The content includes the subsections. The result has no CR.
func Section(body, ref string) (string, bool) {
	lines := strings.Split(lf(body), "\n")
	start, end, ok := bounds(lines, ref)
	if !ok {
		return "", false
	}
	return strings.Trim(strings.Join(lines[start:end], "\n"), "\n"), true
}

// SectionNumber returns the number of the section that ref names. A section with no number gives "".
func SectionNumber(body, ref string) (string, bool) {
	lines := strings.Split(lf(body), "\n")
	start, _, ok := bounds(lines, ref)
	if !ok {
		return "", false
	}
	return headingPattern.FindStringSubmatch(lines[start-1])[2], true
}

// ReplaceSection replaces the content of the section that ref names. The heading stays.
// It refuses content with a heading of the same or a higher level, because that heading starts a new section.
// It also refuses content with an unclosed code fence, and content with a heading that has the title or the
// number of a section of level 2 of body, because the next read would find the wrong section.
// A body with CRLF line ends keeps them.
func ReplaceSection(body, ref, content string) (string, error) {
	crlf := strings.Contains(body, "\r\n")
	lines := strings.Split(lf(body), "\n")
	start, end, ok := bounds(lines, ref)
	if !ok {
		return "", fmt.Errorf("section %q does not exist", ref)
	}
	content = lf(content)
	level := len(headingPattern.FindStringSubmatch(lines[start-1])[1])
	if err := CheckContent(strings.Join(lines, "\n"), content, ref, level); err != nil {
		return "", err
	}
	middle := []string{"", strings.Trim(content, "\n")}
	if end < len(lines) {
		middle = append(middle, "")
	}
	out := strings.Join(append(append(append([]string{}, lines[:start]...), middle...), lines[end:]...), "\n")
	if crlf {
		out = strings.ReplaceAll(out, "\n", "\r\n")
	}
	return out, nil
}

// CheckContent refuses text that would change the structure of body when it is the content of a section
// of level `level`: an unclosed code fence, a heading of level `level` or higher, and a heading that has the
// title or number of a section of level 2 of body. name is the section name for the message. Pass level 2
// to check text for a new document.
func CheckContent(body, content, name string, level int) error {
	lines := strings.Split(lf(content), "\n")
	if unclosedFence(lines) {
		return fmt.Errorf("the content of section %q has a code fence that is not closed. Close it with the same fence", name)
	}
	hs := headings(lines)
	for _, h := range hs {
		if h.level <= level {
			return fmt.Errorf("the content of section %q has a heading of level %d; use only headings of level %d or more", name, h.level, level+1)
		}
	}
	for _, h := range headings(strings.Split(lf(body), "\n")) {
		if h.level != 2 {
			continue
		}
		for _, c := range hs {
			if strings.EqualFold(c.title, h.title) || (c.number != "" && c.number == h.number) {
				return fmt.Errorf("the content of section %q has the heading %q, and the document has a section with the same title or number. Rename the heading", name, c.title)
			}
		}
	}
	return nil
}

// CheckNew refuses text for a section of level `level` of a new document: an unclosed code fence and a
// heading of level `level` or higher. name is the field name for the message.
func CheckNew(content, name string, level int) error {
	return CheckContent("", content, name, level)
}

// CheckFence refuses text with a code fence that is not closed. name is the field name for the message.
func CheckFence(content, name string) error {
	if unclosedFence(strings.Split(lf(content), "\n")) {
		return fmt.Errorf("%s has a code fence that is not closed. Close it with the same fence", name)
	}
	return nil
}

// lf removes the CR of CRLF line ends.
func lf(text string) string {
	return strings.ReplaceAll(text, "\r\n", "\n")
}

// bounds returns the first and the end line of the content of the section ref. A heading of level 2
// wins. Only if there is none, a heading of level 3 matches by title or by a dotted number ("7.1").
func bounds(lines []string, ref string) (start, end int, ok bool) {
	hs := headings(lines)
	for _, level := range []int{2, 3} {
		for i, h := range hs {
			if h.level != level || !matches(h, ref) {
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
	}
	return 0, 0, false
}

// matches reports if h is the section ref, by number or by title. A number of level 3 must be dotted.
func matches(h heading, ref string) bool {
	if h.number == ref && (h.level == 2 || strings.Contains(ref, ".")) {
		return true
	}
	return strings.EqualFold(h.title, ref)
}

// fence tracks a code fence (CommonMark): the opening run, and the line state.
type fence struct {
	char byte
	n    int
}

// step updates f with line and reports if line is in a code block (a fence line included).
func (f *fence) step(line string) bool {
	m := fencePattern.FindStringSubmatch(line)
	if f.n > 0 {
		if m != nil && m[1][0] == f.char && len(m[1]) >= f.n && strings.TrimSpace(m[2]) == "" {
			f.n = 0
		}
		return true
	}
	if m == nil || (m[1][0] == '`' && strings.Contains(m[2], "`")) {
		return false
	}
	f.char, f.n = m[1][0], len(m[1])
	return true
}

// unclosedFence reports if lines end inside a code fence.
func unclosedFence(lines []string) bool {
	var f fence
	for _, line := range lines {
		f.step(strings.TrimSuffix(line, "\r"))
	}
	return f.n > 0
}

// headings returns the headings that are not in a code block.
func headings(lines []string) []heading {
	var hs []heading
	var f fence
	for i, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		if f.step(line) {
			continue
		}
		m := headingPattern.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		hs = append(hs, heading{level: len(m[1]), line: i, number: m[2], title: m[3]})
	}
	return hs
}
