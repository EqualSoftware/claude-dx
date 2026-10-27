package launcher

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestArgs(t *testing.T) {
	dir := filepath.Join("cache", "plugin", "dev")
	got := Args(dir, []string{"--resume", "-p", "hi"})
	want := []string{
		"--plugin-dir", dir, "--settings", settings,
		"--append-system-prompt-file", filepath.Join(dir, "rules.md"),
		"--resume", "-p", "hi",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Args = %q, want %q", got, want)
	}
}
