package spec

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/EqualSoftware/claude-dx/internal/docs/docfile"
	"github.com/EqualSoftware/claude-dx/internal/project"
)

// Delete deletes the folder of a spec (ADR-016). Git keeps the old version, so it refuses changes that
// are not committed. moveTo is the slug of a spec that gets the other files of the folder; discard
// deletes them. The spec must have no other files if both are empty, so a delete never loses a file by chance.
func (s Store) Delete(slug, reason, moveTo string, discard bool) error {
	if strings.TrimSpace(reason) == "" {
		return errors.New("give the reason")
	}
	if moveTo != "" && discard {
		return errors.New("give move_to or discard_files, not both")
	}
	if moveTo == slug {
		return errors.New("the argument move_to must be another spec")
	}
	return docfile.Locked(s.Lock, func() error {
		path, err := s.find(slug)
		if err != nil {
			return err
		}
		dir := filepath.Dir(path)
		if err := s.checkDelete(slug, dir); err != nil {
			return err
		}
		others, err := otherFiles(dir)
		if err != nil {
			return err
		}
		if err := s.keep(slug, dir, others, moveTo, discard); err != nil {
			return err
		}
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
		return s.WriteIndex()
	})
}

// checkDelete refuses a spec that has tasks or changes that are not committed.
func (s Store) checkDelete(slug, dir string) error {
	if err := s.checkTasks(slug); err != nil {
		return err
	}
	return s.checkCommitted(slug, dir)
}

// checkTasks refuses a spec that has local tasks in tasks/<slug>/ of the project data directory, which
// is the folder of the docs lock (ADR-015). The spec package does not read the tasks (TASK-064).
func (s Store) checkTasks(slug string) error {
	entries, err := os.ReadDir(filepath.Join(filepath.Dir(s.Lock), project.Tasks, slug))
	if docfile.NotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	n := 0
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".md" {
			n++
		}
	}
	switch n {
	case 0:
		return nil
	case 1:
		return fmt.Errorf("spec %s has 1 task. Move or delete the task first, with task_move or task_delete", slug)
	}
	return fmt.Errorf("spec %s has %d tasks. Move or delete the tasks first, with task_move or task_delete", slug, n)
}

// checkCommitted refuses a folder with changes or files that git does not have.
func (s Store) checkCommitted(slug, dir string) error {
	out, err := exec.Command("git", "-C", s.Root, "status", "--porcelain", "--untracked-files=all", "--", dir).Output()
	if err != nil {
		return fmt.Errorf("git status failed: %w", gitError(err))
	}
	if len(out) > 0 {
		return fmt.Errorf("the folder docs/specs/%s has changes that are not committed. Commit them first", slug)
	}
	return nil
}

// otherFiles returns the files of dir other than SPEC.md, as slash paths relative to dir.
func otherFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, e fs.DirEntry, err error) error {
		if err != nil || e.IsDir() {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err == nil && rel != fileName {
			files = append(files, filepath.ToSlash(rel))
		}
		return err
	})
	return files, err
}

// keep moves the other files to the spec moveTo, or refuses them if there is no moveTo and no discard.
func (s Store) keep(slug, dir string, others []string, moveTo string, discard bool) error {
	if len(others) == 0 || discard {
		return nil
	}
	if moveTo == "" {
		return fmt.Errorf("the folder docs/specs/%s has other files: %s. Give move_to or discard_files", slug, strings.Join(others, ", "))
	}
	target, err := s.find(moveTo)
	if err != nil {
		return err
	}
	target = filepath.Dir(target)
	for _, f := range others {
		if _, err := os.Lstat(filepath.Join(target, filepath.FromSlash(f))); err == nil {
			return fmt.Errorf("file %q exists in docs/specs/%s", f, moveTo)
		}
	}
	for _, f := range others {
		to := filepath.Join(target, filepath.FromSlash(f))
		if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
			return err
		}
		if err := docfile.Move(filepath.Join(dir, filepath.FromSlash(f)), to); err != nil {
			return err
		}
	}
	return nil
}
