package features

import (
	"github.com/apollogeddon/forgego/internal/config"
	"github.com/apollogeddon/forgego/internal/console"
	"github.com/apollogeddon/forgego/internal/templates"
)

// Workflow generates the project's CI, which calls forgego's reusable workflows.
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
	return CreateFile(ctx, ".github/workflows/index.yml", templates.RenderWorkflow(templates.WorkflowOptions{
		Mode:       string(ctx.Cfg.Mode),
		Docker:     ctx.Cfg.Docker,
		Testing:    ctx.Cfg.Testing,
		Linting:    ctx.Cfg.Linting,
		Versioning: ctx.Cfg.Versioning,
	})) && ok
}
