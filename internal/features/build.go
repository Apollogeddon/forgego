package features

import (
	"strings"

	"github.com/apollogeddon/forgego/internal/config"
	"github.com/apollogeddon/forgego/internal/console"
	"github.com/apollogeddon/forgego/internal/taskfile"
	"github.com/apollogeddon/forgego/internal/templates"
)

// Build adds the build tasks: a binary and its release config for a backend, a compile
// check for a library, and the Hugo site for a website.
type Build struct{}

func (Build) Name() string               { return "build" }
func (Build) ShouldRun(config.Init) bool { return true }
func (Build) Cleanup(*Context)           {}

func (Build) Apply(ctx *Context) bool {
	switch {
	case ctx.Cfg.IsWebsite():
		return buildWebsite(ctx)
	case ctx.Cfg.IsLibrary():
		ctx.Tasks.Add(taskfile.Task{Name: "build", Desc: "Compile every package", Cmds: []string{"go build ./..."}})
		return true
	default:
		return buildBackend(ctx)
	}
}

func buildBackend(ctx *Context) bool {
	name := ctx.Project.Name
	values := map[string]string{"NAME": name, "SLUG": ctx.Project.Slug, "NFPMS": ""}
	if ctx.Cfg.Debian {
		values["NFPMS"] = templates.Render(templates.GoreleaserNfpms, values)
	}
	ok := CreateFile(ctx, ".goreleaser.yaml", templates.Render(templates.Goreleaser, values))
	if existing, err := ctx.FS.ReadFile(ctx.path(".goreleaser.yaml")); err == nil && ctx.Cfg.Debian &&
		!ctx.Cfg.Force && !strings.Contains(existing, "nfpms:") {
		console.Warn(".goreleaser.yaml builds no .deb: re-run with --force, or add its nfpms section yourself")
	}

	ctx.Tasks.Add(taskfile.Task{
		Name: "build",
		Desc: "Build the binary into dist/",
		Cmds: []string{`go build -trimpath -ldflags "-X main.version={{.VERSION}}" -o dist/` + name + `{{exeExt}} ./cmd/` + name},
	})
	ctx.Tasks.Var("VERSION", "dev")
	ctx.Tasks.Add(taskfile.Task{Name: "start", Desc: "Run the service", Cmds: []string{"go run ./cmd/" + name + " {{.CLI_ARGS}}"}})
	ctx.Tasks.Add(taskfile.Task{
		Name: "release:snapshot",
		Desc: "Build every release binary" + debSuffix(ctx.Cfg) + " into dist/ without publishing",
		Cmds: []string{"go run github.com/goreleaser/goreleaser/v2@" + templates.GoreleaserVersion + " release --snapshot --clean"},
	})
	return ok
}

func debSuffix(cfg config.Init) string {
	if cfg.Debian {
		return " and the .deb"
	}
	return ""
}

func buildWebsite(ctx *Context) bool {
	values := map[string]string{"NAME": ctx.Project.Name}
	ok := useTool(ctx, templates.Hugo)
	ok = CreateFile(ctx, "hugo.toml", templates.Render(templates.HugoConfig, values)) && ok
	ok = CreateIfMissing(ctx, "content/_index.md", templates.Render(templates.SiteHome, values)) && ok
	ok = CreateIfMissing(ctx, "content/docs/_index.md", templates.Render(templates.SiteDocsIndex, values)) && ok

	hugo := ref(templates.Hugo)
	ctx.Tasks.Add(taskfile.Task{Name: "dev", Desc: "Serve the site with live reload", Cmds: []string{hugo + " server"}})
	ctx.Tasks.Add(taskfile.Task{Name: "build", Desc: "Build the site into public/", Cmds: []string{hugo + " --gc --minify"}})
	return ok
}

// goMinor is the Go image tag for a version: 1.27.1 and 1.27 both give 1.27.
func goMinor(version string) string {
	parts := strings.SplitN(version, ".", 3)
	if len(parts) < 2 {
		return version
	}
	return parts[0] + "." + parts[1]
}
