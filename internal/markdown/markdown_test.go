package markdown

import (
	"strings"
	"testing"
)

const task = `---
id: TASK-001
title: "Fix: the # sign"  # quoted title
status: Draft  # Draft / In Review / Approved
area: [cli]
---

# TASK-001: Fix

## 1. Summary
Old summary.

## 7. Design
### 7.1 Architecture
Flow.

` + "```text" + `
## Not a heading
` + "```" + `

### 7.2 Public API
API.

## 13. Open Questions
None.
`

func TestSplitJoin(t *testing.T) {
	front, body := Split(task)
	if front == "" || body == "" {
		t.Fatal("Split found no frontmatter")
	}
	if Join(front, body) != task {
		t.Fatal("Join(Split(x)) != x")
	}
	if f, b := Split("# No front\n"); f != "" || b != "# No front\n" {
		t.Fatalf("Split with no frontmatter = %q, %q", f, b)
	}
}

func TestDecode(t *testing.T) {
	front, _ := Split(task)
	var v struct {
		ID     string   `yaml:"id"`
		Title  string   `yaml:"title"`
		Status string   `yaml:"status"`
		Area   []string `yaml:"area"`
	}
	if err := Decode(front, &v); err != nil {
		t.Fatal(err)
	}
	if v.ID != "TASK-001" || v.Title != "Fix: the # sign" || v.Status != "Draft" || len(v.Area) != 1 {
		t.Fatalf("Decode = %+v", v)
	}
}

func TestSetFieldReplacesLine(t *testing.T) {
	front, _ := Split(task)
	tests := []struct {
		key   string
		value any
		want  string
	}{
		{"status", "In Review", "status: In Review\n"},
		{"title", "New: title", "title: 'New: title'\n"},
		{"area", []string{"cli", "docs"}, "area: [cli, docs]\n"},
		{"depends_on", []string{}, "depends_on: []\n"},
		{"platforms", []string{"Linux", "Windows (amd64, arm64)"}, "platforms: [Linux, 'Windows (amd64, arm64)']\n"},
	}
	for _, tt := range tests {
		got := SetField(front, tt.key, tt.value)
		if !strings.Contains(got, tt.want) {
			t.Errorf("SetField(%s) =\n%s\nwant line %q", tt.key, got, tt.want)
		}
	}
}

func TestSetFieldReplacesBlockValue(t *testing.T) {
	tests := []struct {
		name  string
		front string
		want  string
	}{
		{"indented list", "id: T\narea:\n  - a\n  - b\nstatus: Draft\n", "id: T\narea: [c]\nstatus: Draft\n"},
		{"list with no indent", "id: T\narea:\n- a\n- b\nstatus: Draft\n", "id: T\narea: [c]\nstatus: Draft\n"},
		{"list as last key", "id: T\narea:\n  - a\n  - b\n", "id: T\narea: [c]\n"},
		{"scalar with empty line", "id: T\narea: |\n  one\n\n  two\n\nstatus: Draft\n", "id: T\narea: [c]\n\nstatus: Draft\n"},
		{"flow list", "id: T\narea: [a, b]\nstatus: Draft\n", "id: T\narea: [c]\nstatus: Draft\n"},
	}
	for _, tt := range tests {
		got := SetField(tt.front, "area", []string{"c"})
		if got != tt.want {
			t.Errorf("%s: SetField =\n%q\nwant\n%q", tt.name, got, tt.want)
		}
		var v struct {
			Area   []string `yaml:"area"`
			Status string   `yaml:"status"`
		}
		if err := Decode(got, &v); err != nil {
			t.Errorf("%s: Decode: %v", tt.name, err)
		}
	}
}

func TestSection(t *testing.T) {
	_, body := Split(task)
	tests := map[string]string{
		"1":              "Old summary.",
		"summary":        "Old summary.",
		"7.2":            "API.",
		"Public API":     "API.",
		"Open Questions": "None.",
	}
	for ref, want := range tests {
		got, ok := Section(body, ref)
		if !ok || got != want {
			t.Errorf("Section(%q) = %q, %v; want %q", ref, got, ok, want)
		}
	}
	design, _ := Section(body, "7")
	if !strings.Contains(design, "### 7.1 Architecture") || !strings.Contains(design, "## Not a heading") {
		t.Errorf("Section(7) must include subsections and code blocks:\n%s", design)
	}
	if _, ok := Section(body, "99"); ok {
		t.Error("Section(99) found a section")
	}
}

func TestReplaceSection(t *testing.T) {
	_, body := Split(task)
	got, err := ReplaceSection(body, "1", "New summary.\nTwo lines.")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "## 1. Summary\n\nNew summary.\nTwo lines.\n\n## 7. Design") {
		t.Fatalf("ReplaceSection =\n%s", got)
	}
	last, err := ReplaceSection(body, "13", "Q-1 open.")
	if err != nil || !strings.Contains(last, "## 13. Open Questions\n\nQ-1 open.") {
		t.Fatalf("ReplaceSection(last) = %q, %v", last, err)
	}
	if _, err := ReplaceSection(body, "99", "x"); err == nil {
		t.Error("ReplaceSection(99) did not fail")
	}
}
