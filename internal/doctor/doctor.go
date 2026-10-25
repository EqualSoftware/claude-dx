// Package doctor checks the setup of claude-dx and gives a fix for each problem.
package doctor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/mod/semver"
)

// MinClaude is the oldest Claude Code version on which the launch flags are verified (TASK-002).
const MinClaude = "2.1.308"

// Run writes one line for each check and reports whether all checks pass.
func Run(w io.Writer, version, cacheDir string) bool {
	checks := []func() (bool, string){
		claude,
		func() (bool, string) { return writable(cacheDir) },
		func() (bool, string) { return plugin(cacheDir, version) },
	}
	all := true
	for _, check := range checks {
		ok, text := check()
		status := "OK  "
		if !ok {
			status, all = "FAIL", false
		}
		fmt.Fprintf(w, "%s  %s\n", status, text)
	}
	return all
}

func claude() (bool, string) {
	path, err := exec.LookPath("claude")
	if err != nil {
		return false, "claude is not in PATH. Fix: install Claude Code: https://docs.claude.com/claude-code."
	}
	out, err := exec.Command(path, "--version").Output()
	if err != nil {
		return false, fmt.Sprintf("claude --version failed: %v. Fix: install Claude Code again.", err)
	}
	return claudeVersion(string(out))
}

// claudeVersion checks the output of claude --version, for example "2.1.308 (Claude Code)".
func claudeVersion(out string) (bool, string) {
	fields := strings.Fields(out)
	if len(fields) == 0 || !semver.IsValid("v"+fields[0]) {
		return false, fmt.Sprintf("cannot read the claude version from %q. Fix: install Claude Code again.", strings.TrimSpace(out))
	}
	v := fields[0]
	if semver.Compare("v"+v, "v"+MinClaude) < 0 {
		return false, fmt.Sprintf("claude %s is older than %s. Fix: run claude update.", v, MinClaude)
	}
	return true, fmt.Sprintf("claude %s in PATH", v)
}

func writable(dir string) (bool, string) {
	err := os.MkdirAll(dir, 0o755)
	if err == nil {
		var f *os.File
		if f, err = os.CreateTemp(dir, "doctor-*"); err == nil {
			_ = f.Close()
			err = os.Remove(f.Name())
		}
	}
	if err != nil {
		return false, fmt.Sprintf("cache %s is not writable: %v. Fix: give your user write permission to it.", dir, err)
	}
	return true, fmt.Sprintf("cache %s is writable", dir)
}

func plugin(cacheDir, version string) (bool, string) {
	if _, err := os.Stat(filepath.Join(cacheDir, "plugin", version, ".claude-plugin", "plugin.json")); err != nil {
		return false, fmt.Sprintf("plugin %s is not extracted. Fix: run claude-dx once.", version)
	}
	return true, fmt.Sprintf("plugin %s is extracted", version)
}
