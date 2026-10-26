package task

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/EqualSoftware/claude-dx/internal/markdown"
)

var codeSpan = regexp.MustCompile("`([^`]+)`")

// planFiles returns the paths in backticks in the Files column of the Implementation Plan table.
func planFiles(body string) []string {
	section, _ := markdown.Section(body, "8")
	var paths []string
	for line := range strings.SplitSeq(section, "\n") {
		cells := strings.Split(line, "|")
		if len(cells) < 4 {
			continue
		}
		for _, m := range codeSpan.FindAllStringSubmatch(cells[2], -1) {
			paths = append(paths, m[1])
		}
	}
	return paths
}

// overlap reports whether two plan paths change the same code: they are equal, or one is a
// directory (it ends with "/") that contains the other.
func overlap(a, b string) bool {
	return a == b || strings.HasSuffix(a, "/") && strings.HasPrefix(b, a) || strings.HasSuffix(b, "/") && strings.HasPrefix(a, b)
}

// planTask is a task with the files of its plan.
type planTask struct {
	front Front
	files []string
}

// overlapWarnings returns one warning line for each other open task that changes a file of the plan
// of task f (with body) and has no depends_on path to or from f (ADR-013, step 2).
func (s Store) overlapWarnings(f Front, body string) (string, error) {
	mine := planFiles(body)
	if len(mine) == 0 {
		return "", nil
	}
	tasks, err := s.planTasks()
	if err != nil {
		return "", err
	}
	byID := map[string]Front{f.ID: f}
	for _, t := range tasks {
		byID[t.front.ID] = t.front
	}
	var warnings string
	for _, t := range tasks {
		if !unordered(f, t.front, byID) {
			continue
		}
		if path, ok := shared(mine, t.files); ok {
			warnings += fmt.Sprintf("Warning: %s also changes %s. Set the order with depends_on, or merge the tasks.\n", t.front.ID, path)
		}
	}
	return warnings, nil
}

// unordered reports whether other is another open task with no depends_on path to or from f.
func unordered(f, other Front, byID map[string]Front) bool {
	if other.ID == f.ID || other.Status == "Done" || other.Status == "Cancelled" {
		return false
	}
	return depPath(f.ID, other.ID, byID, nil) == nil && depPath(other.ID, f.ID, byID, nil) == nil
}

// shared returns the first path of mine that overlaps a path of theirs.
func shared(mine, theirs []string) (string, bool) {
	for _, a := range mine {
		for _, b := range theirs {
			if overlap(a, b) {
				return a, true
			}
		}
	}
	return "", false
}

// planTasks reads each task file one time and returns its frontmatter and plan files.
func (s Store) planTasks() ([]planTask, error) {
	files, err := kind.Files(s.Dir)
	if err != nil {
		return nil, err
	}
	var tasks []planTask
	for _, name := range sortedNames(files) {
		data, err := os.ReadFile(filepath.Join(s.Dir, name))
		if err != nil {
			return nil, err
		}
		front, body := markdown.Split(string(data))
		var f Front
		if markdown.Decode(front, &f) != nil {
			continue // task_list warns about a bad frontmatter.
		}
		tasks = append(tasks, planTask{front: f, files: planFiles(body)})
	}
	return tasks, nil
}
