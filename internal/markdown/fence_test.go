package markdown

import (
	"strings"
	"testing"
)

const fenceDoc = "# T\n\n## One\nText.\n\n## Two\nMore.\n"

func TestHeadingsFences(t *testing.T) {
	for name, body := range map[string]string{
		"tilde":    "## A\n~~~\n## Fake\n~~~\n## B\nx",
		"indented": "## A\n   ```\n## Fake\n   ```\n## B\nx",
		"long":     "## A\n````\n```\n## Fake\n````\n## B\nx",
		"info":     "## A\n```go\n## Fake\n```\n## B\nx",
	} {
		if _, ok := Section(body, "Fake"); ok {
			t.Errorf("%s: a heading in a code block is a section", name)
		}
		if text, ok := Section(body, "B"); !ok || text != "x" {
			t.Errorf("%s: Section(B) = %q, %v", name, text, ok)
		}
	}
	// A shorter run or another char does not close the fence: the next heading stays in the block.
	if _, ok := Section("## A\n````\n```\n## Fake\nx", "Fake"); ok {
		t.Error("a shorter run closed the fence")
	}
	if _, ok := Section("## A\n~~~\n```\n## Fake\nx", "Fake"); ok {
		t.Error("another char closed the fence")
	}
}

func TestReplaceSectionOddFence(t *testing.T) {
	if _, err := ReplaceSection(fenceDoc, "One", "```\ncode"); err == nil || !strings.Contains(err.Error(), "not closed") {
		t.Errorf("odd fence: err = %v", err)
	}
	if _, err := ReplaceSection(fenceDoc, "One", "~~~\ncode"); err == nil {
		t.Error("an open tilde fence is accepted")
	}
	got, err := ReplaceSection(fenceDoc, "One", "~~~\n## code\n~~~")
	if err != nil || !strings.Contains(got, "~~~\n## code\n~~~\n\n## Two") {
		t.Errorf("closed fence: %q, %v", got, err)
	}
}

func TestReplaceSectionDuplicateHeading(t *testing.T) {
	body := "# T\n\n## 1. Brief\nx\n\n### Notes\nnote\n\n## 5. Requirements\ny\n\n## Notes\nz\n"
	if got, _ := Section(body, "Notes"); got != "z" {
		t.Errorf("Notes = %q, want the level 2 section", got)
	}
	if _, err := ReplaceSection(body, "Brief", "### Notes\nx"); err == nil {
		t.Error("### Notes in a Brief is accepted")
	}
	if _, err := ReplaceSection(body, "1", "### 5 Requirements\nx"); err == nil {
		t.Error("### 5 Requirements in section 1 is accepted")
	}
	if got, _ := Section(body, "5"); got != "y" {
		t.Errorf("section 5 = %q", got)
	}
	if _, err := ReplaceSection(body, "Brief", "### Other\nx"); err != nil {
		t.Errorf("a new subheading is refused: %v", err)
	}
}

func TestCRLF(t *testing.T) {
	body := strings.ReplaceAll(fenceDoc, "\n", "\r\n")
	if got, ok := Section(body, "One"); !ok || got != "Text." {
		t.Errorf("Section = %q, %v", got, ok)
	}
	if n, ok := SectionNumber(strings.ReplaceAll("## 3. Goals\r\nx\r\n", "\n", "\n"), "Goals"); !ok || n != "3" {
		t.Errorf("SectionNumber = %q, %v", n, ok)
	}
	got, err := ReplaceSection(body, "One", "New.\nLine.")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.ReplaceAll(got, "\r\n", ""), "\n") {
		t.Errorf("mixed line ends: %q", got)
	}
	if !strings.Contains(got, "## One\r\n\r\nNew.\r\nLine.\r\n\r\n## Two") {
		t.Errorf("ReplaceSection = %q", got)
	}
}
