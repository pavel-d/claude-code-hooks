// codex-block-bash-antipatterns is a Codex PreToolUse hook for the Bash tool.
package main

import (
	"fmt"
	"os"

	"github.com/pavel-d/claude-code-hooks/internal/policy"
)

func main() {
	if err := policy.Run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
