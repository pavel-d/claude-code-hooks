package policy

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRunDeniesWithJSON(t *testing.T) {
	stdin := strings.NewReader(`{"tool_name":"Bash","tool_input":{"command":"cd /tmp && ls"}}`)
	var stdout bytes.Buffer
	if err := Run(stdin, &stdout); err != nil {
		t.Fatalf("Run: %v", err)
	}

	var out output
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if out.HookSpecificOutput.HookEventName != "PreToolUse" {
		t.Errorf("HookEventName = %q, want PreToolUse", out.HookSpecificOutput.HookEventName)
	}
	if out.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("PermissionDecision = %q, want deny", out.HookSpecificOutput.PermissionDecision)
	}
	if out.HookSpecificOutput.PermissionDecisionReason == "" {
		t.Errorf("PermissionDecisionReason is empty, want a non-empty explanation")
	}
}

func TestRunAllowsBenignWithEmptyOutput(t *testing.T) {
	stdin := strings.NewReader(`{"tool_name":"Bash","tool_input":{"command":"ls /tmp"}}`)
	var stdout bytes.Buffer
	if err := Run(stdin, &stdout); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if stdout.Len() != 0 {
		t.Errorf("benign command produced output: %q", stdout.String())
	}
}

func TestRunCombinesMultipleReasonsWithNewline(t *testing.T) {
	stdin := strings.NewReader(`{"tool_name":"Bash","tool_input":{"command":"cd /tmp && git -C /r status"}}`)
	var stdout bytes.Buffer
	if err := Run(stdin, &stdout); err != nil {
		t.Fatalf("Run: %v", err)
	}
	var out output
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	reason := out.HookSpecificOutput.PermissionDecisionReason
	lines := strings.Split(reason, "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 newline-separated reasons, got %d: %q", len(lines), reason)
	}
}

func TestRunMissingFieldsTreatedAsEmpty(t *testing.T) {
	stdin := strings.NewReader(`{}`)
	var stdout bytes.Buffer
	if err := Run(stdin, &stdout); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if stdout.Len() != 0 {
		t.Errorf("missing fields should produce no output, got: %q", stdout.String())
	}
}

func TestRunInvalidJSONReturnsError(t *testing.T) {
	stdin := strings.NewReader(`not json`)
	var stdout bytes.Buffer
	if err := Run(stdin, &stdout); err == nil {
		t.Errorf("expected error for invalid JSON, got nil")
	}
}
