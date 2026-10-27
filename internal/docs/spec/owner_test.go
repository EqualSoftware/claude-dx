package spec

import (
	"strings"
	"testing"
)

func TestQuestionDefaultOwner(t *testing.T) {
	tests := []struct {
		name   string
		author func() string
		owner  string
		want   string
	}{
		{"author", func() string { return "Tester" }, "", "| Tester | Open |"},
		{"no author name", func() string { return "" }, "", "| Maintainer | Open |"},
		{"given owner", func() string { return "Tester" }, "Equal", "| Equal | Open |"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newStore(t)
			s.Author = tt.author
			if err := s.Create(New{Slug: "a", Title: "A", Milestone: "v0.1.0", Priority: "P1"}); err != nil {
				t.Fatal(err)
			}
			if _, _, err := s.AddQuestion("a", "Q?", tt.owner); err != nil {
				t.Fatal(err)
			}
			text, err := s.Get("a", "")
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(text, tt.want) {
				t.Errorf("SPEC.md has no %q:\n%s", tt.want, text)
			}
		})
	}
}
