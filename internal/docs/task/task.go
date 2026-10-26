// Package task manages the task specs in docs/tasks and the generated INDEX.md (ADR-006).
package task

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/EqualSoftware/claude-dx/internal/docs/docfile"
	"github.com/EqualSoftware/claude-dx/internal/markdown"
	"github.com/EqualSoftware/claude-dx/internal/numbering"
)

//go:embed template.md
var templateText string

// ErrNoDir means that docs/tasks does not exist. It is a normal state, not a failure.
var ErrNoDir = errors.New("no tasks. docs/tasks does not exist")

var (
	kind       = docfile.NewKind("TASK")
	headerLine = regexp.MustCompile(`(?m)^# TASK-\S+: .*$`)
	dateRow    = regexp.MustCompile(`(?m)^\| (\d{4}-\d{2}-\d{2}) \|`)
)

// Statuses are the valid task statuses, in workflow order.
var Statuses = []string{"Draft", "In Review", "Approved", "In Development", "Blocked", "Done", "Cancelled"}

// Priorities are the valid priorities.
var Priorities = []string{"P0", "P1", "P2", "P3"}

// Types are the valid task types. They are the Conventional Commits types, because the type sets the commit type (ADR-012).
var Types = []string{"feat", "fix", "refactor", "perf", "test", "docs", "build", "ci", "chore"}

// Sizes are the valid sizes. An L task must be split.
var Sizes = []string{"S", "M"}

// Store reads and writes the task files of one project.
type Store struct {
	Dir    string              // absolute path of docs/tasks
	Alloc  numbering.Allocator // number allocation (ADR-007)
	Now    func() time.Time    // clock; nil means time.Now
	Author func() string       // author of Document History rows; nil means "Agent"
}

// Front is the YAML frontmatter of a task. It is the only source of these fields.
type Front struct {
	ID        string   `yaml:"id"`
	Title     string   `yaml:"title"`
	Status    string   `yaml:"status"`
	Priority  string   `yaml:"priority"`
	Type      string   `yaml:"type"`
	Size      string   `yaml:"size"`
	Owner     string   `yaml:"owner"`
	Platforms []string `yaml:"platforms"`
	Area      []string `yaml:"area"`
	DependsOn []string `yaml:"depends_on"`
	Related   []string `yaml:"related"`
}

// has reports if the field with the YAML key has the value, with no case difference.
// A list field has the value if one item is equal to it.
func (f Front) has(key, value string) bool {
	v := reflect.ValueOf(f)
	for i := range v.NumField() {
		if v.Type().Field(i).Tag.Get("yaml") != key {
			continue
		}
		switch field := v.Field(i).Interface().(type) {
		case string:
			return strings.EqualFold(field, value)
		case []string:
			return slices.ContainsFunc(field, func(item string) bool { return strings.EqualFold(item, value) })
		}
	}
	return false
}

// Task is one task in a list.
type Task struct {
	File   string
	Front  Front
	Closed string // date of the last Document History row
}

// Filter selects tasks in List: a frontmatter key and the value to find. Empty values select all.
// Status "open" selects all but Done and Cancelled, and "all" selects all.
type Filter map[string]string

func (f Filter) match(t Task) bool {
	for key, value := range f {
		if !matchField(t.Front, key, value) {
			return false
		}
	}
	return true
}

func matchField(f Front, key, value string) bool {
	switch {
	case value == "", key == "status" && value == "all":
		return true
	case key == "status" && value == "open":
		return f.Status != "Done" && f.Status != "Cancelled"
	}
	return f.has(key, value)
}

// List returns the tasks that match f, sorted by ID, and one warning line for each file that it cannot read
// or each number with many files.
func (s Store) List(f Filter) ([]Task, string, error) {
	files, err := kind.Files(s.Dir)
	if docfile.NotExist(err) {
		return nil, "", ErrNoDir
	}
	if err != nil {
		return nil, "", err
	}
	var tasks []Task
	warnings := docfile.Duplicates(kind, files)
	for _, name := range sortedNames(files) {
		t, err := s.read(name)
		if err != nil {
			warnings += fmt.Sprintf("Warning: %s: %v.\n", name, err)
			continue
		}
		if f.match(t) {
			tasks = append(tasks, t)
		}
	}
	return tasks, warnings, nil
}

func (s Store) read(name string) (Task, error) {
	data, err := os.ReadFile(filepath.Join(s.Dir, name))
	if err != nil {
		return Task{}, err
	}
	front, body := markdown.Split(string(data))
	var f Front
	if err := markdown.Decode(front, &f); err != nil {
		return Task{}, fmt.Errorf("bad frontmatter: %w", err)
	}
	history, _ := markdown.Section(body, "Document History")
	return Task{File: name, Front: f, Closed: lastDate(history)}, nil
}

// Get returns the task file, or one of its sections.
func (s Store) Get(id, section string) (string, error) {
	path, err := s.find(id)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil || section == "" {
		return string(data), err
	}
	_, body := markdown.Split(string(data))
	text, ok := markdown.Section(body, section)
	if !ok {
		return "", fmt.Errorf("section %q does not exist. Use a number (example: 7.1) or a title", section)
	}
	return text, nil
}

func (s Store) find(id string) (string, error) {
	if _, err := os.Stat(s.Dir); docfile.NotExist(err) {
		return "", ErrNoDir
	}
	return kind.Find(s.Dir, id)
}

// doc is a task file in memory.
type doc struct {
	path  string
	front string
	body  string
}

// edit reads the task, changes it with fn, writes it and makes INDEX.md again.
func (s Store) edit(id string, fn func(*doc) error) error {
	path, err := s.find(id)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	d := &doc{path: path}
	d.front, d.body = markdown.Split(string(data))
	if err := fn(d); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(markdown.Join(d.front, d.body)), 0o644); err != nil {
		return err
	}
	return s.WriteIndex()
}

func (s Store) today() string {
	if s.Now != nil {
		return s.Now().Format(time.DateOnly)
	}
	return time.Now().Format(time.DateOnly)
}

func (s Store) author() string {
	if s.Author != nil {
		if name := s.Author(); name != "" {
			return name
		}
	}
	return "Agent"
}

// addHistory adds a Document History row with today's date and the author.
func (s Store) addHistory(d *doc, note string) error {
	history, ok := markdown.Section(d.body, "Document History")
	if !ok {
		return errors.New("the task has no Document History section")
	}
	row := fmt.Sprintf("| %s | %s | %s |", s.today(), s.author(), cell(note))
	body, err := markdown.ReplaceSection(d.body, "Document History", history+"\n"+row)
	d.body = body
	return err
}

func lastDate(history string) string {
	m := dateRow.FindAllStringSubmatch(history, -1)
	if m == nil {
		return ""
	}
	return m[len(m)-1][1]
}

// cell makes text safe for one Markdown table cell.
func cell(text string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(text), "\n", " "), "|", `\|`)
}

func sortedNames(files map[int][]string) []string {
	var names []string
	for _, list := range files {
		names = append(names, list...)
	}
	sort.Strings(names)
	return names
}

// oneOf returns an error if value is not in valid. name is the field name for the message.
func oneOf(name, value string, valid []string) error {
	if slices.Contains(valid, value) {
		return nil
	}
	return fmt.Errorf("%s %q is not valid. Use one of: %s", name, value, strings.Join(valid, ", "))
}
