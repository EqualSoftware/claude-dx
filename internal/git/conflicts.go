package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Conflicts returns each conflicted path with its hunks and line numbers (git-tools#R-11). A hunk runs from
// the <<<<<<< line to the >>>>>>> line. It keeps the ||||||| base lines of the diff3 style.
func (r Runner) Conflicts(ctx context.Context, dir string) (string, error) {
	out, err := r.Run(ctx, dir, "diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, path := range strings.Fields(out) {
		data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(path)))
		if err != nil {
			fmt.Fprintf(&b, "%s: no text hunks\n", path)
			continue
		}
		hunks := hunksOf(string(data))
		if len(hunks) == 0 {
			fmt.Fprintf(&b, "%s: no text hunks\n", path)
		}
		for i, h := range hunks {
			fmt.Fprintf(&b, "%s: hunk %d lines %d-%d\n%s\n", path, i+1, h.start, h.end, h.text)
		}
	}
	if b.Len() == 0 {
		return "No conflicts.", nil
	}
	return b.String(), nil
}

type hunk struct {
	start, end int
	text       string
}

// hunksOf finds the conflict hunks of a file. Line numbers start at 1.
func hunksOf(content string) []hunk {
	var hunks []hunk
	var cur []string
	start := 0
	for i, line := range strings.Split(content, "\n") {
		line = strings.TrimRight(line, "\r")
		switch {
		case strings.HasPrefix(line, "<<<<<<<"):
			start, cur = i+1, []string{line}
		case start == 0:
		case strings.HasPrefix(line, ">>>>>>>"):
			hunks = append(hunks, hunk{start: start, end: i + 1, text: strings.Join(append(cur, line), "\n")})
			start = 0
		default:
			cur = append(cur, line)
		}
	}
	return hunks
}
