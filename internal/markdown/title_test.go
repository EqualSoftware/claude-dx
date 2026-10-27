package markdown

import "testing"

func TestReplaceTitle(t *testing.T) {
	tests := []struct{ name, body, heading, want string }{
		{"fenced line after title", "# Old\n\n```sh\n# install\n```\n", "# New", "# New\n\n```sh\n# install\n```\n"},
		{"no title line", "text\n## Section\n", "# New", "text\n## Section\n"},
		{"dollar in title", "# Old\n", "# Cost $1 and ${x}", "# Cost $1 and ${x}\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReplaceTitle(tt.body, tt.heading); got != tt.want {
				t.Errorf("ReplaceTitle = %q, want %q", got, tt.want)
			}
		})
	}
}
