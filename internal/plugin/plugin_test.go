package plugin

import (
	"encoding/json"
	"os"
	"path/filepath"
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
	if want := filepath.Join(cache, "plugin", "v1.2.3"); dir != want {
		t.Fatalf("dir = %q, want %q", dir, want)
	}
	m := readManifest(t, dir)
	if m["name"] != "dx" || m["version"] != "v1.2.3" {
		t.Fatalf("manifest = %v", m)
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
	if _, err := Ensure(cache, "v1.0.0", "/bin/claude-dx"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("a release version was extracted again")
	}
}

func TestEnsureExtractsDevAgain(t *testing.T) {
	cache := t.TempDir()
	dir, err := Ensure(cache, "dev", "/bin/claude-dx")
	if err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(dir, "marker")
	if err := os.WriteFile(marker, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Ensure(cache, "dev", "/bin/claude-dx"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("a dev version was not extracted again")
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
	readManifest(t, filepath.Join(cache, "plugin", "v2.0.0"))
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
