package spec

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/EqualSoftware/claude-dx/internal/docs/docfile"
	"github.com/EqualSoftware/claude-dx/internal/project"
)

func TestRequirementsAndSetDoneLocked(t *testing.T) {
	s := newStore(t)
	session := connect(t, s)
	run(t, session, []step{
		{"spec_create", createArgs("a", "A"), "Created a."},
		{"spec_req_add", map[string]any{"slug": "a", "text": "One."}, "Added R-1."},
		{"spec_req_add", map[string]any{"slug": "a", "text": "Two."}, "Added R-2."},
		{"spec_req_add", map[string]any{"slug": "a", "text": "Three."}, "Added R-3."},
	})
	if err := docfile.Locked(s.Lock, func() error { return checkLocked(t, s) }); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(s.path("a"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{"| R-1 | One. | Done |", "| R-2 | Two. | Planned |", "| R-3 | Three. | Done |", "| Tester | R-1, R-3: Done. |"} {
		if !strings.Contains(text, want) {
			t.Errorf("spec has no %q:\n%s", want, text)
		}
	}
	if n := strings.Count(text, "R-1, R-3: Done."); n != 1 {
		t.Errorf("%d Done rows, want 1", n)
	}
	index, _ := os.ReadFile(filepath.Join(s.Dir, "README.md"))
	if !strings.Contains(string(index), "2/3") {
		t.Errorf("README.md does not show 2/3 requirements:\n%s", index)
	}
}

// checkLocked tests RequirementsLocked and sets R-1 and R-3 to Done two times. The caller holds the lock.
func checkLocked(t *testing.T, s Store) error {
	status, rids, err := s.RequirementsLocked("a")
	if err != nil {
		return err
	}
	if status != "Draft" || !slices.Equal(rids, []string{"R-1", "R-2", "R-3"}) {
		t.Errorf("RequirementsLocked = %q, %q", status, rids)
	}
	if _, _, err := s.RequirementsLocked("nope"); err == nil {
		t.Error("RequirementsLocked of a missing spec: no error")
	}
	if err := s.SetDoneLocked("a", []string{"R-1", "R-3"}); err != nil {
		return err
	}
	// Done on Done changes nothing, so a retry adds no row.
	return s.SetDoneLocked("a", []string{"R-1", "R-3"})
}

func TestDeleteRefusesTasks(t *testing.T) {
	s := newStore(t)
	session := connect(t, s)
	run(t, session, []step{
		{"spec_create", createArgs("a", "A"), "Created a."},
	})
	commitAll(t, s.Root, "docs: spec a")
	tasks := filepath.Join(filepath.Dir(s.Lock), project.Tasks, "a")
	writeFile(t, filepath.Join(tasks, "notes.txt"), "not a task")
	writeFile(t, filepath.Join(tasks, "one.md"), "# One\n")
	del := map[string]any{"slug": "a", "reason": "Gone."}
	run(t, session, []step{
		{"spec_delete", del, "ERR Spec a has 1 task. Move or delete the task first, with task_move or task_delete."},
	})
	writeFile(t, filepath.Join(tasks, "two.md"), "# Two\n")
	run(t, session, []step{
		{"spec_delete", del, "ERR Spec a has 2 tasks. Move or delete the tasks first, with task_move or task_delete."},
	})
	for _, name := range []string{"one.md", "two.md"} {
		if err := os.Remove(filepath.Join(tasks, name)); err != nil {
			t.Fatal(err)
		}
	}
	run(t, session, []step{
		{"spec_delete", del, "Deleted a. Give this reason in the commit message: Gone."},
	})
}
