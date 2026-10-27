package update

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// Apply downloads the archive of r for this platform, verifies its SHA-256 with checksums.txt
// and replaces binary with the binary in the archive. It changes nothing if a step fails.
func (u Updater) Apply(ctx context.Context, r Release, binary string) error {
	name := ArchiveName(r.Tag, runtime.GOOS, runtime.GOARCH)
	archive, err := u.download(ctx, r, name)
	if err != nil {
		return err
	}
	sums, err := u.download(ctx, r, "checksums.txt")
	if err != nil {
		return err
	}
	if err := verify(archive, sums, name); err != nil {
		return err
	}
	data, err := extract(archive, name, exeName())
	if err != nil {
		return err
	}
	if err := install(data, binary); err != nil {
		return err
	}
	u.saveState(r.Tag)
	return nil
}

// exeName is the name of the binary in the release archive.
func exeName() string {
	if runtime.GOOS == "windows" {
		return "claude-dx.exe"
	}
	return "claude-dx"
}

// saveState writes update.json. The binary is already good here, so an error is only a warning.
func (u Updater) saveState(tag string) {
	if err := u.writeState(tag); err != nil {
		fmt.Fprintf(os.Stderr, "warning: cannot write %s: %v\n", stateFile, err)
	}
}

// ArchiveName returns the name of the release archive, as GoReleaser makes it (.goreleaser.yaml).
func ArchiveName(tag, goos, goarch string) string {
	ext := "tar.gz"
	if goos == "windows" {
		ext = "zip"
	}
	return fmt.Sprintf("claude-dx_%s_%s_%s.%s", strings.TrimPrefix(tag, "v"), goos, goarch, ext)
}

func (u Updater) download(ctx context.Context, r Release, name string) ([]byte, error) {
	for _, a := range r.Assets {
		if a.Name == name {
			if requireHTTPS && !strings.HasPrefix(a.URL, "https://") {
				return nil, fmt.Errorf("the URL of %s is not https", name)
			}
			return u.get(ctx, a.URL)
		}
	}
	return nil, fmt.Errorf("release %s has no file %s", r.Tag, name)
}

// verify compares the SHA-256 of archive with the line of name in checksums.txt.
func verify(archive, sums []byte, name string) error {
	sum := sha256.Sum256(archive)
	scanner := bufio.NewScanner(bytes.NewReader(sums))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 || fields[1] != name {
			continue
		}
		if fields[0] != hex.EncodeToString(sum[:]) {
			return fmt.Errorf("the SHA-256 of %s does not match checksums.txt. The binary is not changed", name)
		}
		return nil
	}
	return fmt.Errorf("checksums.txt has no line for %s. The binary is not changed", name)
}

// extract returns the file name from the root of the archive.
func extract(archive []byte, archiveName, name string) ([]byte, error) {
	if strings.HasSuffix(archiveName, ".zip") {
		return extractZip(archive, name)
	}
	return extractTarGz(archive, name)
}

func extractZip(archive []byte, name string) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return nil, err
	}
	f, err := zr.Open(name)
	if err != nil {
		return nil, fmt.Errorf("the archive has no %s: %w", name, err)
	}
	defer func() { _ = f.Close() }()
	if info, err := f.Stat(); err != nil || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s in the archive is not a regular file", name)
	}
	return readFile(f, name)
}

// readFile reads a file of the archive. It refuses empty data.
func readFile(r io.Reader, name string) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxSize+1))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("%s in the archive is empty", name)
	}
	if int64(len(data)) > maxSize {
		return nil, fmt.Errorf("%s in the archive is larger than %d bytes", name, maxSize)
	}
	return data, nil
}

func extractTarGz(archive []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, err
	}
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("the archive has no %s", name)
		}
		if err != nil {
			return nil, err
		}
		if h.Name == name && h.Typeflag == tar.TypeReg {
			return readFile(tr, name)
		}
	}
}

// install writes data to a temporary file next to binary and replaces binary with it.
func install(data []byte, binary string) error {
	dir := filepath.Dir(binary)
	f, err := os.CreateTemp(dir, ".claude-dx-*")
	if err != nil {
		return fmt.Errorf("cannot write to %s: %w. Run the update as a user that can write to this directory", dir, err)
	}
	_, err = f.Write(data)
	err = errors.Join(err, f.Close(), os.Chmod(f.Name(), 0o755))
	if err == nil {
		err = replace(f.Name(), binary)
	}
	if err != nil {
		_ = os.Remove(f.Name())
	}
	return err
}

// replace moves tmp to binary. Windows locks a running .exe but lets it be renamed,
// so on Windows the current binary becomes <binary>.old. The launcher deletes it at the next start.
func replace(tmp, binary string) error {
	if runtime.GOOS != "windows" {
		return os.Rename(tmp, binary)
	}
	return replaceKeepingOld(tmp, binary)
}

const closeSessions = "close the running claude-dx sessions and run the update again"

// replaceKeepingOld renames binary to <binary>.old, or to <binary>.old.<n> if a locked .old file stays,
// and then moves tmp to binary.
func replaceKeepingOld(tmp, binary string) error {
	RemoveOld(binary)
	return renameOld(tmp, binary)
}

// rename is os.Rename. Tests replace it.
var rename = os.Rename

func renameOld(tmp, binary string) error {
	old, err := oldName(binary)
	if err != nil {
		return err
	}
	if err := rename(binary, old); err != nil {
		return fmt.Errorf("cannot rename %s: %w. %s", binary, err, closeSessions)
	}
	err = rename(tmp, binary)
	if err == nil {
		return nil
	}
	err = fmt.Errorf("cannot write %s: %w. %s", binary, err, closeSessions)
	if rerr := rename(old, binary); rerr != nil {
		return errors.Join(err, fmt.Errorf("the update failed; the previous binary is at %s: %w", old, rerr))
	}
	return err
}

// oldName returns <binary>.old after it removes a file with that name. It returns a free
// <binary>.old.<n> if .old is not a regular file or cannot be removed.
func oldName(binary string) (string, error) {
	old := binary + ".old"
	info, err := os.Lstat(old)
	if errors.Is(err, fs.ErrNotExist) {
		return old, nil
	}
	if err == nil && info.Mode().IsRegular() && os.Remove(old) == nil {
		return old, nil
	}
	return freeOld(binary)
}

// freeOld returns the first <binary>.old.<n> that does not exist. It stops after 100 tries
// and at an error that is not "does not exist".
func freeOld(binary string) (string, error) {
	for n := 1; n <= 100; n++ {
		p := fmt.Sprintf("%s.old.%d", binary, n)
		_, err := os.Lstat(p)
		if errors.Is(err, fs.ErrNotExist) {
			return p, nil
		}
		if err != nil {
			return "", fmt.Errorf("cannot check %s: %w. %s", p, err, closeSessions)
		}
	}
	return "", fmt.Errorf("%s.old.1 to %s.old.100 exist. %s", binary, binary, closeSessions)
}

// RemoveOld deletes the <binary>.old and <binary>.old.<n> files that an update left. It ignores errors,
// because Windows keeps a file that a running process uses.
func RemoveOld(binary string) {
	dir := filepath.Dir(binary)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	re := regexp.MustCompile(`^` + regexp.QuoteMeta(filepath.Base(binary)) + `\.old(\.[0-9]+)?$`)
	for _, e := range entries {
		if !e.IsDir() && re.MatchString(e.Name()) {
			_ = os.Remove(filepath.Join(dir, e.Name()))
		}
	}
}
