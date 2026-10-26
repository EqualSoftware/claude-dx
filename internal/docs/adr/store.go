// Package adr manages the ADRs in docs/decisions (ADR-006).
package adr

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/EqualSoftware/claude-dx/internal/docs/docfile"
	"github.com/EqualSoftware/claude-dx/internal/markdown"
	"github.com/EqualSoftware/claude-dx/internal/numbering"
)

//go:embed template.md
var templateText string

var (
	tmpl       = template.Must(template.New("adr").Parse(templateText))
	kind       = docfile.NewKind("ADR")
	statusLine = regexp.MustCompile(`(?m)^- Status: (.*)$`)
	titleLine  = regexp.MustCompile(`(?m)^# ADR-\d{3}: (.*)$`)
	dateLine   = regexp.MustCompile(`(?m)^- Date: (.*)$`)
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
	Date   string
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

// List returns the ADRs by number, and one warning line for each file that it cannot read or each number
// with many files. It skips a bad file, so one bad file does not hide the other ADRs.
func (s Store) List() ([]Entry, string, error) {
	files, err := kind.Files(s.Dir)
	if docfile.NotExist(err) {
		return nil, "", ErrNoDir
	}
	if err != nil {
		return nil, "", err
	}
	var list []Entry
	warnings := docfile.Duplicates(kind, files)
	for _, n := range sortedKeys(files) {
		for _, name := range files[n] {
			entry, err := s.read(name, n)
			if err != nil {
				warnings += fmt.Sprintf("Warning: %s: %v.\n", name, err)
				continue
			}
			list = append(list, entry)
		}
	}
	return list, warnings, nil
}

func (s Store) read(name string, n int) (Entry, error) {
	data, err := os.ReadFile(filepath.Join(s.Dir, name))
	if err != nil {
		return Entry{}, err
	}
	return Entry{Number: n, File: name, Status: first(statusLine, data), Title: first(titleLine, data), Date: first(dateLine, data)}, nil
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

// Create writes a new ADR with the next number, makes README.md again and returns its path.
func (s Store) Create(adr New) (string, error) {
	adr = s.defaults(adr)
	if err := validStatus(adr.Status); err != nil {
		return "", err
	}
	if err := s.checkNewTitle(adr.Title); err != nil {
		return "", err
	}
	var path string
	_, err := s.Alloc.Next("adr", s.Dir, kind.Pattern, func(n int) error {
		adr.Number = n
		path = filepath.Join(s.Dir, kind.FileName(n, adr.Title))
		var b bytes.Buffer
		if err := tmpl.Execute(&b, adr); err != nil {
			return err
		}
		return docfile.Create(path, b.Bytes())
	})
	if err != nil {
		return "", err
	}
	return path, s.WriteIndex()
}

// checkNewTitle refuses a long title, and a title that another ADR has, in all statuses.
func (s Store) checkNewTitle(title string) error {
	list, _, err := s.List()
	if err != nil && !errors.Is(err, ErrNoDir) {
		return err
	}
	titles := make(map[string]string, len(list))
	for _, e := range list {
		titles[kind.ID(e.Number)] = e.Title
	}
	return docfile.CheckNewTitle(title, titles)
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

// Update changes the status, or replaces one section, or both, and makes README.md again.
func (s Store) Update(id, status, section, body string) error {
	if status == "" && section == "" {
		return errors.New("give a status, or a section and a body")
	}
	if status != "" {
		if err := s.canSetStatus(id, status); err != nil {
			return err
		}
	}
	err := s.edit(id, func(doc string) (string, error) {
		if status != "" {
			doc = statusLine.ReplaceAllLiteralString(doc, "- Status: "+status)
		}
		if section == "" {
			return doc, nil
		}
		return markdown.ReplaceSection(doc, section, body)
	})
	if err != nil {
		return err
	}
	return s.WriteIndex()
}

// Supersede marks the old ADR as replaced by the new ADR, adds the link to the new ADR and makes
// README.md again.
func (s Store) Supersede(oldID, newID string) error {
	oldN, err := kind.Parse(oldID)
	if err != nil {
		return err
	}
	newN, err := kind.Parse(newID)
	if err != nil {
		return err
	}
	if oldN == newN {
		return errors.New("an ADR cannot replace itself")
	}
	if err := s.canSupersede(oldID, newID); err != nil {
		return err
	}
	err = s.edit(oldID, func(doc string) (string, error) {
		return statusLine.ReplaceAllLiteralString(doc, fmt.Sprintf("- Status: Superseded by ADR-%03d", newN)), nil
	})
	if err != nil {
		return err
	}
	err = s.edit(newID, func(doc string) (string, error) {
		return addReplaces(doc, oldN)
	})
	if err != nil {
		return err
	}
	return s.WriteIndex()
}

// canSetStatus refuses a status that is not valid, and a status change of a superseded ADR,
// because only adr_supersede sets that status.
func (s Store) canSetStatus(id, status string) error {
	if err := validStatus(status); err != nil {
		return err
	}
	current, err := s.status(id)
	if err != nil {
		return err
	}
	if strings.HasPrefix(current, "Superseded") {
		return fmt.Errorf("%s is %q. Its status cannot change. You can change its sections", id, current)
	}
	return nil
}

// canSupersede refuses a chain that is not correct: the old ADR has one replacement, and the new ADR
// is the current decision, so it is Accepted. This also stops cycles.
func (s Store) canSupersede(oldID, newID string) error {
	oldStatus, err := s.status(oldID)
	if err != nil {
		return err
	}
	if strings.HasPrefix(oldStatus, "Superseded") {
		return fmt.Errorf("%s is already %q. One ADR has one replacement", oldID, oldStatus)
	}
	newStatus, err := s.status(newID)
	if err != nil {
		return err
	}
	if newStatus != "Accepted" {
		return fmt.Errorf("%s is %q. The new ADR must be Accepted", newID, newStatus)
	}
	return nil
}

// status returns the status line value of the ADR.
func (s Store) status(id string) (string, error) {
	path, err := s.find(id)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return first(statusLine, data), nil
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
	if _, err := os.Stat(s.Dir); docfile.NotExist(err) {
		return "", ErrNoDir
	}
	return kind.Find(s.Dir, id)
}

func sortedKeys(files map[int][]string) []int {
	keys := make([]int, 0, len(files))
	for n := range files {
		keys = append(keys, n)
	}
	sort.Ints(keys)
	return keys
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
