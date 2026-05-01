package main

import (
	"bytes"
	"encoding/json"
	"slices"
	"sort"
	"strings"
	"testing"
)

func TestCheck(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		want []string // names of rules that should fire, order-independent
	}{
		// find-exec
		{name: "find -exec blocked", cmd: `find . -name "*.tmp" -exec rm {} \;`, want: []string{"find-exec"}},
		{name: "find -exec at end of string", cmd: `find . -exec`, want: []string{"find-exec"}},
		{name: "find without -exec allowed", cmd: `find . -name "*.go"`},
		{name: "find piped to xargs allowed", cmd: `find . -name "*.go" | xargs rg foo`},
		{name: "filename containing find-exec not blocked", cmd: `cat block-find-exec.sh`},

		// git-C
		{name: "git -C blocked", cmd: `git -C /tmp status`, want: []string{"git-C"}},
		{name: "git -C at end of string", cmd: `git -C`, want: []string{"git-C"}},
		{name: "plain git allowed", cmd: `git status`},
		{name: "git with -c (lowercase) allowed", cmd: `git -c user.name=foo commit`},

		// go-C
		{name: "go -C blocked", cmd: `go -C ./svc test ./...`, want: []string{"go-C"}},
		{name: "plain go test allowed", cmd: `go test ./...`},

		// cd-and-chain
		{name: "cd && simple chain blocked", cmd: `cd /tmp && ls`, want: []string{"cd-and-chain"}},
		{name: "cd && with leading whitespace blocked", cmd: `  cd /tmp && ls`, want: []string{"cd-and-chain"}},
		{name: "cd && in middle of chain blocked", cmd: `foo && cd /tmp && ls`, want: []string{"cd-and-chain"}},
		{name: "cd && in subshell blocked", cmd: `(cd /tmp && ls)`, want: []string{"cd-and-chain"}},
		{name: "cd alone allowed", cmd: `cd /tmp`},
		{name: "cd at end of chain (no && after) allowed", cmd: `foo && cd /tmp`},
		{name: "cd with semicolon allowed", cmd: `cd /tmp; ls`},
		{name: "mycd && bar not blocked (word boundary)", cmd: `mycd /tmp && bar`},

		// Multiple rules firing at once
		{name: "cd && plus git -C", cmd: `cd /tmp && git -C /repo status`, want: []string{"cd-and-chain", "git-C"}},

		// Benign
		{name: "ls allowed", cmd: `ls -la`},
		{name: "empty allowed", cmd: ``},
		{name: "rg allowed", cmd: `rg pattern src/`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hits := check(tt.cmd)
			got := make([]string, len(hits))
			for i, h := range hits {
				got[i] = h.name
			}
			sort.Strings(got)
			want := slices.Clone(tt.want)
			sort.Strings(want)
			if !slices.Equal(got, want) {
				t.Errorf("check(%q) fired %v, want %v", tt.cmd, got, want)
			}
		})
	}
}

func TestRunDeniesWithJSON(t *testing.T) {
	stdin := strings.NewReader(`{"tool_name":"Bash","tool_input":{"command":"cd /tmp && ls"}}`)
	var stdout bytes.Buffer
	if err := run(stdin, &stdout); err != nil {
		t.Fatalf("run: %v", err)
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
	if err := run(stdin, &stdout); err != nil {
		t.Fatalf("run: %v", err)
	}
	if stdout.Len() != 0 {
		t.Errorf("benign command produced output: %q", stdout.String())
	}
}

func TestRunCombinesMultipleReasonsWithNewline(t *testing.T) {
	stdin := strings.NewReader(`{"tool_name":"Bash","tool_input":{"command":"cd /tmp && git -C /r status"}}`)
	var stdout bytes.Buffer
	if err := run(stdin, &stdout); err != nil {
		t.Fatalf("run: %v", err)
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
	if err := run(stdin, &stdout); err != nil {
		t.Fatalf("run: %v", err)
	}
	if stdout.Len() != 0 {
		t.Errorf("missing fields should produce no output, got: %q", stdout.String())
	}
}

func TestRunInvalidJSONReturnsError(t *testing.T) {
	stdin := strings.NewReader(`not json`)
	var stdout bytes.Buffer
	if err := run(stdin, &stdout); err == nil {
		t.Errorf("expected error for invalid JSON, got nil")
	}
}
