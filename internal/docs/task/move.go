package task

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/EqualSoftware/claude-dx/internal/docs/docfile"
	"github.com/EqualSoftware/claude-dx/internal/markdown"
)

// writeNew writes the file of a new task atomically, so a crash never leaves a part of a file (A-2).
func writeNew(path, data string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return docfile.Write(path, []byte(data))
}

// Delete deletes the tasks ids. It checks all of them before it deletes a file (R-9).
func (s Store) Delete(ids []string) error {
	if len(ids) == 0 {
		return errors.New("give one or more task IDs")
	}
	return docfile.Locked(s.Lock, func() error {
		tasks, _, err := s.all()
		if err != nil {
			return err
		}
		for _, id := range ids {
			if _, ok := tasks[id]; !ok && !s.exists(id) { // a bad file of the named task may go
				return fmt.Errorf("task %s does not exist", id)
			}
		}
		if err := checkDependents(tasks, ids); err != nil {
			return err
		}
		for _, id := range ids {
			path, _ := s.path(id)
			if err := os.Remove(path); err != nil {
				return err
			}
			if err := removeEmpty(filepath.Dir(path)); err != nil {
				return err
			}
		}
		return nil
	})
}

// checkDependents refuses a task out of ids that depends on a task of ids.
func checkDependents(tasks map[string]Front, ids []string) error {
	for _, id := range sortedIDs(tasks) {
		if slices.Contains(ids, id) {
			continue
		}
		for _, dep := range tasks[id].DependsOn {
			if slices.Contains(ids, dep) {
				return fmt.Errorf("task %s depends on task %s", id, dep)
			}
		}
	}
	return nil
}

// removeEmpty removes the spec folder dir of the tasks if it has no entry.
func removeEmpty(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) > 0 {
		return err
	}
	return os.Remove(dir)
}

// sortedIDs returns the IDs of tasks in order, so an error names the same task each time.
func sortedIDs(tasks map[string]Front) []string {
	ids := make([]string, 0, len(tasks))
	for id := range tasks {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	return ids
}

// move is one task of a Move: its slug and what to do.
type move struct {
	slug  string
	write bool // write the new file
	drop  bool // remove the old file
}

// Move moves the tasks ids of one spec to the spec to (R-9). It checks all of them before it changes a
// file. A retry after a crash finishes the work, because a task in to with the Notes line
// "Moved from <from>." counts as moved (Q-29, Q-43).
func (s Store) Move(ids []string, to string) error {
	from, err := oneSpec(ids)
	if err != nil {
		return err
	}
	if to == from {
		return fmt.Errorf("the tasks are in %s already", to)
	}
	if !validSlug(to) {
		return fmt.Errorf("no spec has the slug %s", to)
	}
	return docfile.Locked(s.Lock, func() error {
		moves, err := s.planMove(ids, from, to)
		if err != nil {
			return err
		}
		for _, m := range moves {
			if err := s.apply(m, from, to); err != nil {
				return err
			}
		}
		return removeEmpty(filepath.Join(s.Dir, from))
	})
}

// oneSpec returns the spec of ids, or an error if ids is empty or has tasks of more than one spec.
func oneSpec(ids []string) (string, error) {
	if len(ids) == 0 {
		return "", errors.New("give one or more task IDs")
	}
	from, _, _ := strings.Cut(ids[0], "/")
	for _, id := range ids[1:] {
		if slug, _, _ := strings.Cut(id, "/"); slug != from {
			return "", errors.New("the tasks must be of one spec")
		}
	}
	return from, nil
}

// planMove checks the move of ids from the spec from to the spec to, and returns the work to do.
func (s Store) planMove(ids []string, from, to string) ([]move, error) {
	if to != general {
		if _, _, err := s.specLocked(to); err != nil {
			return nil, err
		}
	}
	tasks, _, err := s.all()
	if err != nil {
		return nil, err
	}
	moves := make([]move, 0, len(ids))
	slugs := make([]string, 0, len(ids))
	for _, id := range ids {
		m, err := s.planOne(id, to, from, tasks)
		if err != nil {
			return nil, err
		}
		moves = append(moves, m)
		slugs = append(slugs, m.slug)
	}
	return moves, checkSplit(tasks, from, to, slugs)
}

// planOne returns the work to move the task id to the spec to.
func (s Store) planOne(id, to, from string, tasks map[string]Front) (move, error) {
	_, slug, _ := strings.Cut(id, "/")
	target := to + "/" + slug
	_, inSource := tasks[id]
	_, inTarget := tasks[target]
	moved := inTarget && s.movedFrom(target, from)
	switch {
	case inSource && !inTarget:
		return move{slug: slug, write: true, drop: true}, nil
	case inSource && moved:
		return move{slug: slug, drop: true}, nil
	case inTarget && !moved:
		return move{}, fmt.Errorf("task %s already exists", target)
	case moved:
		return move{slug: slug}, nil
	}
	if s.exists(id) {
		_, err := s.load(id) // the file is bad: say why
		return move{}, err
	}
	return move{}, fmt.Errorf("task %s does not exist", id)
}

// movedFrom reports whether the Notes of the task id have the line of a move from the spec from.
func (s Store) movedFrom(id, from string) bool {
	d, err := s.load(id)
	if err != nil {
		return false
	}
	notes, _ := markdown.Section(d.body, "Notes")
	return strings.Contains(notes+"\n", "- Moved from "+from+".")
}

// checkSplit refuses a move that makes a dependency between the specs from and to: a task that stays in
// from and depends on a task that goes, or a task that goes and depends on a task that stays.
func checkSplit(tasks map[string]Front, from, to string, slugs []string) error {
	for _, id := range sortedIDs(tasks) {
		f := tasks[id]
		if f.Spec() != from {
			continue
		}
		_, slug, _ := strings.Cut(id, "/")
		goes := slices.Contains(slugs, slug)
		for _, dep := range f.DependsOn {
			_, depSlug, _ := strings.Cut(dep, "/")
			switch depGoes := slices.Contains(slugs, depSlug); {
			case goes && !depGoes:
				return fmt.Errorf("task %s depends on task %s, which stays in %s", id, dep, from)
			case !goes && depGoes:
				return fmt.Errorf("task %s depends on task %s, which goes to %s", id, dep, to)
			}
		}
	}
	return nil
}

// apply does one move: it writes the task in the spec to with a new ID, new dependencies, no covers and a
// Notes line, and then removes the old file.
func (s Store) apply(m move, from, to string) error {
	old := filepath.Join(s.Dir, from, m.slug+".md")
	if m.write {
		d, err := s.load(from + "/" + m.slug)
		if err != nil {
			return err
		}
		if err := retarget(d, from, to); err != nil {
			return err
		}
		if err := writeNew(filepath.Join(s.Dir, to, m.slug+".md"), markdown.Join(d.front, d.body)); err != nil {
			return err
		}
	}
	if !m.drop {
		return nil
	}
	return os.Remove(old)
}

// retarget changes the task d of the spec from into a task of the spec to (Q-17, Q-41).
func retarget(d *doc, from, to string) error {
	note := "Moved from " + from + "."
	if len(d.f.Covers) > 0 {
		note += " Old covers: " + strings.Join(d.f.Covers, ", ") + "."
	}
	deps := make([]string, len(d.f.DependsOn))
	for i, dep := range d.f.DependsOn {
		deps[i] = to + strings.TrimPrefix(dep, from)
	}
	_, slug, _ := strings.Cut(d.f.ID, "/")
	d.front = markdown.SetField(d.front, "id", to+"/"+slug)
	d.front = markdown.SetField(d.front, "covers", []string{})
	d.front = markdown.SetField(d.front, "depends_on", deps)
	return d.addNote(note)
}
