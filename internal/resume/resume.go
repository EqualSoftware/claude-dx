// Package resume prints the work state (tasks and git) for a new session (resume-command spec, ADR-008).
// It only reads. It uses no network and no MCP server.
package resume

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// Run writes the work state of the project that contains dir to w. cacheDir is the cache directory of
// claude-dx. A part that fails is skipped and shows one line (R-10).
func Run(w io.Writer, dir, cacheDir string) {
	g, gitErr := readGit(dir)
	tasks, approved, errs := readTasks(dir, cacheDir)
	v := view{
		inProgress: progressList(inProgress(tasks)),
		next:       taskLines(next(tasks, approved), false),
		blocked:    taskLines(blocked(tasks), true),
		errs:       errs,
	}
	v.addGit(g, gitErr)
	for _, line := range render(v) {
		fmt.Fprintln(w, line)
	}
}

// addGit adds the git state to the view.
func (v *view) addGit(g gitState, err error) {
	switch {
	case errors.Is(err, errNoRepo):
		v.header = "dx: no git repository"
		return
	case err != nil:
		v.errs = append([]error{skip("git", err)}, v.errs...)
		return
	case g.branch == "":
		v.header = "dx: detached at " + g.hash
		v.changed = g.changed
		return
	}
	if len(g.changed) == 0 && len(g.unpushed)+len(v.inProgress)+len(v.next)+len(v.blocked) == 0 {
		v.header = fmt.Sprintf("dx: branch %s, clean tree, no active tasks.", g.branch)
		return
	}
	v.header = headerText(g)
	v.changed = g.changed
	v.unpushed = unpushedLines(g.unpushed)
}

// headerText is the first line for a branch (R-1).
func headerText(g gitState) string {
	ahead := "no upstream"
	if g.upstream != "" {
		ahead = fmt.Sprintf("%d ahead of %s", len(g.unpushed), g.upstream)
	}
	return fmt.Sprintf("dx: branch %s, %s, %d changed files", g.branch, ahead, len(g.changed))
}

// unpushedLines writes one line for each unpushed change (R-8).
func unpushedLines(list []commit) []string {
	var lines []string
	for _, c := range list {
		line := "- " + c.hash + " " + c.subject
		if len(c.refs) > 0 {
			line += " [" + strings.Join(c.refs, ", ") + "]"
		}
		lines = append(lines, line)
	}
	return lines
}

// taskHead is the first line of a task: ID, priority and title.
func taskHead(t taskInfo) string {
	return fmt.Sprintf("- %s (%s) %s", t.ID, t.Priority, t.Title)
}

// taskLines writes one line for each task. A Blocked task shows the last line of its Notes (R-7).
func taskLines(list []taskInfo, blocker bool) []string {
	var lines []string
	for _, t := range list {
		line := taskHead(t)
		if notes := noteLines(t.body); blocker && len(notes) > 0 {
			line += ": " + notes[len(notes)-1]
		}
		lines = append(lines, line)
	}
	return lines
}

// progressList makes the In Progress entries (R-6).
func progressList(list []taskInfo) []progress {
	var out []progress
	for _, t := range list {
		out = append(out, progress{head: taskHead(t), step: firstStep(t.body), notes: noteLines(t.body)})
	}
	return out
}
