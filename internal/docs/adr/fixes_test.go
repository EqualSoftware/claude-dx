package adr

import (
	"os"
	"strings"
	"testing"
)

func TestCreateRefusesBadText(t *testing.T) {
	s := newStore(t)
	cases := map[string]func(*New){
		"owners newline":      func(a *New) { a.Owners = "A\nB" },
		"option newline":      func(a *New) { a.Alternatives[0].Option = "A\nB" },
		"reason newline":      func(a *New) { a.Alternatives[0].Reason = "A\r\nB" },
		"title newline":       func(a *New) { a.Title = "A\nB" },
		"empty context":       func(a *New) { a.Context = " " },
		"empty decision":      func(a *New) { a.Decision = "" },
		"unclosed fence":      func(a *New) { a.Context = "```\ncode" },
		"level 2 heading":     func(a *New) { a.Decision = "## Evidence\nx" },
		"repeated section":    func(a *New) { a.Context = "### Evidence\nx" },
		"fence in follow-up":  func(a *New) { a.Consequences.FollowUp = "~~~\nx" },
		"heading in positive": func(a *New) { a.Consequences.Positive = "ok\n## Fake" },
	}
	for name, change := range cases {
		a := sample("Case")
		a.Alternatives = []Alternative{{Option: "Other", Reason: "slow"}}
		change(&a)
		if _, err := s.Create(a); err == nil {
			t.Errorf("%s: Create succeeded", name)
		}
	}
	if list, _, _ := s.List(); len(list) != 0 {
		t.Errorf("%d ADRs were made", len(list))
	}
	a := sample("Fine")
	a.Context = "Text.\n\n### Details\nMore."
	if _, err := s.Create(a); err != nil {
		t.Errorf("a subheading is refused: %v", err)
	}
}

func TestStatusOnlyInHeader(t *testing.T) {
	s := newStore(t)
	path, err := s.Create(sample("Two status"))
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	doc := strings.Replace(string(data), "The context.", "The context.\n\n- Status: Fake", 1)
	if err := os.WriteFile(path, []byte(doc), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := s.Update("1", "Accepted", "", ""); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(path)
	if !strings.Contains(string(data), "- Status: Accepted\n- Date") || !strings.Contains(string(data), "\n- Status: Fake\n") {
		t.Errorf("file =\n%s", data)
	}
}

func TestCRLFReadAndEdit(t *testing.T) {
	s := newStore(t)
	path, err := s.Create(sample("Windows"))
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if err := os.WriteFile(path, []byte(strings.ReplaceAll(string(data), "\n", "\r\n")), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, err := s.Get("1", "Decision"); err != nil || got != "The decision." {
		t.Fatalf("Get = %q, %v", got, err)
	}
	if err := s.Update("1", "Accepted", "Decision", "New decision."); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(path)
	text := string(data)
	if strings.Contains(strings.ReplaceAll(text, "\r\n", ""), "\n") {
		t.Errorf("mixed line ends:\n%q", text)
	}
	if !strings.Contains(text, "- Status: Accepted\r\n") || !strings.Contains(text, "## Decision\r\n\r\nNew decision.\r\n") {
		t.Errorf("file = %q", text)
	}
	if got, _ := s.Get("1", "Decision"); got != "New decision." {
		t.Errorf("Get after edit = %q", got)
	}
}
