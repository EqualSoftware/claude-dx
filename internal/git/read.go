package git

import (
	"context"
	"fmt"
	"strings"
)

// Read tools. Each function returns the text of the tool result.

// Status returns the branch line and the changed paths in groups.
func (r Runner) Status(ctx context.Context, dir string) (string, error) {
	out, err := r.Run(ctx, dir, "status", "--porcelain=v2", "--branch")
	if err != nil {
		return "", err
	}
	return formatStatus(out), nil
}

// formatStatus groups the output of git status --porcelain=v2 --branch.
func formatStatus(out string) string {
	var head string
	groups := map[string][]string{}
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		if rest, ok := strings.CutPrefix(line, "# "); ok {
			head = branchLine(head, rest)
			continue
		}
		addChange(groups, line)
	}
	var b strings.Builder
	b.WriteString(head + "\n")
	order := []string{"Staged", "Unstaged", "Conflicted", "Untracked"}
	empty := true
	for _, name := range order {
		if len(groups[name]) == 0 {
			continue
		}
		empty = false
		fmt.Fprintf(&b, "%s:\n  %s\n", name, strings.Join(groups[name], "\n  "))
	}
	if empty {
		b.WriteString("Clean.\n")
	}
	return b.String()
}

// branchLine adds one "# branch.*" header to the branch line.
func branchLine(head, rest string) string {
	key, value, _ := strings.Cut(rest, " ")
	switch key {
	case "branch.head":
		return "## " + value + head
	case "branch.upstream":
		return head + "..." + value
	case "branch.ab":
		return head + " " + value
	}
	return head
}

func addChange(groups map[string][]string, line string) {
	kind, rest, _ := strings.Cut(line, " ")
	switch kind {
	case "?":
		groups["Untracked"] = append(groups["Untracked"], "? "+rest)
	case "1":
		f := strings.SplitN(rest, " ", 8)
		addXY(groups, f[0], f[len(f)-1])
	case "2":
		f := strings.SplitN(rest, " ", 9)
		addXY(groups, f[0], strings.Replace(f[len(f)-1], "\t", " <- ", 1))
	case "u":
		f := strings.SplitN(rest, " ", 10)
		groups["Conflicted"] = append(groups["Conflicted"], f[0]+" "+f[len(f)-1])
	}
}

func addXY(groups map[string][]string, xy, path string) {
	if len(xy) != 2 {
		return
	}
	if xy[0] != '.' {
		groups["Staged"] = append(groups["Staged"], string(xy[0])+" "+path)
	}
	if xy[1] != '.' {
		groups["Unstaged"] = append(groups["Unstaged"], string(xy[1])+" "+path)
	}
}

// DiffOptions are the options of Diff.
type DiffOptions struct {
	Staged bool
	Ref    string
	Paths  []string
	Stat   bool
}

// Diff returns the diff without color, or the file stats.
func (r Runner) Diff(ctx context.Context, dir string, o DiffOptions) (string, error) {
	if err := noDash(o.Ref); err != nil {
		return "", err
	}
	args := []string{"diff", "--no-color", "--no-ext-diff"}
	if o.Staged {
		args = append(args, "--staged")
	}
	if o.Stat {
		args = append(args, "--stat")
	}
	if o.Ref != "" {
		args = append(args, o.Ref)
	}
	args = append(args, "--")
	args = append(args, o.Paths...)
	out, err := r.Run(ctx, dir, args...)
	if err == nil && out == "" {
		return "No changes.", nil
	}
	return out, err
}

// LogOptions are the options of Log.
type LogOptions struct {
	Ref   string
	N     int
	Paths []string
}

// DefaultLog is the number of commits that Log returns when N is not set.
const DefaultLog = 20

// Log returns one line for each commit: "<short-hash> <YYYY-MM-DD> <author> <subject>".
func (r Runner) Log(ctx context.Context, dir string, o LogOptions) (string, error) {
	if err := noDash(o.Ref); err != nil {
		return "", err
	}
	if o.N <= 0 {
		o.N = DefaultLog
	}
	args := []string{"log", "--no-color", fmt.Sprintf("-n%d", o.N), "--date=short", "--format=%h %ad %an %s"}
	if o.Ref != "" {
		args = append(args, o.Ref)
	}
	args = append(args, "--")
	args = append(args, o.Paths...)
	return r.Run(ctx, dir, args...)
}

// Show returns the header line of a commit and its diff, or its file stats.
func (r Runner) Show(ctx context.Context, dir, ref string, stat bool) (string, error) {
	if err := noDash(ref); err != nil {
		return "", err
	}
	if ref == "" {
		ref = "HEAD"
	}
	args := []string{"show", "--no-color", "--no-ext-diff", "--date=short", "--format=%h %ad %an %s%n"}
	if stat {
		args = append(args, "--stat")
	}
	args = append(args, ref, "--")
	return r.Run(ctx, dir, args...)
}

// BlameOptions are the options of Blame.
type BlameOptions struct {
	File       string
	Ref        string
	Start, End int
}

// Blame returns one line for each source line: "<short-hash> <author> <line-no>: <text>".
func (r Runner) Blame(ctx context.Context, dir string, o BlameOptions) (string, error) {
	if o.File == "" {
		return "", fmt.Errorf("file is required")
	}
	if err := noDash(o.Ref); err != nil {
		return "", err
	}
	args := []string{"blame", "--line-porcelain"}
	if o.Start > 0 || o.End > 0 {
		args = append(args, "-L", lineRange(o.Start, o.End))
	}
	if o.Ref != "" {
		args = append(args, o.Ref)
	}
	args = append(args, "--", o.File)
	out, err := r.Run(ctx, dir, args...)
	if err != nil {
		return "", err
	}
	return formatBlame(out), nil
}

func lineRange(start, end int) string {
	if start <= 0 {
		start = 1
	}
	if end <= 0 {
		return fmt.Sprintf("%d,", start)
	}
	return fmt.Sprintf("%d,%d", start, end)
}

// formatBlame converts the output of git blame --line-porcelain.
func formatBlame(out string) string {
	var b strings.Builder
	var hash, author, line string
	for _, l := range strings.Split(out, "\n") {
		if text, ok := strings.CutPrefix(l, "\t"); ok {
			fmt.Fprintf(&b, "%s %s %s: %s\n", hash, author, line, text)
			hash = ""
			continue
		}
		if name, ok := strings.CutPrefix(l, "author "); ok {
			author = name
			continue
		}
		if hash != "" {
			continue
		}
		f := strings.Fields(l)
		if len(f) >= 3 {
			hash, line = shortHash(f[0]), f[2]
		}
	}
	return b.String()
}

func shortHash(h string) string {
	if len(h) > 7 {
		return h[:7]
	}
	return h
}

// noDash refuses a value that git can read as an option.
func noDash(ref string) error {
	if strings.HasPrefix(ref, "-") {
		return fmt.Errorf("ref %q must not start with a dash", ref)
	}
	return nil
}
