// Package hook decides if a Bash command of Claude Code runs git or a file operation on project paths
// (git-tools#R-14, R-15, filesystem-tools#R-5).
package hook

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/EqualSoftware/claude-dx/internal/project"
)

// maxDepth limits the nesting of sh -c, eval and command substitutions.
const maxDepth = 8

// DenyReason is the reason that the hook gives when a Bash command runs git.
const DenyReason = "Do not run git in Bash. Use the git_* tools of the dx MCP server (git_status, git_diff, git_commit and others). If no tool maps your operation, use git_exec."

// shells run a command string with -c.
var shells = map[string]bool{"sh": true, "bash": true, "zsh": true, "dash": true, "ksh": true}

// skipped are keywords and wrappers that run the next word as a command. The value is the set of
// options that take an argument.
var skipped = map[string]string{
	"!": "", "{": "", "}": "", "if": "", "then": "", "else": "", "elif": "", "do": "", "while": "", "until": "",
	"time": "", "command": "", "builtin": "", "exec": "", "nohup": "",
	"env": "-u -C -S", "xargs": "-I -n -P -L -d -s -E -a", "sudo": "-u -g -C -h -p", "nice": "-n",
}

// matcher reports whether one simple command (program name and arguments) is denied.
type matcher func(name string, args []string) bool

// RunsGit reports whether the Bash command runs git. A command that does not parse gives false (R-15).
func RunsGit(command string) bool {
	return runs(command, 0, func(name string, _ []string) bool { return name == "git" })
}

// fileReasons maps a file program to its deny reason (filesystem-tools#R-5).
var fileReasons = map[string]string{
	"rm":    "Do not run rm on project files in Bash. Use fs_delete of the dx MCP server.",
	"mv":    "Do not run mv on project files in Bash. Use fs_move of the dx MCP server.",
	"cp":    "Do not run cp on project files in Bash. Use fs_copy of the dx MCP server.",
	"mkdir": "Do not run mkdir on project files in Bash. Use fs_mkdir of the dx MCP server.",
}

// Deny returns the deny reason for a Bash command, or an empty string to allow it. cwd is the working
// directory of the command. Git has priority over file operations.
func Deny(command, cwd string) string {
	if RunsGit(command) {
		return DenyReason
	}
	p, err := project.Open(cwd)
	if err != nil {
		return ""
	}
	var reason string
	runs(command, 0, func(name string, args []string) bool {
		r, ok := fileReasons[name]
		if !ok || !anyInside(p, cwd, args) {
			return false
		}
		reason = r
		return true
	})
	return reason
}

// anyInside reports whether one argument is a path inside the project. Options are not paths. A path
// with a variable or a glob counts as inside.
func anyInside(p project.Project, cwd string, args []string) bool {
	for _, a := range args {
		if a == "" || strings.HasPrefix(a, "-") {
			continue
		}
		if strings.ContainsAny(a, "$*?[") {
			return true
		}
		abs := a
		if !filepath.IsAbs(abs) {
			abs = filepath.Join(cwd, abs)
		}
		if _, err := p.Resolve(abs); err == nil {
			return true
		}
	}
	return false
}

func runs(command string, depth int, match matcher) bool {
	if depth > maxDepth {
		return false
	}
	cmds, subs, ok := split(command)
	if !ok {
		return false
	}
	for _, sub := range subs {
		if runs(sub, depth+1, match) {
			return true
		}
	}
	for _, words := range cmds {
		if commandRuns(words, depth, match) {
			return true
		}
	}
	return false
}

// commandRuns checks one simple command.
func commandRuns(words []string, depth int, match matcher) bool {
	words = skipPrefix(words)
	if len(words) == 0 {
		return false
	}
	name := baseName(words[0])
	switch {
	case match(name, words[1:]):
		return true
	case name == "eval":
		return runs(strings.Join(words[1:], " "), depth+1, match)
	case shells[name]:
		return shellRuns(words[1:], depth, match)
	}
	return false
}

// skipPrefix removes assignments, keywords and wrappers (with their options) before the command.
func skipPrefix(words []string) []string {
	for len(words) > 0 {
		w := words[0]
		if isAssignment(w) {
			words = words[1:]
			continue
		}
		withArg, ok := skipped[baseName(w)]
		if !ok {
			return words
		}
		words = skipOptions(words[1:], withArg)
	}
	return words
}

// skipOptions removes the options of a wrapper, and the assignments of env.
func skipOptions(words []string, withArg string) []string {
	for len(words) > 0 && (strings.HasPrefix(words[0], "-") || isAssignment(words[0])) {
		takesArg := strings.Contains(" "+withArg+" ", " "+words[0]+" ")
		words = words[1:]
		if takesArg && len(words) > 0 {
			words = words[1:]
		}
	}
	return words
}

// shellRuns checks the string after the -c option of a shell.
func shellRuns(args []string, depth int, match matcher) bool {
	for i, a := range args {
		if !strings.HasPrefix(a, "-") {
			return false
		}
		if !strings.HasPrefix(a, "--") && strings.Contains(a, "c") {
			return i+1 < len(args) && runs(args[i+1], depth+1, match)
		}
	}
	return false
}

func isAssignment(w string) bool {
	i := strings.IndexByte(w, '=')
	if i <= 0 {
		return false
	}
	for _, r := range w[:i] {
		if r != '_' && (r < '0' || r > '9') && (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
			return false
		}
	}
	return true
}

// baseName returns the lowercase name of a program, without its directory and .exe suffix.
func baseName(w string) string {
	name := path.Base(strings.ReplaceAll(w, `\`, "/"))
	return strings.TrimSuffix(strings.ToLower(name), ".exe")
}
