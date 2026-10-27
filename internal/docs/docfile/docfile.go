// Package docfile holds the file rules that ADRs and tasks share: IDs, numbered file names and exclusive create.
package docfile

import (
	"errors"
	"fmt"
	"io/fs"
	"math/rand/v2"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

// MaxTitle is the maximum length of a task or ADR title, so a file name is short (ADR-002).
const MaxTitle = 40

// CheckTitle refuses an empty title, a title of more than MaxTitle characters and a title with an empty
// slug, because the file name must not be "TASK-NNN-.md".
func CheckTitle(title string) error {
	if err := CheckLine(title, "title"); err != nil {
		return err
	}
	n := utf8.RuneCountInString(strings.TrimSpace(title))
	if n == 0 {
		return errors.New("give a title")
	}
	if n > MaxTitle {
		return fmt.Errorf("the title has %d characters. The limit is %d. Write a shorter title", n, MaxTitle)
	}
	if Slug(title) == "" {
		return errors.New("give a title with ASCII letters or digits")
	}
	return nil
}

// CheckLine refuses text with a line break, because name must be one line.
func CheckLine(text, name string) error {
	if strings.ContainsAny(text, "\r\n") {
		return fmt.Errorf("the %s has a line break. Write the %s on one line", name, name)
	}
	return nil
}

// CheckNewTitle does CheckTitle, and refuses a title that another document has with no case
// difference, because one title names one thing. titles maps a document ID to its title.
func CheckNewTitle(title string, titles map[string]string) error {
	if err := CheckTitle(title); err != nil {
		return err
	}
	ids := make([]string, 0, len(titles))
	for id := range titles {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	for _, id := range ids {
		if strings.EqualFold(strings.TrimSpace(titles[id]), strings.TrimSpace(title)) {
			return fmt.Errorf("%s has the same title. Give a different title, or change %s", id, id)
		}
	}
	return nil
}

// Kind describes one type of numbered document, for example ADR-NNN-<slug>.md.
type Kind struct {
	Prefix  string         // "ADR" or "TASK"
	Pattern *regexp.Regexp // file name pattern; group 1 is the number
	id      *regexp.Regexp
}

// NewKind returns the kind for prefix.
func NewKind(prefix string) Kind {
	return Kind{
		Prefix:  prefix,
		Pattern: regexp.MustCompile(`^` + prefix + `-(\d{3})-.*\.md$`),
		id:      regexp.MustCompile(`^(?i:` + prefix + `-)?(\d{1,3})$`),
	}
}

// ID formats n, for example ADR-007.
func (k Kind) ID(n int) string {
	return fmt.Sprintf("%s-%03d", k.Prefix, n)
}

// FileName returns the file name for n and title.
func (k Kind) FileName(n int, title string) string {
	return k.ID(n) + "-" + Slug(title) + ".md"
}

// Parse returns the number in id. It accepts "ADR-007", "adr-7", "7" and "007".
func (k Kind) Parse(id string) (int, error) {
	m := k.id.FindStringSubmatch(strings.TrimSpace(id))
	if m == nil {
		return 0, fmt.Errorf("%q is not a %s ID. Example: %s", id, k.Prefix, k.ID(7))
	}
	n, _ := strconv.Atoi(m[1])
	return n, nil
}

// Files returns the numbered files in dir, by number. fs.ErrNotExist means that dir does not exist.
func (k Kind) Files(dir string) (map[int][]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	files := map[int][]string{}
	for _, e := range entries {
		m := k.Pattern.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		n, _ := strconv.Atoi(m[1])
		files[n] = append(files[n], e.Name())
	}
	return files, nil
}

// Find returns the path of the file with the number in id. It fails if no file or many files have it.
func (k Kind) Find(dir, id string) (string, error) {
	n, err := k.Parse(id)
	if err != nil {
		return "", err
	}
	files, err := k.Files(dir)
	if err != nil {
		return "", err
	}
	switch len(files[n]) {
	case 0:
		return "", fmt.Errorf("%s does not exist", k.ID(n))
	case 1:
		return filepath.Join(dir, files[n][0]), nil
	}
	return "", fmt.Errorf("%s has %d files: %s. Give each file a different number first", k.ID(n), len(files[n]), strings.Join(files[n], ", "))
}

// Duplicates returns one warning line for each number that has more than one file.
func Duplicates(k Kind, files map[int][]string) string {
	var b strings.Builder
	for n, names := range files {
		if len(names) > 1 {
			fmt.Fprintf(&b, "Warning: %s has %d files. Give each file a different number.\n", k.ID(n), len(names))
		}
	}
	return b.String()
}

// link makes a hard link. A test replaces it.
var link = os.Link

// Create makes a new file with data. It fails if the file exists, so two writers never use one number.
// It writes a temporary file, then links it to path: a crash never leaves a partial file at path.
func Create(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	cleanStale(filepath.Dir(path))
	tmp := filepath.Join(filepath.Dir(path), fmt.Sprintf(".%s.tmp-%d", filepath.Base(path), rand.Uint64()))
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644) // the umask applies
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp) }()
	_, err = f.Write(data)
	err = errors.Join(err, f.Sync(), f.Close())
	if err != nil {
		return err
	}
	err = link(tmp, path)
	if err == nil || errors.Is(err, fs.ErrExist) {
		return err
	}
	return createExclusive(path, data) // the file system has no hard links
}

// createExclusive makes path with O_EXCL. A write failure removes the partial file.
func createExclusive(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	err = errors.Join(err, f.Sync(), f.Close())
	if err != nil {
		_ = os.Remove(path)
	}
	return err
}

// NotExist reports whether err means that a file or directory does not exist.
func NotExist(err error) bool {
	return errors.Is(err, fs.ErrNotExist)
}

// Cell makes text safe for one Markdown table cell: it escapes "|" and puts a space for each line break.
func Cell(text string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(text), "\r", ""), "\n", " "), "|", `\|`)
}

// Slug makes the file name part from a title: lowercase ASCII words joined by "-", maximum MaxTitle
// characters, cut at a word boundary. The title limit makes the cut a guard for old titles only.
func Slug(title string) string {
	words := strings.FieldsFunc(strings.ToLower(title), func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9')
	})
	slug := strings.Join(words, "-")
	if len(slug) <= MaxTitle {
		return slug
	}
	cut := strings.LastIndexByte(slug[:MaxTitle+1], '-')
	if cut <= 0 {
		return slug[:MaxTitle]
	}
	return slug[:cut]
}
