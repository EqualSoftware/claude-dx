package task

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/EqualSoftware/claude-dx/internal/docs/docfile"
	"github.com/EqualSoftware/claude-dx/internal/markdown"
)

var (
	questionRow = regexp.MustCompile(`^\| (Q-(\d+)) \| (.*?) \| (.*?) \| (.*?) \|$`)
	dodItem     = regexp.MustCompile(`^- \[([ x])\] (.*)$`)
	sizeCell    = regexp.MustCompile("(?m)^\\| Size \\| `?([SML])`?(.*) \\|$")
)

// leafSections are the sections with no subsections. In Review needs all of them filled.
var leafSections = []string{"1", "2", "3", "4", "5.1", "5.2", "5.3", "6.1", "6.2", "7.1", "7.2", "7.3", "7.4", "7.5", "8", "9", "10", "11", "12", "13", "14"}

// pristine holds the template text of each leaf section.
var pristine = func() map[string]string {
	_, body := markdown.Split(templateText)
	m := map[string]string{}
	for _, ref := range leafSections {
		m[ref], _ = markdown.Section(body, ref)
	}
	return m
}()

// UpdateSection replaces one section.
func (s Store) UpdateSection(id, section, content string) error {
	return s.edit(id, func(d *doc) error {
		body, err := markdown.ReplaceSection(d.body, section, content)
		d.body = body
		return err
	})
}

// Update holds the frontmatter and table changes of UpdateFields. Nil or empty fields do not change.
type Update struct {
	Title     string
	Priority  string
	Area      []string
	DependsOn []string
	Fields    Fields
}

// UpdateFields changes the frontmatter and the table values.
func (s Store) UpdateFields(id string, u Update) error {
	if u.Priority != "" {
		if err := validPriority(u.Priority); err != nil {
			return err
		}
	}
	return s.edit(id, func(d *doc) error {
		if u.Title != "" {
			d.front = markdown.SetField(d.front, "title", u.Title)
			d.body = setHeader(d.body, kind.ID(mustNumber(d.path)), u.Title)
		}
		if u.Priority != "" {
			d.front = markdown.SetField(d.front, "priority", u.Priority)
		}
		if u.Area != nil {
			d.front = markdown.SetField(d.front, "area", u.Area)
		}
		if u.DependsOn != nil {
			d.front = markdown.SetField(d.front, "depends_on", u.DependsOn)
		}
		d.body = setFields(d.body, u.Fields)
		return nil
	})
}

// SetStatus changes the status, adds a Document History row and enforces the status rules.
func (s Store) SetStatus(id, status, note string) error {
	if !slices.Contains(Statuses, status) {
		return fmt.Errorf("status %q is not valid. Use one of: %s", status, strings.Join(Statuses, ", "))
	}
	return s.edit(id, func(d *doc) error {
		if problems := rules[status](d.body, note); len(problems) > 0 {
			return fmt.Errorf("cannot set %s: %s", status, strings.Join(problems, "; "))
		}
		d.front = markdown.SetField(d.front, "status", status)
		if note == "" {
			note = "Status: " + status + "."
		}
		return s.addHistory(d, note)
	})
}

// rules returns the problems that stop a change to each status.
var rules = map[string]func(body, note string) []string{
	"Draft":          none,
	"In Review":      inReview,
	"Approved":       func(body, _ string) []string { return openQuestions(body) },
	"In Development": none,
	"Blocked":        needNote,
	"Done":           func(body, _ string) []string { return uncheckedDoD(body) },
	"Cancelled":      needNote,
}

func none(string, string) []string { return nil }

func needNote(_, note string) []string {
	if strings.TrimSpace(note) == "" {
		return []string{"give a note with the reason"}
	}
	return nil
}

func inReview(body, _ string) []string {
	problems := openQuestions(body)
	for _, ref := range leafSections {
		if text, _ := markdown.Section(body, ref); text == pristine[ref] {
			problems = append(problems, "section "+ref+" has the template text. Fill it or write Not applicable")
		}
	}
	if m := sizeCell.FindStringSubmatch(body); m == nil || m[1] == "L" || strings.Contains(m[2], " / ") {
		problems = append(problems, "Size must be S or M. Split an L task into smaller tasks")
	}
	return problems
}

func openQuestions(body string) []string {
	var problems []string
	for _, q := range questions(body) {
		if q[5] == "Open" {
			problems = append(problems, q[1]+" is Open")
		}
	}
	return problems
}

func uncheckedDoD(body string) []string {
	var problems []string
	for i, item := range dodItems(body) {
		if item[1] == " " {
			problems = append(problems, fmt.Sprintf("Definition of Done item %d is not checked", i+1))
		}
	}
	return problems
}

// questions returns the rows of the Open Questions table.
func questions(body string) [][]string {
	section, _ := markdown.Section(body, "13")
	var rows [][]string
	for line := range strings.SplitSeq(section, "\n") {
		if m := questionRow.FindStringSubmatch(line); m != nil {
			rows = append(rows, m)
		}
	}
	return rows
}

func dodItems(body string) [][]string {
	section, _ := markdown.Section(body, "14")
	var items [][]string
	for line := range strings.SplitSeq(section, "\n") {
		if m := dodItem.FindStringSubmatch(line); m != nil {
			items = append(items, m)
		}
	}
	return items
}

// AddQuestion adds the next Q-n row with the answer Open and returns its ID.
func (s Store) AddQuestion(id, question, owner string) (string, error) {
	if owner == "" {
		owner = "Maintainer"
	}
	var qid string
	err := s.edit(id, func(d *doc) error {
		section, _ := markdown.Section(d.body, "13")
		lines, next := withoutPlaceholders(strings.Split(section, "\n"))
		qid = fmt.Sprintf("Q-%d", next)
		row := fmt.Sprintf("| %s | %s | %s | Open |", qid, cell(question), cell(owner))
		body, err := markdown.ReplaceSection(d.body, "13", strings.Join(insertRow(lines, row), "\n"))
		d.body = body
		return err
	})
	return qid, err
}

// withoutPlaceholders removes template rows ("[Question]", "None.") and returns the next question number.
func withoutPlaceholders(lines []string) ([]string, int) {
	var out []string
	next := 1
	for _, line := range lines {
		m := questionRow.FindStringSubmatch(line)
		if m != nil && (strings.HasPrefix(m[3], "[") || m[3] == "None.") {
			continue
		}
		if m != nil {
			n, _ := strconv.Atoi(m[2]) // The pattern allows only digits.
			next = max(next, n+1)
		}
		out = append(out, line)
	}
	return out, next
}

// insertRow puts row after the last row of the first table in lines.
func insertRow(lines []string, row string) []string {
	at := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "|") {
			at = i
		}
	}
	return slices.Insert(lines, at+1, row)
}

// AnswerQuestion writes the answer of one question.
func (s Store) AnswerQuestion(id, qid, answer string) error {
	return s.edit(id, func(d *doc) error {
		section, _ := markdown.Section(d.body, "13")
		lines := strings.Split(section, "\n")
		for i, line := range lines {
			m := questionRow.FindStringSubmatch(line)
			if m == nil || !strings.EqualFold(m[1], qid) {
				continue
			}
			lines[i] = fmt.Sprintf("| %s | %s | %s | %s |", m[1], m[3], m[4], cell(answer))
			body, err := markdown.ReplaceSection(d.body, "13", strings.Join(lines, "\n"))
			d.body = body
			return err
		}
		return fmt.Errorf("question %s does not exist", qid)
	})
}

// CheckDoD checks one Definition of Done item (1 is the first item) and adds the evidence.
func (s Store) CheckDoD(id string, item int, evidence string) error {
	if strings.TrimSpace(evidence) == "" {
		return errors.New("give the evidence")
	}
	return s.edit(id, func(d *doc) error {
		section, _ := markdown.Section(d.body, "14")
		lines := strings.Split(section, "\n")
		count := 0
		for i, line := range lines {
			m := dodItem.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			if count++; count != item {
				continue
			}
			if m[1] == "x" {
				return fmt.Errorf("item %d is already checked", item)
			}
			lines[i] = "- [x] " + m[2] + " " + strings.TrimSpace(evidence)
			body, err := markdown.ReplaceSection(d.body, "14", strings.Join(lines, "\n"))
			d.body = body
			return err
		}
		return fmt.Errorf("item %d does not exist. The Definition of Done has %d items", item, count)
	})
}

// Renumber gives the task in file the next free number. It returns the new ID and the tasks
// that refer to the old ID. It does not change them, because the old ID can name more than one file.
func (s Store) Renumber(file string) (string, []string, error) {
	file = filepath.Base(file)
	if !kind.Pattern.MatchString(file) {
		return "", nil, fmt.Errorf("%s is not a task file name", file)
	}
	data, err := os.ReadFile(filepath.Join(s.Dir, file))
	if err != nil {
		return "", nil, err
	}
	oldID := kind.ID(mustNumber(file))
	front, body := markdown.Split(string(data))
	var f Front
	_ = markdown.Decode(front, &f)
	var newID string
	_, err = s.Alloc.Next("task", s.Dir, kind.Pattern, func(n int) error {
		newID = kind.ID(n)
		doc := markdown.Join(markdown.SetField(front, "id", newID), setHeader(body, newID, f.Title))
		return docfile.Create(filepath.Join(s.Dir, newID+strings.TrimPrefix(file, oldID)), []byte(doc))
	})
	if err != nil {
		return "", nil, err
	}
	if err := os.Remove(filepath.Join(s.Dir, file)); err != nil {
		return "", nil, err
	}
	refs, err := s.references(oldID)
	if err != nil {
		return "", nil, err
	}
	return newID, refs, s.WriteIndex()
}

// references returns the task files that contain id.
func (s Store) references(id string) ([]string, error) {
	files, err := kind.Files(s.Dir)
	if err != nil {
		return nil, err
	}
	var refs []string
	for _, name := range sortedNames(files) {
		data, err := os.ReadFile(filepath.Join(s.Dir, name))
		if err != nil {
			return nil, err
		}
		if strings.Contains(string(data), id) {
			refs = append(refs, name)
		}
	}
	return refs, nil
}

// mustNumber returns the number in a task file name or path that matches the pattern.
func mustNumber(path string) int {
	m := kind.Pattern.FindStringSubmatch(filepath.Base(path))
	if m == nil {
		return 0
	}
	n, _ := strconv.Atoi(m[1]) // The pattern allows only digits.
	return n
}
