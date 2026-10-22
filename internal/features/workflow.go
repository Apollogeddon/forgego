package features

import (
	"strings"

	"github.com/apollogeddon/forgego/v2/internal/config"
	"github.com/apollogeddon/forgego/v2/internal/console"
	"github.com/apollogeddon/forgego/v2/internal/templates"
)

// Workflow generates the project's CI, which calls forgego's pipelines.
type Workflow struct{}

func (Workflow) Name() string               { return "workflows" }
func (Workflow) ShouldRun(config.Init) bool { return true }
func (Workflow) Cleanup(*Context)           {}

func (Workflow) Apply(ctx *Context) bool {
	ok := CreateFile(ctx, ".github/dependabot.yml", templates.Dependabot(ctx.Cfg.Docker))
	if owner := ctx.Project.GitHubOwner(); owner != "" {
		ok = CreateFile(ctx, ".github/CODEOWNERS", templates.Render(templates.Codeowners, map[string]string{"OWNER": owner})) && ok
	} else {
		console.Info("No GitHub module path or remote yet, so no .github/CODEOWNERS. Run init again once the project has one.")
	}
	// only once .index.yml is written: a kept or failed one would leave the project with no CI, or the old one
	replacing := !ctx.FS.Exists(ctx.path(templates.IndexWorkflow)) || ctx.Cfg.Force
	written := CreateFile(ctx, templates.IndexWorkflow, templates.RenderWorkflow(templates.WorkflowOptions{
		Mode:       string(ctx.Cfg.Mode),
		Docker:     ctx.Cfg.Docker,
		Testing:    ctx.Cfg.Testing,
		Linting:    ctx.Cfg.Linting,
		Versioning: ctx.Cfg.Versioning,
		Windows:    ctx.Cfg.Windows,
		MacOS:      ctx.Cfg.MacOS,
	}))
	if written && replacing {
		removeOldWorkflows(ctx)
	}
	return written && ok
}

// Before .index.yml, forgego generated index.yml; left in place, it runs every pipeline twice, or fails
// calling ones that are gone. One exactly as an earlier forgego wrote it goes; one calling forgego's pipelines
// that's been changed goes only with --force, as RemoveFile does; one calling none of them is the project's own.
func removeOldWorkflows(ctx *Context) {
	for file := range templates.LegacyWorkflows {
		content, err := ctx.FS.ReadFile(ctx.path(file))
		if err != nil {
			continue
		}
		switch {
		case templates.IsLegacyWorkflow(file, content):
			force := *ctx
			force.Cfg.Force = true
			RemoveFile(&force, file)
		case templates.CallsForgego(content):
			if !ctx.Cfg.Force {
				console.Warn("%s calls forgego's old pipelines and has been changed since init wrote it.", file)
			}
			RemoveFile(ctx, file)
		case strings.HasSuffix(file, "/index.yml"):
			console.Warn("Left %s, which doesn't call forgego's pipelines: check it doesn't repeat %s's jobs.", file, templates.IndexWorkflow)
		}
	}
}
