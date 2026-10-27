package doctor

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClaudeVersion(t *testing.T) {
	cases := map[string]string{
		"2.1.308 (Claude Code)\n": "OK claude 2.1.308 in PATH",
		"2.2.0 (Claude Code)":     "OK claude 2.2.0 in PATH",
		"2.1.99 (Claude Code)":    "FAIL claude 2.1.99 is older than 2.1.308. Fix: run claude update.",
		"":                        `FAIL cannot read the claude version from "". Fix: install Claude Code again.`,
	}
	for out, want := range cases {
		ok, text := claudeVersion(out)
		got := map[bool]string{true: "OK ", false: "FAIL "}[ok] + text
		if got != want {
			t.Errorf("claudeVersion(%q) = %q, want %q", out, got, want)
		}
	}
}

func TestRun(t *testing.T) {
	t.Setenv("PATH", t.TempDir()) // no claude
	cache := filepath.Join(t.TempDir(), "claude-dx")
	var b bytes.Buffer
	if Run(&b, "v0.1.0", cache) {
		t.Fatal("Run passed with no claude and no plugin")
	}
	want := "FAIL  claude is not in PATH. Fix: install Claude Code: https://docs.claude.com/claude-code.\n" +
		"OK    cache " + cache + " is writable\n" +
		"FAIL  plugin v0.1.0 is not extracted. Fix: run claude-dx once.\n"
	if b.String() != want {
		t.Fatalf("output =\n%s\nwant\n%s", b.String(), want)
	}
	manifest := filepath.Join(cache, "plugin", "v0.1.0", ".claude-plugin")
	if err := os.MkdirAll(manifest, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(manifest, "plugin.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	b.Reset()
	Run(&b, "v0.1.0", cache)
	if !strings.Contains(b.String(), "OK    plugin v0.1.0 is extracted") {
		t.Fatalf("output =\n%s", b.String())
	}
}
