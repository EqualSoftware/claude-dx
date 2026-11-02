package hook

import (
	"encoding/json"
	"io"
)

type input struct {
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		Command string `json:"command"`
	} `json:"tool_input"`
}

type output struct {
	HookSpecificOutput struct {
		HookEventName            string `json:"hookEventName"`
		PermissionDecision       string `json:"permissionDecision"`
		PermissionDecisionReason string `json:"permissionDecisionReason"`
	} `json:"hookSpecificOutput"`
}

// PreToolUse reads the PreToolUse event from in. It writes a deny decision to out when a Bash command
// runs git. Otherwise it writes nothing, and Claude Code allows the call. An input that does not
// parse is allowed (R-15).
func PreToolUse(in io.Reader, out io.Writer) error {
	var ev input
	if err := json.NewDecoder(in).Decode(&ev); err != nil {
		return nil
	}
	if ev.ToolName != "Bash" || !RunsGit(ev.ToolInput.Command) {
		return nil
	}
	var o output
	o.HookSpecificOutput.HookEventName = "PreToolUse"
	o.HookSpecificOutput.PermissionDecision = "deny"
	o.HookSpecificOutput.PermissionDecisionReason = DenyReason
	return json.NewEncoder(out).Encode(o)
}
