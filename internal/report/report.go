package report

import (
	"fmt"
	"io"
	"net/url"
	"os/exec"
	"runtime"
	"sort"
	"strings"
)

// IssueURL is the issue form for an unmapped git operation.
const IssueURL = "https://github.com/EqualSoftware/claude-dx/issues/new?template=unmapped-git-operation.yml"

const (
	maxURL      = 6000
	perSub      = 5
	trimOperand = 200
)

// Reporter runs the command claude-dx report.
type Reporter struct {
	Path    string             // the log file
	Version string             // the claude-dx version
	GOOS    string             // the operating system, as runtime.GOOS
	Open    func(string) error // opens a URL in the browser
}

// Run runs report with its arguments and returns the exit code.
func (r Reporter) Run(args []string, stdout, stderr io.Writer) int {
	if len(args) > 1 || (len(args) == 1 && args[0] != "--clear") {
		fmt.Fprintf(stderr, "claude-dx report: unknown argument %q.\n", strings.Join(args, " "))
		return 2
	}
	entries, err := Read(r.Path)
	if err != nil {
		fmt.Fprintf(stderr, "claude-dx report: %v\n", err)
		return 1
	}
	if len(args) == 1 {
		return r.clear(len(entries), stdout, stderr)
	}
	if len(entries) == 0 {
		fmt.Fprintln(stdout, "No unmapped git operations.")
		return 0
	}
	groups := group(entries)
	for _, g := range groups {
		fmt.Fprintf(stdout, "%s: %d call(s)\n", g.sub, g.count)
	}
	link := form(groups, r.Version, r.GOOS)
	if err := r.Open(link); err != nil {
		fmt.Fprintf(stdout, "Could not open the browser. Open this URL:\n%s\n", link)
		return 0
	}
	fmt.Fprintf(stdout, "Opening the issue form:\n%s\n", link)
	return 0
}

func (r Reporter) clear(n int, stdout, stderr io.Writer) int {
	if err := Clear(r.Path); err != nil {
		fmt.Fprintf(stderr, "claude-dx report: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "Cleared %d entries.\n", n)
	return 0
}

// bucket is the entries of one git subcommand.
type bucket struct {
	sub   string
	count int
	last  []Entry // the latest entries with different args, oldest first
}

// group groups the entries by subcommand, sorted by name.
func group(entries []Entry) []bucket {
	bySub := map[string][]Entry{}
	for _, e := range entries {
		bySub[e.Sub] = append(bySub[e.Sub], e)
	}
	groups := make([]bucket, 0, len(bySub))
	for sub, list := range bySub {
		groups = append(groups, bucket{sub: sub, count: len(list), last: latest(list)})
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].sub < groups[j].sub })
	return groups
}

// latest returns the last perSub entries that have different args, oldest first.
func latest(list []Entry) []Entry {
	seen := map[string]bool{}
	var out []Entry
	for i := len(list) - 1; i >= 0 && len(out) < perSub; i-- {
		key := strings.Join(list[i].Args, "\x00")
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append([]Entry{list[i]}, out...)
	}
	return out
}

// form returns the URL of the issue form with the values filled in. It keeps the URL under maxURL.
func form(groups []bucket, version, goos string) string {
	subs := make([]string, len(groups))
	var lines []string
	for i, g := range groups {
		subs[i] = g.sub
		lines = append(lines, fmt.Sprintf("%s (%d call(s))", g.sub, g.count))
		for _, e := range g.last {
			lines = append(lines, fmt.Sprintf("  git %s -- %s", strings.Join(e.Args, " "), e.Reason))
		}
	}
	build := func(ops string) string {
		v := url.Values{}
		v.Set("title", "git: add a tool for "+strings.Join(subs, ", "))
		v.Set("operations", ops)
		v.Set("version", version)
		v.Set("os", osName(goos))
		return IssueURL + "&" + v.Encode()
	}
	for len(lines) > 1 && len(build(strings.Join(lines, "\n"))) > maxURL {
		lines = lines[:len(lines)-1]
	}
	ops := strings.Join(lines, "\n")
	for len(ops) > trimOperand && len(build(ops)) > maxURL {
		ops = ops[:len(ops)-trimOperand]
	}
	return build(ops)
}

func osName(goos string) string {
	switch goos {
	case "darwin":
		return "macOS"
	case "windows":
		return "Windows"
	}
	return "Linux"
}

// OpenBrowser opens a URL in the default browser.
func OpenBrowser(link string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", link).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", link).Start()
	}
	return exec.Command("xdg-open", link).Start()
}
