package policy

import (
	"encoding/json"
	"fmt"
	"io"
)

type input struct {
	ToolInput struct {
		Command string `json:"command"`
	} `json:"tool_input"`
}

type output struct {
	HookSpecificOutput hookSpecificOutput `json:"hookSpecificOutput"`
}

type hookSpecificOutput struct {
	HookEventName            string `json:"hookEventName"`
	PermissionDecision       string `json:"permissionDecision"`
	PermissionDecisionReason string `json:"permissionDecisionReason"`
}

// Run reads a PreToolUse hook payload, evaluates the shared rule engine, and
// writes a deny decision when the command should be blocked.
func Run(stdin io.Reader, stdout io.Writer) error {
	raw, err := io.ReadAll(stdin)
	if err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}

	var in input
	if err := json.Unmarshal(raw, &in); err != nil {
		return fmt.Errorf("parse json: %w", err)
	}

	hits := Check(in.ToolInput.Command)
	if len(hits) == 0 {
		return nil
	}

	out := output{
		HookSpecificOutput: hookSpecificOutput{
			HookEventName:            "PreToolUse",
			PermissionDecision:       "deny",
			PermissionDecisionReason: Reasons(hits),
		},
	}
	return json.NewEncoder(stdout).Encode(out)
}
