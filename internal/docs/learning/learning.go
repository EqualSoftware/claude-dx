// Package learning manages the learnings in docs/learnings (ADR-002).
package learning

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/EqualSoftware/claude-dx/internal/docs/docfile"
	"github.com/EqualSoftware/claude-dx/internal/markdown"
)

// ErrNoDir means that docs/learnings does not exist. It is a normal state, not a failure.
var ErrNoDir = errors.New("no learnings. docs/learnings does not exist")

var topicPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// Store reads and writes the learning files of one project.
type Store struct {
	Dir  string // absolute path of docs/learnings
	Lock string // path of the docs lock (docs-common#R-10, R-11); required
}

// Learning is one learning file in a list.
type Learning struct {
	Topic       string
	Description string
}

// List returns the learnings by topic, and one warning line for each file that it cannot read or parse. It
// skips a bad file, so one bad file does not hide the other learnings.
func (s Store) List() ([]Learning, string, error) {
	entries, err := os.ReadDir(s.Dir)
	if docfile.NotExist(err) {
		return nil, "", ErrNoDir
	}
	if err != nil {
		return nil, "", err
	}
	var list []Learning
	var warnings string
	for _, e := range entries {
		topic, ok := strings.CutSuffix(e.Name(), ".md")
		if !ok || e.IsDir() || strings.EqualFold(e.Name(), indexFile) {
			continue
		}
		description, err := s.description(e.Name())
		if err != nil {
			warnings += fmt.Sprintf("Warning: %s: %v.\n", e.Name(), err)
			continue
		}
		list = append(list, Learning{Topic: topic, Description: description})
	}
	return list, warnings, nil
}

func (s Store) description(name string) (string, error) {
	data, err := os.ReadFile(filepath.Join(s.Dir, name))
	if err != nil {
		return "", err
	}
	front, _ := markdown.Split(string(data))
	var f struct {
		Description string `yaml:"description"`
	}
	if err := markdown.Decode(front, &f); err != nil {
		return "", fmt.Errorf("the frontmatter is not valid: %w", err)
	}
	return f.Description, nil
}

// Get returns the learning file.
func (s Store) Get(topic string) (string, error) {
	path, err := s.find(topic)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	return string(data), err
}

// Create writes a new learning file and makes README.md again. It fails if the topic exists.
func (s Store) Create(topic, description, body string) error {
	return docfile.Locked(s.Lock, func() error { return s.create(topic, description, body) })
}

func (s Store) create(topic, description, body string) error {
	if err := validTopic(topic); err != nil {
		return err
	}
	if strings.TrimSpace(description) == "" {
		return errors.New("give the description")
	}
	if strings.TrimSpace(body) == "" {
		return errors.New("give the body")
	}
	if err := markdown.CheckFence(body, "the body"); err != nil {
		return err
	}
	front := markdown.SetField("topic: "+topic+"\n", "description", oneLine(description))
	err := docfile.Create(filepath.Join(s.Dir, topic+".md"), []byte(markdown.Join(front, content(body))))
	if errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("learning %q exists. Use learning_update to change it", topic)
	}
	if err != nil {
		return err
	}
	return s.WriteIndex()
}

// Update changes the description, or replaces the body, or both, and makes README.md again.
func (s Store) Update(topic, description, body string) error {
	return docfile.Locked(s.Lock, func() error { return s.update(topic, description, body) })
}

func (s Store) update(topic, description, body string) error {
	if description == "" && body == "" {
		return errors.New("give a description, or a body, or both")
	}
	if err := checkUpdate(description, body); err != nil {
		return err
	}
	path, err := s.find(topic)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	front, old := markdown.Split(string(data))
	if description != "" {
		front = markdown.SetField(front, "description", oneLine(description))
	}
	if body != "" {
		old = content(body)
	}
	if err := docfile.Write(path, []byte(markdown.Join(front, old))); err != nil {
		return err
	}
	return s.WriteIndex()
}

// checkUpdate refuses a blank description and a body with an unclosed code fence. Empty means no change.
func checkUpdate(description, body string) error {
	if description != "" && strings.TrimSpace(description) == "" {
		return errors.New("give the description")
	}
	return markdown.CheckFence(body, "the body")
}

// Delete removes the learning file and makes README.md again.
func (s Store) Delete(topic string) error {
	return docfile.Locked(s.Lock, func() error { return s.delete(topic) })
}

func (s Store) delete(topic string) error {
	path, err := s.find(topic)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		return err
	}
	return s.WriteIndex()
}

// find returns the path of the learning file of topic.
func (s Store) find(topic string) (string, error) {
	if err := validTopic(topic); err != nil {
		return "", err
	}
	if _, err := os.Stat(s.Dir); docfile.NotExist(err) {
		return "", ErrNoDir
	}
	path := filepath.Join(s.Dir, topic+".md")
	if _, err := os.Stat(path); docfile.NotExist(err) {
		return "", fmt.Errorf("learning %q does not exist. Use learning_list", topic)
	}
	return path, nil
}

func validTopic(topic string) error {
	if strings.EqualFold(topic+".md", indexFile) {
		return fmt.Errorf("topic %q is not valid, because %s is the index of docs/learnings", topic, indexFile)
	}
	if topicPattern.MatchString(topic) {
		return nil
	}
	return fmt.Errorf("topic %q is not kebab-case. Example: go-embed-paths", topic)
}

// oneLine joins the lines of a description, because the frontmatter value is one line.
func oneLine(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

// content returns the body of a file: one empty line after the frontmatter, and one new line at the end.
func content(body string) string {
	return "\n" + strings.TrimSpace(body) + "\n"
}
