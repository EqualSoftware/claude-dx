package markdown

import "regexp"

var titleLine = regexp.MustCompile(`(?m)^# .*$`)

// ReplaceTitle replaces the first "# " line of body, the title heading. Other lines that start with
// "# ", for example comments in a fenced code block, do not change.
func ReplaceTitle(body, heading string) string {
	loc := titleLine.FindStringIndex(body)
	if loc == nil {
		return body
	}
	return body[:loc[0]] + heading + body[loc[1]:]
}
