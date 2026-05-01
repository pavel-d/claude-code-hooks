// Package policy evaluates hook guardrail rules.
package policy

import "regexp"

// Rule describes one blocked Bash pattern and the explanation shown to users.
type Rule struct {
	Name    string
	Pattern *regexp.Regexp
	Reason  string
}

// Word boundary that allows shell-meaningful preceding characters.
const boundary = `(^|[\s;|&(])`

// Rules is the single source of truth for blocked Bash antipatterns.
var Rules = []Rule{
	{
		Name:    "find-exec",
		Pattern: regexp.MustCompile(boundary + `find\s.*\s-exec(\s|$)`),
		Reason:  "`find -exec` is blocked (triggers unskippable safety prompt). Use Glob/Grep tools or `find ... | xargs rg ...`.",
	},
	{
		Name:    "git-C",
		Pattern: regexp.MustCompile(boundary + `git\s+-C(\s|$)`),
		Reason:  "`git -C <path>` is blocked. The cwd is already the target repo, or you should `cd <dir>` once in a separate Bash call and then run plain `git ...` from there.",
	},
	{
		Name:    "go-C",
		Pattern: regexp.MustCompile(boundary + `go\s+-C(\s|$)`),
		Reason:  "`go -C <path>` is blocked. The cwd is already the target dir, or you should `cd <dir>` once in a separate Bash call and then run plain `go ...` from there.",
	},
	{
		Name:    "cd-and-chain",
		Pattern: regexp.MustCompile(boundary + `cd\s.*&&`),
		Reason:  "Chained `cd <dir> && ...` is blocked (triggers unskippable safety prompt EVERY time). Run `cd <dir>` ONCE in its own Bash call, then run the next command in a separate Bash call from there.",
	},
	{
		Name:    "printf-redirect",
		Pattern: regexp.MustCompile(boundary + `printf\s+[^;&|]*>{1,2}\s*\S+`),
		Reason:  "`printf ... > <file>` is blocked. Use the Edit/Write tools for file changes instead of shell redirection.",
	},
	{
		Name:    "python-inline-write",
		Pattern: regexp.MustCompile(boundary + `python[0-9.]*\s+[\s\S]*-c\s+[\s\S]*\b(write_text|write_bytes)\s*\(`),
		Reason:  "`python -c` file writes are blocked. Use the Edit/Write tools for file changes instead.",
	},
	{
		Name:    "cat-heredoc-redirect",
		Pattern: regexp.MustCompile(boundary + `cat\s+>{1,2}\s*\S+\s*<<-?`),
		Reason:  "`cat > <file> <<EOF` is blocked. Use the Edit/Write tools for file changes instead of heredoc redirection.",
	},
}

// Check returns all rules whose pattern matches cmd.
func Check(cmd string) []Rule {
	var hits []Rule
	for _, r := range Rules {
		if r.Pattern.MatchString(cmd) {
			hits = append(hits, r)
		}
	}
	return hits
}

// Reasons joins rule explanations in rule order.
func Reasons(hits []Rule) string {
	combined := ""
	for i, h := range hits {
		if i > 0 {
			combined += "\n"
		}
		combined += h.Reason
	}
	return combined
}
