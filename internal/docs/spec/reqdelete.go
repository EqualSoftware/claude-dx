package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/EqualSoftware/claude-dx/internal/docs/docfile"
	"github.com/EqualSoftware/claude-dx/internal/markdown"
	"github.com/EqualSoftware/claude-dx/internal/project"
)

// usedID is a history note that shows an ID: "Added R-N." or "Deleted R-N.".
var usedID = regexp.MustCompile(`(?:Added|Deleted) R-(\d+)\.`)

// historyMax returns the highest requirement number that the Document History shows, or 0.
func historyMax(body string) int {
	history, _ := markdown.Section(body, "Document History")
	highest := 0
	for _, m := range usedID.FindAllStringSubmatch(history, -1) {
		n, _ := strconv.Atoi(m[1]) // The pattern allows only digits.
		highest = max(highest, n)
	}
	return highest
}

// DeleteRequirement deletes one Planned requirement that has no commit and no task. It adds the note
// "Deleted R-N." to the Document History. review is true if the spec was Approved and is now In Review.
// A refusal names the reasons.
func (s Store) DeleteRequirement(slug, rid string) (review bool, err error) {
	rid = strings.ToUpper(strings.TrimSpace(rid))
	list, err := s.commits(slug) // git data does not change with the docs, so it is read before the lock.
	if err != nil {
		return false, err
	}
	err = s.edit(slug, func(d *doc) error {
		if err := s.checkRequirementDelete(d, rid, list); err != nil {
			return err
		}
		if err := removeRow(d, rid); err != nil {
			return err
		}
		if err := s.addHistory(d, "Deleted "+rid+"."); err != nil {
			return err
		}
		review, err = s.toReview(d, scopeNote("5"))
		return err
	})
	return review, err
}

// checkRequirementDelete returns an error with all the reasons why rid cannot be deleted.
func (s Store) checkRequirementDelete(d *doc, rid string, list []commit) error {
	status := requirementStatus(d, rid)
	if status == "" {
		return fmt.Errorf("requirement %s does not exist", rid)
	}
	var reasons []string
	if status != "Planned" {
		reasons = append(reasons, "its status is "+status+". Only a Planned requirement is deleted")
	}
	if hashes := commitsOf(rid, list); len(hashes) > 0 {
		reasons = append(reasons, "commits refer to it: "+strings.Join(hashes, ", "))
	}
	tasks, err := s.tasksCovering(d.slug, rid)
	if err != nil {
		return err
	}
	if len(tasks) > 0 {
		reasons = append(reasons, "tasks cover it: "+strings.Join(tasks, ", ")+". Change their covers first")
	}
	if len(reasons) > 0 {
		return fmt.Errorf("cannot delete %s: %s", rid, strings.Join(reasons, "; "))
	}
	return nil
}

// commitsOf returns the short hashes of the commits that refer to rid.
func commitsOf(rid string, list []commit) []string {
	var out []string
	for _, c := range list {
		if slices.Contains(c.refs, rid) {
			out = append(out, c.hash)
		}
	}
	return out
}

// tasksCovering returns the IDs of the local tasks of slug that cover rid. Like checkTasks, it reads the
// task folder of the docs lock, because this package does not import the task package. The caller holds
// the docs lock.
func (s Store) tasksCovering(slug, rid string) ([]string, error) {
	dir := filepath.Join(filepath.Dir(s.Lock), project.Tasks, slug)
	entries, err := os.ReadDir(dir)
	if docfile.NotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		var f struct {
			Covers []string `yaml:"covers"`
		}
		front, _ := markdown.Split(string(data))
		if err := markdown.Decode(front, &f); err != nil {
			return nil, fmt.Errorf("bad frontmatter in task %s/%s: %w", slug, e.Name(), err)
		}
		if slices.ContainsFunc(f.Covers, func(c string) bool { return strings.EqualFold(c, rid) }) {
			ids = append(ids, slug+"/"+strings.TrimSuffix(e.Name(), ".md"))
		}
	}
	return ids, nil
}

// removeRow removes the row of rid from section 5. An empty table gets the placeholder row again.
func removeRow(d *doc, rid string) error {
	section, _ := markdown.Section(d.body, "5")
	var lines []string
	for line := range strings.SplitSeq(section, "\n") {
		if m := requirementRow.FindStringSubmatch(line); m == nil || m[1] != rid {
			lines = append(lines, line)
		}
	}
	if !slices.ContainsFunc(lines, requirementRow.MatchString) {
		lines = insertRow(lines, noRequirements)
	}
	return replaceTable(d, "5", lines)
}
