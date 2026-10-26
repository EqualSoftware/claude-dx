// Package launcher prepares the dx plugin and starts claude (ADR-001, ADR-003).
package launcher

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/EqualSoftware/claude-dx/internal/cache"
	"github.com/EqualSoftware/claude-dx/internal/plugin"
	"github.com/EqualSoftware/claude-dx/internal/session"
	"github.com/EqualSoftware/claude-dx/internal/update"
)

// installURL is shown when claude is not in PATH.
const installURL = "https://docs.claude.com/claude-code"

// ErrClaudeNotFound means that claude is not in PATH.
var ErrClaudeNotFound = fmt.Errorf("claude is not in PATH. Install Claude Code: %s", installURL)

// Run starts claude with the dx flags and args, and returns the exit code of claude.
// On macOS and Linux, Run does not return when claude starts: claude replaces this process.
func Run(version string, args []string) (int, error) {
	binary, err := Executable()
	if err != nil {
		return 1, err
	}
	// Windows cannot replace a running binary, so "update" leaves a .old file (ADR-004).
	_ = os.Remove(binary + ".old")

	cacheDir, err := cache.Dir()
	if err != nil {
		return 1, err
	}
	// The launcher only reads update.json. The dx server checks for releases (ADR-004).
	if notice := update.Notice(cacheDir, version); notice != "" {
		fmt.Fprintln(os.Stderr, notice)
	}
	pluginDir, end, err := session.Start(cacheDir, func() (string, error) {
		dir, err := plugin.Ensure(cacheDir, version, binary)
		if err != nil {
			return "", fmt.Errorf("cannot extract the dx plugin: %w", err)
		}
		return dir, nil
	})
	if err != nil {
		return 1, err
	}
	// On Unix, end runs only if exec fails. On Windows, it runs when claude ends.
	defer end()
	claude, err := exec.LookPath("claude")
	if err != nil {
		return 1, ErrClaudeNotFound
	}
	return start(claude, Args(pluginDir, args))
}

// settings allows all dx tools with no prompt (ADR-005). Tools that change git state
// will be in an "ask" list, because "ask" wins over "allow".
const settings = `{"permissions":{"allow":["mcp__plugin_dx_mcp"]}}`

// Args returns the arguments for claude: the dx flags, then the arguments of the user.
// The rules go in a file flag, because claude uses only the last --append-system-prompt,
// and the user can give one (ADR-009).
func Args(pluginDir string, user []string) []string {
	dx := []string{
		"--plugin-dir", pluginDir,
		"--settings", settings,
		"--append-system-prompt-file", filepath.Join(pluginDir, "rules.md"),
	}
	return append(dx, user...)
}

// Executable returns the absolute path of claude-dx, with symlinks resolved.
func Executable() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(path)
}
