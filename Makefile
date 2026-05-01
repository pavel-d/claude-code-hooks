HOOKS_DIR ?= $(HOME)/.claude/hooks
BINARIES := block-bash-antipatterns

.PHONY: all build install test clean

all: build

build:
	@for bin in $(BINARIES); do \
		echo "Building $$bin..."; \
		go build -o ./bin/$$bin ./cmd/$$bin; \
	done

install: build
	@mkdir -p $(HOOKS_DIR)
	@for bin in $(BINARIES); do \
		echo "Installing $$bin to $(HOOKS_DIR)/$$bin..."; \
		install -m 0755 ./bin/$$bin $(HOOKS_DIR)/$$bin; \
	done

test:
	go test ./...

clean:
	rm -rf ./bin
