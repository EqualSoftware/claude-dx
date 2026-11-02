package resume

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"syscall"

	"github.com/EqualSoftware/claude-dx/internal/docs/spike"
	"github.com/EqualSoftware/claude-dx/internal/markdown"
	"github.com/EqualSoftware/claude-dx/internal/project"
)

// notePattern matches the date of a spike note: "- YYYY-MM-DD: text".
var notePattern = regexp.MustCompile(`^- (\d{4}-\d{2}-\d{2}): `)

// spikeInfo is an Open spike that the command read.
type spikeInfo struct {
	spike.Spike
	last string // the date of the last note, or the created date
}

// readSpikes reads the Open spikes of the project that contains dir. As readTasks, it takes no lock,
// writes nothing and creates nothing (R-11). A missing data directory or spikes folder means no spikes.
// A project or a cache directory that readTasks already reported gives no spikes and no second error.
func readSpikes(dir, cacheDir string) (spikes []spikeInfo, errs []error) {
	if cacheDir == "" {
		return nil, nil
	}
	p, err := project.Open(dir)
	if err != nil {
		return nil, nil
	}
	spikesDir := filepath.Join(p.DataDir(cacheDir), project.Spikes)
	entries, err := readSpikeDir(spikesDir)
	if err != nil {
		return nil, []error{skip("spikes", err)}
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		sp, err := readSpike(filepath.Join(spikesDir, e.Name()))
		if err != nil {
			errs = append(errs, skip("spikes", fmt.Errorf("%s: %w", e.Name(), err)))
			continue
		}
		if sp.Status == spike.Open {
			spikes = append(spikes, sp)
		}
	}
	return spikes, errs
}

// readSpikeDir reads the spikes folder. A missing folder gives no entries and no error.
// On Windows, ReadDir of a file gives a "not exist" error, so Stat tells a missing folder from a file.
func readSpikeDir(dir string) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(dir)
	if !errors.Is(err, os.ErrNotExist) {
		return entries, err
	}
	if _, statErr := os.Stat(dir); statErr == nil {
		return nil, &os.PathError{Op: "open", Path: dir, Err: syscall.ENOTDIR}
	}
	return nil, nil
}

// readSpike reads one spike file.
func readSpike(path string) (spikeInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return spikeInfo{}, err
	}
	front, body := markdown.Split(string(data))
	var s spikeInfo
	if err := markdown.Decode(front, &s.Spike); err != nil {
		return spikeInfo{}, fmt.Errorf("bad frontmatter: %w", err)
	}
	s.last = lastNote(body, s.Created)
	return s, nil
}

// lastNote returns the date of the last note of the Notes section, or created when there is none.
func lastNote(body, created string) string {
	notes, _ := markdown.Section(body, "Notes")
	last := created
	for _, line := range splitLines(notes) {
		if m := notePattern.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
			last = m[1]
		}
	}
	return last
}

// newestSpikes orders the spikes by created date, newest first, and then by slug (R-13).
func newestSpikes(spikes []spikeInfo) []spikeInfo {
	list := slices.Clone(spikes)
	slices.SortFunc(list, func(a, b spikeInfo) int {
		return cmp.Or(cmp.Compare(b.Created, a.Created), cmp.Compare(a.Slug, b.Slug))
	})
	return list
}

// spikeLines writes one line for each spike: slug, title and date.
func spikeLines(spikes []spikeInfo) []string {
	var lines []string
	for _, s := range spikes {
		lines = append(lines, fmt.Sprintf("- %s %s (%s)", s.Slug, s.Title, s.last))
	}
	return lines
}
