package spec

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/EqualSoftware/claude-dx/internal/docs/docfile"
	"github.com/EqualSoftware/claude-dx/internal/markdown"
)

var (
	milestonePattern = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`)
	adrKind          = docfile.NewKind("ADR")
)

// reserved is the slug of the tasks that have no spec (ADR-016).
const reserved = "general"

// New holds the content of a new spec.
type New struct {
	Slug      string
	Title     string
	Milestone string
	Priority  string
	Area      []string
	Related   []string
	SplitFrom string // slug of the old spec of a split, for the Document History
}

// check returns an error for a missing or invalid field.
func (n New) check() error {
	if err := checkSlug(n.Slug); err != nil {
		return err
	}
	if n.SplitFrom != "" && !slugPattern.MatchString(n.SplitFrom) {
		return fmt.Errorf("slug %q is not kebab-case. Example: docs-lock", n.SplitFrom)
	}
	if !milestonePattern.MatchString(n.Milestone) {
		return fmt.Errorf("milestone %q is not valid. Use vMAJOR.MINOR.PATCH, example v0.1.0", n.Milestone)
	}
	return oneOf("priority", n.Priority, Priorities)
}

// checkSlug refuses a slug that is not kebab-case, is longer than the title limit, or is reserved.
func checkSlug(slug string) error {
	if slug == reserved {
		return fmt.Errorf("slug %q is reserved", slug)
	}
	if !slugPattern.MatchString(slug) {
		return fmt.Errorf("slug %q is not kebab-case. Example: docs-lock", slug)
	}
	if len(slug) > docfile.MaxTitle {
		return fmt.Errorf("the slug has %d characters. The limit is %d. Write a shorter slug", len(slug), docfile.MaxTitle)
	}
	return nil
}

// Create writes a new Draft spec. It holds the docs lock, so the slug and title checks stay true until the write.
func (s Store) Create(n New) error {
	if err := n.check(); err != nil {
		return err
	}
	return docfile.Locked(s.Lock, func() error {
		if err := s.checkNew(n); err != nil {
			return err
		}
		if err := docfile.Create(s.path(n.Slug), []byte(s.fill(n))); err != nil {
			return err
		}
		return s.WriteIndex()
	})
}

// checkNew refuses a slug that exists, a title that another spec has and an ADR that does not exist.
func (s Store) checkNew(n New) error {
	if _, err := os.Stat(s.path(n.Slug)); err == nil {
		return fmt.Errorf("spec %q exists", n.Slug)
	}
	specs, _, err := s.List(nil)
	if err != nil && !errors.Is(err, ErrNoDir) {
		return err
	}
	titles := make(map[string]string, len(specs))
	for _, sp := range specs {
		titles[sp.Front.Slug] = sp.Front.Title
	}
	if err := docfile.CheckNewTitle(n.Title, titles); err != nil {
		return err
	}
	for _, id := range n.Related {
		if _, err := adrKind.Find(s.ADRDir, id); err != nil {
			return fmt.Errorf("related %q: %w", id, err)
		}
	}
	return nil
}

// fill makes a new spec document from the template.
func (s Store) fill(n New) string {
	front, body := markdown.Split(templateText)
	front = markdown.SetField(front, "slug", n.Slug)
	front = markdown.SetField(front, "title", n.Title)
	front = markdown.SetField(front, "milestone", n.Milestone)
	front = markdown.SetField(front, "priority", n.Priority)
	front = markdown.SetField(front, "area", nonNil(n.Area))
	front = markdown.SetField(front, "related", nonNil(n.Related))
	body = headerLine.ReplaceAllLiteralString(body, "# "+n.Title)
	note := "Initial draft specification"
	if n.SplitFrom != "" {
		note += ". Split from " + n.SplitFrom + "."
	}
	body = strings.Replace(body, "| YYYY-MM-DD | [Name] | Initial draft specification |",
		fmt.Sprintf("| %s | %s | %s |", s.today(), docfile.Cell(s.author()), note), 1)
	return markdown.Join(front, body)
}

func nonNil(list []string) []string {
	if list == nil {
		return []string{}
	}
	return list
}
