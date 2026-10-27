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
	"os"
	"path/filepath"
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
	data, err := extract(archive, name, filepath.Base(binary))
	if err != nil {
		return err
	}
	if err := install(data, binary); err != nil {
		return err
	}
	return u.writeState(r.Tag)
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
	return io.ReadAll(io.LimitReader(f, maxSize))
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
		if h.Name == name {
			return io.ReadAll(io.LimitReader(tr, maxSize))
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
	old := binary + ".old"
	_ = os.Remove(old)
	if err := os.Rename(binary, old); err != nil {
		return err
	}
	if err := os.Rename(tmp, binary); err != nil {
		return errors.Join(err, os.Rename(old, binary))
	}
	return nil
}
