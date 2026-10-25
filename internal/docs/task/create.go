package task

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/EqualSoftware/claude-dx/internal/docs/docfile"
	"github.com/EqualSoftware/claude-dx/internal/markdown"
)

// Fields are the values of the table at the top of a task. Empty fields do not change.
type Fields struct {
	Type      string `json:"type,omitempty" jsonschema:"feat, fix, refactor, perf, test, docs, build, ci or chore"`
	Owner     string `json:"owner,omitempty" jsonschema:"name or role"`
	Reviewers string `json:"reviewers,omitempty" jsonschema:"names or roles"`
	Platforms string `json:"platforms,omitempty" jsonschema:"operating systems, architectures or runtimes, or Not applicable"`
	Size      string `json:"size,omitempty" jsonschema:"S (less than 1 day) or M (1 to 3 days). Split an L task"`
	Blocks    string `json:"blocks,omitempty" jsonschema:"TASK-NNN list, or None"`
	Related   string `json:"related,omitempty" jsonschema:"ADR-NNN, docs, issues, external references"`
}

func (f Fields) values() map[string]string {
	return map[string]string{
		"Type": f.Type, "Owner": f.Owner, "Reviewers": f.Reviewers, "Platforms": f.Platforms,
		"Size": f.Size, "Blocks": f.Blocks, "Related": f.Related,
	}
}

// New holds the content of a new task.
type New struct {
	Title     string
	Priority  string
	Area      []string
	DependsOn []string
	Fields    Fields
	Sections  map[string]string // section number or title -> body
}

// Create writes a new Draft task with the next number and returns its path.
func (s Store) Create(t New) (string, error) {
	if strings.TrimSpace(t.Title) == "" {
		return "", errors.New("give a title")
	}
	if err := validPriority(t.Priority); err != nil {
		return "", err
	}
	front, body, err := s.fill(t)
	if err != nil {
		return "", err
	}
	var path string
	_, err = s.Alloc.Next("task", s.Dir, kind.Pattern, func(n int) error {
		id := kind.ID(n)
		path = filepath.Join(s.Dir, kind.FileName(n, t.Title))
		doc := markdown.Join(markdown.SetField(front, "id", id), setHeader(body, id, t.Title))
		return docfile.Create(path, []byte(doc))
	})
	if err != nil {
		return "", err
	}
	return path, s.WriteIndex()
}

// fill makes the frontmatter and the body of a new task from the template. The number comes later.
func (s Store) fill(t New) (front, body string, err error) {
	front, body = markdown.Split(templateText)
	front = markdown.SetField(front, "title", t.Title)
	front = markdown.SetField(front, "status", "Draft")
	front = markdown.SetField(front, "priority", t.Priority)
	front = markdown.SetField(front, "area", nonNil(t.Area))
	front = markdown.SetField(front, "depends_on", nonNil(t.DependsOn))
	body = setFields(body, t.Fields)
	body = strings.Replace(body, "| YYYY-MM-DD | [Name] | Initial draft specification |",
		fmt.Sprintf("| %s | %s | Initial draft specification |", s.today(), s.author()), 1)
	for ref, content := range t.Sections {
		if body, err = markdown.ReplaceSection(body, ref, content); err != nil {
			return "", "", err
		}
	}
	return front, body, nil
}

// setFields writes the non-empty fields into the table at the top of the body.
func setFields(body string, f Fields) string {
	for label, value := range f.values() {
		if value == "" {
			continue
		}
		row := regexp.MustCompile(`(?m)^\| ` + label + ` \| .* \|$`)
		body = row.ReplaceAllLiteralString(body, "| "+label+" | "+cell(value)+" |")
	}
	return body
}

func setHeader(body, id, title string) string {
	return headerLine.ReplaceAllLiteralString(body, "# "+id+": "+title)
}

func nonNil(list []string) []string {
	if list == nil {
		return []string{}
	}
	return list
}
