package features

import (
	"github.com/apollogeddon/forgego/internal/config"
	"github.com/apollogeddon/forgego/internal/sync"
	"github.com/apollogeddon/forgego/internal/taskfile"
	"github.com/apollogeddon/forgego/internal/templates"
	"github.com/apollogeddon/forgego/internal/version"
)

// Base creates go.mod and the starter source, and pins Task, which runs everything else.
type Base struct{}

func (Base) Name() string               { return "base" }
func (Base) ShouldRun(config.Init) bool { return true }
func (Base) Cleanup(*Context)           {}

func (Base) Apply(ctx *Context) bool {
	values := map[string]string{
		"MODULE":  ctx.Project.Module,
		"NAME":    ctx.Project.Name,
		"PACKAGE": ctx.Project.Package,
		"GO":      ctx.Cfg.GoDirective(),
	}
	// go.mod is the project's own file: created once, never rewritten.
	ok := CreateIfMissing(ctx, "go.mod", templates.Render(templates.GoMod, values))

	switch {
	case ctx.Cfg.IsBackend():
		ok = starter(ctx, "cmd/"+ctx.Project.Name+"/main.go", templates.Render(templates.BackendMain, values)) && ok
	case ctx.Cfg.IsLibrary():
		ok = starter(ctx, ctx.Project.Package+".go", templates.Render(templates.LibrarySource, values)) && ok
	}

	ignore := templates.Gitignore
	if ctx.Cfg.IsWebsite() {
		ignore = templates.GitignoreWebsite
	}
	ok = CreateIfMissing(ctx, ".gitignore", ignore) && ok

	ok = useTool(ctx, templates.Task) && ok
	// Security scanning doesn't depend on linting: CI's patch job upgrades what govulncheck finds.
	if !ctx.Cfg.IsWebsite() {
		ok = useTool(ctx, templates.Govulncheck) && ok
		ctx.Tasks.Add(taskfile.Task{Name: "security", Desc: "Report known vulnerabilities the code calls", Cmds: []string{ref(templates.Govulncheck) + " ./..."}})
	}
	ctx.Tasks.Var(sync.TaskfileVar, version.RunCommand())
	ctx.Tasks.Add(taskfile.Task{
		Name: "sync",
		Desc: "Refresh the files forgego manages in .forgego/ and .golangci.yml",
		Cmds: []string{"{{." + sync.TaskfileVar + "}} sync"},
	})
	ctx.Tasks.Add(taskfile.Task{
		Name: "sync-check",
		Desc: "Fail if a file forgego manages is out of date",
		Cmds: []string{"{{." + sync.TaskfileVar + "}} sync --check"},
	})
	return ok
}

// starter creates the starter source when it's missing, and notes whether the project's
// source is still the starter.
func starter(ctx *Context, rel, content string) bool {
	ctx.Starter = isStarter(ctx, rel, content)
	return CreateIfMissing(ctx, rel, content)
}
