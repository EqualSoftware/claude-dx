package task

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strconv"
)

var milestonePattern = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`)

// checkMilestone returns an error if m is not vMAJOR.MINOR.PATCH, because the ready order compares versions.
func checkMilestone(m string) error {
	if milestonePattern.MatchString(m) {
		return nil
	}
	return fmt.Errorf("milestone %q is not valid. Use vMAJOR.MINOR.PATCH, example v0.1.0", m)
}

// version returns the numbers of a milestone. A milestone that is not valid comes after all others.
func version(m string) []int {
	parts := milestonePattern.FindStringSubmatch(m)
	if parts == nil {
		return []int{1 << 30}
	}
	v := make([]int, 3)
	for i := range v {
		v[i], _ = strconv.Atoi(parts[i+1]) // The pattern allows only digits.
	}
	return v
}

// sortByMilestone sorts tasks by milestone in SemVer order, then by priority, then by ID.
func sortByMilestone(tasks []Task) {
	sort.SliceStable(tasks, func(i, j int) bool {
		a, b := tasks[i].Front, tasks[j].Front
		if c := slices.Compare(version(a.Milestone), version(b.Milestone)); c != 0 {
			return c < 0
		}
		if a.Priority != b.Priority {
			return a.Priority < b.Priority
		}
		return a.ID < b.ID
	})
}

// ready returns the Approved tasks that have all dependencies in done, in milestone order.
func ready(tasks []Task, done map[string]bool) []Task {
	var out []Task
	for _, t := range tasks {
		if t.Front.Status != "Approved" {
			continue
		}
		if slices.ContainsFunc(t.Front.DependsOn, func(id string) bool { return !done[id] }) {
			continue
		}
		out = append(out, t)
	}
	sortByMilestone(out)
	return out
}
