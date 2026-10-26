//go:build windows

package session

import "os"

// inherit does nothing on Windows: the launcher keeps the file open until claude ends.
func inherit(*os.File) error { return nil }
