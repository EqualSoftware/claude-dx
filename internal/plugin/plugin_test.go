package plugin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func readManifest(t *testing.T, dir string) map[string]string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, ".claude-plugin", "plugin.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("plugin.json is not valid JSON: %v", err)
	}
	return m
}

func TestEnsureExtracts(t *testing.T) {
	cache := t.TempDir()
	dir, err := Ensure(cache, "v1.2.3", "/bin/claude-dx")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(cache, "plugin", DirName("v1.2.3", "/bin/claude-dx")); dir != want || !strings.HasPrefix(filepath.Base(dir), "v1.2.3-") {
		t.Fatalf("dir = %q, want %q", dir, want)
	}
	m := readManifest(t, dir)
	if m["name"] != "dx" || m["version"] != "v1.2.3" {
		t.Fatalf("manifest = %v", m)
	}
}

// TestRulesExtracted checks that the launcher can give rules.md, and plugin#R-1, R-2, R-7: 15 lines or less.
func TestRulesExtracted(t *testing.T) {
	dir, err := Ensure(t.TempDir(), "v1.2.3", "/bin/claude-dx")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "rules.md"))
	if err != nil {
		t.Fatal(err)
	}
	if lines := strings.Count(string(data), "\n"); !strings.HasPrefix(string(data), "# dx rules\n") || lines > 15 {
		t.Fatalf("rules.md has %d lines:\n%s", lines, data)
	}
}

func TestEnsureReusesRelease(t *testing.T) {
	cache := t.TempDir()
	dir, err := Ensure(cache, "v1.0.0", "/bin/claude-dx")
	if err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(dir, "marker")
	if err := os.WriteFile(marker, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	// plugin#R-9: one version and one path give one directory.
	if again, err := Ensure(cache, "v1.0.0", "/bin/claude-dx"); err != nil || again != dir {
		t.Fatalf("Ensure again = %q, %v; want %q", again, err, dir)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("a release version was extracted again")
	}
}

// TestEnsureDevKeepsRunningSession checks plugin#R-10: a dev build with other files (here, another
// binary path) extracts to a new directory, and the directory of the first session does not change.
func TestEnsureDevKeepsRunningSession(t *testing.T) {
	cache := t.TempDir()
	first, err := Ensure(cache, "dev", "/a/claude-dx")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(filepath.Base(first), "dev-") {
		t.Fatalf("dev directory = %q", first)
	}
	marker := filepath.Join(first, "marker")
	if err := os.WriteFile(marker, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	second, err := Ensure(cache, "dev", "/b/claude-dx")
	if err != nil || second == first {
		t.Fatalf("second dev directory = %q, %v", second, err)
	}
	again, err := Ensure(cache, "dev", "/a/claude-dx")
	if err != nil || again != first {
		t.Fatalf("same dev build = %q, %v; want %q", again, err, first)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("a dev start changed the files of the first session")
	}
}

// TestEnsureReleasePerBinary checks plugin#R-9: one release version and two binary paths give two
// directories, and each .mcp.json starts its own path.
func TestEnsureReleasePerBinary(t *testing.T) {
	cache := t.TempDir()
	first, err := Ensure(cache, "v1.0.0", "/a/claude-dx")
	if err != nil {
		t.Fatal(err)
	}
	second, err := Ensure(cache, "v1.0.0", "/b/claude-dx")
	if err != nil || second == first {
		t.Fatalf("second directory = %q, %v; first %q", second, err, first)
	}
	for dir, binary := range map[string]string{first: "/a/claude-dx", second: "/b/claude-dx"} {
		data, err := os.ReadFile(filepath.Join(dir, ".mcp.json"))
		if err != nil || !strings.Contains(string(data), binary) {
			t.Errorf("%s/.mcp.json = %s, %v; want %s", dir, data, err, binary)
		}
	}
}

func TestEnsureParallel(t *testing.T) {
	cache := t.TempDir()
	var wg sync.WaitGroup
	errs := make(chan error, 10)
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := Ensure(cache, "v2.0.0", "/bin/claude-dx")
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	readManifest(t, filepath.Join(cache, "plugin", DirName("v2.0.0", "/bin/claude-dx")))
	entries, _ := os.ReadDir(filepath.Join(cache, "plugin"))
	if len(entries) != 1 {
		t.Fatalf("plugin dir has %d entries, want 1 (no temp dirs left)", len(entries))
	}
}

func TestJSONEscapeWindowsPath(t *testing.T) {
	raw := `{"command":"` + jsonEscape(`C:\Users\me\claude-dx.exe`) + `"}`
	var v map[string]string
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		t.Fatalf("invalid JSON %s: %v", raw, err)
	}
	if v["command"] != `C:\Users\me\claude-dx.exe` {
		t.Fatalf("command = %q", v["command"])
	}
}

// TestHooksExtracted checks resume-command#R-3: the SessionStart hook has the 4 matchers and runs the
// absolute binary path.
func TestHooksExtracted(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "claude-dx")
	dir, err := Ensure(t.TempDir(), "v1.2.3", binary)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "hooks", "hooks.json"))
	if err != nil {
		t.Fatal(err)
	}
	var file struct {
		Hooks map[string][]struct {
			Matcher string
			Hooks   []struct{ Type, Command string }
		}
	}
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatalf("hooks.json is not valid JSON: %v", err)
	}
	want := `"` + binary + `" resume`
	var matchers []string
	for _, entry := range file.Hooks["SessionStart"] {
		matchers = append(matchers, entry.Matcher)
		if len(entry.Hooks) != 1 || entry.Hooks[0].Type != "command" || entry.Hooks[0].Command != want {
			t.Fatalf("hook of %s = %+v, want the command %q", entry.Matcher, entry.Hooks, want)
		}
	}
	if wantMatchers := []string{"startup", "resume", "clear", "compact"}; !slices.Equal(matchers, wantMatchers) {
		t.Fatalf("matchers = %v, want %v", matchers, wantMatchers)
	}
}

// TestPreToolUseHookExtracted checks git-tools#R-14: the Bash PreToolUse hook runs the absolute binary path.
func TestPreToolUseHookExtracted(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "claude-dx")
	dir, err := Ensure(t.TempDir(), "v1.2.3", binary)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "hooks", "hooks.json"))
	if err != nil {
		t.Fatal(err)
	}
	var file struct {
		Hooks map[string][]struct {
			Matcher string
			Hooks   []struct{ Type, Command string }
		}
	}
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatal(err)
	}
	entries := file.Hooks["PreToolUse"]
	want := `"` + binary + `" hook pre-tool-use`
	if len(entries) != 1 || entries[0].Matcher != "Bash" || len(entries[0].Hooks) != 1 || entries[0].Hooks[0].Type != "command" || entries[0].Hooks[0].Command != want {
		t.Fatalf("PreToolUse = %+v, want one Bash command %q", entries, want)
	}
}

// TestRulesHaveGitLine checks git-tools#R-18.
func TestRulesHaveGitLine(t *testing.T) {
	dir, err := Ensure(t.TempDir(), "v1.2.3", "/bin/claude-dx")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "rules.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "- Use the git_* tools for git. Do not run git in Bash.\n") {
		t.Fatalf("rules.md has no git line:\n%s", data)
	}
}

func TestEnsureRefusesUnsafeBinary(t *testing.T) {
	for _, binary := range []string{`/a"b/claude-dx`, "/a\rb/claude-dx", `\\\\host\\share\\claude-dx`, "/a$b/claude-dx", "/a`b/claude-dx", "/a\nb/claude-dx"} {
		cache := t.TempDir()
		if _, err := Ensure(cache, "v1.0.0", binary); err == nil {
			t.Errorf("Ensure accepted %q", binary)
		}
		if exists(filepath.Join(cache, "plugin")) {
			t.Errorf("Ensure wrote files for %q", binary)
		}
	}
}

func TestEnsureRepairsIncompleteDir(t *testing.T) {
	for _, missing := range required {
		cache := t.TempDir()
		dir, err := Ensure(cache, "v1.0.0", "/bin/claude-dx")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(filepath.Join(dir, filepath.FromSlash(missing))); err != nil {
			t.Fatal(err)
		}
		again, err := Ensure(cache, "v1.0.0", "/bin/claude-dx")
		if err != nil || again != dir {
			t.Fatalf("Ensure = %q, %v; want %q", again, err, dir)
		}
		if !exists(filepath.Join(dir, filepath.FromSlash(missing))) {
			t.Errorf("%s is not extracted again", missing)
		}
	}
}

func TestEnsureConcurrentRepair(t *testing.T) {
	cache := t.TempDir()
	dir, err := Ensure(cache, "v1.0.0", "/bin/claude-dx")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir, "rules.md")); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	var returned atomic.Int32
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, err := Ensure(cache, "v1.0.0", "/bin/claude-dx")
			if err != nil || got != dir || !complete(got) {
				t.Errorf("Ensure = %q, %v", got, err)
			}
			returned.Add(1)
			// After Ensure returned, the directory must stay complete while the other callers repair.
			for returned.Load() < 8 {
				if !complete(dir) {
					t.Error("the directory was missing after Ensure returned")
					return
				}
				runtime.Gosched()
			}
		}()
	}
	wg.Wait()
	if !complete(dir) {
		t.Fatal("the directory is not complete")
	}
}

// TestCommitSkill checks skill-commit#R-1 to R-4: the embedded commit skill has the steps and the tools.
func TestCommitSkill(t *testing.T) {
	dir, err := Ensure(t.TempDir(), "v1.2.3", "/bin/claude-dx")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "skills", "commit", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{
		"name: commit", "git_unstage", "`hunks: true`", "git_stage_hunks", "git_commit",
		"`all: true`", "Write the plan as text in your reply", "Stop and wait for the reply", "Refs: <slug>#R-N", "no trailer", "Do not run the acceptance check", "Do not push",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("commit skill lacks %q", want)
		}
	}
	if strings.Index(text, "Stop and wait for the reply") > strings.Index(text, "Run `git_commit`") {
		t.Error("the plan step is after the commit step")
	}
}

// TestSpikeText checks plugin#R-15 and R-16: the spike skill and the spike rule.
func TestSpikeText(t *testing.T) {
	dir, err := Ensure(t.TempDir(), "v1.2.3", "/bin/claude-dx")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "skills", "spike", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"name: spike", "spike_create", "spike_note", "spike_close", "`/dx:spec` makes the spec from the spike and the code"} {
		if !strings.Contains(string(data), want) {
			t.Errorf("spike skill lacks %q", want)
		}
	}
	rules, err := os.ReadFile(filepath.Join(dir, "rules.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(rules), "use a spike") || !strings.Contains(string(rules), "does not need an Approved spec") {
		t.Error("rules.md lacks the spike rule")
	}
}

// TestEurekaText checks skill-eureka#R-1 to R-4: the eureka skill text.
func TestEurekaText(t *testing.T) {
	dir, err := Ensure(t.TempDir(), "v1.2.3", "/bin/claude-dx")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "skills", "eureka", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{
		"name: eureka", "verified, reusable and not yet documented", "adr_list", "learning_list", "glossary_lookup",
		"No new lessons", "Ask the user to select", "Record nothing that the user did not select",
		"`dx:adr`", "`dx:learning`", "`dx:glossary`", "Do not commit", "/dx:commit",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("eureka skill lacks %q", want)
		}
	}
	for _, bad := range []string{"git_commit", "spec_update", "task_set_status", "spike_note"} {
		if strings.Contains(text, bad) {
			t.Errorf("eureka skill names %q", bad)
		}
	}
}

// TestADR018Text checks plugin#R-13 and R-14: the rules stay short and the skills do not name a Document History.
func TestADR018Text(t *testing.T) {
	dir, err := Ensure(t.TempDir(), "v1.2.3", "/bin/claude-dx")
	if err != nil {
		t.Fatal(err)
	}
	rules, err := os.ReadFile(filepath.Join(dir, "rules.md"))
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(string(rules), "\n"); n > 15 {
		t.Errorf("rules.md has %d lines, want 15 or less", n)
	}
	if !strings.Contains(string(rules), "Decision: X, because Y.") {
		t.Error("rules.md lacks the Decision line rule")
	}
	for _, name := range []string{"adr", "spec", "task"} {
		data, err := os.ReadFile(filepath.Join(dir, "skills", name, "SKILL.md"))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), "Document History") {
			t.Errorf("skill %s names the Document History", name)
		}
	}
}
