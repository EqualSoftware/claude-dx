//go:build windows

package session

import "os"

// inherit does nothing on Windows: the launcher keeps the file open until claude ends.
func inherit(*os.File) error { return nil }

// alive is always true on Windows: no child of claude gets the session handle, so a held session
// file is of a running session.
func alive(string) bool { return true }
