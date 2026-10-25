package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// release serves a GitHub release with tag, an archive for this platform with the file name
// and the content data, and checksums.txt. A non-empty sum replaces the correct checksum.
type release struct {
	tag  string
	name string
	data []byte
	sum  string
}

func (r release) serve(t *testing.T) *httptest.Server {
	t.Helper()
	archiveName := ArchiveName(r.tag, runtime.GOOS, runtime.GOARCH)
	archive := makeArchive(t, archiveName, r.name, r.data)
	sum := sha256.Sum256(archive)
	if r.sum == "" {
		r.sum = hex.EncodeToString(sum[:])
	}
	files := map[string][]byte{
		archiveName:     archive,
		"checksums.txt": []byte(r.sum + "  " + archiveName + "\n0000  other.zip\n"),
	}
	mux := http.NewServeMux()
	var srv *httptest.Server
	mux.HandleFunc("/latest", func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get("User-Agent") == "" {
			http.Error(w, "no User-Agent", http.StatusForbidden)
			return
		}
		rel := Release{Tag: r.tag}
		for name := range files {
			rel.Assets = append(rel.Assets, Asset{Name: name, URL: srv.URL + "/files/" + name})
		}
		_ = json.NewEncoder(w).Encode(rel)
	})
	mux.HandleFunc("/files/", func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write(files[strings.TrimPrefix(req.URL.Path, "/files/")])
	})
	srv = httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func makeArchive(t *testing.T, archiveName, name string, data []byte) []byte {
	t.Helper()
	var b bytes.Buffer
	if strings.HasSuffix(archiveName, ".zip") {
		zw := zip.NewWriter(&b)
		w, _ := zw.Create(name)
		_, _ = w.Write(data)
		must(t, zw.Close())
		return b.Bytes()
	}
	gz := gzip.NewWriter(&b)
	tw := tar.NewWriter(gz)
	must(t, tw.WriteHeader(&tar.Header{Name: "LICENSE", Mode: 0o644, Size: 1}))
	_, _ = tw.Write([]byte("x"))
	must(t, tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(data))}))
	_, _ = tw.Write(data)
	must(t, tw.Close())
	must(t, gz.Close())
	return b.Bytes()
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func exe(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func TestNewer(t *testing.T) {
	cases := []struct {
		latest, current string
		want            bool
	}{
		{"v0.2.0", "v0.1.0", true},
		{"v0.1.0", "v0.1.0", false},
		{"v0.1.0", "v0.2.0", false},
		{"v0.10.0", "v0.9.0", true},
		{"v0.2.0", "dev", false},
		{"", "v0.1.0", false},
	}
	for _, c := range cases {
		if got := Newer(c.latest, c.current); got != c.want {
			t.Errorf("Newer(%q, %q) = %v", c.latest, c.current, got)
		}
	}
}

func TestUpdateReplacesBinary(t *testing.T) {
	srv := release{tag: "v0.2.0", name: exe("claude-dx"), data: []byte("new")}.serve(t)
	dir := t.TempDir()
	binary := filepath.Join(dir, exe("claude-dx"))
	must(t, os.WriteFile(binary, []byte("old"), 0o755))
	u := Updater{URL: srv.URL + "/latest", Version: "v0.1.0", Cache: filepath.Join(dir, "cache")}
	msg, err := u.Update(context.Background(), binary)
	if err != nil || msg != "claude-dx is updated to v0.2.0." {
		t.Fatalf("Update = %q, %v", msg, err)
	}
	if data, _ := os.ReadFile(binary); string(data) != "new" {
		t.Fatalf("binary = %q", data)
	}
	if s, err := ReadState(u.Cache); err != nil || s.Latest != "v0.2.0" {
		t.Fatalf("state = %+v, %v", s, err)
	}
	if n := Notice(u.Cache, "v0.2.0"); n != "" {
		t.Fatalf("Notice after the update = %q", n)
	}
}

// TestBadChecksum is AC-2: the binary does not change.
func TestBadChecksum(t *testing.T) {
	srv := release{tag: "v0.2.0", name: exe("claude-dx"), data: []byte("new"), sum: strings.Repeat("0", 64)}.serve(t)
	dir := t.TempDir()
	binary := filepath.Join(dir, exe("claude-dx"))
	must(t, os.WriteFile(binary, []byte("old"), 0o755))
	u := Updater{URL: srv.URL + "/latest", Version: "v0.1.0", Cache: dir}
	if _, err := u.Update(context.Background(), binary); err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("Update = %v", err)
	}
	if data, _ := os.ReadFile(binary); string(data) != "old" {
		t.Fatalf("binary = %q", data)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("files in the binary directory: %v", entries)
	}
}

func TestUpToDateAndDev(t *testing.T) {
	srv := release{tag: "v0.2.0", name: exe("claude-dx"), data: []byte("new")}.serve(t)
	u := Updater{URL: srv.URL + "/latest", Version: "v0.2.0", Cache: t.TempDir()}
	if msg, err := u.Update(context.Background(), "unused"); err != nil || msg != "claude-dx is up to date (v0.2.0)." {
		t.Fatalf("Update = %q, %v", msg, err)
	}
	u.Version = "dev"
	if _, err := u.Update(context.Background(), "unused"); err == nil || err.Error() != "dev builds do not update. Install a release" {
		t.Fatalf("Update of dev = %v", err)
	}
}

// TestCheck is AC-3: an update.json from 2 days ago gets a new checked_at. A new file is not checked again.
func TestCheck(t *testing.T) {
	srv := release{tag: "v0.3.0", name: exe("claude-dx"), data: []byte("new")}.serve(t)
	now := time.Date(2026, 10, 25, 10, 0, 0, 0, time.UTC)
	u := Updater{URL: srv.URL + "/latest", Version: "v0.1.0", Cache: t.TempDir(), Now: func() time.Time { return now }}
	old := []byte(`{"checked_at":"2026-10-23T10:00:00Z","latest":"v0.2.0"}`)
	must(t, os.WriteFile(filepath.Join(u.Cache, "update.json"), old, 0o644))
	u.Check(context.Background())
	if s, _ := ReadState(u.Cache); !s.CheckedAt.Equal(now) || s.Latest != "v0.3.0" {
		t.Fatalf("state = %+v", s)
	}
	if n := Notice(u.Cache, "v0.1.0"); n != "claude-dx v0.3.0 is available. Run: claude-dx update" {
		t.Fatalf("Notice = %q", n)
	}
	u.URL = "http://127.0.0.1:1/latest" // a check in less than 24 hours does no request
	u.Now = func() time.Time { return now.Add(time.Hour) }
	u.Check(context.Background())
	if s, _ := ReadState(u.Cache); !s.CheckedAt.Equal(now) {
		t.Fatalf("state after a second check = %+v", s)
	}
}

func TestCheckIgnoresErrors(t *testing.T) {
	cache := t.TempDir()
	Updater{URL: "http://127.0.0.1:1/latest", Version: "v0.1.0", Cache: cache}.Check(context.Background())
	if _, err := ReadState(cache); err == nil {
		t.Fatal("a failed check wrote update.json")
	}
}

// TestUpdateRunningBinary is AC-1: it replaces a claude-dx that runs (Windows locks it),
// then the new binary runs, and its next start deletes the .old file.
func TestUpdateRunningBinary(t *testing.T) {
	if testing.Short() {
		t.Skip("builds binaries")
	}
	dir := t.TempDir()
	binary := filepath.Join(dir, "bin", exe("claude-dx"))
	build(t, binary, "v0.1.0")
	newBinary := filepath.Join(dir, "new", exe("claude-dx"))
	build(t, newBinary, "v0.2.0")
	data, err := os.ReadFile(newBinary)
	must(t, err)
	srv := release{tag: "v0.2.0", name: exe("claude-dx"), data: data}.serve(t)

	freshState(t, dir)
	running := exec.Command(binary, "mcp")
	running.Dir = dir
	running.Env = cacheEnv(dir)
	stdin, err := running.StdinPipe()
	must(t, err)
	must(t, running.Start())
	u := Updater{URL: srv.URL + "/latest", Version: "v0.1.0", Cache: filepath.Join(dir, "cache")}
	_, err = u.Update(context.Background(), binary)
	_ = stdin.Close()
	_ = running.Wait()
	must(t, err)

	cmd := exec.Command(binary, "--version")
	cmd.Env = cacheEnv(dir)
	out, err := cmd.Output()
	if err != nil || string(out) != "claude-dx v0.2.0\n" {
		t.Fatalf("new binary --version = %q, %v", out, err)
	}
	start := exec.Command(binary, "--resume") // claude is not in PATH, so it stops after the delete
	start.Env = append(cacheEnv(dir), "PATH="+dir)
	_ = start.Run()
	if _, err := os.Stat(binary + ".old"); err == nil {
		t.Fatal("the start did not delete the .old file")
	}
}

func build(t *testing.T, out, version string) {
	t.Helper()
	cmd := exec.Command("go", "build", "-o", out, "-ldflags=-X main.version="+version, "../../cmd/claude-dx")
	if msg, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, msg)
	}
}

// cacheEnv returns the environment of a child process with its cache in dir/base.
func cacheEnv(dir string) []string {
	base := filepath.Join(dir, "base")
	return append(os.Environ(), "HOME="+base, "XDG_CACHE_HOME="+base, "LocalAppData="+base)
}

// freshState writes a new update.json in the cache of cacheEnv, so that the child does no network request.
// os.UserCacheDir is base/Library/Caches on macOS and base on Linux and Windows.
func freshState(t *testing.T, dir string) {
	t.Helper()
	base := filepath.Join(dir, "base")
	for _, cache := range []string{filepath.Join(base, "claude-dx"), filepath.Join(base, "Library", "Caches", "claude-dx")} {
		u := Updater{Version: "v0.1.0", Cache: cache}
		must(t, u.writeState("v0.1.0"))
	}
}
