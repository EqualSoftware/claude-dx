---
topic: go-windows-file-sharing
description: How Go opens files on Windows and why a rename over an open file can fail. Use it when code replaces a file that another process can read.
---

- On Windows, Go `os.Open` and `os.ReadFile` open a file with `FILE_SHARE_READ | FILE_SHARE_WRITE` only, with no `FILE_SHARE_DELETE` (`GOROOT/src/syscall/syscall_windows.go:395`, go1.27.3). Thus `os.Rename` over that file fails while another handle has it open, also from another process.
- To replace a file that readers can have open, write a temporary file in the same directory and rename it, and on Windows try the rename again for a short time. Example: `docfile.Write` (`internal/docs/docfile/write.go`) tries again for 1 second. The test `TestWriteOpenFile` checks it on the Windows CI.
