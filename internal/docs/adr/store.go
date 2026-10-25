// Package adr manages the ADRs in docs/decisions (ADR-006).
package adr

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/is-equal/claude-dx/internal/markdown"
	"github.com/is-equal/claude-dx/internal/numbering"
)

//go:embed template.md
var templateText string

var (
	tmpl        = template.Must(template.New("adr").Parse(templateText))
	filePattern = regexp.MustCompile(`^ADR-(\d{3})-.*\.md$`)
	idPattern   = regexp.MustCompile(`^(?i:ADR-)?(\d{1,3})$`)
	statusLine  = regexp.MustCompile(`(?m)^- Status: (.*)$`)
	titleLine   = regexp.MustCompile(`(?m)^# ADR-\d{3}: (.*)$`)
)

// ErrNoDir means that docs/decisions does not exist. It is a normal state, not a failure.
var ErrNoDir = errors.New("no ADRs. docs/decisions does not exist")

// Sections are the sections that adr_get and adr_update accept.
var Sections = []string{"Context", "Decision", "Alternatives", "Consequences", "Evidence"}

// Store reads and writes the ADR files of one project.
type Store struct {
	Dir   string              // absolute path of docs/decisions
	Alloc numbering.Allocator // number allocation (ADR-007)
	Now   func() time.Time    // clock; nil means time.Now
}

// Entry is one ADR in a list.
type Entry struct {
	Number int
	Status string
	Title  string
	File   string
}

// Alternative is one option that the ADR did not select, or selected.
type Alternative struct {
	Option string `json:"option" jsonschema:"the option"`
	Reason string `json:"reason" jsonschema:"why the option is accepted or rejected"`
}

// Consequences are the effects of the decision.
type Consequences struct {
	Positive string `json:"positive" jsonschema:"benefits"`
	Negative string `json:"negative" jsonschema:"costs or risks"`
	FollowUp string `json:"follow_up" jsonschema:"required actions, or None"`
}

// New holds the content of a new ADR.
type New struct {
	Number       int
	Title        string
	Status       string
	Date         string
	Owners       string
	Context      string
	Decision     string
	Alternatives []Alternative
	Consequences Consequences
	Evidence     string
}

// List returns the ADRs sorted by number. Two files with one number are both in the list.
func (s Store) List() ([]Entry, error) {
	entries, err := os.ReadDir(s.Dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrNoDir
	}
	if err != nil {
		return nil, err
	}
	var list []Entry
	for _, e := range entries {
		m := filePattern.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		entry, err := s.read(e.Name(), m[1])
		if err != nil {
			return nil, err
		}
		list = append(list, entry)
	}
	return list, nil
}

func (s Store) read(name, number string) (Entry, error) {
	data, err := os.ReadFile(filepath.Join(s.Dir, name))
	if err != nil {
		return Entry{}, err
	}
	n, _ := strconv.Atoi(number)
	return Entry{Number: n, File: name, Status: first(statusLine, data), Title: first(titleLine, data)}, nil
}

// Get returns the ADR, or one of its sections.
func (s Store) Get(id, section string) (string, error) {
	path, err := s.find(id)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil || section == "" {
		return string(data), err
	}
	text, ok := markdown.Section(string(data), section)
	if !ok {
		return "", fmt.Errorf("section %q does not exist. Use one of: %s", section, strings.Join(Sections, ", "))
	}
	return text, nil
}

// Create writes a new ADR with the next number and returns its path relative to the docs root.
func (s Store) Create(adr New) (string, error) {
	adr = s.defaults(adr)
	if err := validStatus(adr.Status); err != nil {
		return "", err
	}
	var path string
	_, err := s.Alloc.Next("adr", s.Dir, filePattern, func(n int) error {
		adr.Number = n
		path = filepath.Join(s.Dir, fmt.Sprintf("ADR-%03d-%s.md", n, Slug(adr.Title)))
		return s.write(path, adr)
	})
	return path, err
}

func (s Store) defaults(adr New) New {
	if adr.Status == "" {
		adr.Status = "Proposed"
	}
	if adr.Owners == "" {
		adr.Owners = "Maintainer"
	}
	now := time.Now
	if s.Now != nil {
		now = s.Now
	}
	adr.Date = now().Format(time.DateOnly)
	return adr
}

// write renders the template into a new file. Exclusive create makes a duplicate number fail.
func (s Store) write(path string, adr New) error {
	var b bytes.Buffer
	if err := tmpl.Execute(&b, adr); err != nil {
		return err
	}
	if err := os.MkdirAll(s.Dir, 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	_, err = f.Write(b.Bytes())
	return errors.Join(err, f.Close())
}

// Update changes the status, or replaces one section, or both.
func (s Store) Update(id, status, section, body string) error {
	if status == "" && section == "" {
		return errors.New("give a status, or a section and a body")
	}
	if status != "" {
		if err := validStatus(status); err != nil {
			return err
		}
	}
	return s.edit(id, func(doc string) (string, error) {
		if status != "" {
			doc = statusLine.ReplaceAllLiteralString(doc, "- Status: "+status)
		}
		if section == "" {
			return doc, nil
		}
		return markdown.ReplaceSection(doc, section, body)
	})
}

// Supersede marks the old ADR as replaced by the new ADR, and adds the link to the new ADR.
func (s Store) Supersede(oldID, newID string) error {
	oldN, err := parseID(oldID)
	if err != nil {
		return err
	}
	newN, err := parseID(newID)
	if err != nil {
		return err
	}
	if oldN == newN {
		return errors.New("an ADR cannot replace itself")
	}
	if _, err := s.find(newID); err != nil {
		return err
	}
	err = s.edit(oldID, func(doc string) (string, error) {
		return statusLine.ReplaceAllLiteralString(doc, fmt.Sprintf("- Status: Superseded by ADR-%03d", newN)), nil
	})
	if err != nil {
		return err
	}
	return s.edit(newID, func(doc string) (string, error) {
		return addReplaces(doc, oldN)
	})
}

// addReplaces puts "Replaces ADR-NNN." at the start of Context, one time only.
func addReplaces(doc string, oldN int) (string, error) {
	line := fmt.Sprintf("Replaces ADR-%03d.", oldN)
	context, _ := markdown.Section(doc, "Context")
	if strings.Contains(context, line) {
		return doc, nil
	}
	return markdown.ReplaceSection(doc, "Context", line+"\n\n"+context)
}

// edit reads the ADR, changes it with fn and writes it again.
func (s Store) edit(id string, fn func(string) (string, error)) error {
	path, err := s.find(id)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	doc, err := fn(string(data))
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(doc), 0o644)
}

// find returns the path of the ADR with the number in id.
func (s Store) find(id string) (string, error) {
	n, err := parseID(id)
	if err != nil {
		return "", err
	}
	list, err := s.List()
	if err != nil {
		return "", err
	}
	var matches []string
	for _, e := range list {
		if e.Number == n {
			matches = append(matches, e.File)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("ADR-%03d does not exist", n)
	case 1:
		return filepath.Join(s.Dir, matches[0]), nil
	}
	return "", fmt.Errorf("ADR-%03d has %d files: %s. Give each file a different number first", n, len(matches), strings.Join(matches, ", "))
}

func parseID(id string) (int, error) {
	m := idPattern.FindStringSubmatch(strings.TrimSpace(id))
	if m == nil {
		return 0, fmt.Errorf("%q is not an ADR ID. Example: ADR-007", id)
	}
	n, _ := strconv.Atoi(m[1])
	return n, nil
}

func validStatus(status string) error {
	if status == "Proposed" || status == "Accepted" {
		return nil
	}
	return fmt.Errorf("status %q is not valid. Use Proposed or Accepted. Use adr_supersede to replace an ADR", status)
}

func first(re *regexp.Regexp, data []byte) string {
	m := re.FindSubmatch(data)
	if m == nil {
		return ""
	}
	return strings.TrimSpace(string(m[1]))
}

// Slug makes the file name part from a title: lowercase ASCII words joined by "-", maximum 60 characters.
func Slug(title string) string {
	words := strings.FieldsFunc(strings.ToLower(title), func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9')
	})
	slug := strings.Join(words, "-")
	if len(slug) <= 60 {
		return slug
	}
	cut := strings.LastIndexByte(slug[:61], '-')
	if cut <= 0 {
		return slug[:60]
	}
	return slug[:cut]
}
