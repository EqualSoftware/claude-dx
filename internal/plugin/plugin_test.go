package plugin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func readManifest(t *testing.T, dir string) map[string]string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, ".claude-plugin", "plugin.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("plugin.json is not valid JSON: %v", err)
	}
	return m
}

func TestEnsureExtracts(t *testing.T) {
	cache := t.TempDir()
	dir, err := Ensure(cache, "v1.2.3", "/bin/claude-dx")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(cache, "plugin", DirName("v1.2.3", "/bin/claude-dx")); dir != want || !strings.HasPrefix(filepath.Base(dir), "v1.2.3-") {
		t.Fatalf("dir = %q, want %q", dir, want)
	}
	m := readManifest(t, dir)
	if m["name"] != "dx" || m["version"] != "v1.2.3" {
		t.Fatalf("manifest = %v", m)
	}
}

// TestRulesExtracted checks that the launcher can give rules.md, and plugin#R-1, R-2, R-7: 15 lines or less.
func TestRulesExtracted(t *testing.T) {
	dir, err := Ensure(t.TempDir(), "v1.2.3", "/bin/claude-dx")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "rules.md"))
	if err != nil {
		t.Fatal(err)
	}
	if lines := strings.Count(string(data), "\n"); !strings.HasPrefix(string(data), "# dx rules\n") || lines > 15 {
		t.Fatalf("rules.md has %d lines:\n%s", lines, data)
	}
}

func TestEnsureReusesRelease(t *testing.T) {
	cache := t.TempDir()
	dir, err := Ensure(cache, "v1.0.0", "/bin/claude-dx")
	if err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(dir, "marker")
	if err := os.WriteFile(marker, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	// plugin#R-9: one version and one path give one directory.
	if again, err := Ensure(cache, "v1.0.0", "/bin/claude-dx"); err != nil || again != dir {
		t.Fatalf("Ensure again = %q, %v; want %q", again, err, dir)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("a release version was extracted again")
	}
}

// TestEnsureDevKeepsRunningSession checks plugin#R-10: a dev build with other files (here, another
// binary path) extracts to a new directory, and the directory of the first session does not change.
func TestEnsureDevKeepsRunningSession(t *testing.T) {
	cache := t.TempDir()
	first, err := Ensure(cache, "dev", "/a/claude-dx")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(filepath.Base(first), "dev-") {
		t.Fatalf("dev directory = %q", first)
	}
	marker := filepath.Join(first, "marker")
	if err := os.WriteFile(marker, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	second, err := Ensure(cache, "dev", "/b/claude-dx")
	if err != nil || second == first {
		t.Fatalf("second dev directory = %q, %v", second, err)
	}
	again, err := Ensure(cache, "dev", "/a/claude-dx")
	if err != nil || again != first {
		t.Fatalf("same dev build = %q, %v; want %q", again, err, first)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("a dev start changed the files of the first session")
	}
}

// TestEnsureReleasePerBinary checks plugin#R-9: one release version and two binary paths give two
// directories, and each .mcp.json starts its own path.
func TestEnsureReleasePerBinary(t *testing.T) {
	cache := t.TempDir()
	first, err := Ensure(cache, "v1.0.0", "/a/claude-dx")
	if err != nil {
		t.Fatal(err)
	}
	second, err := Ensure(cache, "v1.0.0", "/b/claude-dx")
	if err != nil || second == first {
		t.Fatalf("second directory = %q, %v; first %q", second, err, first)
	}
	for dir, binary := range map[string]string{first: "/a/claude-dx", second: "/b/claude-dx"} {
		data, err := os.ReadFile(filepath.Join(dir, ".mcp.json"))
		if err != nil || !strings.Contains(string(data), binary) {
			t.Errorf("%s/.mcp.json = %s, %v; want %s", dir, data, err, binary)
		}
	}
}

func TestEnsureParallel(t *testing.T) {
	cache := t.TempDir()
	var wg sync.WaitGroup
	errs := make(chan error, 10)
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := Ensure(cache, "v2.0.0", "/bin/claude-dx")
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	readManifest(t, filepath.Join(cache, "plugin", DirName("v2.0.0", "/bin/claude-dx")))
	entries, _ := os.ReadDir(filepath.Join(cache, "plugin"))
	if len(entries) != 1 {
		t.Fatalf("plugin dir has %d entries, want 1 (no temp dirs left)", len(entries))
	}
}

func TestJSONEscapeWindowsPath(t *testing.T) {
	raw := `{"command":"` + jsonEscape(`C:\Users\me\claude-dx.exe`) + `"}`
	var v map[string]string
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		t.Fatalf("invalid JSON %s: %v", raw, err)
	}
	if v["command"] != `C:\Users\me\claude-dx.exe` {
		t.Fatalf("command = %q", v["command"])
	}
}
