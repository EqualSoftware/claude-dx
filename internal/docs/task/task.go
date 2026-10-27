// Package task manages the local tasks in tasks/<spec>/<task-slug>.md of the project data directory
// (ADR-016). A task points to one spec, or to "general" for work that has no spec. Git does not track the
// tasks, so a task has no number and no Document History.
package task

import (
	"cmp"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/EqualSoftware/claude-dx/internal/docs/docfile"
	"github.com/EqualSoftware/claude-dx/internal/docs/spec"
	"github.com/EqualSoftware/claude-dx/internal/markdown"
)

//go:embed template.md
var templateText string

// general is the spec slug of the tasks that have no spec (ADR-016).
const general = "general"

// Statuses are the valid task statuses, in the order of task_list.
var Statuses = []string{"In Progress", "Blocked", "Todo", "Done"}

// Priorities are the valid priorities.
var Priorities = []string{"P0", "P1", "P2", "P3"}

var (
	slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	headerLine  = regexp.MustCompile(`(?m)^# .*$`)
)

// Store reads and writes the local tasks of one project.
type Store struct {
	Dir   string     // <data dir>/tasks
	Specs spec.Store // the specs of the project; Specs.Author gives the author of a spec row
	Lock  string     // path of the docs lock; required
}

// Front is the YAML frontmatter of a task. It is the only source of these fields.
type Front struct {
	ID        string   `yaml:"id"`
	Title     string   `yaml:"title"`
	Status    string   `yaml:"status"`
	Priority  string   `yaml:"priority"`
	Covers    []string `yaml:"covers"`
	DependsOn []string `yaml:"depends_on"`
}

// Spec returns the spec slug of the task.
func (f Front) Spec() string {
	slug, _, _ := strings.Cut(f.ID, "/")
	return slug
}

// doc is a task file in memory.
type doc struct {
	path  string
	front string
	body  string
	f     Front
}

// validSlug reports whether slug is kebab-case and not longer than a title.
func validSlug(slug string) bool {
	return slugPattern.MatchString(slug) && len(slug) <= docfile.MaxTitle
}

// path returns the file of the task id, or an error if id is not <spec>/<task-slug>.
func (s Store) path(id string) (string, error) {
	sp, slug, ok := strings.Cut(id, "/")
	if !ok || !validSlug(sp) || !validSlug(slug) {
		return "", fmt.Errorf("task %s does not exist", id)
	}
	return filepath.Join(s.Dir, sp, slug+".md"), nil
}

// load reads the task id.
func (s Store) load(id string) (*doc, error) {
	path, err := s.path(id)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if docfile.NotExist(err) {
		return nil, fmt.Errorf("task %s does not exist", id)
	}
	if err != nil {
		return nil, err
	}
	return parse(id, path, data)
}

// parse reads the task file data of id. It refuses a frontmatter id that is not the path.
func parse(id, path string, data []byte) (*doc, error) {
	d := &doc{path: path}
	d.front, d.body = markdown.Split(string(data))
	if err := markdown.Decode(d.front, &d.f); err != nil {
		return nil, fmt.Errorf("task file %s.md has a bad frontmatter: %w", id, err)
	}
	if d.f.ID != id {
		return nil, fmt.Errorf("task file %s.md has the id %s", id, d.f.ID)
	}
	return d, nil
}

// exists reports whether the file of the task id exists.
func (s Store) exists(id string) bool {
	path, err := s.path(id)
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// all returns the frontmatter of all tasks, by ID.
func (s Store) all() (map[string]Front, error) {
	tasks := map[string]Front{}
	specs, err := os.ReadDir(s.Dir)
	if docfile.NotExist(err) {
		return tasks, nil
	}
	if err != nil {
		return nil, err
	}
	for _, e := range specs {
		if !e.IsDir() {
			continue
		}
		if err := s.readSpec(e.Name(), tasks); err != nil {
			return nil, err
		}
	}
	return tasks, nil
}

// readSpec adds the tasks of the spec slug to tasks.
func (s Store) readSpec(slug string, tasks map[string]Front) error {
	dir := filepath.Join(s.Dir, slug)
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range files {
		name, ok := strings.CutSuffix(e.Name(), ".md")
		if e.IsDir() || !ok {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return err
		}
		d, err := parse(slug+"/"+name, filepath.Join(dir, e.Name()), data)
		if err != nil {
			return err
		}
		tasks[d.f.ID] = d.f
	}
	return nil
}

// List returns the tasks of the spec slug (all specs if empty) with the status (all if empty), in the
// order of R-12: spec slug with general last, status, priority, ID.
func (s Store) List(specSlug, status string) ([]Front, error) {
	if status != "" {
		if err := oneOf("status", status, Statuses); err != nil {
			return nil, err
		}
	}
	tasks, err := s.all()
	if err != nil {
		return nil, err
	}
	var out []Front
	for _, f := range tasks {
		if (specSlug == "" || f.Spec() == specSlug) && (status == "" || f.Status == status) {
			out = append(out, f)
		}
	}
	slices.SortFunc(out, compare)
	return out, nil
}

// compare orders two tasks for List.
func compare(a, b Front) int {
	return cmp.Or(
		cmp.Compare(last(a), last(b)),
		cmp.Compare(a.Spec(), b.Spec()),
		cmp.Compare(slices.Index(Statuses, a.Status), slices.Index(Statuses, b.Status)),
		cmp.Compare(a.Priority, b.Priority),
		cmp.Compare(a.ID, b.ID),
	)
}

// last is 1 for a general task, so List puts it after the tasks of the specs.
func last(f Front) int {
	if f.Spec() == general {
		return 1
	}
	return 0
}

// Get returns the task file, or one of its sections.
func (s Store) Get(id, section string) (string, error) {
	d, err := s.load(id)
	if err != nil {
		return "", err
	}
	if section == "" {
		return markdown.Join(d.front, d.body), nil
	}
	text, ok := markdown.Section(d.body, section)
	if !ok {
		return "", fmt.Errorf("section %s does not exist. Use Brief, Plan, Notes or Definition of Done", section)
	}
	return text, nil
}

// edit reads the task id, changes it with fn and writes it, with the docs lock. It refuses a Done task,
// because Done is final (R-3).
func (s Store) edit(id string, fn func(*doc) error) error {
	return docfile.Locked(s.Lock, func() error {
		d, err := s.load(id)
		if err != nil {
			return err
		}
		if d.f.Status == "Done" {
			return fmt.Errorf("task %s is Done and does not change", id)
		}
		if err := fn(d); err != nil {
			return err
		}
		return d.write()
	})
}

// write writes d. The file ends with one line break, also after a change of the last section.
func (d *doc) write() error {
	return docfile.Write(d.path, []byte(markdown.Join(d.front, strings.TrimRight(d.body, "\n")+"\n")))
}

// addNote adds the line as a list item to the Notes section (Q-41).
func (d *doc) addNote(line string) error {
	notes, ok := markdown.Section(d.body, "Notes")
	if !ok {
		return errors.New("the task has no Notes section")
	}
	body, err := markdown.ReplaceSection(d.body, "Notes", strings.TrimSpace(notes+"\n- "+line))
	d.body = body
	return err
}

// specLocked returns the status and the requirement IDs of the spec slug. The caller holds the docs lock.
func (s Store) specLocked(slug string) (status string, rids []string, err error) {
	status, rids, err = s.Specs.RequirementsLocked(slug)
	var missing spec.NotFoundError
	if errors.Is(err, spec.ErrNoDir) || errors.As(err, &missing) {
		return "", nil, fmt.Errorf("no spec has the slug %s", slug)
	}
	return status, rids, err
}

// oneOf returns an error if value is not in valid. name is the field name for the message.
func oneOf(name, value string, valid []string) error {
	if slices.Contains(valid, value) {
		return nil
	}
	return fmt.Errorf("%s %q is not valid. Use one of: %s", name, value, strings.Join(valid, ", "))
}

// oneLine joins the lines of text with one space.
func oneLine(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

func nonNil(list []string) []string {
	if list == nil {
		return []string{}
	}
	return list
}
