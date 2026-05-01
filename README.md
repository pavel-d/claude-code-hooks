# claude-code-hooks

Small deterministic hook binaries for blocking Bash commands that create
unskippable prompts or bypass the normal file editing tools.

## Build

```sh
make build
```

This creates:

- `bin/block-bash-antipatterns` for Claude Code
- `bin/codex-block-bash-antipatterns` for Codex

Both binaries use the same rule engine and the same rules from
`internal/policy`.

## Install

Install both hook binaries:

```sh
make install
```

This prints the installed hook paths and where to configure them.

Or install them separately:

```sh
make install-claude
make install-codex
```

By default, Claude installs to `~/.claude/hooks` and Codex installs to
`~/.codex/hooks`. Override the destinations with:

```sh
make install-claude CLAUDE_HOOKS_DIR=/path/to/claude/hooks
make install-codex CODEX_HOOKS_DIR=/path/to/codex/hooks
```

## Configure Claude Code

Add this to `~/.claude/settings.json`, or use `/hooks` and add a
`PreToolUse` matcher for `Bash` with the same command:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "/Users/pavlo/.claude/hooks/block-bash-antipatterns"
          }
        ]
      }
    ]
  }
}
```

Claude Code hook config details are documented at:
https://docs.claude.com/en/docs/claude-code/hooks

## Configure Codex

Codex must also have hooks enabled and registered in `~/.codex/config.toml`
or a `hooks.json` file.

Codex config:

```toml
[features]
codex_hooks = true

[[hooks.PreToolUse]]
matcher = "^Bash$"

[[hooks.PreToolUse.hooks]]
type = "command"
command = "/Users/pavlo/.codex/hooks/codex-block-bash-antipatterns"
timeout = 30
statusMessage = "Checking Bash command"
```

Codex hook config details are documented at:
https://developers.openai.com/codex/hooks

## Test

```sh
make test
```
