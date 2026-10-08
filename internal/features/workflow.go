package features

import (
	"github.com/apollogeddon/forgego/internal/config"
	"github.com/apollogeddon/forgego/internal/templates"
)

// Workflow generates the project's CI, which calls forgego's reusable workflows.
type Workflow struct{}

func (Workflow) Name() string               { return "workflows" }
func (Workflow) ShouldRun(config.Init) bool { return true }
func (Workflow) Cleanup(*Context)           {}

func (Workflow) Apply(ctx *Context) bool {
	return CreateFile(ctx, ".github/workflows/index.yml", templates.RenderWorkflow(templates.WorkflowOptions{
		Mode:       string(ctx.Cfg.Mode),
		Go:         ctx.Cfg.Go,
		Docker:     ctx.Cfg.Docker,
		Testing:    ctx.Cfg.Testing,
		Versioning: ctx.Cfg.Versioning,
	}))
}
