---
topic: go-file-modes
description: How Go sets file and directory modes, and how to test a failed os.Stat. Use it when code writes, copies or replaces files and must keep their permission bits.
---

- `os.CreateTemp` makes a file with mode 0600 and `os.MkdirTemp` makes a directory with mode 0700 (Go documentation). A file that replaces another file by rename gets these modes unless the code sets the mode.
- `os.OpenFile`, `os.WriteFile` and `os.MkdirAll` apply the umask to the mode. `os.Chmod` and `File.Chmod` set the exact bits. To keep a mode such as 0664, call `Chmod` after the create. Examples: `docfile.Write` (`internal/docs/docfile/write.go`) and `copyFile` (`internal/project/data.go`), tested by `TestWriteKeepsMode` and `TestMoveDataKeepsMode`.
- To make `os.Stat` fail with an error that is not `fs.ErrNotExist` in a writable directory, use a symlink loop (`os.Symlink(path, path)`): it gives ELOOP. A directory with no search permission gives a permission error, but root ignores the permission, so skip that test for root. Examples: `TestWriteStatLoop` and `TestWriteStatFails` (`internal/docs/docfile/write_unix_test.go`).
