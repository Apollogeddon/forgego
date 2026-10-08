package commitmsg

import (
	"strings"
	"testing"
)

func TestCheck(t *testing.T) {
	valid := []string{
		"feat: add sync",
		"fix(cli): reject --debian outside backend mode",
		"feat(api)!: drop v1 routes",
		"chore(deps): bump golangci-lint\n\nBody text.",
		"# comment git leaves above\n\ndocs: explain tools",
		"Merge pull request #12 from acme/feat",
		"Revert \"feat: add sync\"",
		"fixup! fix: typo",
	}
	for _, msg := range valid {
		if err := Check(msg); err != nil {
			t.Errorf("Check(%q) = %v, want nil", msg, err)
		}
	}

	invalid := []string{
		"",
		"# only a comment",
		"added sync",
		"feature: add sync",
		"feat:add sync",
		"feat(): add sync",
		"feat: " + strings.Repeat("x", MaxHeader),
	}
	for _, msg := range invalid {
		if Check(msg) == nil {
			t.Errorf("Check(%q) = nil, want an error", msg)
		}
	}
}
