package policy

import (
	"slices"
	"sort"
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

		// file-editing shell hacks
		{name: "printf redirect blocked", cmd: "printf '%s\n' foo > /tmp/config.toml", want: []string{"printf-redirect"}},
		{name: "printf append redirect blocked", cmd: "printf '%s\n' foo >> /tmp/config.toml", want: []string{"printf-redirect"}},
		{name: "printf without redirect allowed", cmd: "printf '%s\n' foo"},
		{name: "python inline write_text blocked", cmd: `python3 -c 'from pathlib import Path; Path("/tmp/config.toml").write_text("x")'`, want: []string{"python-inline-write"}},
		{name: "python inline write_bytes blocked", cmd: `python -c 'from pathlib import Path; Path("/tmp/config.toml").write_bytes(b"x")'`, want: []string{"python-inline-write"}},
		{name: "python inline print allowed", cmd: `python3 -c 'print("hello")'`},
		{name: "cat heredoc redirect blocked", cmd: "cat > /tmp/config.toml <<'EOF'\nfoo\nEOF", want: []string{"cat-heredoc-redirect"}},
		{name: "cat heredoc append redirect blocked", cmd: "cat >> /tmp/config.toml <<EOF\nfoo\nEOF", want: []string{"cat-heredoc-redirect"}},
		{name: "plain cat allowed", cmd: `cat /tmp/config.toml`},

		// Multiple rules firing at once
		{name: "cd && plus git -C", cmd: `cd /tmp && git -C /repo status`, want: []string{"cd-and-chain", "git-C"}},
		{name: "cd && plus printf redirect", cmd: "cd /tmp && printf '%s\n' foo > config.toml", want: []string{"cd-and-chain", "printf-redirect"}},

		// Benign
		{name: "ls allowed", cmd: `ls -la`},
		{name: "empty allowed", cmd: ``},
		{name: "rg allowed", cmd: `rg pattern src/`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hits := Check(tt.cmd)
			got := make([]string, len(hits))
			for i, h := range hits {
				got[i] = h.Name
			}
			sort.Strings(got)
			want := slices.Clone(tt.want)
			sort.Strings(want)
			if !slices.Equal(got, want) {
				t.Errorf("Check(%q) fired %v, want %v", tt.cmd, got, want)
			}
		})
	}
}
