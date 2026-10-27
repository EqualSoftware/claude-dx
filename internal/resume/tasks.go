package resume

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/EqualSoftware/claude-dx/internal/docs/spec"
	"github.com/EqualSoftware/claude-dx/internal/docs/task"
	"github.com/EqualSoftware/claude-dx/internal/markdown"
	"github.com/EqualSoftware/claude-dx/internal/project"
)

// general is the spec slug of the tasks that have no spec.
const general = "general"

// stepPattern matches an unchecked step of the Plan section.
var stepPattern = regexp.MustCompile(`^- \[ \] (.+)$`)

// taskInfo is a task file that the command read.
type taskInfo struct {
	task.Front
	mtime time.Time
	body  string
}

// readTasks reads all tasks of the project that contains dir. It writes nothing and creates nothing (R-11).
// It returns the tasks, the slugs of the Approved specs and one error for each part that it skipped.
func readTasks(dir, cacheDir string) (tasks []taskInfo, approved map[string]bool, errs []error) {
	if cacheDir == "" {
		return nil, nil, []error{skip("tasks", errors.New("no cache directory"))}
	}
	p, err := project.Open(dir)
	if err != nil {
		return nil, nil, []error{skip("tasks", err)}
	}
	data := p.DataDir(cacheDir)
	if _, err := os.Stat(data); errors.Is(err, os.ErrNotExist) {
		return nil, nil, nil // no data directory means no tasks
	} else if err != nil {
		return nil, nil, []error{skip("tasks", err)}
	}
	tasks, errs = loadTasks(filepath.Join(data, project.Tasks))
	return tasks, approvedSpecs(p.Root), errs
}

// skip makes the error of a part that the command skips (R-10).
func skip(part string, err error) error {
	return fmt.Errorf("dx: skipped %s: %w", part, err)
}

// loadTasks reads <tasksDir>/<spec>/<slug>.md. A folder that does not exist has no tasks.
func loadTasks(tasksDir string) (tasks []taskInfo, errs []error) {
	specs, err := os.ReadDir(tasksDir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, []error{skip("tasks", err)}
	}
	for _, s := range specs {
		if !s.IsDir() {
			continue
		}
		more, moreErrs := loadSpec(tasksDir, s.Name())
		tasks = append(tasks, more...)
		errs = append(errs, moreErrs...)
	}
	return tasks, errs
}

// loadSpec reads the task files of one spec folder.
func loadSpec(tasksDir, slug string) (tasks []taskInfo, errs []error) {
	files, err := os.ReadDir(filepath.Join(tasksDir, slug))
	if err != nil {
		return nil, []error{skip("tasks", err)}
	}
	for _, f := range files {
		name, ok := strings.CutSuffix(f.Name(), ".md")
		if f.IsDir() || !ok {
			continue
		}
		id := slug + "/" + name
		t, err := readTask(filepath.Join(tasksDir, slug, f.Name()), id)
		if err != nil {
			errs = append(errs, skip(id, err))
			continue
		}
		tasks = append(tasks, t)
	}
	return tasks, errs
}

// readTask reads one task file. The frontmatter id must be id.
func readTask(path, id string) (taskInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return taskInfo{}, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return taskInfo{}, err
	}
	front, body := markdown.Split(string(data))
	t := taskInfo{mtime: info.ModTime(), body: body}
	if err := markdown.Decode(front, &t.Front); err != nil {
		return taskInfo{}, fmt.Errorf("bad frontmatter: %w", err)
	}
	if t.ID != id {
		return taskInfo{}, fmt.Errorf("the frontmatter id is %q", t.ID)
	}
	return t, nil
}

// approvedSpecs returns the slugs of the Approved specs of the project at root. A project with no
// specs, or specs that it cannot read, has none.
func approvedSpecs(root string) map[string]bool {
	specs, _, err := spec.Store{Dir: filepath.Join(root, "docs", "specs")}.List(spec.Filter{"status": "Approved"})
	approved := map[string]bool{general: true}
	if err != nil {
		return approved
	}
	for _, s := range specs {
		approved[s.Front.Slug] = true
	}
	return approved
}

// byStatus returns the tasks with the status.
func byStatus(tasks []taskInfo, status string) []taskInfo {
	return slices.DeleteFunc(slices.Clone(tasks), func(t taskInfo) bool { return t.Status != status })
}

// inProgress returns the In Progress tasks, newest modification time first (R-6).
func inProgress(tasks []taskInfo) []taskInfo {
	list := byStatus(tasks, "In Progress")
	slices.SortFunc(list, func(a, b taskInfo) int {
		return cmp.Or(b.mtime.Compare(a.mtime), cmp.Compare(a.ID, b.ID))
	})
	return list
}

// blocked returns the Blocked tasks, by priority and then ID.
func blocked(tasks []taskInfo) []taskInfo {
	list := byStatus(tasks, "Blocked")
	sortByPriority(list)
	return list
}

// next returns the Todo tasks of Approved specs and of general, with all depends_on tasks Done (R-7).
func next(tasks []taskInfo, approved map[string]bool) []taskInfo {
	status := map[string]string{}
	for _, t := range tasks {
		status[t.ID] = t.Status
	}
	list := byStatus(tasks, "Todo")
	list = slices.DeleteFunc(list, func(t taskInfo) bool {
		return !approved[t.Spec()] || !depsDone(t, status)
	})
	sortByPriority(list)
	return list
}

// depsDone reports whether all depends_on tasks of t are Done. A task that does not exist is not Done.
func depsDone(t taskInfo, status map[string]string) bool {
	return !slices.ContainsFunc(t.DependsOn, func(dep string) bool { return status[dep] != "Done" })
}

func sortByPriority(list []taskInfo) {
	slices.SortFunc(list, func(a, b taskInfo) int {
		return cmp.Or(cmp.Compare(a.Priority, b.Priority), cmp.Compare(a.ID, b.ID))
	})
}

// firstStep returns the text of the first unchecked step of the Plan section, or "".
func firstStep(body string) string {
	plan, _ := markdown.Section(body, "Plan")
	for _, line := range splitLines(plan) {
		if m := stepPattern.FindStringSubmatch(strings.TrimRight(line, " \t\r")); m != nil {
			return m[1]
		}
	}
	return ""
}

// noteLines returns the non-empty lines of the Notes section, sub-items included, with no list marker.
func noteLines(body string) []string {
	notes, _ := markdown.Section(body, "Notes")
	var lines []string
	for _, line := range splitLines(notes) {
		line = strings.TrimSpace(line)
		line = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* "))
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}
