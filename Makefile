CLAUDE_HOOKS_DIR ?= $(HOME)/.claude/hooks
CODEX_HOOKS_DIR ?= $(HOME)/.codex/hooks
CLAUDE_BINARIES := block-bash-antipatterns
CODEX_BINARIES := codex-block-bash-antipatterns
BINARIES := $(CLAUDE_BINARIES) $(CODEX_BINARIES)

.PHONY: help all build install install-claude install-codex install-instructions test clean

help:
	@echo "This makefile contains the following targets,"
	@echo "from most commonly used to least:"
	@echo "  make install          build and install both Claude Code and Codex hooks"
	@echo "  make test             run the Go test suite"
	@echo "  make build            build hook binaries into ./bin"
	@echo "  make install-claude   install only the Claude Code hook"
	@echo "  make install-codex    install only the Codex hook"
	@echo "  make clean            remove ./bin"

all: build

build:
	@for bin in $(BINARIES); do \
		echo "Building $$bin..."; \
		go build -o ./bin/$$bin ./cmd/$$bin; \
	done

install: install-claude install-codex install-instructions

install-claude: build
	@mkdir -p $(CLAUDE_HOOKS_DIR)
	@for bin in $(CLAUDE_BINARIES); do \
		echo "Installing $$bin to $(CLAUDE_HOOKS_DIR)/$$bin..."; \
		install -m 0755 ./bin/$$bin $(CLAUDE_HOOKS_DIR)/$$bin; \
	done

install-codex: build
	@mkdir -p $(CODEX_HOOKS_DIR)
	@for bin in $(CODEX_BINARIES); do \
		echo "Installing $$bin to $(CODEX_HOOKS_DIR)/$$bin..."; \
		install -m 0755 ./bin/$$bin $(CODEX_HOOKS_DIR)/$$bin; \
	done

install-instructions:
	@echo ""
	@echo "Hook binaries installed."
	@echo "Claude Code: run /hooks, add a PreToolUse hook for Bash, command: $(CLAUDE_HOOKS_DIR)/block-bash-antipatterns"
	@echo "Codex: enable hooks in ~/.codex/config.toml and add a PreToolUse Bash command hook: $(CODEX_HOOKS_DIR)/codex-block-bash-antipatterns"
	@echo "Full config snippets are in README.md."

test:
	go test ./...

clean:
	rm -rf ./bin
