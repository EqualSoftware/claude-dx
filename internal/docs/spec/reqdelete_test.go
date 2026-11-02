package spec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EqualSoftware/claude-dx/internal/project"
)

// writeTask writes a task file that covers rids.
func writeTask(t *testing.T, s Store, slug, name, covers string) {
	t.Helper()
	dir := filepath.Join(filepath.Dir(s.Lock), project.Tasks, slug)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	text := "---\nid: " + slug + "/" + name + "\ncovers: [" + covers + "]\n---\n# T\n"
	if err := os.WriteFile(filepath.Join(dir, name+".md"), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestReqDelete(t *testing.T) {
	s := newStore(t)
	session := connect(t, s)
	del := func(rid string) map[string]any { return map[string]any{"slug": "a", "rid": rid} }
	run(t, session, []step{
		{"spec_create", createArgs("a", "A"), "Created a."},
		{"spec_req_add", map[string]any{"slug": "a", "text": "One."}, "Added R-1."},
		{"spec_req_add", map[string]any{"slug": "a", "text": "Two."}, "Added R-2."},
		{"spec_req_add", map[string]any{"slug": "a", "text": "Three."}, "Added R-3."},
		{"spec_req_add", map[string]any{"slug": "a", "text": "Four."}, "Added R-4."},
		{"spec_req_set", map[string]any{"slug": "a", "rid": "R-1", "status": "Done"}, "Updated R-1."},
		{"spec_req_delete", del("R-9"), "ERR Requirement R-9 does not exist."},
		{"spec_req_delete", del("R-1"), "ERR Cannot delete R-1: its status is Done. Only a Planned requirement is deleted."},
	})
	writeTask(t, s, "a", "work", "R-2, R-3")
	writeTask(t, s, "a", "other", "R-4")
	c := commitAll(t, s.Root, "feat: x\n\nRefs: a#R-3")
	run(t, session, []step{
		{"spec_req_delete", del("R-2"), "ERR Cannot delete R-2: tasks cover it: a/work. Change their covers first."},
		{"spec_req_delete", del("r-3"), "ERR Cannot delete R-3: commits refer to it: " + c + "; tasks cover it: a/work. Change their covers first."},
	})
	if err := os.Remove(filepath.Join(filepath.Dir(s.Lock), project.Tasks, "a", "other.md")); err != nil {
		t.Fatal(err)
	}
	run(t, session, []step{{"spec_req_delete", del("R-4"), "Deleted R-4."}})
	text, _ := s.Get("a", "")
	if strings.Contains(text, "| R-4 |") || !strings.Contains(text, "| Deleted R-4. |") {
		t.Errorf("R-4 not deleted:\n%s", text)
	}
}

func TestReqDeleteNoReuse(t *testing.T) {
	s := newStore(t)
	session := connect(t, s)
	run(t, session, []step{
		{"spec_create", createArgs("a", "A"), "Created a."},
		{"spec_req_add", map[string]any{"slug": "a", "text": "One."}, "Added R-1."},
		{"spec_req_delete", map[string]any{"slug": "a", "rid": "R-1"}, "Deleted R-1."},
		{"spec_req_add", map[string]any{"slug": "a", "text": "Two."}, "Added R-2."},
		{"spec_req_delete", map[string]any{"slug": "a", "rid": "R-2"}, "Deleted R-2."},
		{"spec_req_add", map[string]any{"slug": "a", "text": "Three."}, "Added R-3."},
	})
	text, _ := s.Get("a", "5")
	if strings.Contains(text, "| R-1 |") || !strings.Contains(text, "| R-3 | Three. | Planned |") {
		t.Errorf("section 5 =\n%s", text)
	}
}

func TestReqDeleteReview(t *testing.T) {
	s := newStore(t)
	session := connect(t, s)
	steps := []step{
		{"spec_create", createArgs("a", "A"), "Created a."},
		{"spec_req_add", map[string]any{"slug": "a", "text": "One."}, "Added R-1."},
		{"spec_req_add", map[string]any{"slug": "a", "text": "Two."}, "Added R-2."},
	}
	for _, ref := range []string{"1", "2", "3", "4", "6", "7"} {
		steps = append(steps, step{"spec_update_section", map[string]any{"slug": "a", "section": ref, "body": filled[ref]}, "Updated section " + ref + " of a."})
	}
	steps = append(steps,
		step{"spec_set_status", map[string]any{"slug": "a", "status": "In Review"}, "a is In Review."},
		step{"spec_set_status", map[string]any{"slug": "a", "status": "Approved"}, "a is Approved."},
		step{"spec_req_delete", map[string]any{"slug": "a", "rid": "R-2"}, "Deleted R-2. The scope changed, so a is In Review. The user must approve it again."},
	)
	run(t, session, steps)
	if f, _ := s.Get("a", ""); !strings.Contains(f, "status: In Review") {
		t.Errorf("spec is not In Review:\n%s", f)
	}
}
