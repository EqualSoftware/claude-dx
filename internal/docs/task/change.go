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
)

// leafSections are the sections with no subsections. In Review needs all of them filled.
// Sections 13 and 14 are not in the list, because their template text is valid and only their tools change them.
var leafSections = []string{"1", "2", "3", "4", "5.1", "5.2", "5.3", "6.1", "6.2", "7.1", "7.2", "7.3", "7.4", "7.5", "8", "9", "10", "11", "12"}

// noQuestions is the template row of section 13. task_question_add removes it.
const noQuestions = "| — | None. | — | — |"

// protected are the sections that only their own tools change, so the flow checks of ADR-013 stay true.
var protected = []struct{ number, title, tools string }{
	{"", "Document History", "task_set_status"},
	{"13", "Open Questions", "task_question_add and task_question_answer"},
	{"14", "Definition of Done", "task_dod_add, task_dod_remove, task_dod_check and task_dod_uncheck"},
}

// checkSection refuses a protected section.
func checkSection(ref string) error {
	for _, p := range protected {
		if ref == p.number || strings.EqualFold(ref, p.title) {
			return fmt.Errorf("section %q changes only with %s", ref, p.tools)
		}
	}
	return nil
}

// pristine holds the template text of each leaf section.
var pristine = func() map[string]string {
	_, body := markdown.Split(templateText)
	m := map[string]string{}
	for _, ref := range leafSections {
		m[ref], _ = markdown.Section(body, ref)
	}
	return m
}()

// scopeSections define the scope of a task: Goals, Out of Scope, Requirements, Design and
// Acceptance Criteria (ADR-013, step 3).
var scopeSections = []string{"3", "4", "6", "7", "9"}

// UpdateSection replaces one section. A scope change of an Approved task sets it to In Review,
// because the user must approve the new scope. review is true after that status change.
func (s Store) UpdateSection(id, section, content string) (review bool, err error) {
	if err := checkSection(section); err != nil {
		return false, err
	}
	err = s.edit(id, func(d *doc) error {
		body, err := markdown.ReplaceSection(d.body, section, content)
		if err != nil {
			return err
		}
		var f Front
		if err := markdown.Decode(d.front, &f); err != nil {
			return fmt.Errorf("bad frontmatter: %w", err)
		}
		old := d.body
		d.body = body
		if f.Status != "Approved" || !scopeChanged(old, body) {
			return nil
		}
		review = true
		d.front = markdown.SetField(d.front, "status", "In Review")
		return s.addHistory(d, "Status: In Review. The scope changed in section "+section+".")
	})
	return review, err
}

func scopeChanged(old, body string) bool {
	return slices.ContainsFunc(scopeSections, func(ref string) bool {
		a, _ := markdown.Section(old, ref)
		b, _ := markdown.Section(body, ref)
		return a != b
	})
}

// Update holds the frontmatter changes of UpdateFields. Nil or empty fields do not change.
type Update struct {
	Title     string
	Priority  string
	Milestone string
	Type      string
	Size      string
	Owner     string
	Platforms []string
	Area      []string
	DependsOn []string
	Related   []string
}

// check returns an error for an invalid field.
func (u Update) check() error {
	if u.Title != "" {
		if err := docfile.CheckTitle(u.Title); err != nil {
			return err
		}
	}
	if u.Milestone != "" {
		if err := checkMilestone(u.Milestone); err != nil {
			return err
		}
	}
	for _, c := range []struct {
		name, value string
		valid       []string
	}{{"priority", u.Priority, Priorities}, {"type", u.Type, Types}, {"size", u.Size, Sizes}} {
		if c.value == "" {
			continue
		}
		if err := oneOf(c.name, c.value, c.valid); err != nil {
			return err
		}
	}
	return nil
}

// field is one frontmatter change. set is false for a field that does not change.
type field struct {
	key   string
	value any
	set   bool
}

// fields returns the frontmatter changes in frontmatter order.
func (u Update) fields() []field {
	return []field{
		{"title", u.Title, u.Title != ""},
		{"priority", u.Priority, u.Priority != ""},
		{"milestone", u.Milestone, u.Milestone != ""},
		{"type", u.Type, u.Type != ""},
		{"size", u.Size, u.Size != ""},
		{"owner", u.Owner, u.Owner != ""},
		{"platforms", u.Platforms, u.Platforms != nil},
		{"area", u.Area, u.Area != nil},
		{"depends_on", u.DependsOn, u.DependsOn != nil},
		{"related", u.Related, u.Related != nil},
	}
}

// UpdateFields changes the frontmatter fields and returns the new file name and the warnings. A new depends_on
// list is checked against the milestone and the priority after the change. A new title renames the file to
// the slug of the title, so the file name and the title agree. file is empty if the name does not change.
func (s Store) UpdateFields(id string, u Update) (file, warnings string, err error) {
	if err := u.check(); err != nil {
		return "", "", err
	}
	err = s.edit(id, func(d *doc) error {
		var old Front
		if u.DependsOn != nil {
			_ = markdown.Decode(d.front, &old) // a bad frontmatter fails after the change
		}
		for _, f := range u.fields() {
			if f.set {
				d.front = markdown.SetField(d.front, f.key, f.value)
			}
		}
		if u.Title != "" {
			file = retitle(d, u.Title)
		}
		if u.DependsOn == nil {
			return nil
		}
		var f Front
		if err := markdown.Decode(d.front, &f); err != nil {
			return fmt.Errorf("bad frontmatter: %w", err)
		}
		f.ID = kind.ID(mustNumber(d.path))
		var err error
		warnings, err = s.checkDeps(f, old.DependsOn)
		return err
	})
	if err != nil {
		return "", "", err
	}
	return file, warnings, nil
}

// retitle sets the header of d to title and the path of d to the file name of title. It returns the new
// file name, or "" if the name does not change.
func retitle(d *doc, title string) string {
	n := mustNumber(d.path)
	d.body = setHeader(d.body, kind.ID(n), title)
	name := kind.FileName(n, title)
	if name == filepath.Base(d.path) {
		return ""
	}
	d.path = filepath.Join(filepath.Dir(d.path), name)
	return name
}

// SetStatus changes the status, adds a Document History row and enforces the status rules.
// For In Review and Approved, it returns a warning for each open task with an overlap of plan files.
func (s Store) SetStatus(id, status, note string) (string, error) {
	if !slices.Contains(Statuses, status) {
		return "", fmt.Errorf("status %q is not valid. Use one of: %s", status, strings.Join(Statuses, ", "))
	}
	var warnings string
	err := s.edit(id, func(d *doc) error {
		var f Front
		if err := markdown.Decode(d.front, &f); err != nil {
			return fmt.Errorf("bad frontmatter: %w", err)
		}
		if err := allowedChange(f.Status, status); err != nil {
			return err
		}
		if problems := rules[status](f, d.body, note); len(problems) > 0 {
			return fmt.Errorf("cannot set %s: %s", status, strings.Join(problems, "; "))
		}
		if status == "In Review" || status == "Approved" {
			var err error
			if warnings, err = s.overlapWarnings(f, d.body); err != nil {
				return err
			}
		}
		d.front = markdown.SetField(d.front, "status", status)
		if note == "" {
			note = "Status: " + status + "."
		}
		return s.addHistory(d, note)
	})
	return warnings, err
}

// transitions gives the allowed new statuses for each current status (ADR-013). Done is final,
// because a new finding is a new task.
var transitions = map[string][]string{
	"Draft":          {"In Review", "Cancelled"},
	"In Review":      {"Approved", "Draft", "Cancelled"},
	"Approved":       {"In Development", "In Review", "Blocked", "Cancelled"},
	"In Development": {"Done", "Blocked", "Cancelled"},
	"Blocked":        {"Approved", "In Development", "Cancelled"},
	"Cancelled":      {"Draft"},
	"Done":           {},
}

// allowedChange refuses a status change that skips a step of the flow.
func allowedChange(from, to string) error {
	allowed, ok := transitions[from]
	if !ok || slices.Contains(allowed, to) {
		return nil // A task with an unknown status can get any valid status.
	}
	if len(allowed) == 0 {
		return fmt.Errorf("cannot set %s from %s. %s is final: make a new task", to, from, from)
	}
	return fmt.Errorf("cannot set %s from %s. Allowed: %s", to, from, strings.Join(allowed, ", "))
}

// rules returns the problems that stop a change to each status.
var rules = map[string]func(f Front, body, note string) []string{
	"Draft":          none,
	"In Review":      inReview,
	"Approved":       func(_ Front, body, _ string) []string { return openQuestions(body) },
	"In Development": none,
	"Blocked":        needNote,
	"Done":           func(_ Front, body, _ string) []string { return uncheckedDoD(body) },
	"Cancelled":      needNote,
}

func none(Front, string, string) []string { return nil }

func needNote(_ Front, _, note string) []string {
	if strings.TrimSpace(note) == "" {
		return []string{"give a note with the reason"}
	}
	return nil
}

func inReview(f Front, body, _ string) []string {
	problems := openQuestions(body)
	for _, ref := range leafSections {
		if text, _ := markdown.Section(body, ref); text == pristine[ref] {
			problems = append(problems, "section "+ref+" has the template text. Fill it or write Not applicable")
		}
	}
	if f.Size == "" {
		problems = append(problems, "give a size: S or M. Split a larger task into smaller tasks")
	}
	if len(f.Platforms) == 0 {
		problems = append(problems, "give the platforms, or [Not applicable]")
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

// uncheckedDoD returns one problem with the numbers of the unchecked Definition of Done items.
func uncheckedDoD(body string) []string {
	var numbers []string
	for i, item := range dodItems(body) {
		if item[1] == " " {
			numbers = append(numbers, strconv.Itoa(i+1))
		}
	}
	switch len(numbers) {
	case 0:
		return nil
	case 1:
		return []string{"Definition of Done item " + numbers[0] + " is not checked"}
	}
	return []string{"Definition of Done items " + strings.Join(numbers, ", ") + " are not checked"}
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

// withoutPlaceholders removes template rows (noQuestions, "[Question]", "None.") and returns the next question number.
func withoutPlaceholders(lines []string) ([]string, int) {
	var out []string
	next := 1
	for _, line := range lines {
		if line == noQuestions {
			continue
		}
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

// evidencePrefix starts the sub-item that holds the evidence of a checked Definition of Done item.
const evidencePrefix = "  - Evidence: "

// CheckDoD checks one Definition of Done item (1 is the first item) and adds the evidence as a sub-item.
func (s Store) CheckDoD(id string, item int, evidence string) error {
	evidence = oneLine(evidence)
	if evidence == "" {
		return errors.New("give the evidence")
	}
	return s.editDoD(id, func(_ *doc, lines []string) ([]string, error) {
		start, end, err := findItem(lines, item)
		if err != nil {
			return nil, err
		}
		m := dodItem.FindStringSubmatch(lines[start])
		if m[1] == "x" {
			return nil, fmt.Errorf("item %d is already checked", item)
		}
		return slices.Concat(lines[:start], []string{"- [x] " + m[2]}, lines[start+1:end], []string{evidencePrefix + evidence}, lines[end:]), nil
	})
}

// UncheckDoD unchecks one Definition of Done item, removes its evidence and adds the reason to the Document History.
// An item that was checked before the evidence sub-item keeps its full text.
func (s Store) UncheckDoD(id string, item int, reason string) error {
	if strings.TrimSpace(reason) == "" {
		return errors.New("give the reason")
	}
	return s.editDoD(id, func(d *doc, lines []string) ([]string, error) {
		start, end, err := findItem(lines, item)
		if err != nil {
			return nil, err
		}
		m := dodItem.FindStringSubmatch(lines[start])
		if m[1] != "x" {
			return nil, fmt.Errorf("item %d is not checked", item)
		}
		subItems := slices.DeleteFunc(slices.Clone(lines[start+1:end]), func(line string) bool {
			return strings.HasPrefix(line, evidencePrefix)
		})
		if err := s.addHistory(d, fmt.Sprintf("Definition of Done item %d unchecked: %s", item, reason)); err != nil {
			return nil, err
		}
		return slices.Concat(lines[:start], []string{"- [ ] " + m[2]}, subItems, lines[end:]), nil
	})
}

// AddDoD adds an unchecked Definition of Done item after the last item and returns its number.
func (s Store) AddDoD(id, text string) (int, error) {
	text = oneLine(text)
	if text == "" {
		return 0, errors.New("give the text of the item")
	}
	var n int
	err := s.editDoD(id, func(_ *doc, lines []string) ([]string, error) {
		at := 0
		for i, line := range lines {
			if dodItem.MatchString(line) {
				n++
				at = i + 1
			} else if n > 0 && at == i && strings.HasPrefix(line, " ") {
				at = i + 1 // a sub-item of the last item
			}
		}
		n++
		return slices.Insert(lines, at, "- [ ] "+text), nil
	})
	return n, err
}

// RemoveDoD removes one Definition of Done item with its sub-items, and adds its text and the reason to the Document History.
func (s Store) RemoveDoD(id string, item int, reason string) error {
	if strings.TrimSpace(reason) == "" {
		return errors.New("give the reason")
	}
	return s.editDoD(id, func(d *doc, lines []string) ([]string, error) {
		start, end, err := findItem(lines, item)
		if err != nil {
			return nil, err
		}
		text := dodItem.FindStringSubmatch(lines[start])[2]
		if err := s.addHistory(d, fmt.Sprintf("Definition of Done item %d removed (%s): %s", item, text, reason)); err != nil {
			return nil, err
		}
		return slices.Delete(lines, start, end), nil
	})
}

// editDoD changes the lines of the Definition of Done with fn and writes the task.
// It refuses a Done task, because Done is final (ADR-013).
func (s Store) editDoD(id string, fn func(d *doc, lines []string) ([]string, error)) error {
	return s.edit(id, func(d *doc) error {
		var f Front
		if err := markdown.Decode(d.front, &f); err != nil {
			return fmt.Errorf("bad frontmatter: %w", err)
		}
		if f.Status == "Done" {
			return fmt.Errorf("%s is Done. Done is final: make a new task", kind.ID(mustNumber(d.path)))
		}
		section, _ := markdown.Section(d.body, "14")
		lines, err := fn(d, strings.Split(section, "\n"))
		if err != nil {
			return err
		}
		body, err := markdown.ReplaceSection(d.body, "14", strings.Join(lines, "\n"))
		d.body = body
		return err
	})
}

// findItem returns the lines [start, end) of Definition of Done item n (1 is the first item):
// the item line and its indented sub-items.
func findItem(lines []string, n int) (start, end int, err error) {
	count := 0
	for i, line := range lines {
		if !dodItem.MatchString(line) {
			continue
		}
		if count++; count != n {
			continue
		}
		end = i + 1
		for end < len(lines) && strings.HasPrefix(lines[end], " ") {
			end++
		}
		return i, end, nil
	}
	return 0, 0, fmt.Errorf("item %d does not exist. The Definition of Done has %d items", n, count)
}

// oneLine joins the lines of text with one space, because a Definition of Done line is one line.
func oneLine(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

// Renumber gives the task in file the next free number. It returns the new ID and the tasks
// that refer to the old ID. It does not change them, because the old ID can name more than one file.
// It holds the docs lock.
func (s Store) Renumber(file string) (newID string, refs []string, err error) {
	err = docfile.Locked(s.Lock, func() error {
		newID, refs, err = s.renumber(file)
		return err
	})
	return newID, refs, err
}

func (s Store) renumber(file string) (string, []string, error) {
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
	// The file moves with one rename, so a failure never leaves a third file (TASK-049).
	_, err = s.Alloc.Next("task", s.Dir, kind.Pattern, func(n int) error {
		newID = kind.ID(n)
		doc := markdown.Join(markdown.SetField(front, "id", newID), setHeader(body, newID, f.Title))
		old := filepath.Join(s.Dir, file)
		if err := docfile.Write(old, []byte(doc)); err != nil {
			return err
		}
		return docfile.Move(old, filepath.Join(s.Dir, newID+strings.TrimPrefix(file, oldID)))
	})
	if err != nil {
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
