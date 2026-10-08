package features

import (
	"github.com/apollogeddon/forgego/internal/config"
	"github.com/apollogeddon/forgego/internal/taskfile"
	"github.com/apollogeddon/forgego/internal/templates"
)

// Testing runs go test through gotestsum, which also writes a JUnit report for CI.
type Testing struct{}

func (Testing) Name() string { return "testing" }

// ShouldRun is false for a website: a Hugo site has no Go to test.
func (Testing) ShouldRun(cfg config.Init) bool { return cfg.Testing && !cfg.IsWebsite() }

func (Testing) Apply(ctx *Context) bool {
	values := map[string]string{"NAME": ctx.Project.Name, "PACKAGE": ctx.Project.Package}
	ok := useTool(ctx, templates.Gotestsum)
	switch {
	case !ctx.Starter:
		// the project's own source: the starter test would test code that isn't there
	case ctx.Cfg.IsBackend():
		ok = CreateIfMissing(ctx, "cmd/"+ctx.Project.Name+"/main_test.go", templates.Render(templates.BackendMainTest, values)) && ok
	case ctx.Cfg.IsLibrary():
		ok = CreateIfMissing(ctx, ctx.Project.Package+"_test.go", templates.Render(templates.LibraryTest, values)) && ok
	}
	ctx.Tasks.Add(taskfile.Task{
		Name: "test",
		Desc: "Run the tests with coverage and a JUnit report",
		Cmds: []string{ref(templates.Gotestsum) + " --junitfile junit-report.xml -- -coverprofile=coverage.out -covermode=atomic ./..."},
	})
	return ok
}

func (Testing) Cleanup(ctx *Context) {
	if !ctx.Cfg.Testing || ctx.Cfg.IsWebsite() {
		dropTool(ctx, templates.Gotestsum)
	}
}
