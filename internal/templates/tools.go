package templates

import (
	"embed"

	"golang.org/x/mod/modfile"
)

//go:generate go run ./gen

// Each tool lives in its own module, copied from tools/<name>/ by `go generate`, so no
// two tools' dependencies are ever resolved together, and none touch the project's go.mod.
//
//go:embed tools configs
var toolFiles embed.FS

// Tool is a command forgego pins and runs with `go tool -modfile=.forgego/<name>/go.mod`.
type Tool struct {
	Name    string // directory under .forgego/, and the command
	Package string // the package `go tool` builds
}

var (
	Task         = Tool{"task", "github.com/go-task/task/v3/cmd/task"}
	Lefthook     = Tool{"lefthook", "github.com/evilmartians/lefthook/v2"}
	GolangciLint = Tool{"golangci-lint", "github.com/golangci/golangci-lint/v2/cmd/golangci-lint"}
	Gotestsum    = Tool{"gotestsum", "gotest.tools/gotestsum"}
	Govulncheck  = Tool{"govulncheck", "golang.org/x/vuln/cmd/govulncheck"}
	Hugo         = Tool{"hugo", "github.com/gohugoio/hugo"}
)

// Tools is every tool forgego knows, in a stable order.
var Tools = []Tool{Task, Lefthook, GolangciLint, Gotestsum, Govulncheck, Hugo}

// ModPath is where the tool's module file lives in a generated project. Naming it go.mod
// keeps its go.sum out of secret scanners, which skip go.sum but flag its h1: hashes
// under any other name.
func (t Tool) ModPath() string { return ".forgego/" + t.Name + "/go.mod" }

// SumPath is the checksum file `go tool -modfile` reads next to ModPath.
func (t Tool) SumPath() string { return ".forgego/" + t.Name + "/go.sum" }

// LegacyModPath and LegacySumPath are where forgego pinned the tool before it had a
// directory of its own; `forgego sync` moves them.
func (t Tool) LegacyModPath() string { return ".forgego/" + t.Name + ".mod" }
func (t Tool) LegacySumPath() string { return ".forgego/" + t.Name + ".sum" }

// Command is how a Taskfile or hook runs the tool.
func (t Tool) Command() string {
	return "go tool -modfile=" + t.ModPath() + " " + t.Name
}

// ModFile is the pinned go.mod for the tool.
func (t Tool) ModFile() string { return mustRead("tools/" + t.Name + ".mod") }

// SumFile is the pinned go.sum for the tool.
func (t Tool) SumFile() string { return mustRead("tools/" + t.Name + ".sum") }

func mustRead(name string) string {
	b, err := toolFiles.ReadFile(name)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// Config returns one of forgego's embedded base configs, such as golangci.yml.
func Config(name string) string { return mustRead("configs/" + name) }

// GoVersion is the Go the tool's module needs, from its go line, such as 1.27.0.
func (t Tool) GoVersion() string {
	f, err := modfile.ParseLax(t.ModPath(), []byte(t.ModFile()), nil)
	if err != nil || f.Go == nil {
		panic("no go line in " + t.ModPath())
	}
	return f.Go.Version
}
