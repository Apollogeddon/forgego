// Package commitmsg checks a commit message against Conventional Commits, so the hook
// needs nothing beyond Go on any platform.
package commitmsg

import (
	"fmt"
	"regexp"
	"strings"
)

// Types are the commit types the hook accepts.
var Types = []string{"build", "chore", "ci", "docs", "feat", "fix", "perf", "refactor", "revert", "style", "test"}

var header = regexp.MustCompile(`^(` + strings.Join(Types, "|") + `)(\([\w.,/ -]+\))?!?: \S.*$`)

// exempt are messages git or a tool writes, which a hook shouldn't reject.
var exempt = regexp.MustCompile(`^(Merge |Revert "|fixup! |squash! |amend! )`)

// MaxHeader is the longest subject line accepted.
const MaxHeader = 100

// Check returns why message isn't a Conventional Commit, or nil.
func Check(message string) error {
	first := ""
	for _, line := range strings.Split(message, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		first = strings.TrimRight(line, " \t\r")
		break
	}
	switch {
	case first == "":
		return fmt.Errorf("the commit message is empty")
	case exempt.MatchString(first):
		return nil
	case !header.MatchString(first):
		return fmt.Errorf("%q isn't a Conventional Commit: use <type>(<scope>): <subject>, where type is one of %s",
			first, strings.Join(Types, ", "))
	case len(first) > MaxHeader:
		return fmt.Errorf("the subject line is %d characters long; keep it to %d", len(first), MaxHeader)
	}
	return nil
}
