// Package launcher prepares the dx plugin and starts claude (ADR-001, ADR-003).
package launcher

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/is-equal/claude-dx/internal/cache"
	"github.com/is-equal/claude-dx/internal/plugin"
)

// installURL is shown when claude is not in PATH.
const installURL = "https://docs.claude.com/claude-code"

// ErrClaudeNotFound means that claude is not in PATH.
var ErrClaudeNotFound = fmt.Errorf("claude is not in PATH. Install Claude Code: %s", installURL)

// Run starts claude with the dx flags and args, and returns the exit code of claude.
// On macOS and Linux, Run does not return when claude starts: claude replaces this process.
func Run(version string, args []string) (int, error) {
	binary, err := executable()
	if err != nil {
		return 1, err
	}
	// Windows cannot replace a running binary, so "update" leaves a .old file (ADR-004).
	_ = os.Remove(binary + ".old")

	cacheDir, err := cache.Dir()
	if err != nil {
		return 1, err
	}
	pluginDir, err := plugin.Ensure(cacheDir, version, binary)
	if err != nil {
		return 1, fmt.Errorf("cannot extract the dx plugin: %w", err)
	}
	claude, err := exec.LookPath("claude")
	if err != nil {
		return 1, ErrClaudeNotFound
	}
	return start(claude, Args(pluginDir, args))
}

// Args returns the arguments for claude: the dx flags, then the arguments of the user.
func Args(pluginDir string, user []string) []string {
	return append([]string{"--plugin-dir", pluginDir}, user...)
}

// executable returns the absolute path of claude-dx, with symlinks resolved.
func executable() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(path)
}
