// Package glossary manages the term table in docs/GLOSSARY.md (ADR-006).
package glossary

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/EqualSoftware/claude-dx/internal/docs/docfile"
)

// templateText is the start of a new glossary: the header text and the table header, with no rows.
//
//go:embed template.md
var templateText string

// ErrNoFile means that docs/GLOSSARY.md does not exist. It is a normal state, not a failure.
var ErrNoFile = errors.New("no glossary. docs/GLOSSARY.md does not exist")

const (
	tableHeader = "| Term | Definition | Area | Do not use | Source |"
	empty       = "—" // the Do not use cell with no words
	pipe        = "\x00"
)

// Store reads and writes the glossary of one project.
type Store struct {
	Path string // absolute path of docs/GLOSSARY.md
}

// Row is one term of the glossary.
type Row struct {
	Term       string
	Definition string
	Area       string
	DoNotUse   []string
	Source     string
}

// table is the glossary file: the text before the rows (with the table header), the rows and the text after them.
type table struct {
	head string
	rows []Row
	tail string
}

// Lookup returns the row of term, or a message if term is a word in Do not use.
// It compares the full term with no case difference. ok is false if no row matches.
func (s Store) Lookup(term string) (text string, ok bool, err error) {
	t, err := s.read()
	if err != nil {
		return "", false, err
	}
	for _, r := range t.rows {
		if same(r.Term, term) {
			return r.line(), true, nil
		}
		if i := slices.IndexFunc(r.DoNotUse, func(w string) bool { return same(w, term) }); i >= 0 {
			return fmt.Sprintf("Use %q, not %q.", r.Term, r.DoNotUse[i]), true, nil
		}
	}
	return "", false, nil
}

// List returns the rows of the glossary in file order. A non-empty area keeps only the rows of
// that area, with no case difference.
func (s Store) List(area string) ([]Row, error) {
	t, err := s.read()
	if err != nil || area == "" {
		return t.rows, err
	}
	var rows []Row
	for _, r := range t.rows {
		if same(r.Area, area) {
			rows = append(rows, r)
		}
	}
	return rows, nil
}

// Upsert adds the row, or replaces the row with the same term, and sorts the rows by term.
// A replaced row keeps the old spelling of the term, and the old Do not use words if row.DoNotUse
// is nil. An empty list removes the words. It makes the file if it does not exist. added is true
// for a new term.
func (s Store) Upsert(row Row) (added bool, err error) {
	t, err := s.read()
	if errors.Is(err, ErrNoFile) {
		t, err = parse(templateText)
	}
	if err != nil {
		return false, err
	}
	i := t.index(row.Term)
	if i >= 0 {
		row.Term = t.rows[i].Term
		if row.DoNotUse == nil {
			row.DoNotUse = t.rows[i].DoNotUse
		}
	}
	if err := conflict(t.rows, row); err != nil {
		return false, err
	}
	if i < 0 {
		t.rows = append(t.rows, row)
	} else {
		t.rows[i] = row
	}
	slices.SortStableFunc(t.rows, func(a, b Row) int {
		return strings.Compare(strings.ToLower(a.Term), strings.ToLower(b.Term))
	})
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		return false, err
	}
	return i < 0, s.write(t)
}

// Delete removes the row of term.
func (s Store) Delete(term string) error {
	t, err := s.read()
	if err != nil {
		return err
	}
	i := t.index(term)
	if i < 0 {
		return fmt.Errorf("%q is not a term of the glossary. Use glossary_lookup", term)
	}
	t.rows = slices.Delete(t.rows, i, i+1)
	return s.write(t)
}

// conflict refuses a term that is a Do not use word of another row, and a Do not use word that
// is the term of another row, because one meaning has one term.
func conflict(rows []Row, row Row) error {
	for _, r := range rows {
		if same(r.Term, row.Term) {
			continue
		}
		if slices.ContainsFunc(r.DoNotUse, func(w string) bool { return same(w, row.Term) }) {
			return fmt.Errorf("%q is in Do not use of the term %q. Use %q, or change that row first", row.Term, r.Term, r.Term)
		}
		if slices.ContainsFunc(row.DoNotUse, func(w string) bool { return same(w, r.Term) }) {
			return fmt.Errorf("do_not_use word %q is a term of the glossary. Change or delete that row first", r.Term)
		}
	}
	return nil
}

func (s Store) read() (table, error) {
	data, err := os.ReadFile(s.Path)
	if docfile.NotExist(err) {
		return table{}, ErrNoFile
	}
	if err != nil {
		return table{}, err
	}
	return parse(string(data))
}

func (s Store) write(t table) error {
	var b strings.Builder
	b.WriteString(t.head)
	for _, r := range t.rows {
		b.WriteString(r.line())
		b.WriteByte('\n')
	}
	b.WriteString(t.tail)
	return os.WriteFile(s.Path, []byte(b.String()), 0o644)
}

// parse splits the file into the text before the rows, the rows and the text after the rows.
func parse(src string) (table, error) {
	lines := strings.SplitAfter(strings.ReplaceAll(src, "\r\n", "\n"), "\n")
	start := slices.IndexFunc(lines, func(l string) bool { return strings.TrimSpace(l) == tableHeader })
	if start < 0 || start+1 >= len(lines) {
		return table{}, fmt.Errorf("docs/GLOSSARY.md has no table with the header %s", tableHeader)
	}
	t := table{head: strings.Join(lines[:start+2], "")}
	end := start + 2
	for ; end < len(lines) && strings.HasPrefix(lines[end], "|"); end++ {
		row, err := parseRow(lines[end])
		if err != nil {
			return table{}, fmt.Errorf("docs/GLOSSARY.md line %d: %w", end+1, err)
		}
		t.rows = append(t.rows, row)
	}
	t.tail = strings.Join(lines[end:], "")
	return t, nil
}

func parseRow(line string) (Row, error) {
	line = strings.TrimSpace(strings.ReplaceAll(line, `\|`, pipe))
	cells := strings.Split(strings.TrimSuffix(strings.TrimPrefix(line, "|"), "|"), "|")
	if len(cells) != 5 {
		return Row{}, fmt.Errorf("a row must have 5 cells, it has %d", len(cells))
	}
	for i, c := range cells {
		cells[i] = strings.ReplaceAll(strings.TrimSpace(c), pipe, "|")
	}
	return Row{Term: cells[0], Definition: cells[1], Area: cells[2], DoNotUse: words(cells[3]), Source: cells[4]}, nil
}

// words splits a Do not use cell into its words.
func words(cell string) []string {
	if cell == empty || cell == "" {
		return nil
	}
	list := strings.Split(cell, ",")
	for i, w := range list {
		list[i] = strings.TrimSpace(w)
	}
	return list
}

// line returns the row as a Markdown table line.
func (r Row) line() string {
	doNotUse := strings.Join(r.DoNotUse, ", ")
	if doNotUse == "" {
		doNotUse = empty
	}
	cells := []string{r.Term, r.Definition, r.Area, doNotUse, r.Source}
	for i, c := range cells {
		cells[i] = cell(c)
	}
	return "| " + strings.Join(cells, " | ") + " |"
}

// cell escapes "|" and puts all lines of text on one line, because a table row is one line.
func cell(text string) string {
	text = strings.ReplaceAll(text, "|", `\|`)
	return strings.Join(strings.Fields(text), " ")
}

func (t table) index(term string) int {
	return slices.IndexFunc(t.rows, func(r Row) bool { return same(r.Term, term) })
}

func same(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}
