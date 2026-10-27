package task

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

// checkItem matches one line of a checklist: a Plan step or a Definition of Done item.
var checkItem = regexp.MustCompile(`^- \[([ x])\] (.*)$`)

// dod is the title of the Definition of Done section.
const dod = "Definition of Done"

// sections are the sections that task_update_section changes.
var sections = []string{"Brief", "Plan", "Notes"}

// UpdateSection replaces the Brief, the Plan or the Notes. The Brief and the Plan must not be empty (Q-32).
func (s Store) UpdateSection(id, ref, body string) error {
	i := slices.IndexFunc(sections, func(name string) bool { return strings.EqualFold(name, ref) })
	if strings.EqualFold(ref, dod) {
		return errors.New("section Definition of Done changes only with task_dod_add, task_dod_remove, task_dod_check and task_dod_uncheck")
	}
	if i < 0 {
		return fmt.Errorf("section %s does not exist. Use Brief, Plan or Notes", ref)
	}
	if i < 2 && strings.TrimSpace(body) == "" {
		return [...]error{errors.New("give the brief"), errors.New("give one or more plan steps")}[i]
	}
	return s.edit(id, func(d *doc) error {
		text, err := markdown.ReplaceSection(d.body, sections[i], strings.TrimSpace(body))
		d.body = text
		return err
	})
}

// Update holds the new frontmatter fields. An empty string or a nil list keeps the field. An empty
// list removes all items.
type Update struct {
	Title     string
	Priority  string
	Covers    []string
	DependsOn []string
}

// check returns an error for an invalid field, or for an Update with no field.
func (u Update) check() error {
	if u.Title == "" && u.Priority == "" && u.Covers == nil && u.DependsOn == nil {
		return errors.New("give a field to change")
	}
	if u.Title != "" {
		if err := docfile.CheckTitle(u.Title); err != nil {
			return err
		}
	}
	if u.Priority != "" {
		return oneOf("priority", u.Priority, Priorities)
	}
	return nil
}

// UpdateFields changes the frontmatter fields of the task id.
func (s Store) UpdateFields(id string, u Update) error {
	if err := u.check(); err != nil {
		return err
	}
	return s.edit(id, func(d *doc) error {
		u.apply(d)
		return s.checkFront(d.f)
	})
}

// apply sets the fields of u in d.
func (u Update) apply(d *doc) {
	if u.Title != "" {
		d.f.Title = strings.TrimSpace(u.Title)
		d.front = markdown.SetField(d.front, "title", d.f.Title)
		d.body = headerLine.ReplaceAllLiteralString(d.body, "# "+d.f.Title)
	}
	if u.Priority != "" {
		d.f.Priority = u.Priority
		d.front = markdown.SetField(d.front, "priority", u.Priority)
	}
	if u.Covers != nil {
		d.f.Covers = u.Covers
		d.front = markdown.SetField(d.front, "covers", u.Covers)
	}
	if u.DependsOn != nil {
		d.f.DependsOn = u.DependsOn
		d.front = markdown.SetField(d.front, "depends_on", u.DependsOn)
	}
}

// transitions are the allowed status changes (R-3). Done is final.
var transitions = map[string][]string{
	"Todo":        {"In Progress", "Blocked"},
	"In Progress": {"Todo", "Blocked", "Done"},
	"Blocked":     {"Todo", "In Progress"},
}

// SetStatus changes the status of the task id and adds the note to Notes (R-3, R-7). For Done, it sets the
// covers requirements to Done in the spec first, so a retry after a failure is safe (R-6). It returns
// the covers of a Done task.
func (s Store) SetStatus(id, status, note string) (covers []string, err error) {
	if err := oneOf("status", status, Statuses); err != nil {
		return nil, err
	}
	note = oneLine(note)
	if status == "Blocked" && note == "" {
		return nil, errors.New("give the blocker in note")
	}
	err = s.edit(id, func(d *doc) error {
		if err := s.checkStatus(d, status); err != nil {
			return err
		}
		if note != "" {
			if err := d.addNote(status + ": " + note); err != nil {
				return err
			}
		}
		d.front = markdown.SetField(d.front, "status", status)
		if status != "Done" {
			return nil
		}
		covers = d.f.Covers
		return s.Specs.SetDoneLocked(d.f.Spec(), covers)
	})
	return covers, err
}

// checkStatus refuses a status change that R-3, R-6 or R-7 does not allow.
func (s Store) checkStatus(d *doc, to string) error {
	if allowed := transitions[d.f.Status]; !slices.Contains(allowed, to) {
		return fmt.Errorf("a task does not go from %s to %s. Allowed: %s", d.f.Status, to, strings.Join(allowed, ", "))
	}
	switch to {
	case "In Progress":
		return s.checkStart(d.f)
	case "Done":
		return checkDone(d)
	}
	return nil
}

// checkStart refuses a dependency that is not Done and a spec that is not Approved (R-7).
func (s Store) checkStart(f Front) error {
	for _, dep := range f.DependsOn {
		d, err := s.load(dep)
		if err != nil {
			return err
		}
		if d.f.Status != "Done" {
			return fmt.Errorf("task %s is not Done", dep)
		}
	}
	if f.Spec() == general {
		return nil
	}
	status, _, err := s.specLocked(f.Spec())
	if err != nil || status == "Approved" {
		return err
	}
	return fmt.Errorf("spec %s is %s. A task starts only when its spec is Approved", f.Spec(), status)
}

// checkDone refuses a task with a Plan step or a Definition of Done item that is not checked (R-6).
func checkDone(d *doc) error {
	var parts []string
	if open := unchecked(d.body, "Plan"); len(open) > 0 {
		parts = append(parts, plural("Plan step", open))
	}
	if open := unchecked(d.body, dod); len(open) > 0 {
		parts = append(parts, plural(dod+" item", open))
	}
	if len(parts) == 0 {
		return nil
	}
	verb := "are"
	if len(parts) == 1 && !strings.Contains(parts[0], ",") && !strings.Contains(parts[0], " and ") {
		verb = "is"
	}
	return fmt.Errorf("%s %s not checked", strings.Join(parts, " and "), verb)
}

// plural returns "Plan step 2" or "Plan steps 2, 4".
func plural(noun string, numbers []string) string {
	if len(numbers) == 1 {
		return noun + " " + numbers[0]
	}
	return noun + "s " + strings.Join(numbers, ", ")
}

// unchecked returns the numbers of the unchecked items of the checklist in the section ref.
func unchecked(body, ref string) []string {
	section, _ := markdown.Section(body, ref)
	var open []string
	n := 0
	for line := range strings.SplitSeq(section, "\n") {
		m := checkItem.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		n++
		if m[1] == " " {
			open = append(open, strconv.Itoa(n))
		}
	}
	return open
}

// CheckStep checks one Plan step (1 is the first step).
func (s Store) CheckStep(id string, step int) error {
	return s.editList(id, "Plan", func(_ *doc, lines []string) ([]string, error) {
		start, _, err := findItem(lines, step, "step %d does not exist. The Plan has %d steps")
		if err != nil {
			return nil, err
		}
		m := checkItem.FindStringSubmatch(lines[start])
		if m[1] == "x" {
			return nil, fmt.Errorf("step %d is already checked", step)
		}
		lines[start] = "- [x] " + m[2]
		return lines, nil
	})
}

// evidencePrefix starts the sub-item that holds the evidence of a checked Definition of Done item.
const evidencePrefix = "  - Evidence: "

// noItem is the error text for a Definition of Done item that does not exist.
const noItem = "item %d does not exist. The Definition of Done has %d items"

// CheckDoD checks one Definition of Done item (1 is the first item) and adds the evidence as a sub-item.
func (s Store) CheckDoD(id string, item int, evidence string) error {
	evidence = oneLine(evidence)
	if evidence == "" {
		return errors.New("give the evidence")
	}
	return s.editList(id, dod, func(_ *doc, lines []string) ([]string, error) {
		start, end, err := findItem(lines, item, noItem)
		if err != nil {
			return nil, err
		}
		m := checkItem.FindStringSubmatch(lines[start])
		if m[1] == "x" {
			return nil, fmt.Errorf("item %d is already checked", item)
		}
		return slices.Concat(lines[:start], []string{"- [x] " + m[2]}, lines[start+1:end], []string{evidencePrefix + evidence}, lines[end:]), nil
	})
}

// UncheckDoD unchecks one Definition of Done item and removes its evidence. It needs no reason, because an
// uncheck hides nothing (Q-38).
func (s Store) UncheckDoD(id string, item int) error {
	return s.editList(id, dod, func(_ *doc, lines []string) ([]string, error) {
		start, end, err := findItem(lines, item, noItem)
		if err != nil {
			return nil, err
		}
		m := checkItem.FindStringSubmatch(lines[start])
		if m[1] != "x" {
			return nil, fmt.Errorf("item %d is not checked", item)
		}
		subItems := slices.DeleteFunc(slices.Clone(lines[start+1:end]), func(line string) bool {
			return strings.HasPrefix(line, evidencePrefix)
		})
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
	err := s.editList(id, dod, func(_ *doc, lines []string) ([]string, error) {
		at := 0
		for i, line := range lines {
			if checkItem.MatchString(line) {
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

// RemoveDoD removes one Definition of Done item with its sub-items, and adds its text and the reason to
// Notes, because a removal makes the Definition of Done weaker (Q-38).
func (s Store) RemoveDoD(id string, item int, reason string) error {
	reason = oneLine(reason)
	if reason == "" {
		return errors.New("give the reason")
	}
	return s.editList(id, dod, func(d *doc, lines []string) ([]string, error) {
		start, end, err := findItem(lines, item, noItem)
		if err != nil {
			return nil, err
		}
		text := checkItem.FindStringSubmatch(lines[start])[2]
		if err := d.addNote(fmt.Sprintf("Definition of Done item %d removed (%s): %s", item, text, reason)); err != nil {
			return nil, err
		}
		return slices.Delete(lines, start, end), nil
	})
}

// editList changes the lines of the section ref with fn and writes the task.
func (s Store) editList(id, ref string, fn func(d *doc, lines []string) ([]string, error)) error {
	return s.edit(id, func(d *doc) error {
		section, _ := markdown.Section(d.body, ref)
		lines, err := fn(d, strings.Split(section, "\n"))
		if err != nil {
			return err
		}
		body, err := markdown.ReplaceSection(d.body, ref, strings.Join(lines, "\n"))
		d.body = body
		return err
	})
}

// findItem returns the lines [start, end) of checklist item n (1 is the first item): the item line and
// its indented sub-items. missing is the error format for an item that does not exist: n, count.
func findItem(lines []string, n int, missing string) (start, end int, err error) {
	count := 0
	for i, line := range lines {
		if !checkItem.MatchString(line) {
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
	return 0, 0, fmt.Errorf(missing, n, count)
}
