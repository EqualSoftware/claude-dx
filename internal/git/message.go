package git

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// maxHeader is the maximum length of the commit header (ADR-003).
const maxHeader = 72

// commitTypes are the Conventional Commit types of ADR-003.
var commitTypes = map[string]bool{
	"feat": true, "fix": true, "refactor": true, "perf": true, "test": true, "docs": true,
	"build": true, "ci": true, "style": true, "chore": true, "revert": true,
}

var (
	scopePattern    = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
	coAuthorPattern = regexp.MustCompile(`^[^<>\n]+ <[^<>\s]+@[^<>\s]+>$`)
)

// notImperative lists first words in the third person or the past tense that the -ed and -ing rule
// does not find.
var notImperative = map[string]bool{
	"adds": true, "fixes": true, "updates": true, "removes": true, "changes": true, "uses": true,
	"makes": true, "sets": true, "gets": true, "moves": true, "renames": true, "creates": true, "deletes": true,
	"ran": true, "made": true, "got": true,
}

// edIngExceptions are imperative verbs and nouns that end in -ed or -ing.
var edIngExceptions = map[string]bool{
	"embed": true, "need": true, "speed": true, "feed": true, "seed": true, "proceed": true,
	"exceed": true, "bring": true, "string": true, "spring": true, "swing": true, "bed": true,
}

// Message is the input of a Conventional Commit (ADR-003).
type Message struct {
	Type      string
	Scope     string
	Summary   string
	Body      string
	Breaking  bool
	CoAuthors []string
}

// Header returns the commit header: type(scope)!: summary.
func (m Message) Header() string {
	scope := ""
	if m.Scope != "" {
		scope = "(" + m.Scope + ")"
	}
	bang := ""
	if m.Breaking {
		bang = "!"
	}
	return m.Type + scope + bang + ": " + m.Summary
}

// Args validates the message and returns the -m arguments of git commit: the header, the body and the
// footers. The footers are BREAKING CHANGE, then Co-Authored-By (git-tools#R-5).
func (m Message) Args() ([]string, error) {
	if err := m.validate(); err != nil {
		return nil, err
	}
	args := []string{"-m", m.Header()}
	if body := strings.TrimSpace(m.Body); body != "" {
		args = append(args, "-m", body)
	}
	var footers []string
	if m.Breaking {
		footers = append(footers, "BREAKING CHANGE: "+m.Summary)
	}
	for _, c := range m.CoAuthors {
		footers = append(footers, "Co-Authored-By: "+c)
	}
	if len(footers) > 0 {
		args = append(args, "-m", strings.Join(footers, "\n"))
	}
	return args, nil
}

func (m Message) validate() error {
	if !commitTypes[m.Type] {
		return fmt.Errorf("git_commit refused: type %q is not one of feat, fix, refactor, perf, test, docs, build, ci, style, chore, revert", m.Type)
	}
	if m.Scope != "" && !scopePattern.MatchString(m.Scope) {
		return fmt.Errorf("git_commit refused: scope %q must be lowercase letters, digits and dashes", m.Scope)
	}
	if err := validSummary(m.Summary); err != nil {
		return fmt.Errorf("git_commit refused: %w", err)
	}
	if n := utf8.RuneCountInString(m.Header()); n > maxHeader {
		return fmt.Errorf("git_commit refused: the header has %d characters and the maximum is %d", n, maxHeader)
	}
	for _, c := range m.CoAuthors {
		if !coAuthorPattern.MatchString(c) {
			return fmt.Errorf("git_commit refused: co-author %q must have the form Name <email>", c)
		}
	}
	return nil
}

func validSummary(s string) error {
	if strings.TrimSpace(s) == "" || strings.ContainsAny(s, "\r\n") {
		return fmt.Errorf("the summary must be one line and not empty")
	}
	if strings.HasSuffix(s, ".") {
		return fmt.Errorf("the summary must not end with a period")
	}
	if first, _ := utf8.DecodeRuneInString(s); unicode.IsUpper(first) {
		return fmt.Errorf("the summary must start with a lowercase letter")
	}
	word, _, _ := strings.Cut(s, " ")
	if !imperative(word) {
		return fmt.Errorf("the summary must start with an imperative verb, not %q", word)
	}
	return nil
}

// imperative is a heuristic: it refuses the words of notImperative and the words that end in -ed or -ing.
func imperative(word string) bool {
	if notImperative[word] {
		return false
	}
	if edIngExceptions[word] {
		return true
	}
	return !strings.HasSuffix(word, "ed") && !strings.HasSuffix(word, "ing")
}
