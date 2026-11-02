package git

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Hunk tools (git-tools#R-23, git-tools#R-24, ADR-017). The numbers and the token come from the unstaged
// diff of the working tree.

// HunkSelection is the hunks of one file to stage. The numbers start at 1.
type HunkSelection struct {
	Path    string `json:"path" jsonschema:"the file path as git_diff shows it"`
	Numbers []int  `json:"numbers" jsonschema:"the hunk numbers of the file, from git_diff with hunks: true"`
}

type fileDiff struct {
	path   string
	header string
	hunks  []string
}

// unstagedDiff returns the diff that the hunk numbers and the token come from.
func (r Runner) unstagedDiff(ctx context.Context, dir string) (string, error) {
	return r.Run(ctx, dir, "diff", "--no-color", "--no-ext-diff", "--no-renames")
}

// diffToken returns the first 12 hex characters of the sha256 of the diff.
func diffToken(diff string) string {
	sum := sha256.Sum256([]byte(diff))
	return hex.EncodeToString(sum[:])[:12]
}

// hunkDiff returns the unstaged diff with a token line and a number before each hunk (git-tools#R-23).
func (r Runner) hunkDiff(ctx context.Context, dir string, o DiffOptions) (string, error) {
	if o.Staged || o.Ref != "" || o.Stat {
		return "", errors.New("git_diff refused: hunks works only on the unstaged working tree diff. Do not use staged, ref or stat with it")
	}
	diff, err := r.unstagedDiff(ctx, dir)
	if err != nil {
		return "", err
	}
	files := parseDiff(diff)
	if len(files) == 0 {
		return "No changes.", nil
	}
	var b strings.Builder
	b.WriteString("Token: " + diffToken(diff) + "\n\n")
	for _, f := range files {
		if len(o.Paths) > 0 && !hasPath(o.Paths, f.path) {
			continue
		}
		b.WriteString(f.header)
		for i, h := range f.hunks {
			fmt.Fprintf(&b, "[%d] %s", i+1, h)
		}
	}
	return b.String(), nil
}

func hasPath(paths []string, p string) bool {
	for _, x := range paths {
		if filepath.ToSlash(filepath.Clean(x)) == p {
			return true
		}
	}
	return false
}

// parseDiff splits a diff into files and hunks. Each part keeps its line ends.
func parseDiff(diff string) []fileDiff {
	var files []fileDiff
	for _, line := range strings.SplitAfter(diff, "\n") {
		switch {
		case strings.HasPrefix(line, "diff --git "):
			files = append(files, fileDiff{path: diffPath(line), header: line})
		case len(files) == 0 || line == "":
			continue
		case strings.HasPrefix(line, "@@ "):
			f := &files[len(files)-1]
			f.hunks = append(f.hunks, line)
		default:
			f := &files[len(files)-1]
			if len(f.hunks) == 0 {
				f.header += line
				continue
			}
			f.hunks[len(f.hunks)-1] += line
		}
	}
	return files
}

// diffPath returns the path of a "diff --git a/<p> b/<p>" line. The diff has no renames, so both are the same.
func diffPath(line string) string {
	rest := strings.TrimRight(strings.TrimPrefix(line, "diff --git a/"), "\r\n")
	return rest[:max(0, (len(rest)-3)/2)]
}

// StageHunks stages only the chosen hunks of the unstaged diff (git-tools#R-24).
func (r Runner) StageHunks(ctx context.Context, dir, token string, sel []HunkSelection) (string, error) {
	if token == "" || len(sel) == 0 {
		return "", errors.New("git_stage_hunks refused: give the token and the hunks from git_diff with hunks: true")
	}
	diff, err := r.unstagedDiff(ctx, dir)
	if err != nil {
		return "", err
	}
	if got := diffToken(diff); got != token {
		return "", fmt.Errorf("git_stage_hunks refused: the token %q is not the token of the current diff (%s). The diff changed. Call git_diff with hunks: true again", token, got)
	}
	patch, count, err := buildPatch(parseDiff(diff), sel)
	if err != nil {
		return "", err
	}
	if err := r.applyCached(ctx, dir, patch); err != nil {
		return "", err
	}
	return fmt.Sprintf("Staged %d hunks in %d files.", count, len(sel)), nil
}

// buildPatch joins the header and the chosen hunks of each file. It returns the patch and the hunk count.
func buildPatch(files []fileDiff, sel []HunkSelection) (string, int, error) {
	var b strings.Builder
	count := 0
	for _, s := range sel {
		f, err := findFile(files, s.Path)
		if err != nil {
			return "", 0, err
		}
		nums, err := checkNumbers(f, s.Numbers)
		if err != nil {
			return "", 0, err
		}
		b.WriteString(f.header)
		for _, n := range nums {
			b.WriteString(f.hunks[n-1])
		}
		count += len(nums)
	}
	return b.String(), count, nil
}

func findFile(files []fileDiff, path string) (fileDiff, error) {
	want := filepath.ToSlash(filepath.Clean(path))
	for _, f := range files {
		if f.path == want {
			return f, nil
		}
	}
	return fileDiff{}, fmt.Errorf("git_stage_hunks refused: %s has no unstaged changes. For an untracked file use git_stage", path)
}

// checkNumbers returns the numbers sorted and without duplicates.
func checkNumbers(f fileDiff, nums []int) ([]int, error) {
	if len(nums) == 0 {
		return nil, fmt.Errorf("git_stage_hunks refused: give hunk numbers for %s", f.path)
	}
	for _, n := range nums {
		if n < 1 || n > len(f.hunks) {
			return nil, fmt.Errorf("git_stage_hunks refused: hunk %d does not exist in %s (it has %d hunks)", n, f.path, len(f.hunks))
		}
	}
	out := append([]int(nil), nums...)
	sort.Ints(out)
	uniq := out[:1]
	for _, n := range out[1:] {
		if n != uniq[len(uniq)-1] {
			uniq = append(uniq, n)
		}
	}
	return uniq, nil
}

// applyCached applies a patch to the index. The patch goes in a temporary file, as the runner has no stdin.
func (r Runner) applyCached(ctx context.Context, dir, patch string) error {
	f, err := os.CreateTemp("", "claude-dx-hunks-*.patch")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(f.Name()) }()
	_, werr := f.WriteString(patch)
	if cerr := f.Close(); werr == nil {
		werr = cerr
	}
	if werr != nil {
		return werr
	}
	_, err = r.Run(ctx, dir, "apply", "--cached", "--recount", f.Name())
	return err
}
