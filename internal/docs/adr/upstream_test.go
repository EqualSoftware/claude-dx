package adr

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// git runs git in dir with a test identity and no commit signature.
func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	args = append([]string{"-C", dir, "-c", "user.name=Tester", "-c", "user.email=t@example.com", "-c", "commit.gpgsign=false"}, args...)
	if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// pushStore returns a store in a repository whose upstream is a bare remote. Its ADR-001 is pushed.
func pushStore(t *testing.T) (Store, string) {
	t.Helper()
	remote := t.TempDir()
	git(t, remote, "init", "-q", "--bare", "-b", "main")
	root := t.TempDir()
	git(t, root, "init", "-q", "-b", "main")
	git(t, root, "remote", "add", "origin", remote)
	s := newStore(t)
	s.Dir = filepath.Join(root, "docs", "decisions")
	_, err := s.Create(sample("Use stdio"))
	must(t, err)
	push(t, root)
	return s, root
}

// push commits all files and pushes them to the upstream.
func push(t *testing.T, root string) {
	t.Helper()
	git(t, root, "add", "-A")
	git(t, root, "commit", "-q", "-m", "docs: change")
	git(t, root, "push", "-q", "-u", "origin", "main")
}

// TestUpdateUpstream is the R-1 test of spec pushed-adr-lock.
func TestUpdateUpstream(t *testing.T) {
	s, root := pushStore(t)
	// Proposed in the upstream: the change is allowed, also when it is Accepted locally.
	must(t, s.Update("ADR-001", "Accepted", "", ""))
	must(t, s.Update("ADR-001", "", "Evidence", "More."))
	push(t, root)
	refused(t, s.Update("ADR-001", "", "Evidence", "Again."),
		"ADR-001 is Accepted in the upstream branch. A pushed decision does not change. Record a new ADR and use adr_supersede.")
	// A local ADR that is not pushed can change.
	_, err := s.Create(sample("Local"))
	must(t, err)
	must(t, s.Update("ADR-002", "Accepted", "", ""))
}

func TestUpdateUpstreamSuperseded(t *testing.T) {
	s, root := pushStore(t)
	_, err := s.Create(sample("New"))
	must(t, err)
	must(t, s.Update("ADR-002", "Accepted", "", ""))
	must(t, s.Supersede("ADR-001", "ADR-002"))
	push(t, root)
	refused(t, s.Update("ADR-001", "", "Evidence", "Again."), "is Superseded by ADR-002 in the upstream branch")
}

func TestUpdateNoUpstream(t *testing.T) {
	root := t.TempDir()
	git(t, root, "init", "-q", "-b", "main")
	s := newStore(t)
	s.Dir = filepath.Join(root, "docs", "decisions")
	_, err := s.Create(sample("Use stdio"))
	must(t, err)
	git(t, root, "add", "-A")
	git(t, root, "commit", "-q", "-m", "docs: add")
	must(t, s.Update("ADR-001", "Accepted", "", ""))
	must(t, s.Update("ADR-001", "", "Evidence", "More."))
}

// TestUpdateOutsideGit: a directory outside git allows the change.
func TestUpdateOutsideGit(t *testing.T) {
	s := newStore(t)
	_, err := s.Create(sample("Use stdio"))
	must(t, err)
	must(t, s.Update("ADR-001", "Accepted", "", ""))
	must(t, s.Update("ADR-001", "", "Evidence", "More."))
}

// TestSupersedePushed is the R-2 test: the old ADR changes only its status line, and a pushed new ADR gets the line.
func TestSupersedePushed(t *testing.T) {
	s, root := pushStore(t)
	_, err := s.Create(sample("New"))
	must(t, err)
	must(t, s.Update("ADR-002", "Accepted", "", ""))
	push(t, root)
	before, _ := s.Get("ADR-001", "")
	must(t, s.Supersede("ADR-001", "ADR-002"))
	after, _ := s.Get("ADR-001", "")
	if after != strings.Replace(before, "- Status: Proposed", "- Status: Superseded by ADR-002", 1) {
		t.Fatalf("old =\n%s", after)
	}
	doc, _ := s.Get("ADR-002", "")
	if strings.Count(doc, "- Supersedes: ADR-001") != 1 {
		t.Fatalf("new =\n%s", doc)
	}
}
