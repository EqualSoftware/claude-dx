package task

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// checkDeps checks the dependencies of the task f before a write (ADR-013, step 1). f.ID is empty
// for a new task. It refuses an ID with no task file, the same task, a Done or Cancelled task, a later
// milestone and a cycle. It returns one warning line for each dependency with a lower priority. It checks
// only the IDs that are not in old, because an old dependency that became Done must not stop a change.
// The cycle check uses all IDs.
func (s Store) checkDeps(f Front, old []string) (string, error) {
	if len(f.DependsOn) == 0 {
		return "", nil
	}
	tasks, _, err := s.List(Filter{"status": "all"})
	if err != nil && !errors.Is(err, ErrNoDir) {
		return "", err
	}
	byID := map[string]Front{}
	for _, t := range tasks {
		byID[t.Front.ID] = t.Front
	}
	var warnings string
	for _, id := range f.DependsOn {
		if slices.Contains(old, id) {
			continue
		}
		warning, err := checkDep(f, id, byID)
		if err != nil {
			return "", err
		}
		warnings += warning
	}
	if f.ID == "" {
		return warnings, nil
	}
	byID[f.ID] = f
	if path := depPath(f.ID, f.ID, byID, nil); path != nil {
		return "", fmt.Errorf("depends_on makes a cycle: %s", strings.Join(path, " -> "))
	}
	return warnings, nil
}

// checkDep checks one dependency id of the task f.
func checkDep(f Front, id string, byID map[string]Front) (string, error) {
	if strings.EqualFold(id, f.ID) {
		return "", fmt.Errorf("%s cannot depend on itself", f.ID)
	}
	dep, ok := byID[id]
	if !ok {
		return "", fmt.Errorf("dependency %s does not exist. Use task_list", id)
	}
	if dep.Status == "Done" || dep.Status == "Cancelled" {
		return "", fmt.Errorf("dependency %s is %s, so it sets no order. Put it in related", id, dep.Status)
	}
	if slices.Compare(version(dep.Milestone), version(f.Milestone)) > 0 {
		return "", fmt.Errorf("dependency %s has the later milestone %s. Change the milestone of one task", id, dep.Milestone)
	}
	if dep.Priority > f.Priority {
		return fmt.Sprintf("Warning: dependency %s has the lower priority %s. This task waits for it.\n", id, dep.Priority), nil
	}
	return "", nil
}

// depPath returns the depends_on path from start to target, or nil.
func depPath(start, target string, byID map[string]Front, seen map[string]bool) []string {
	if seen == nil {
		seen = map[string]bool{}
	}
	for _, dep := range byID[start].DependsOn {
		if dep == target {
			return []string{start, dep}
		}
		if seen[dep] {
			continue
		}
		seen[dep] = true
		if path := depPath(dep, target, byID, seen); path != nil {
			return append([]string{start}, path...)
		}
	}
	return nil
}
