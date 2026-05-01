// block-bash-antipatterns is a Claude Code PreToolUse hook for the Bash tool.
// It denies commands that either trigger an unskippable safety prompt or just
// create avoidable approval friction. See README for the full list.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
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

type rule struct {
	name    string
	pattern *regexp.Regexp
	reason  string
}

// Word boundary that allows shell-meaningful preceding characters.
const boundary = `(^|[\s;|&(])`

var rules = []rule{
	{
		name:    "find-exec",
		pattern: regexp.MustCompile(boundary + `find\s.*\s-exec(\s|$)`),
		reason:  "`find -exec` is blocked (triggers unskippable safety prompt). Use Glob/Grep tools or `find ... | xargs rg ...`.",
	},
	{
		name:    "git-C",
		pattern: regexp.MustCompile(boundary + `git\s+-C(\s|$)`),
		reason:  "`git -C <path>` is blocked. The cwd is already the target repo, or you should `cd <dir>` once in a separate Bash call and then run plain `git ...` from there.",
	},
	{
		name:    "go-C",
		pattern: regexp.MustCompile(boundary + `go\s+-C(\s|$)`),
		reason:  "`go -C <path>` is blocked. The cwd is already the target dir, or you should `cd <dir>` once in a separate Bash call and then run plain `go ...` from there.",
	},
	{
		name:    "cd-and-chain",
		pattern: regexp.MustCompile(boundary + `cd\s.*&&`),
		reason:  "Chained `cd <dir> && ...` is blocked (triggers unskippable safety prompt EVERY time). Run `cd <dir>` ONCE in its own Bash call, then run the next command in a separate Bash call from there.",
	},
}

func main() {
	if err := run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(stdin io.Reader, stdout io.Writer) error {
	raw, err := io.ReadAll(stdin)
	if err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}

	var in input
	if err := json.Unmarshal(raw, &in); err != nil {
		return fmt.Errorf("parse json: %w", err)
	}

	hits := check(in.ToolInput.Command)
	if len(hits) == 0 {
		return nil
	}

	combined := ""
	for i, h := range hits {
		if i > 0 {
			combined += "\n"
		}
		combined += h.reason
	}

	out := output{
		HookSpecificOutput: hookSpecificOutput{
			HookEventName:            "PreToolUse",
			PermissionDecision:       "deny",
			PermissionDecisionReason: combined,
		},
	}
	return json.NewEncoder(stdout).Encode(out)
}

func check(cmd string) []rule {
	var hits []rule
	for _, r := range rules {
		if r.pattern.MatchString(cmd) {
			hits = append(hits, r)
		}
	}
	return hits
}
