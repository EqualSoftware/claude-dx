package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// A state write error after a good replace is not a failure.
func TestStateErrorIsNotFailure(t *testing.T) {
	dir := t.TempDir()
	cache := filepath.Join(dir, "cache")
	must(t, os.WriteFile(cache, []byte("x"), 0o644)) // a file, so MkdirAll fails
	srv := release{tag: "v0.2.0", name: exe("claude-dx"), data: []byte("new")}.serve(t)
	binary := filepath.Join(dir, exe("claude-dx"))
	must(t, os.WriteFile(binary, []byte("old"), 0o755))
	u := Updater{URL: srv.URL + "/latest", Version: "v0.1.0", Cache: cache}
	if msg, err := u.Update(context.Background(), binary); err != nil || !strings.Contains(msg, "updated") {
		t.Fatalf("Update = %q, %v", msg, err)
	}
	if data, _ := os.ReadFile(binary); string(data) != "new" {
		t.Fatalf("binary = %q", data)
	}
	u.Version = "v0.2.0"
	if msg, err := u.Update(context.Background(), binary); err != nil || !strings.Contains(msg, "up to date") {
		t.Fatalf("Update up to date = %q, %v", msg, err)
	}
}

func TestApplyDeadline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/latest" {
			_, _ = w.Write([]byte(`{"tag_name":"v0.2.0","assets":[{"name":"` + ArchiveName("v0.2.0", runtime.GOOS, runtime.GOARCH) +
				`","browser_download_url":"http://` + r.Host + `/slow"}]}`))
			return
		}
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
	}))
	defer srv.Close()
	old := applyTime
	applyTime = 100 * time.Millisecond
	t.Cleanup(func() { applyTime = old })
	u := Updater{URL: srv.URL + "/latest", Version: "v0.1.0", Cache: t.TempDir()}
	start := time.Now()
	_, err := u.Update(context.Background(), filepath.Join(t.TempDir(), "x"))
	if err == nil || !strings.Contains(err.Error(), "the download took more than") || time.Since(start) > 3*time.Second {
		t.Fatalf("Update = %v after %v", err, time.Since(start))
	}
}

func TestRefuseNonHTTPS(t *testing.T) {
	requireHTTPS = true
	t.Cleanup(func() { requireHTTPS = false })
	u := Updater{Version: "v0.1.0", Cache: t.TempDir()}
	r := Release{Tag: "v0.2.0", Assets: []Asset{{Name: ArchiveName("v0.2.0", runtime.GOOS, runtime.GOARCH), URL: "http://example.com/a"}}}
	if err := u.Apply(context.Background(), r, filepath.Join(t.TempDir(), "x")); err == nil || !strings.Contains(err.Error(), "not https") {
		t.Fatalf("Apply = %v", err)
	}
	srv := httptest.NewServer(http.RedirectHandler("http://example.com/b", http.StatusFound))
	defer srv.Close()
	if _, err := u.get(context.Background(), srv.URL); err == nil || !strings.Contains(err.Error(), "not https") {
		t.Fatalf("get with a redirect to http = %v", err)
	}
}

func tarWith(t *testing.T, h *tar.Header, data []byte) []byte {
	t.Helper()
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	tw := tar.NewWriter(gz)
	h.Size = int64(len(data))
	must(t, tw.WriteHeader(h))
	_, _ = tw.Write(data)
	must(t, tw.Close())
	must(t, gz.Close())
	return b.Bytes()
}

func TestExtractRefusals(t *testing.T) {
	link := tarWith(t, &tar.Header{Name: "claude-dx", Typeflag: tar.TypeSymlink, Linkname: "/bin/sh"}, nil)
	if _, err := extractTarGz(link, "claude-dx"); err == nil {
		t.Fatal("a symlink entry was accepted")
	}
	empty := tarWith(t, &tar.Header{Name: "claude-dx", Typeflag: tar.TypeReg, Mode: 0o755}, nil)
	if _, err := extractTarGz(empty, "claude-dx"); err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("empty entry = %v", err)
	}
	if _, err := extractZip(makeArchive(t, "a.zip", "claude-dx", nil), "claude-dx"); err == nil {
		t.Fatal("an empty zip entry was accepted")
	}
}

// The new binary comes from the fixed name in the archive, whatever the name of the old binary.
func TestRenamedBinary(t *testing.T) {
	srv := release{tag: "v0.2.0", name: exeName(), data: []byte("new")}.serve(t)
	dir := t.TempDir()
	binary := filepath.Join(dir, "renamed")
	must(t, os.WriteFile(binary, []byte("old"), 0o755))
	u := Updater{URL: srv.URL + "/latest", Version: "v0.1.0", Cache: dir}
	if _, err := u.Update(context.Background(), binary); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(binary); string(data) != "new" {
		t.Fatalf("binary = %q", data)
	}
}

// A .old name that cannot be deleted (here a full directory, which RemoveOld skips) gets a unique name.
func TestReplaceKeepingOld(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "claude-dx")
	must(t, os.WriteFile(binary, []byte("old"), 0o755))
	locked := binary + ".old"
	must(t, os.MkdirAll(filepath.Join(locked, "x"), 0o755))
	tmp := filepath.Join(dir, "tmp")
	must(t, os.WriteFile(tmp, []byte("new"), 0o755))
	if err := renameOld(tmp, binary); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(binary); string(data) != "new" {
		t.Fatalf("binary = %q", data)
	}
	if data, _ := os.ReadFile(binary + ".old.1"); string(data) != "old" {
		t.Fatalf(".old.1 = %q", data)
	}
}

func TestRemoveOld(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "claude-dx")
	for _, n := range []string{"claude-dx.older", "claude-dx.old-x", "claude-dx.old", "claude-dx.old.2"} {
		must(t, os.WriteFile(filepath.Join(dir, n), []byte("x"), 0o644))
	}
	must(t, os.MkdirAll(filepath.Join(dir, "claude-dx.old.3", "x"), 0o755))
	RemoveOld(binary)
	for n, want := range map[string]bool{"claude-dx.older": true, "claude-dx.old-x": true, "claude-dx.old.3": true, "claude-dx.old": false, "claude-dx.old.2": false} {
		if _, err := os.Lstat(filepath.Join(dir, n)); (err == nil) != want {
			t.Errorf("%s exists = %v, want %v", n, err == nil, want)
		}
	}
}

func TestFreeOldLimit(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "claude-dx")
	for n := 1; n <= 100; n++ {
		must(t, os.WriteFile(binary+".old."+strconv.Itoa(n), nil, 0o644))
	}
	if _, err := freeOld(binary); err == nil || !strings.Contains(err.Error(), closeSessions) {
		t.Fatalf("freeOld = %v", err)
	}
}

// The second rename fails: the old binary comes back. If the restore fails too, the error names the old path.
func TestRenameOldRestore(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "claude-dx")
	must(t, os.WriteFile(binary, []byte("old"), 0o755))
	err := renameOld(filepath.Join(dir, "missing"), binary)
	if err == nil || !strings.Contains(err.Error(), closeSessions) {
		t.Fatalf("error = %v", err)
	}
	if data, _ := os.ReadFile(binary); string(data) != "old" {
		t.Fatalf("binary after the restore = %q", data)
	}
	orig := rename
	t.Cleanup(func() { rename = orig })
	calls := 0
	rename = func(a, b string) error {
		calls++
		if calls == 1 {
			return os.Rename(a, b)
		}
		return errors.New("locked")
	}
	err = renameOld(filepath.Join(dir, "tmp"), binary)
	if err == nil || !strings.Contains(err.Error(), "the update failed; the previous binary is at "+binary+".old") {
		t.Fatalf("error = %v", err)
	}
}

func TestExtractNotRegular(t *testing.T) {
	dir := tarWith(t, &tar.Header{Name: "claude-dx", Typeflag: tar.TypeDir, Mode: 0o755}, nil)
	if _, err := extractTarGz(dir, "claude-dx"); err == nil {
		t.Fatal("a tar directory entry was accepted")
	}
	var b bytes.Buffer
	zw := zip.NewWriter(&b)
	_, _ = zw.Create("claude-dx/")
	h := &zip.FileHeader{Name: "link"}
	h.SetMode(os.ModeSymlink | 0o777)
	w, _ := zw.CreateHeader(h)
	_, _ = w.Write([]byte("/bin/sh"))
	must(t, zw.Close())
	for _, n := range []string{"claude-dx", "link"} {
		if _, err := extractZip(b.Bytes(), n); err == nil {
			t.Fatalf("zip entry %s was accepted", n)
		}
	}
}

func TestExtractTooLarge(t *testing.T) {
	old := maxSize
	maxSize = 4
	t.Cleanup(func() { maxSize = old })
	if _, err := extractTarGz(tarWith(t, &tar.Header{Name: "claude-dx", Typeflag: tar.TypeReg, Mode: 0o755}, []byte("12345")), "claude-dx"); err == nil || !strings.Contains(err.Error(), "larger") {
		t.Fatalf("error = %v", err)
	}
	if _, err := extractTarGz(tarWith(t, &tar.Header{Name: "claude-dx", Typeflag: tar.TypeReg, Mode: 0o755}, []byte("1234")), "claude-dx"); err != nil {
		t.Fatal(err)
	}
}

func TestReplaceKeepingOldError(t *testing.T) {
	dir := t.TempDir()
	err := replaceKeepingOld(filepath.Join(dir, "tmp"), filepath.Join(dir, "missing"))
	if err == nil || !strings.Contains(err.Error(), "close the running claude-dx sessions and run the update again") {
		t.Fatalf("error = %v", err)
	}
}

// An empty directory named <binary>.old is not removed: the old binary gets a unique name.
func TestRenameOldEmptyDir(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "claude-dx")
	must(t, os.WriteFile(binary, []byte("old"), 0o755))
	must(t, os.Mkdir(binary+".old", 0o755))
	tmp := filepath.Join(dir, "tmp")
	must(t, os.WriteFile(tmp, []byte("new"), 0o755))
	must(t, renameOld(tmp, binary))
	if info, err := os.Lstat(binary + ".old"); err != nil || !info.IsDir() {
		t.Fatalf(".old = %v, %v", info, err)
	}
	if data, _ := os.ReadFile(binary + ".old.1"); string(data) != "old" {
		t.Fatalf(".old.1 = %q", data)
	}
}

func TestCheckCancelled(t *testing.T) {
	cache := t.TempDir()
	srv := release{tag: "v0.3.0", name: exeName(), data: []byte("new")}.serve(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	Updater{URL: srv.URL + "/latest", Version: "v0.1.0", Cache: cache}.Check(ctx)
	if _, err := ReadState(cache); err == nil {
		t.Fatal("a cancelled check wrote update.json")
	}
}
