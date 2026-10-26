package task

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/EqualSoftware/claude-dx/internal/docs/docfile"
	"github.com/EqualSoftware/claude-dx/internal/markdown"
)

// New holds the content of a new task.
type New struct {
	Title     string
	Priority  string
	Milestone string
	Type      string
	Area      []string
	DependsOn []string
	Size      string
	Platforms []string
	Related   []string
	Sections  map[string]string // section number or title -> body
}

// check returns an error for a missing or invalid field.
func (t New) check() error {
	if err := docfile.CheckTitle(t.Title); err != nil {
		return err
	}
	if err := oneOf("priority", t.Priority, Priorities); err != nil {
		return err
	}
	if err := checkMilestone(t.Milestone); err != nil {
		return err
	}
	if err := oneOf("type", t.Type, Types); err != nil {
		return err
	}
	if t.Size == "" {
		return nil
	}
	return oneOf("size", t.Size, Sizes)
}

// Create writes a new Draft task with the next number and returns its path and the warnings.
func (s Store) Create(t New) (path, warnings string, err error) {
	if err := t.check(); err != nil {
		return "", "", err
	}
	if err := s.checkNewTitle(t.Title); err != nil {
		return "", "", err
	}
	warnings, err = s.checkDeps(Front{Milestone: t.Milestone, Priority: t.Priority, DependsOn: t.DependsOn}, nil)
	if err != nil {
		return "", "", err
	}
	front, body, err := s.fill(t)
	if err != nil {
		return "", "", err
	}
	_, err = s.Alloc.Next("task", s.Dir, kind.Pattern, func(n int) error {
		id := kind.ID(n)
		path = filepath.Join(s.Dir, kind.FileName(n, t.Title))
		doc := markdown.Join(markdown.SetField(front, "id", id), setHeader(body, id, t.Title))
		return docfile.Create(path, []byte(doc))
	})
	if err != nil {
		return "", "", err
	}
	return path, warnings, s.WriteIndex()
}

// fill makes the frontmatter and the body of a new task from the template. The number comes later.
func (s Store) fill(t New) (front, body string, err error) {
	front, body = markdown.Split(templateText)
	front = markdown.SetField(front, "title", t.Title)
	front = markdown.SetField(front, "status", "Draft")
	front = markdown.SetField(front, "priority", t.Priority)
	front = markdown.SetField(front, "milestone", t.Milestone)
	front = markdown.SetField(front, "type", t.Type)
	front = markdown.SetField(front, "size", t.Size)
	front = markdown.SetField(front, "owner", s.author())
	front = markdown.SetField(front, "platforms", nonNil(t.Platforms))
	front = markdown.SetField(front, "area", nonNil(t.Area))
	front = markdown.SetField(front, "depends_on", nonNil(t.DependsOn))
	front = markdown.SetField(front, "related", nonNil(t.Related))
	body = strings.Replace(body, "| YYYY-MM-DD | [Name] | Initial draft specification |",
		fmt.Sprintf("| %s | %s | Initial draft specification |", s.today(), s.author()), 1)
	for ref, content := range t.Sections {
		if err := checkSection(ref); err != nil {
			return "", "", err
		}
		if body, err = markdown.ReplaceSection(body, ref, content); err != nil {
			return "", "", err
		}
	}
	return front, body, nil
}

// checkNewTitle refuses a title that another task has, in all statuses.
func (s Store) checkNewTitle(title string) error {
	tasks, _, err := s.List(Filter{"status": "all"})
	if err != nil && !errors.Is(err, ErrNoDir) {
		return err
	}
	titles := make(map[string]string, len(tasks))
	for _, t := range tasks {
		titles[t.Front.ID] = t.Front.Title
	}
	return docfile.CheckNewTitle(title, titles)
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
