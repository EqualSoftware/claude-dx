// Package spike manages the local spikes in the project data directory (ADR-018).
package spike

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/EqualSoftware/claude-dx/internal/docs/docfile"
	"github.com/EqualSoftware/claude-dx/internal/markdown"
)

// Status values of a spike.
const (
	Open   = "Open"
	Closed = "Closed"
)

// maxSlug is the maximum length of a spike slug.
const maxSlug = 40

// ErrNoDir means that the spikes folder does not exist. It is a normal state, not a failure.
var ErrNoDir = errors.New("no spikes. The spikes folder does not exist")

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// now gives the current time. A test replaces it.
var now = time.Now

// Store reads and writes the spike files of one project.
type Store struct {
	Dir  string // absolute path of <data dir>/spikes
	Lock string // path of the docs lock (docs-spike#R-9); required
}

// Spike is one spike in a list.
type Spike struct {
	Slug    string `yaml:"slug"`
	Title   string `yaml:"title"`
	Status  string `yaml:"status"`
	Created string `yaml:"created"`
}

func today() string { return now().Format("2006-01-02") }

// Create writes a new Open spike. It fails if the slug exists.
func (s Store) Create(slug, title, question string) error {
	return docfile.Locked(s.Lock, func() error { return s.create(slug, title, question) })
}

func (s Store) create(slug, title, question string) error {
	if err := validSlug(slug); err != nil {
		return err
	}
	if err := docfile.CheckTitle(title); err != nil {
		return err
	}
	if strings.TrimSpace(question) == "" {
		return errors.New("give the question")
	}
	if err := markdown.CheckNew(question, "the question", 2); err != nil {
		return err
	}
	front := "slug: " + slug + "\n"
	front = markdown.SetField(front, "title", strings.TrimSpace(title))
	front += "status: " + Open + "\ncreated: " + today() + "\n"
	body := "\n## Question\n\n" + strings.TrimSpace(question) + "\n\n## Notes\n\n## Result\n"
	err := docfile.Create(filepath.Join(s.Dir, slug+".md"), []byte(markdown.Join(front, body)))
	if errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("spike %q exists. Use spike_note to add a finding", slug)
	}
	return err
}

// Note adds one finding to the Notes of an Open spike.
func (s Store) Note(slug, text string) error {
	return docfile.Locked(s.Lock, func() error { return s.note(slug, text) })
}

func (s Store) note(slug, text string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return errors.New("give the text")
	}
	if err := docfile.CheckLine(text, "text"); err != nil {
		return err
	}
	path, front, body, err := s.open(slug)
	if err != nil {
		return err
	}
	notes, _ := markdown.Section(body, "Notes")
	item := "- " + today() + ": " + text
	if notes != "" {
		item = notes + "\n" + item
	}
	return save(path, front, body, "Notes", item)
}

// Close writes the Result and sets the status of an Open spike to Closed.
func (s Store) Close(slug, result string) error {
	return docfile.Locked(s.Lock, func() error { return s.close(slug, result) })
}

func (s Store) close(slug, result string) error {
	if strings.TrimSpace(result) == "" {
		return errors.New("give the result")
	}
	path, front, body, err := s.open(slug)
	if err != nil {
		return err
	}
	return save(path, markdown.SetField(front, "status", Closed), body, "Result", strings.TrimSpace(result))
}

// save replaces a section of body and writes the file.
func save(path, front, body, section, content string) error {
	if err := markdown.CheckContent(body, content, section, 2); err != nil {
		return err
	}
	body, err := markdown.ReplaceSection(body, section, content)
	if err != nil {
		return err
	}
	return docfile.Write(path, []byte(markdown.Join(front, body)))
}

// open reads an Open spike. It refuses a Closed spike.
func (s Store) open(slug string) (path, front, body string, err error) {
	path, err = s.find(slug)
	if err != nil {
		return "", "", "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", "", err
	}
	front, body = markdown.Split(string(data))
	var f Spike
	if err := markdown.Decode(front, &f); err != nil {
		return "", "", "", fmt.Errorf("the frontmatter of spike %q is not valid: %w", slug, err)
	}
	if f.Status == Closed {
		return "", "", "", fmt.Errorf("spike %q is Closed. A Closed spike does not change. Make a new spike", slug)
	}
	return path, front, body, nil
}

// List returns the spikes, Open first and then by slug, and one warning line for each file that it cannot
// read or parse. status filters the list; empty means all.
func (s Store) List(status string) ([]Spike, string, error) {
	if status != "" && status != Open && status != Closed {
		return nil, "", fmt.Errorf("status %q is not valid. Use %s or %s", status, Open, Closed)
	}
	var list []Spike
	var warnings string
	err := docfile.RLocked(s.Lock, func() (err error) {
		list, warnings, err = s.scan(status)
		return err
	})
	slices.SortFunc(list, compare)
	return list, warnings, err
}

// scan reads all spike files with the status. The caller holds the docs lock.
func (s Store) scan(status string) (list []Spike, warnings string, err error) {
	entries, err := os.ReadDir(s.Dir)
	if docfile.NotExist(err) {
		return nil, "", ErrNoDir
	}
	if err != nil {
		return nil, "", err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		sp, err := readSpike(filepath.Join(s.Dir, e.Name()))
		if err != nil {
			warnings += fmt.Sprintf("Warning: %s: %v.\n", e.Name(), err)
			continue
		}
		if status == "" || sp.Status == status {
			list = append(list, sp)
		}
	}
	return list, warnings, nil
}

// compare orders Open spikes before the others, then by slug.
func compare(a, b Spike) int {
	if (a.Status == Open) != (b.Status == Open) {
		if a.Status == Open {
			return -1
		}
		return 1
	}
	return strings.Compare(a.Slug, b.Slug)
}

func readSpike(path string) (Spike, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Spike{}, err
	}
	front, _ := markdown.Split(string(data))
	var sp Spike
	if err := markdown.Decode(front, &sp); err != nil {
		return Spike{}, fmt.Errorf("the frontmatter is not valid: %w", err)
	}
	return sp, nil
}

// Get returns the spike file, or one section of it: Question, Notes or Result.
func (s Store) Get(slug, section string) (text string, err error) {
	err = docfile.RLocked(s.Lock, func() error {
		path, err := s.find(slug)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text = string(data)
		if section == "" {
			return nil
		}
		_, body := markdown.Split(text)
		var ok bool
		if text, ok = markdown.Section(body, section); !ok {
			return fmt.Errorf("section %q does not exist. Use Question, Notes or Result", section)
		}
		return nil
	})
	return text, err
}

// Delete removes spikes of any status. It changes nothing if one slug does not exist.
func (s Store) Delete(slugs []string) error {
	return docfile.Locked(s.Lock, func() error { return s.delete(slugs) })
}

func (s Store) delete(slugs []string) error {
	if len(slugs) == 0 {
		return errors.New("give at least one slug")
	}
	paths := make([]string, 0, len(slugs))
	for _, slug := range slugs {
		path, err := s.find(slug)
		if err != nil {
			return err
		}
		paths = append(paths, path)
	}
	for _, path := range paths {
		if err := os.Remove(path); err != nil && !docfile.NotExist(err) {
			return err
		}
	}
	return nil
}

// find returns the path of the spike file of slug.
func (s Store) find(slug string) (string, error) {
	if err := validSlug(slug); err != nil {
		return "", err
	}
	if _, err := os.Stat(s.Dir); docfile.NotExist(err) {
		return "", ErrNoDir
	}
	path := filepath.Join(s.Dir, slug+".md")
	if _, err := os.Stat(path); docfile.NotExist(err) {
		return "", fmt.Errorf("spike %q does not exist. Use spike_list", slug)
	}
	return path, nil
}

func validSlug(slug string) error {
	if !slugPattern.MatchString(slug) {
		return fmt.Errorf("slug %q is not kebab-case. Example: lock-on-nfs", slug)
	}
	if len(slug) > maxSlug {
		return fmt.Errorf("the slug has %d characters. The limit is %d", len(slug), maxSlug)
	}
	return nil
}
