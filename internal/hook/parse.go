package hook

import "strings"

// lexer collects the simple commands of a shell command.
type lexer struct {
	cmds         [][]string
	subs         []string
	cur          []string
	word, expand strings.Builder
	inWord       bool
}

func (l *lexer) endWord() {
	if l.inWord {
		l.cur = append(l.cur, l.word.String())
		l.word.Reset()
		l.inWord = false
	}
	l.subs = append(l.subs, substitutions(l.expand.String())...)
	l.expand.Reset()
}

func (l *lexer) endCmd() {
	l.endWord()
	if len(l.cur) > 0 {
		l.cmds = append(l.cmds, l.cur)
		l.cur = nil
	}
}

// quoted adds the text between two quotes to the word and returns the index of the closing quote,
// or -1. Only double quotes expand command substitutions.
func (l *lexer) quoted(rs []rune, i int) int {
	var end int
	if rs[i] == '\'' {
		end = indexRune(rs, i+1, '\'')
	} else {
		end = closingQuote(rs, i+1)
	}
	if end < 0 {
		return -1
	}
	l.word.WriteString(string(rs[i+1 : end]))
	if rs[i] == '"' {
		l.expand.WriteString(string(rs[i+1 : end]))
	}
	l.inWord = true
	return end
}

// split breaks a shell command into simple commands (lists of words). It also returns the text of
// command substitutions inside double quotes. ok is false when a quote is not closed.
func split(s string) (cmds [][]string, subs []string, ok bool) {
	var l lexer
	rs := []rune(s)
	for i := 0; i < len(rs); i++ {
		switch c := rs[i]; c {
		case '\'', '"':
			if i = l.quoted(rs, i); i < 0 {
				return nil, nil, false
			}
		case '\\':
			if i+1 < len(rs) && rs[i+1] != '\n' {
				l.word.WriteRune(rs[i+1])
				l.inWord = true
			}
			i++
		case ' ', '\t', '\r':
			l.endWord()
		case ';', '&', '|', '\n', '(', ')', '`':
			l.endCmd()
		default:
			l.word.WriteRune(c)
			l.inWord = true
		}
	}
	l.endCmd()
	return l.cmds, l.subs, true
}

func indexRune(rs []rune, from int, r rune) int {
	for i := from; i < len(rs); i++ {
		if rs[i] == r {
			return i
		}
	}
	return -1
}

// closingQuote returns the index of the double quote that closes a string, or -1.
func closingQuote(rs []rune, from int) int {
	for i := from; i < len(rs); i++ {
		switch rs[i] {
		case '\\':
			i++
		case '"':
			return i
		}
	}
	return -1
}

// substitutions returns the text of each $(...) and `...` in s.
func substitutions(s string) []string {
	var out []string
	for i := 0; i < len(s); i++ {
		switch {
		case strings.HasPrefix(s[i:], "$("):
			end := matchParen(s, i+2)
			out = append(out, s[i+2:end])
			i = end
		case s[i] == '`':
			end := strings.IndexByte(s[i+1:], '`')
			if end < 0 {
				end = len(s) - i - 1
			}
			out = append(out, s[i+1:i+1+end])
			i += end + 1
		}
	}
	return out
}

// matchParen returns the index of the parenthesis that closes one opened before from.
func matchParen(s string, from int) int {
	depth := 1
	for i := from; i < len(s); i++ {
		switch s[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return len(s)
}
