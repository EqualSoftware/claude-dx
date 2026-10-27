package spec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDelete(t *testing.T) {
	s := newStore(t)
	session := connect(t, s)
	del := func(slug string, extra map[string]any) map[string]any {
		args := map[string]any{"slug": slug, "reason": "Split."}
		for k, v := range extra {
			args[k] = v
		}
		return args
	}
	run(t, session, []step{
		{"spec_create", createArgs("a", "A"), "Created a."},
		{"spec_create", createArgs("b", "B"), "Created b."},
	})
	writeFile(t, filepath.Join(s.Dir, "a", "diagram.svg"), "<svg/>")
	writeFile(t, filepath.Join(s.Dir, "a", "sub", "x.png"), "png")
	run(t, session, []step{
		{"spec_delete", map[string]any{"slug": "a", "reason": " "}, "ERR Give the reason."},
		{"spec_delete", del("a", map[string]any{"move_to": "b", "discard_files": true}), "ERR Give move_to or discard_files, not both."},
		{"spec_delete", del("a", map[string]any{"move_to": "a"}), "ERR The argument move_to must be another spec."},
		{"spec_delete", del("nope", nil), `ERR Spec "nope" does not exist.`},
		{"spec_delete", del("a", nil), "ERR The folder docs/specs/a has changes that are not committed. Commit them first."},
	})
	writeFile(t, filepath.Join(s.Dir, "b", "diagram.svg"), "<svg/>")
	commitAll(t, s.Root, "docs: specs")
	run(t, session, []step{
		{"spec_delete", del("a", nil), "ERR The folder docs/specs/a has other files: diagram.svg, sub/x.png. Give move_to or discard_files."},
		{"spec_delete", del("a", map[string]any{"move_to": "zzz"}), `ERR Spec "zzz" does not exist.`},
		{"spec_delete", del("a", map[string]any{"move_to": "b"}), `ERR File "diagram.svg" exists in docs/specs/b.`},
	})
	if err := os.Remove(filepath.Join(s.Dir, "b", "diagram.svg")); err != nil {
		t.Fatal(err)
	}
	commitAll(t, s.Root, "docs: remove diagram")
	run(t, session, []step{
		{"spec_delete", del("a", map[string]any{"move_to": "b"}), "Deleted a. Give this reason in the commit message: Split."},
		// The moved files are untracked in b.
		{"spec_delete", del("b", map[string]any{"discard_files": true}), "ERR The folder docs/specs/b has changes that are not committed. Commit them first."},
	})
	for _, f := range []string{"b/diagram.svg", "b/sub/x.png"} {
		if _, err := os.Stat(filepath.Join(s.Dir, filepath.FromSlash(f))); err != nil {
			t.Errorf("%s: %v", f, err)
		}
	}
	if _, err := os.Stat(filepath.Join(s.Dir, "a")); !os.IsNotExist(err) {
		t.Errorf("docs/specs/a exists: %v", err)
	}
	commitAll(t, s.Root, "docs: delete a")
	run(t, session, []step{
		{"spec_update_section", map[string]any{"slug": "b", "section": "1", "body": "Changed."}, "Updated section 1 of b."},
		{"spec_delete", del("b", map[string]any{"discard_files": true}), "ERR The folder docs/specs/b has changes that are not committed. Commit them first."},
	})
	commitAll(t, s.Root, "docs: change b")
	run(t, session, []step{
		{"spec_delete", del("b", map[string]any{"discard_files": true}), "Deleted b. Give this reason in the commit message: Split."},
		{"spec_list", map[string]any{}, "No specs match.\n"},
	})
	index, _ := os.ReadFile(filepath.Join(s.Dir, "README.md"))
	if string(index) != indexHeader {
		t.Errorf("README.md =\n%s", index)
	}
	if strings.Contains(string(index), "[a]") {
		t.Error("README.md has a")
	}
}
