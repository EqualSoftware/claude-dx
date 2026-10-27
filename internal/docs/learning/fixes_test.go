package learning

import "testing"

func TestCreateRefusesBadText(t *testing.T) {
	session := connect(t, newStore(t))
	for name, args := range map[string]map[string]any{
		"empty description": {"topic": "a", "description": " ", "body": "x"},
		"empty body":        {"topic": "a", "description": "d", "body": "\n"},
		"unclosed fence":    {"topic": "a", "description": "d", "body": "```\ncode"},
	} {
		if text, isErr := call(t, session, "learning_create", args); !isErr {
			t.Errorf("%s: learning_create = %q, want an error", name, text)
		}
	}
	if text, isErr := call(t, session, "learning_create", map[string]any{"topic": "a", "description": "d", "body": "~~~\n# x\n~~~"}); isErr {
		t.Errorf("a closed fence is refused: %s", text)
	}
}

func TestUpdateRefusesBadText(t *testing.T) {
	session := connect(t, newStore(t))
	call(t, session, "learning_create", map[string]any{"topic": "a", "description": "d", "body": "x"})
	for name, args := range map[string]map[string]any{
		"blank description": {"topic": "a", "description": "  "},
		"unclosed fence":    {"topic": "a", "body": "```\ncode"},
	} {
		if text, isErr := call(t, session, "learning_update", args); !isErr {
			t.Errorf("%s: learning_update = %q, want an error", name, text)
		}
	}
}
