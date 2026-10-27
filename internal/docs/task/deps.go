package task

import (
	"errors"
	"fmt"
	"strings"
)

// checkDeps refuses a dependency on a task of another spec, a task that does not exist, the same task and
// a cycle (R-4). The caller holds the docs lock.
func (s Store) checkDeps(f Front) error {
	if len(f.DependsOn) == 0 {
		return nil
	}
	tasks, err := s.all()
	if err != nil {
		return err
	}
	for _, dep := range f.DependsOn {
		if err := checkDep(f, dep, tasks); err != nil {
			return err
		}
	}
	tasks[f.ID] = f
	if path := depPath(f.ID, f.ID, tasks, map[string]bool{}); path != nil {
		return fmt.Errorf("the dependency makes a cycle: %s", strings.Join(path, " -> "))
	}
	return nil
}

// checkDep checks one dependency dep of the task f.
func checkDep(f Front, dep string, tasks map[string]Front) error {
	if dep == f.ID {
		return errors.New("a task does not depend on itself")
	}
	if slug, _, _ := strings.Cut(dep, "/"); slug != f.Spec() {
		return fmt.Errorf("task %s is in another spec. A task depends only on tasks of its spec", dep)
	}
	if _, ok := tasks[dep]; !ok {
		return fmt.Errorf("task %s does not exist", dep)
	}
	return nil
}

// depPath returns the depends_on path from start to target, or nil.
func depPath(start, target string, tasks map[string]Front, seen map[string]bool) []string {
	for _, dep := range tasks[start].DependsOn {
		if dep == target {
			return []string{start, dep}
		}
		if seen[dep] {
			continue
		}
		seen[dep] = true
		if path := depPath(dep, target, tasks, seen); path != nil {
			return append([]string{start}, path...)
		}
	}
	return nil
}
