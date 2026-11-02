package report

import (
	"bytes"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func logWith(t *testing.T, entries ...Entry) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sub", LogName)
	for _, e := range entries {
		if err := Append(path, e); err != nil {
			t.Fatal(err)
		}
	}
	return path
}

func entry(sub, reason string, args ...string) Entry {
	return Entry{Time: time.Unix(0, 0).UTC(), Sub: sub, Args: append([]string{sub}, args...), Reason: reason}
}

func TestReadSkipsInvalidLines(t *testing.T) {
	path := logWith(t, entry("reflog", "why"))
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString("not json\n{}\n")
	_ = f.Close()
	got, err := Read(path)
	if err != nil || len(got) != 1 || got[0].Sub != "reflog" {
		t.Fatalf("Read = %+v, %v", got, err)
	}
	if got, err := Read(filepath.Join(t.TempDir(), "none")); err != nil || got != nil {
		t.Fatalf("Read of a missing log = %+v, %v", got, err)
	}
}

func run(t *testing.T, r Reporter, args ...string) (int, string, string) {
	t.Helper()
	var out, errOut bytes.Buffer
	code := r.Run(args, &out, &errOut)
	return code, out.String(), errOut.String()
}

func TestRunOpensForm(t *testing.T) {
	path := logWith(t,
		entry("reflog", "see old head", "-n", "5"),
		entry("reflog", "see old head", "-n", "5"),
		entry("bisect", "find the bug", "start"),
	)
	var opened string
	r := Reporter{Path: path, Version: "v1.2.3", GOOS: "darwin", Open: func(u string) error { opened = u; return nil }}
	code, out, _ := run(t, r)
	if code != 0 || !strings.Contains(out, "bisect: 1 call(s)\nreflog: 2 call(s)\nOpening the issue form:\n"+opened+"\n") {
		t.Fatalf("code %d, output %q", code, out)
	}
	u, err := url.Parse(opened)
	if err != nil {
		t.Fatal(err)
	}
	checkQuery(t, u.Query())
}

func checkQuery(t *testing.T, q url.Values) {
	t.Helper()
	for key, want := range map[string]string{"template": "unmapped-git-operation.yml", "version": "v1.2.3", "os": "macOS", "title": "git: add a tool for bisect, reflog"} {
		if q.Get(key) != want {
			t.Errorf("%s = %q, want %q", key, q.Get(key), want)
		}
	}
	ops := q.Get("operations")
	if strings.Count(ops, "git reflog -n 5 -- see old head") != 1 || !strings.Contains(ops, "reflog (2 call(s))") {
		t.Errorf("operations = %q", ops)
	}
	if q.Has("proposal") {
		t.Errorf("proposal is set: %v", q)
	}
}

func TestRunBrowserFails(t *testing.T) {
	r := Reporter{Path: logWith(t, entry("reflog", "x")), GOOS: "linux", Open: func(string) error { return errors.New("no browser") }}
	code, out, _ := run(t, r)
	if code != 0 || !strings.HasPrefix(out, "reflog: 1 call(s)\nCould not open the browser. Open this URL:\nhttps://") {
		t.Fatalf("code %d, output %q", code, out)
	}
}

func TestRunEmptyAndClear(t *testing.T) {
	r := Reporter{Path: filepath.Join(t.TempDir(), LogName), Open: func(string) error { t.Fatal("opened"); return nil }}
	if code, out, _ := run(t, r); code != 0 || out != "No unmapped git operations.\n" {
		t.Fatalf("empty: code %d, output %q", code, out)
	}
	r.Path = logWith(t, entry("a", "x"), entry("b", "y"))
	if code, out, _ := run(t, r, "--clear"); code != 0 || out != "Cleared 2 entries.\n" {
		t.Fatalf("clear: code %d, output %q", code, out)
	}
	if got, _ := Read(r.Path); len(got) != 0 {
		t.Errorf("log after clear = %+v", got)
	}
	if code, out, _ := run(t, Reporter{Path: filepath.Join(t.TempDir(), LogName)}, "--clear"); code != 0 || out != "Cleared 0 entries.\n" {
		t.Fatalf("clear of a missing log: code %d, output %q", code, out)
	}
}

func TestRunUnknownArgument(t *testing.T) {
	code, _, errOut := run(t, Reporter{Path: logWith(t)}, "--nope")
	if code != 2 || !strings.Contains(errOut, `unknown argument "--nope"`) {
		t.Fatalf("code %d, stderr %q", code, errOut)
	}
}

func TestFormCapsURL(t *testing.T) {
	var entries []Entry
	for i := 0; i < 30; i++ {
		entries = append(entries, entry("sub"+string(rune('a'+i%26))+string(rune('a'+i/26)), strings.Repeat("long reason ", 40), "arg", string(rune('a'+i))))
	}
	if got := form(group(entries), "dev", "windows"); len(got) > maxURL {
		t.Errorf("URL length = %d, want at most %d", len(got), maxURL)
	}
}

func TestLatestKeepsFive(t *testing.T) {
	var list []Entry
	for i := 0; i < 8; i++ {
		list = append(list, entry("x", "r", string(rune('a'+i))))
	}
	got := latest(list)
	if len(got) != 5 || got[0].Args[1] != "d" || got[4].Args[1] != "h" {
		t.Errorf("latest = %+v", got)
	}
}
