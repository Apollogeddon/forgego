package templates

import (
	"embed"
	"strings"

	"golang.org/x/mod/modfile"
)

//go:generate go run ./gen

// Each tool lives in its own module, copied from tools/<name>/ by `go generate`, so no
// two tools' dependencies are ever resolved together, and none touch the project's go.mod.
//
//go:embed tools configs
var toolFiles embed.FS

// Tool is a command forgego pins and runs with `go tool -modfile=.forgego/<name>.mod`.
type Tool struct {
	Name    string // file name under .forgego/, and the command
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

// ModPath is where the tool's module file lives in a generated project.
func (t Tool) ModPath() string { return ".forgego/" + t.Name + ".mod" }

// SumPath is the checksum file `go tool -modfile` reads next to ModPath.
func (t Tool) SumPath() string { return ".forgego/" + t.Name + ".sum" }

// Command is how a Taskfile or hook runs the tool.
func (t Tool) Command() string {
	return "go tool -modfile=" + t.ModPath() + " " + t.Name
}

// ModFile is the pinned go.mod for the tool.
func (t Tool) ModFile() string { return mustRead("tools/" + t.Name + ".mod") }

// SumFile is the pinned go.sum for the tool.
func (t Tool) SumFile() string { return mustRead("tools/" + t.Name + ".sum") }

// Version is the pinned version of the module that provides the tool: the
// longest required module path that is a prefix of the tool's package.
func (t Tool) Version() string {
	f, err := modfile.ParseLax(t.ModPath(), []byte(t.ModFile()), nil)
	if err != nil {
		panic(err)
	}
	best, version := "", ""
	for _, r := range f.Require {
		p := r.Mod.Path
		if (t.Package == p || strings.HasPrefix(t.Package, p+"/")) && len(p) > len(best) {
			best, version = p, r.Mod.Version
		}
	}
	return version
}

func mustRead(name string) string {
	b, err := toolFiles.ReadFile(name)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// Config returns one of forgego's embedded base configs, such as golangci.yml.
func Config(name string) string { return mustRead("configs/" + name) }
