// Package adr manages the ADRs in docs/decisions (ADR-002).
package adr

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
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
	statusHead = regexp.MustCompile(`(?m)^- Status:[^\r\n]*`)
	statusLine = regexp.MustCompile(`(?m)^- Status: (.*)$`)
	titleLine  = regexp.MustCompile(`(?m)^# ADR-\d{3}: (.*)$`)
	dateLine   = regexp.MustCompile(`(?m)^- Date: (.*)$`)
	ownersLine = regexp.MustCompile(`(?m)^- Owners:[^\r\n]*\r?$`)
)

// ErrNoDir means that docs/decisions does not exist. It is a normal state, not a failure.
var ErrNoDir = errors.New("no ADRs. docs/decisions does not exist")

// Sections are the sections that adr_get and adr_update accept.
var Sections = []string{"Context", "Decision", "Alternatives", "Consequences", "Evidence"}

// Store reads and writes the ADR files of one project.
type Store struct {
	Dir    string              // absolute path of docs/decisions
	Alloc  numbering.Allocator // number allocation (ADR-007)
	Now    func() time.Time    // clock; nil means time.Now
	Author func() string       // default owner (the user name of the VCS); nil or "" means "Maintainer"
	Lock   string              // path of the docs lock (docs-common#R-10, R-11); required
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
func (s Store) Create(adr New) (path string, err error) {
	err = docfile.Locked(s.Lock, func() error {
		path, err = s.create(adr)
		return err
	})
	return path, err
}

func (s Store) create(adr New) (string, error) {
	adr = s.defaults(adr)
	if err := validStatus(adr.Status); err != nil {
		return "", err
	}
	if err := s.checkNewTitle(adr.Title); err != nil {
		return "", err
	}
	if err := checkText(adr); err != nil {
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

// field is a text of an ADR and its name for the messages.
type field struct{ name, text string }

// checkText refuses an empty required text, a line break in a one-line field, and text that changes the
// structure of the ADR: an unclosed code fence and a heading that starts or repeats a section.
func checkText(adr New) error {
	lines := []field{{"owners", adr.Owners}}
	for _, a := range adr.Alternatives {
		lines = append(lines, field{"alternative option", a.Option}, field{"alternative reason", a.Reason})
	}
	for _, l := range lines {
		if err := docfile.CheckLine(l.text, l.name); err != nil {
			return err
		}
	}
	required := []field{{"context", adr.Context}, {"decision", adr.Decision}, {"evidence", adr.Evidence}}
	for _, f := range required {
		if strings.TrimSpace(f.text) == "" {
			return fmt.Errorf("give the %s", f.name)
		}
	}
	var b bytes.Buffer
	if err := tmpl.Execute(&b, adr); err != nil {
		return err
	}
	c := adr.Consequences
	texts := slices.Clone(required)
	texts = append(texts, field{"positive", c.Positive}, field{"negative", c.Negative},
		field{"follow_up", c.FollowUp})
	for _, f := range texts {
		if err := markdown.CheckContent(b.String(), f.text, f.name, 2); err != nil {
			return err
		}
	}
	return nil
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

// owner returns the default owner of an ADR: the author, else "Maintainer".
func (s Store) owner() string {
	if s.Author != nil {
		if name := s.Author(); name != "" {
			return name
		}
	}
	return "Maintainer"
}

func (s Store) defaults(adr New) New {
	if adr.Status == "" {
		adr.Status = "Proposed"
	}
	if adr.Owners == "" {
		adr.Owners = s.owner()
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
	if path, err := s.find(id); err == nil {
		if err := checkUpstream(id, path); err != nil {
			return err
		}
	}
	return docfile.Locked(s.Lock, func() error { return s.update(id, status, section, body) })
}

func (s Store) update(id, status, section, body string) error {
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
			var err error
			if doc, err = setStatus(doc, status); err != nil {
				return "", err
			}
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
	return docfile.Locked(s.Lock, func() error { return s.supersede(oldID, newID) })
}

func (s Store) supersede(oldID, newID string) error {
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
	// The new ADR is written first: its edit can fail, and then nothing is written.
	err = s.edit(newID, func(doc string) (string, error) {
		return addSupersedes(doc, oldN)
	})
	if err != nil {
		return err
	}
	err = s.edit(oldID, func(doc string) (string, error) {
		return setStatus(doc, fmt.Sprintf("Superseded by ADR-%03d", newN))
	})
	if err != nil {
		return err
	}
	return s.WriteIndex()
}

// setStatus replaces the value of the first "- Status:" line before the first "## " heading. A line
// of the same form in a section stays. The CR of a CRLF line end stays.
func setStatus(doc, status string) (string, error) {
	head := doc
	if i := strings.Index("\n"+doc, "\n## "); i >= 0 {
		head = doc[:i]
	}
	loc := statusHead.FindStringIndex(head)
	if loc == nil {
		return "", errors.New("the ADR has no Status line")
	}
	return doc[:loc[0]] + "- Status: " + status + doc[loc[1]:], nil
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

// addSupersedes puts "- Supersedes: ADR-NNN" after the Owners line, one time only (ADR-009).
func addSupersedes(doc string, oldN int) (string, error) {
	line := fmt.Sprintf("- Supersedes: ADR-%03d", oldN)
	for l := range strings.SplitSeq(doc, "\n") {
		if strings.TrimSuffix(l, "\r") == line {
			return doc, nil
		}
	}
	loc := ownersLine.FindStringIndex(doc)
	if loc == nil {
		return "", errors.New("the new ADR has no Owners line")
	}
	insert := "\n" + line
	if doc[loc[1]-1] == '\r' { // the file has CRLF line ends
		insert += "\r"
	}
	return doc[:loc[1]] + insert + doc[loc[1]:], nil
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
	return docfile.Write(path, []byte(doc))
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
