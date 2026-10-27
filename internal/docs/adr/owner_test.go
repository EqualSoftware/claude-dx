package adr

import (
	"strings"
	"testing"
)

func TestDefaultOwner(t *testing.T) {
	tests := []struct {
		name   string
		author func() string
		owners string
		want   string
	}{
		{"author", func() string { return "Tester" }, "", "- Owners: Tester\n"},
		{"no author name", func() string { return "" }, "", "- Owners: Maintainer\n"},
		{"no author function", nil, "", "- Owners: Maintainer\n"},
		{"given owner", func() string { return "Tester" }, "Team", "- Owners: Team\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newStore(t)
			s.Author = tt.author
			adr := sample("Title")
			adr.Owners = tt.owners
			if _, err := s.Create(adr); err != nil {
				t.Fatal(err)
			}
			doc, err := s.Get("ADR-001", "")
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(doc, tt.want) {
				t.Errorf("ADR has no %q:\n%s", tt.want, doc)
			}
		})
	}
}
