// Package hook decides if a Bash command of Claude Code runs git (git-tools#R-14, R-15).
package hook

import (
	"path"
	"strings"
)

// maxDepth limits the nesting of sh -c, eval and command substitutions.
const maxDepth = 8

// DenyReason is the reason that the hook gives when a Bash command runs git.
const DenyReason = "Do not run git in Bash. Use the git_* tools of the dx MCP server (git_status, git_diff, git_commit and others)."

// shells run a command string with -c.
var shells = map[string]bool{"sh": true, "bash": true, "zsh": true, "dash": true, "ksh": true}

// skipped are keywords and wrappers that run the next word as a command. The value is the set of
// options that take an argument.
var skipped = map[string]string{
	"!": "", "{": "", "}": "", "if": "", "then": "", "else": "", "elif": "", "do": "", "while": "", "until": "",
	"time": "", "command": "", "builtin": "", "exec": "", "nohup": "",
	"env": "-u -C -S", "xargs": "-I -n -P -L -d -s -E -a", "sudo": "-u -g -C -h -p", "nice": "-n",
}

// RunsGit reports whether the Bash command runs git. A command that does not parse gives false (R-15).
func RunsGit(command string) bool {
	return runsGit(command, 0)
}

func runsGit(command string, depth int) bool {
	if depth > maxDepth {
		return false
	}
	cmds, subs, ok := split(command)
	if !ok {
		return false
	}
	for _, sub := range subs {
		if runsGit(sub, depth+1) {
			return true
		}
	}
	for _, words := range cmds {
		if commandRunsGit(words, depth) {
			return true
		}
	}
	return false
}

// commandRunsGit checks one simple command.
func commandRunsGit(words []string, depth int) bool {
	words = skipPrefix(words)
	if len(words) == 0 {
		return false
	}
	name := baseName(words[0])
	switch {
	case name == "git":
		return true
	case name == "eval":
		return runsGit(strings.Join(words[1:], " "), depth+1)
	case shells[name]:
		return shellRunsGit(words[1:], depth)
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

// shellRunsGit checks the string after the -c option of a shell.
func shellRunsGit(args []string, depth int) bool {
	for i, a := range args {
		if !strings.HasPrefix(a, "-") {
			return false
		}
		if !strings.HasPrefix(a, "--") && strings.Contains(a, "c") {
			return i+1 < len(args) && runsGit(args[i+1], depth+1)
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
