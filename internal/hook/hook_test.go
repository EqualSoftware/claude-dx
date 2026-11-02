package hook

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunsGit(t *testing.T) {
	tests := []struct {
		command string
		want    bool
	}{
		{`git status`, true},
		{`cd sub && bash -c "git push"`, true},
		{`echo a; git log`, true},
		{`false || git diff`, true},
		{`cat f | git apply`, true},
		{`(cd x; git status)`, true},
		{`echo $(git rev-parse HEAD)`, true},
		{`echo "$(git rev-parse HEAD)"`, true},
		{"echo `git rev-parse HEAD`", true},
		{`sh -c 'echo hi && git add .'`, true},
		{`bash -lc "git push"`, true},
		{`eval "git status"`, true},
		{`eval git status`, true},
		{`env FOO=1 git status`, true},
		{`env -u X git status`, true},
		{`ls | xargs -n 1 git add`, true},
		{`xargs -I {} git show {}`, true},
		{`FOO=1 git status`, true},
		{`/usr/bin/git status`, true},
		{`sudo git status`, true},
		{`"git" status`, true},
		{`C:/Git/cmd/git.exe status`, true},
		{`bash -c "sh -c 'git push'"`, true},
		{`if true; then git status; fi`, true},
		{`gh pr list`, false},
		{`gh pr create --title "git fix"`, false},
		{`echo git`, false},
		{`echo 'git status'`, false},
		{`echo "(git)"`, false},
		{`grep -r git .`, false},
		{`ls | xargs echo git`, false},
		{`cat .gitignore`, false},
		{`bash -c 'echo hi'`, false},
		{`go test ./...`, false},
		{``, false},
		{`echo "git status`, false},
		{`bash -c "git push`, false},
	}
	for _, tt := range tests {
		if got := RunsGit(tt.command); got != tt.want {
			t.Errorf("RunsGit(%q) = %v, want %v", tt.command, got, tt.want)
		}
	}
}

func TestPreToolUse(t *testing.T) {
	tests := []struct {
		name, in string
		deny     bool
	}{
		{"git in Bash", `{"tool_name":"Bash","tool_input":{"command":"cd sub && bash -c \"git push\""}}`, true},
		{"gh in Bash", `{"tool_name":"Bash","tool_input":{"command":"gh pr list"}}`, false},
		{"other tool", `{"tool_name":"Edit","tool_input":{"command":"git status"}}`, false},
		{"bad JSON", `{`, false},
		{"unparsed command", `{"tool_name":"Bash","tool_input":{"command":"git 'x"}}`, false},
	}
	for _, tt := range tests {
		var out bytes.Buffer
		if err := PreToolUse(strings.NewReader(tt.in), &out); err != nil {
			t.Fatalf("%s: %v", tt.name, err)
		}
		if got := out.Len() > 0; got != tt.deny {
			t.Errorf("%s: deny = %v, want %v (%q)", tt.name, got, tt.deny, out.String())
		}
		if tt.deny && (!strings.Contains(out.String(), `"permissionDecision":"deny"`) || !strings.Contains(out.String(), "git_* tools")) {
			t.Errorf("%s: output = %q", tt.name, out.String())
		}
	}
}
