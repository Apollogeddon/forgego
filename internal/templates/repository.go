package templates

import (
	"fmt"
	"strings"
)

// Editorconfig matches what gofmt writes: tabs in Go files, two spaces elsewhere.
const Editorconfig = `root = true

[*]
indent_style = space
indent_size = 2
end_of_line = lf
charset = utf-8
trim_trailing_whitespace = true
insert_final_newline = true
max_line_length = 120

[{*.go,go.mod,go.sum,Makefile}]
indent_style = tab

[*.md]
trim_trailing_whitespace = false
`

// Codeowners requests the owner's review on every pull request someone else opens,
// Dependabot's and release-please's included.
const Codeowners = `# Every pull request opened by someone else, Dependabot and release-please included,
# requests a review from the owner, so it shows in their review requests.
* @__OWNER__
`

type ecosystem struct {
	name, prefix, group string
	updateTypes         []string // only these are grouped; others get a pull request each
	ignore              []string
}

func (e ecosystem) render() string {
	lines := []string{
		fmt.Sprintf("  - package-ecosystem: %q", e.name),
		`    directory: "/"`,
		"    schedule:",
		`      interval: "weekly"`,
		"    groups:",
		"      " + e.group + ":",
		"        patterns:",
		`          - "*"`,
	}
	if len(e.updateTypes) > 0 {
		lines = append(lines, "        update-types:")
		for _, t := range e.updateTypes {
			lines = append(lines, fmt.Sprintf("          - %q", t))
		}
	}
	lines = append(lines, "    commit-message:", fmt.Sprintf("      prefix: %q", e.prefix))
	if len(e.ignore) > 0 {
		lines = append(lines, "    ignore:")
		for _, d := range e.ignore {
			lines = append(lines, fmt.Sprintf("      - dependency-name: %q", d))
		}
	}
	return strings.Join(append(lines, "    cooldown:", "      default-days: 3"), "\n")
}

// Dependabot proposes the project's module and GitHub Actions updates weekly, and its
// Docker base images with docker. The tools pinned in .forgego/ are left to forgego sync.
func Dependabot(docker bool) string {
	ecosystems := []ecosystem{
		{name: "gomod", prefix: "fix(deps)", group: "dependencies", updateTypes: []string{"minor", "patch"}},
		// the reusable workflows are called at @main, which has no versions to propose
		{name: "github-actions", prefix: "chore(ci)", group: "actions", ignore: []string{"apollogeddon/forgego"}},
	}
	if docker {
		ecosystems = append(ecosystems, ecosystem{name: "docker", prefix: "fix(deps)", group: "docker", updateTypes: []string{"minor", "patch"}})
	}
	rendered := make([]string, len(ecosystems))
	for i, e := range ecosystems {
		rendered[i] = e.render()
	}
	return `version: 2
# Every update waits 3 days after a version is published before it's proposed, so a
# compromised release has time to be caught and yanked upstream first.
updates:
` + strings.Join(rendered, "\n\n") + "\n"
}
