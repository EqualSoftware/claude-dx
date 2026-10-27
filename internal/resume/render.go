package resume

import (
	"fmt"
	"strings"
)

// Limits of the output (R-5, R-7, R-8, R-9).
const (
	maxLines    = 40
	maxNotes    = 3
	maxChanged  = 10
	maxNext     = 3
	maxBlocked  = 5
	maxUnpushed = 10
)

// view is everything that the output shows, before the limits.
type view struct {
	header     string
	errs       []error
	changed    []string
	inProgress []progress
	next       []string
	blocked    []string
	unpushed   []string
}

// progress is an In Progress task.
type progress struct {
	head  string
	step  string
	notes []string
}

// render returns the lines of the output with at most notes Notes lines for each task. The output has
// maxLines lines or less.
func render(v view) []string {
	var lines []string
	for notes := maxNotes; notes >= 0; notes-- {
		if lines = v.lines(notes); len(lines) <= maxLines {
			return lines
		}
	}
	more := len(lines) - (maxLines - 1)
	return append(lines[:maxLines-1], fmt.Sprintf("+%d more", more))
}

// lines writes the output with at most notes Notes lines for each In Progress task.
func (v view) lines(notes int) []string {
	var lines []string
	if v.header != "" {
		lines = append(lines, v.header)
	}
	for _, err := range v.errs {
		lines = append(lines, err.Error())
	}
	if v.clean() {
		return lines
	}
	if len(v.changed) > 0 {
		lines = append(lines, "Changed: "+changedText(v.changed))
	}
	if len(v.inProgress) > 0 {
		lines = append(lines, "In Progress:")
		for _, p := range v.inProgress {
			lines = append(lines, p.lines(notes)...)
		}
	}
	lines = appendList(lines, "Next", v.next, maxNext)
	lines = appendList(lines, "Blocked", v.blocked, maxBlocked)
	return appendList(lines, "Unpushed", v.unpushed, maxUnpushed)
}

// clean reports whether the view has only the header and the errors.
func (v view) clean() bool {
	return len(v.changed) == 0 && len(v.inProgress) == 0 && len(v.next) == 0 &&
		len(v.blocked) == 0 && len(v.unpushed) == 0
}

// lines writes the task, with the last notes Notes lines.
func (p progress) lines(notes int) []string {
	lines := []string{p.head}
	if p.step != "" {
		lines = append(lines, "  Next step: "+p.step)
	}
	n := p.notes[max(0, len(p.notes)-notes):]
	for _, note := range n {
		lines = append(lines, "  Notes: "+note)
	}
	return lines
}

// changedText joins the changed files with commas. It shows maxChanged files.
func changedText(changed []string) string {
	if len(changed) <= maxChanged {
		return strings.Join(changed, ", ")
	}
	return fmt.Sprintf("%s, +%d more", strings.Join(changed[:maxChanged], ", "), len(changed)-maxChanged)
}

// appendList adds a section with at most limit items. A cut list ends with "+N more".
func appendList(lines []string, title string, items []string, limit int) []string {
	if len(items) == 0 {
		return lines
	}
	lines = append(lines, title+":")
	lines = append(lines, items[:min(limit, len(items))]...)
	if len(items) > limit {
		lines = append(lines, fmt.Sprintf("  +%d more", len(items)-limit))
	}
	return lines
}
