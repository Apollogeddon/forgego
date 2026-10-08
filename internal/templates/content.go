// Package templates holds the content of every file forgego generates. Templates are
// plain strings with __TOKEN__ placeholders: GitHub Actions' ${{ }} and Task's {{ }}
// both clash with text/template, so substitution is explicit.
package templates

import "strings"

// Render replaces each __KEY__ token with its value.
func Render(template string, values map[string]string) string {
	pairs := make([]string, 0, len(values)*2)
	for key, value := range values {
		pairs = append(pairs, "__"+key+"__", value)
	}
	return strings.NewReplacer(pairs...).Replace(template)
}

// GoMod starts a project's go.mod.
const GoMod = `module __MODULE__

go __GO__
`

// Gitignore covers what the generated tasks write.
const Gitignore = `# Build output
/dist/

# Test reports
coverage.out
coverage.html
junit-report.xml
`

// GitignoreWebsite covers what Hugo writes.
const GitignoreWebsite = `# Hugo output
/public/
/resources/_gen/
.hugo_build.lock
`

// BackendMain is the starter command for --backend.
const BackendMain = `// Command __NAME__ is the entry point for the service.
package main

import (
	"fmt"
	"io"
	"os"
)

// version is set at build time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if err := run(os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(out io.Writer) error {
	_, err := fmt.Fprintf(out, "__NAME__ %s\n", version)
	return err
}
`

// BackendMainTest tests the starter command.
const BackendMainTest = `package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunPrintsTheVersion(t *testing.T) {
	var out bytes.Buffer
	if err := run(&out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), version) {
		t.Errorf("run() printed %q, want the version", out.String())
	}
}
`

// LibrarySource is the starter package for --library.
const LibrarySource = `// Package __PACKAGE__ is the library's public API.
package __PACKAGE__

// Greeting returns a greeting for name.
func Greeting(name string) string {
	return "Hello, " + name
}
`

// LibraryTest tests the starter package.
const LibraryTest = `package __PACKAGE__

import "testing"

func TestGreeting(t *testing.T) {
	if got := Greeting("Go"); got != "Hello, Go" {
		t.Errorf("Greeting() = %q", got)
	}
}
`

// LefthookConfig runs the formatter on commit, the linter on push, and checks commit
// messages when versioning is on (__COMMIT_MSG__). The git hooks run the pinned
// lefthook through go tool, as there's no lefthook on PATH to find.
const LefthookConfig = `lefthook: go tool -modfile=.forgego/lefthook/go.mod lefthook

pre-commit:
  commands:
__PRE_COMMIT__
pre-push:
  commands:
__PRE_PUSH__
__COMMIT_MSG__`

// LefthookGoPreCommit formats the staged Go files.
const LefthookGoPreCommit = `    format:
      glob: "*.go"
      run: go tool -modfile=.forgego/golangci-lint/go.mod golangci-lint fmt {staged_files}
      stage_fixed: true
    sync-check:
      run: go tool -modfile=.forgego/task/go.mod task sync-check
`

// LefthookGoPrePush lints the whole module before it leaves the machine.
const LefthookGoPrePush = `    lint:
      glob: "*.go"
      run: go tool -modfile=.forgego/golangci-lint/go.mod golangci-lint run
`

// LefthookWebsitePreCommit has no Go to format, so only checks forgego's files.
const LefthookWebsitePreCommit = `    sync-check:
      run: go tool -modfile=.forgego/task/go.mod task sync-check
`

// LefthookWebsitePrePush builds the site, which fails on broken templates and links.
const LefthookWebsitePrePush = `    build:
      run: go tool -modfile=.forgego/task/go.mod task build
`

// LefthookCommitMsg checks each commit message against Conventional Commits.
const LefthookCommitMsg = `commit-msg:
  commands:
    conventional-commit:
      run: go tool -modfile=.forgego/task/go.mod task commit-msg -- {1}
`

// ReleaseConfig configures release-please for a Go module: versions are git tags, so
// there's no file to bump.
const ReleaseConfig = `{
  "packages": {
    ".": {
      "release-type": "go"
    }
  }
}
`

// ReleaseManifest starts the version history at 0.0.0: the first release is v0.1.0
// for a feat, or v0.0.1 if it only has fixes.
const ReleaseManifest = `{
  ".": "0.0.0"
}
`
