package spec

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/EqualSoftware/claude-dx/internal/docs/docfile"
	"github.com/EqualSoftware/claude-dx/internal/markdown"
)

var (
	questionRow    = regexpRow(`(Q-(\d+)) \| (.*?) \| (.*?) \| (.*?)`)
	requirementRow = regexpRow(`(R-(\d+)) \| (.*?) \| (Planned|Done)`)
)

// regexpRow returns the pattern of one table row with the cell patterns.
func regexpRow(cells string) *regexp.Regexp {
	return regexp.MustCompile(`^\| ` + cells + ` \|$`)
}

// leafSections are the sections that In Review needs filled. Sections 5 and 8 are not in the list, because
// their template text is valid and only their tools change them.
var leafSections = []string{"1", "2", "3", "4", "6", "7"}

// Placeholder rows of the template. The add tools remove them.
const (
	noQuestions    = "| — | None. | — | — |"
	noRequirements = "| — | None. | — |"
)

// protected are the sections that only their own tools change, so the flow checks of ADR-013 stay true.
var protected = []struct{ number, title, tools string }{
	{"5", "Requirements", "spec_req_add and spec_req_set"},
	{"8", "Open Questions", "spec_question_add and spec_question_answer"},
	{"9", "Document History", "spec_set_status"},
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

// scopeSections define the scope of a spec: Goals, Out of Scope, Requirements, Design and
// Acceptance Criteria (ADR-016).
var scopeSections = []string{"3", "4", "5", "6", "7"}

// scopeNote is the Document History note when a scope change sets In Review.
func scopeNote(section string) string {
	return "Status: In Review. The scope changed in section " + section + "."
}

// UpdateSection replaces one section. A scope change of an Approved spec sets it to In Review,
// because the user must approve the new scope. review is true after that status change.
func (s Store) UpdateSection(slug, section, content string) (review bool, err error) {
	if err := checkSection(section); err != nil {
		return false, err
	}
	err = s.edit(slug, func(d *doc) error {
		body, err := markdown.ReplaceSection(d.body, section, content)
		if err != nil {
			return err
		}
		old := d.body
		d.body = body
		if !scopeChanged(old, body) {
			return nil
		}
		review, err = s.toReview(d, scopeNote(section))
		return err
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

// transitions gives the allowed new statuses for each current status (ADR-016).
var transitions = map[string][]string{
	"Draft":     {"In Review"},
	"In Review": {"Approved", "Draft"},
	"Approved":  {"In Review"},
}

// SetStatus changes the status, adds a Document History row and enforces the status rules.
func (s Store) SetStatus(slug, status, note string) error {
	if err := oneOf("status", status, Statuses); err != nil {
		return err
	}
	return s.edit(slug, func(d *doc) error {
		from, err := d.status()
		if err != nil {
			return err
		}
		if allowed, ok := transitions[from]; ok && !slices.Contains(allowed, status) {
			return fmt.Errorf("cannot set %s from %s. Allowed: %s", status, from, strings.Join(allowed, ", "))
		}
		if problems := statusProblems(status, d.body); len(problems) > 0 {
			return fmt.Errorf("cannot set %s: %s", status, strings.Join(problems, "; "))
		}
		d.front = markdown.SetField(d.front, "status", status)
		if note == "" {
			note = "Status: " + status + "."
		}
		return s.addHistory(d, note)
	})
}

// statusProblems returns the problems that stop a change to status.
func statusProblems(status, body string) []string {
	switch status {
	case "Approved":
		return openQuestions(body)
	case "In Review":
		return inReview(body)
	}
	return nil
}

func inReview(body string) []string {
	problems := openQuestions(body)
	for _, ref := range leafSections {
		if text, _ := markdown.Section(body, ref); text == pristine[ref] {
			problems = append(problems, "section "+ref+" has the template text. Fill it or write Not applicable")
		}
	}
	if len(requirements(body)) == 0 {
		problems = append(problems, "section 5 has no requirement. Add one with spec_req_add")
	}
	return problems
}

func openQuestions(body string) []string {
	var problems []string
	for _, q := range rows(body, "8", questionRow) {
		if q[5] == "Open" {
			problems = append(problems, q[1]+" is Open")
		}
	}
	return problems
}

// requirements returns the rows of the Requirements table: ID, number, text and status.
func requirements(body string) [][]string {
	return rows(body, "5", requirementRow)
}

// AddRequirement adds the next R-n row with the status Planned and returns its ID. review is true
// if the spec was Approved and is now In Review.
func (s Store) AddRequirement(slug, text string) (rid string, review bool, err error) {
	text = oneLine(text)
	if text == "" {
		return "", false, errors.New("give the text of the requirement")
	}
	err = s.edit(slug, func(d *doc) error {
		lines, next := withoutPlaceholder(d.body, "5", noRequirements, requirementRow)
		rid = fmt.Sprintf("R-%d", next)
		row := fmt.Sprintf("| %s | %s | Planned |", rid, docfile.Cell(text))
		if err := replaceTable(d, "5", insertRow(lines, row)); err != nil {
			return err
		}
		if err := s.addHistory(d, "Added "+rid+"."); err != nil {
			return err
		}
		review, err = s.toReview(d, scopeNote("5"))
		return err
	})
	return rid, review, err
}

// SetRequirement changes the text, the status, or both of one requirement. Only a text change sets an
// Approved spec to In Review, because a status change does not change the scope.
func (s Store) SetRequirement(slug, rid, text, status string) (review bool, err error) {
	text = oneLine(text)
	if text == "" && status == "" {
		return false, errors.New("give text, status or both")
	}
	if status != "" {
		if err := oneOf("status", status, []string{"Planned", "Done"}); err != nil {
			return false, err
		}
	}
	err = s.edit(slug, func(d *doc) error {
		notes, err := setRow(d, rid, text, status)
		if err != nil {
			return err
		}
		for _, note := range notes {
			if err := s.addHistory(d, note); err != nil {
				return err
			}
		}
		if text == "" {
			return nil
		}
		review, err = s.toReview(d, scopeNote("5"))
		return err
	})
	return review, err
}

// RequirementsLocked returns the status of the spec and the IDs of its requirements. The caller holds
// the docs lock (TASK-064).
func (s Store) RequirementsLocked(slug string) (status string, rids []string, err error) {
	d, _, err := s.readDoc(slug)
	if err != nil {
		return "", nil, err
	}
	if status, err = d.status(); err != nil {
		return "", nil, err
	}
	for _, row := range requirements(d.body) {
		rids = append(rids, row[1])
	}
	return status, rids, nil
}

// SetDoneLocked sets the requirements rids to Done and adds one Document History row, for example
// "R-2, R-5: Done.". A requirement that is Done already does not change, so a retry is safe. The caller
// holds the docs lock (TASK-064).
func (s Store) SetDoneLocked(slug string, rids []string) error {
	if len(rids) == 0 {
		return nil
	}
	return s.editLocked(slug, func(d *doc) error {
		if !slices.ContainsFunc(rids, func(rid string) bool { return requirementStatus(d, rid) != "Done" }) {
			return errUnchanged
		}
		for _, rid := range rids {
			if _, err := setRow(d, rid, "", "Done"); err != nil {
				return err
			}
		}
		return s.addHistory(d, strings.Join(rids, ", ")+": Done.")
	})
}

// requirementStatus returns the status of rid in d, or "" if d has no rid.
func requirementStatus(d *doc, rid string) string {
	for _, m := range requirements(d.body) {
		if strings.EqualFold(m[1], rid) {
			return m[4]
		}
	}
	return ""
}

// setRow changes the row of rid and returns the Document History notes.
func setRow(d *doc, rid, text, status string) ([]string, error) {
	section, _ := markdown.Section(d.body, "5")
	lines := strings.Split(section, "\n")
	for i, line := range lines {
		m := requirementRow.FindStringSubmatch(line)
		if m == nil || !strings.EqualFold(m[1], rid) {
			continue
		}
		var notes []string
		if text != "" {
			m[3] = docfile.Cell(text)
			notes = append(notes, m[1]+": text changed.")
		}
		if status != "" {
			m[4] = status
			notes = append(notes, m[1]+": status "+status+".")
		}
		lines[i] = fmt.Sprintf("| %s | %s | %s |", m[1], m[3], m[4])
		return notes, replaceTable(d, "5", lines)
	}
	return nil, fmt.Errorf("requirement %s does not exist", rid)
}

// AddQuestion adds the next Q-n row with the answer Open and returns its ID. An Approved spec goes
// to In Review, because Approved needs no Open question. review is true after that status change.
func (s Store) AddQuestion(slug, question, owner string) (qid string, review bool, err error) {
	if owner == "" {
		owner = "Maintainer"
	}
	err = s.edit(slug, func(d *doc) error {
		lines, next := withoutPlaceholder(d.body, "8", noQuestions, questionRow)
		qid = fmt.Sprintf("Q-%d", next)
		row := fmt.Sprintf("| %s | %s | %s | Open |", qid, docfile.Cell(question), docfile.Cell(owner))
		if err := replaceTable(d, "8", insertRow(lines, row)); err != nil {
			return err
		}
		review, err = s.toReview(d, "Status: In Review. "+qid+" is Open.")
		return err
	})
	return qid, review, err
}

// AnswerQuestion writes the answer of one question.
func (s Store) AnswerQuestion(slug, qid, answer string) error {
	return s.edit(slug, func(d *doc) error {
		section, _ := markdown.Section(d.body, "8")
		lines := strings.Split(section, "\n")
		for i, line := range lines {
			m := questionRow.FindStringSubmatch(line)
			if m == nil || !strings.EqualFold(m[1], qid) {
				continue
			}
			lines[i] = fmt.Sprintf("| %s | %s | %s | %s |", m[1], m[3], m[4], docfile.Cell(answer))
			return replaceTable(d, "8", lines)
		}
		return fmt.Errorf("question %s does not exist", qid)
	})
}

// rows returns the matches of pattern in the lines of one section.
func rows(body, section string, pattern *regexp.Regexp) [][]string {
	text, _ := markdown.Section(body, section)
	var out [][]string
	for line := range strings.SplitSeq(text, "\n") {
		if m := pattern.FindStringSubmatch(line); m != nil {
			out = append(out, m)
		}
	}
	return out
}

// withoutPlaceholder returns the lines of one section with no placeholder row, and the next row number.
// Group 2 of pattern is the row number.
func withoutPlaceholder(body, section, placeholder string, pattern *regexp.Regexp) ([]string, int) {
	text, _ := markdown.Section(body, section)
	var out []string
	next := 1
	for line := range strings.SplitSeq(text, "\n") {
		if line == placeholder {
			continue
		}
		if m := pattern.FindStringSubmatch(line); m != nil {
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

func replaceTable(d *doc, section string, lines []string) error {
	body, err := markdown.ReplaceSection(d.body, section, strings.Join(lines, "\n"))
	d.body = body
	return err
}

// oneLine joins the lines of text with one space, because a table row is one line.
func oneLine(text string) string {
	return strings.Join(strings.Fields(text), " ")
}
