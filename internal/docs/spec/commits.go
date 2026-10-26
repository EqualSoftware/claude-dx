package spec

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/EqualSoftware/claude-dx/internal/markdown"
)

var requirementID = regexp.MustCompile(`^R-\d+$`)

// commit is one commit of HEAD with a Refs trailer for the spec.
type commit struct {
	hash    string
	subject string
	refs    []string // requirement IDs of the spec
}

// commits returns the commits of HEAD with a trailer "Refs: <slug>#R-N", newest first (ADR-016).
// A repository with no commit has no commits.
func (s Store) commits(slug string) ([]commit, error) {
	err := exec.Command("git", "-C", s.Root, "rev-parse", "--verify", "-q", "HEAD").Run()
	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() == 1 {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}
	out, err := exec.Command("git", "-C", s.Root, "log", "HEAD",
		"--format=%h%x1f%s%x1f%(trailers:key=Refs,valueonly,separator=%x2C)%x1e").Output()
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", gitError(err))
	}
	var list []commit
	for record := range strings.SplitSeq(string(out), "\x1e") {
		fields := strings.Split(strings.TrimLeft(record, "\n"), "\x1f")
		if len(fields) != 3 {
			continue
		}
		if refs := specRefs(slug, fields[2]); len(refs) > 0 {
			list = append(list, commit{hash: fields[0], subject: fields[1], refs: refs})
		}
	}
	return list, nil
}

// specRefs returns the requirement IDs of slug in the comma list of Refs values. It ignores other values.
func specRefs(slug, values string) []string {
	var refs []string
	for value := range strings.SplitSeq(values, ",") {
		rid, ok := strings.CutPrefix(strings.TrimSpace(value), slug+"#")
		if ok && requirementID.MatchString(rid) {
			refs = append(refs, rid)
		}
	}
	return refs
}

// gitError adds the standard error output of git to err.
func gitError(err error) error {
	var exit *exec.ExitError
	if errors.As(err, &exit) && len(exit.Stderr) > 0 {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(exit.Stderr)))
	}
	return err
}

// withCommits adds a Commits column to the Requirements table in text, and a list of the commits after it.
// text is the full spec document or the Requirements section.
func withCommits(text string, full bool, list []commit) string {
	if !full {
		return commitTable(text, list)
	}
	front, body := markdown.Split(text)
	section, ok := markdown.Section(body, "5")
	if !ok {
		return text
	}
	body, err := markdown.ReplaceSection(body, "5", commitTable(section, list))
	if err != nil {
		return text
	}
	return markdown.Join(front, body)
}

// commitTable adds the Commits column to the table rows of section, and the commit list after the table.
func commitTable(section string, list []commit) string {
	lines := strings.Split(section, "\n")
	row := 0
	for i, line := range lines {
		if !strings.HasPrefix(line, "|") {
			continue
		}
		switch row++; row {
		case 1:
			lines[i] += " Commits |"
		case 2:
			lines[i] += " :--- |"
		default:
			lines[i] += " " + hashes(line, list) + " |"
		}
	}
	text := strings.Join(lines, "\n")
	if len(list) == 0 {
		return text
	}
	text += "\n\nCommits:"
	for _, c := range list {
		text += "\n- " + c.hash + " " + c.subject
	}
	return text
}

// hashes returns the short hashes of the commits of the requirement in row, or "—".
func hashes(row string, list []commit) string {
	m := requirementRow.FindStringSubmatch(row)
	if m == nil {
		return "—"
	}
	var out []string
	for _, c := range list {
		for _, rid := range c.refs {
			if rid == m[1] {
				out = append(out, c.hash)
				break
			}
		}
	}
	if len(out) == 0 {
		return "—"
	}
	return strings.Join(out, ", ")
}
