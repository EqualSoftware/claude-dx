// Package spec manages the spec documents in docs/specs and the generated README.md (ADR-016).
package spec

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/EqualSoftware/claude-dx/internal/docs/docfile"
	"github.com/EqualSoftware/claude-dx/internal/markdown"
)

//go:embed template.md
var templateText string

// ErrNoDir means that docs/specs does not exist. It is a normal state, not a failure.
var ErrNoDir = errors.New("no specs. docs/specs does not exist")

// fileName is the name of the spec document in the folder of the spec.
const fileName = "SPEC.md"

var (
	slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	headerLine  = regexp.MustCompile(`(?m)^# .*$`)
)

// Statuses are the valid spec statuses, in workflow order.
var Statuses = []string{"Draft", "In Review", "Approved"}

// Priorities are the valid priorities.
var Priorities = []string{"P0", "P1", "P2", "P3"}

// Store reads and writes the specs of one project.
type Store struct {
	Dir    string           // absolute path of docs/specs
	ADRDir string           // absolute path of docs/decisions, for the related check
	Root   string           // project root, where git runs
	Now    func() time.Time // clock; nil means time.Now
	Author func() string    // author of Document History rows; nil means "Agent"
	Lock   string           // path of the docs lock (ADR-015); required
}

// Front is the YAML frontmatter of a spec. It is the only source of these fields.
type Front struct {
	Slug      string   `yaml:"slug"`
	Title     string   `yaml:"title"`
	Status    string   `yaml:"status"`
	Milestone string   `yaml:"milestone"`
	Priority  string   `yaml:"priority"`
	Area      []string `yaml:"area"`
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

// Spec is one spec in a list.
type Spec struct {
	Front Front
	Done  int // requirements with the status Done
	Total int // all requirements
}

// Filter selects specs in List: a frontmatter key and the value to find. Empty values select all.
type Filter map[string]string

func (f Filter) match(front Front) bool {
	for key, value := range f {
		if value != "" && !front.has(key, value) {
			return false
		}
	}
	return true
}

// List returns the specs that match f, sorted by slug, and one warning line for each spec that it cannot read.
func (s Store) List(f Filter) ([]Spec, string, error) {
	entries, err := os.ReadDir(s.Dir)
	if docfile.NotExist(err) {
		return nil, "", ErrNoDir
	}
	if err != nil {
		return nil, "", err
	}
	var specs []Spec
	var warnings string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		sp, err := s.read(e.Name())
		if docfile.NotExist(err) {
			continue // a folder with no SPEC.md is not a spec
		}
		if err != nil {
			warnings += fmt.Sprintf("Warning: %s: %v.\n", e.Name(), err)
			continue
		}
		if f.match(sp.Front) {
			specs = append(specs, sp)
		}
	}
	return specs, warnings, nil
}

func (s Store) read(slug string) (Spec, error) {
	data, err := os.ReadFile(s.path(slug))
	if err != nil {
		return Spec{}, err
	}
	front, body := markdown.Split(string(data))
	var f Front
	if err := markdown.Decode(front, &f); err != nil {
		return Spec{}, fmt.Errorf("bad frontmatter: %w", err)
	}
	sp := Spec{Front: f}
	for _, r := range requirements(body) {
		sp.Total++
		if r[4] == "Done" {
			sp.Done++
		}
	}
	return sp, nil
}

// path returns the path of the spec document of slug.
func (s Store) path(slug string) string {
	return filepath.Join(s.Dir, slug, fileName)
}

// find returns the path of the spec document of slug, or an error if the spec does not exist.
func (s Store) find(slug string) (string, error) {
	if _, err := os.Stat(s.Dir); docfile.NotExist(err) {
		return "", ErrNoDir
	}
	if !slugPattern.MatchString(slug) {
		return "", fmt.Errorf("spec %q does not exist", slug)
	}
	path := s.path(slug)
	if _, err := os.Stat(path); docfile.NotExist(err) {
		return "", fmt.Errorf("spec %q does not exist", slug)
	}
	return path, nil
}

// Get returns the spec document, or one of its sections, with no commits.
func (s Store) Get(slug, ref string) (string, error) {
	path, err := s.find(slug)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil || ref == "" {
		return string(data), err
	}
	text, _, err := section(string(data), ref)
	return text, err
}

// section returns one section of the spec document data and its number.
func section(data, ref string) (text, number string, err error) {
	_, body := markdown.Split(data)
	text, ok := markdown.Section(body, ref)
	if !ok {
		return "", "", fmt.Errorf("section %q does not exist. Use a number (example: 5) or a title", ref)
	}
	number, _ = markdown.SectionNumber(body, ref)
	return text, number, nil
}

// doc is a spec document in memory.
type doc struct {
	slug  string
	front string
	body  string
}

// status returns the status in the frontmatter of d.
func (d *doc) status() (string, error) {
	var f Front
	if err := markdown.Decode(d.front, &f); err != nil {
		return "", fmt.Errorf("bad frontmatter: %w", err)
	}
	return f.Status, nil
}

// edit reads the spec, changes it with fn, writes it and makes README.md again, with the docs lock.
func (s Store) edit(slug string, fn func(*doc) error) error {
	return docfile.Locked(s.Lock, func() error {
		path, err := s.find(slug)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		d := &doc{slug: slug}
		d.front, d.body = markdown.Split(string(data))
		if err := fn(d); err != nil {
			return err
		}
		if err := docfile.Write(path, []byte(markdown.Join(d.front, d.body))); err != nil {
			return err
		}
		return s.WriteIndex()
	})
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
		return errors.New("the spec has no Document History section")
	}
	row := fmt.Sprintf("| %s | %s | %s |", s.today(), s.author(), docfile.Cell(note))
	body, err := markdown.ReplaceSection(d.body, "Document History", history+"\n"+row)
	d.body = body
	return err
}

// toReview sets an Approved spec to In Review and adds the note to the Document History.
// It returns true if the status changed.
func (s Store) toReview(d *doc, note string) (bool, error) {
	status, err := d.status()
	if err != nil || status != "Approved" {
		return false, err
	}
	d.front = markdown.SetField(d.front, "status", "In Review")
	return true, s.addHistory(d, note)
}

// oneOf returns an error if value is not in valid. name is the field name for the message.
func oneOf(name, value string, valid []string) error {
	if slices.Contains(valid, value) {
		return nil
	}
	return fmt.Errorf("%s %q is not valid. Use one of: %s", name, value, strings.Join(valid, ", "))
}
