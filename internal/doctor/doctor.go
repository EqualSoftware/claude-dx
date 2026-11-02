// Package doctor checks the setup of claude-dx and gives a fix for each problem.
package doctor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/mod/semver"
)

// MinClaude is the oldest Claude Code version on which the launch flags are verified (cli-launch#R-1 to R-3).
const MinClaude = "2.1.308"

// MinGit is the oldest git version that the git tools support (git-tools#R-3).
const MinGit = "2.30"

// Run writes one line for each check and reports whether all checks pass.
// pluginName is the name of the plugin directory in <cacheDir>/plugin (plugin.DirName).
func Run(w io.Writer, pluginName, cacheDir string) bool {
	checks := []func() (bool, string){
		claude,
		git,
		func() (bool, string) { return writable(cacheDir) },
		func() (bool, string) { return plugin(cacheDir, pluginName) },
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

func git() (bool, string) {
	path, err := exec.LookPath("git")
	if err != nil {
		return false, "git is not in PATH. Fix: install git " + MinGit + " or later: https://git-scm.com/downloads."
	}
	out, err := exec.Command(path, "--version").Output()
	if err != nil {
		return false, fmt.Sprintf("git --version failed: %v. Fix: install git again.", err)
	}
	return gitVersion(string(out))
}

// gitVersion checks the output of git --version, for example "git version 2.39.3 (Apple Git-146)".
func gitVersion(out string) (bool, string) {
	fields := strings.Fields(out)
	if len(fields) < 3 {
		return false, fmt.Sprintf("cannot read the git version from %q. Fix: install git again.", strings.TrimSpace(out))
	}
	major, minor, ok := majorMinor(fields[2])
	if !ok {
		return false, fmt.Sprintf("cannot read the git version from %q. Fix: install git again.", strings.TrimSpace(out))
	}
	minMajor, minMinor, _ := majorMinor(MinGit)
	if major < minMajor || (major == minMajor && minor < minMinor) {
		return false, fmt.Sprintf("git %s is older than %s. Fix: update git to %s or later.", fields[2], MinGit, MinGit)
	}
	return true, fmt.Sprintf("git %s in PATH", fields[2])
}

func majorMinor(v string) (major, minor int, ok bool) {
	parts := strings.Split(v, ".")
	if len(parts) < 2 {
		return 0, 0, false
	}
	major, err1 := strconv.Atoi(parts[0])
	minor, err2 := strconv.Atoi(parts[1])
	return major, minor, err1 == nil && err2 == nil
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

func plugin(cacheDir, name string) (bool, string) {
	if _, err := os.Stat(filepath.Join(cacheDir, "plugin", name, ".claude-plugin", "plugin.json")); err != nil {
		return false, fmt.Sprintf("plugin %s is not extracted. Fix: run claude-dx once.", name)
	}
	return true, fmt.Sprintf("plugin %s is extracted", name)
}
