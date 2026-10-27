// Package launcher prepares the dx plugin and starts claude (ADR-001).
package launcher

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/EqualSoftware/claude-dx/internal/cache"
	"github.com/EqualSoftware/claude-dx/internal/plugin"
	"github.com/EqualSoftware/claude-dx/internal/project"
	"github.com/EqualSoftware/claude-dx/internal/session"
	"github.com/EqualSoftware/claude-dx/internal/update"
)

// installURL is shown when claude is not in PATH.
const installURL = "https://docs.claude.com/claude-code"

// ErrClaudeNotFound means that claude is not in PATH.
var ErrClaudeNotFound = fmt.Errorf("claude is not in PATH. Install Claude Code: %s", installURL)

// Run starts claude with the dx flags and args, waits for it and returns the exit code of claude.
func Run(version string, args []string) (int, error) {
	binary, err := Executable()
	if err != nil {
		return 1, err
	}
	// Windows cannot replace a running binary, so "update" leaves .old files (ADR-005).
	update.RemoveOld(binary)

	cacheDir, err := cache.Dir()
	if err != nil {
		return 1, err
	}
	// The launcher only reads update.json. The dx server checks for releases (ADR-005).
	if notice := update.Notice(cacheDir, version); notice != "" {
		fmt.Fprintln(os.Stderr, notice)
	}
	alone := func() {
		if dir, err := os.Getwd(); err == nil {
			moveData(cacheDir, dir, os.Stderr)
		}
	}
	pluginDir, end, err := session.Start(cacheDir, plugin.DirName(version, binary), alone, func() (string, error) {
		dir, err := plugin.Ensure(cacheDir, version, binary)
		if err != nil {
			return "", fmt.Errorf("cannot extract the dx plugin: %w", err)
		}
		return dir, nil
	})
	if err != nil {
		return 1, err
	}
	defer end() // It runs when claude ends.
	claude, err := exec.LookPath("claude")
	if err != nil {
		return 1, ErrClaudeNotFound
	}
	return start(claude, Args(pluginDir, args))
}

// moveData moves the data of the project of dir from the cache to the git common directory after git
// init (ADR-007), and writes a warning to w if the move fails. session.Start calls it only while no
// other claude-dx process runs. If the project is not found, it does nothing: the MCP server reports
// the same error.
func moveData(cacheDir, dir string, w io.Writer) {
	p, err := project.Open(dir)
	if err != nil {
		return
	}
	old, dx := p.DataDir(cacheDir), filepath.Join(p.Common, "dx")
	if err := p.MoveData(cacheDir); err != nil {
		fmt.Fprintf(w, "claude-dx: cannot move the project data from %s to %s: %v. The data stays in %s.\n", old, dx, err, old)
	}
}

// settings allows all dx tools with no prompt (ADR-002). Tools that change git state
// will be in an "ask" list, because "ask" wins over "allow".
const settings = `{"permissions":{"allow":["mcp__plugin_dx_mcp"]}}`

// Args returns the arguments for claude: the dx flags, then the arguments of the user.
// The rules go in a file flag, because claude uses only the last --append-system-prompt,
// and the user can give one (ADR-010).
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
