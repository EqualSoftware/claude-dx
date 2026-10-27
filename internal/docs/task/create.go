package task

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/EqualSoftware/claude-dx/internal/docs/docfile"
	"github.com/EqualSoftware/claude-dx/internal/markdown"
)

// generalItem is the first Definition of Done item of a general task, which covers no requirement (Q-42).
const generalItem = "- [ ] The work of the Brief is done, with evidence."

// New holds the content of a new task.
type New struct {
	Spec      string
	Slug      string
	Title     string
	Priority  string
	Brief     string
	Plan      []string
	Covers    []string
	DependsOn []string
}

// check returns an error for a missing or invalid field.
func (n New) check() error {
	if !validSlug(n.Spec) {
		return fmt.Errorf("no spec has the slug %s", n.Spec)
	}
	if !validSlug(n.Slug) {
		return errors.New("the slug must be kebab-case, 40 characters or less")
	}
	if err := docfile.CheckTitle(n.Title); err != nil {
		return err
	}
	if err := oneOf("priority", n.Priority, Priorities); err != nil {
		return err
	}
	if strings.TrimSpace(n.Brief) == "" {
		return errors.New("give the brief")
	}
	if planText(n.Plan) == "" {
		return errors.New("give one or more plan steps")
	}
	return nil
}

// Create writes a new Todo task and returns its ID. It holds the docs lock, so the checks of the spec,
// the covers and the dependencies stay true until the write.
func (s Store) Create(n New) (string, error) {
	if err := n.check(); err != nil {
		return "", err
	}
	f := Front{ID: n.Spec + "/" + n.Slug, Title: strings.TrimSpace(n.Title), Status: "Todo", Priority: n.Priority,
		Covers: nonNil(n.Covers), DependsOn: nonNil(n.DependsOn)}
	return f.ID, docfile.Locked(s.Lock, func() error {
		if s.exists(f.ID) {
			return fmt.Errorf("task %s already exists", f.ID)
		}
		if err := s.checkFront(&f); err != nil {
			return err
		}
		path, _ := s.path(f.ID)
		data, err := fill(f, n)
		if err != nil {
			return err
		}
		return writeNew(path, data)
	})
}

// checkFront checks the spec, the covers and the dependencies of f. It changes f.Covers to the form of
// the spec ("r-2" gives "R-2"). The caller holds the docs lock.
func (s Store) checkFront(f *Front) error {
	if err := s.checkCovers(f); err != nil {
		return err
	}
	return s.checkDeps(*f)
}

// checkCovers refuses a spec that does not exist, a requirement that the spec does not have, and covers
// in a general task (R-5).
func (s Store) checkCovers(f *Front) error {
	if f.Spec() == general {
		if len(f.Covers) > 0 {
			return errors.New("a general task does not cover requirements")
		}
		return nil
	}
	_, rids, err := s.specLocked(f.Spec())
	if err != nil {
		return err
	}
	for i, rid := range f.Covers {
		j := slices.IndexFunc(rids, func(r string) bool { return strings.EqualFold(r, rid) })
		if j < 0 {
			return fmt.Errorf("spec %s has no requirement %s", f.Spec(), rid)
		}
		f.Covers[i] = rids[j]
	}
	return nil
}

// fill makes a new task file from the template.
func fill(f Front, n New) (string, error) {
	front, body := markdown.Split(templateText)
	front = markdown.SetField(front, "id", f.ID)
	front = markdown.SetField(front, "title", f.Title)
	front = markdown.SetField(front, "priority", f.Priority)
	front = markdown.SetField(front, "covers", f.Covers)
	front = markdown.SetField(front, "depends_on", f.DependsOn)
	body = headerLine.ReplaceAllLiteralString(body, "# "+f.Title)
	if f.Spec() == general {
		body = strings.Replace(body, "- [ ] The covered requirements pass their acceptance criteria, with evidence.", generalItem, 1)
	}
	body, err := markdown.ReplaceSection(body, "Brief", strings.TrimSpace(n.Brief))
	if err != nil {
		return "", err
	}
	body, err = markdown.ReplaceSection(body, "Plan", planText(n.Plan))
	return markdown.Join(front, body), err
}

// planText makes the Plan checklist of steps. It drops empty steps.
func planText(steps []string) string {
	var lines []string
	for _, step := range steps {
		if step = oneLine(step); step != "" {
			lines = append(lines, "- [ ] "+step)
		}
	}
	return strings.Join(lines, "\n")
}
