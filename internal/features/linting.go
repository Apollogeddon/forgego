package features

import (
	"github.com/apollogeddon/forgego/internal/config"
	"github.com/apollogeddon/forgego/internal/console"
	"github.com/apollogeddon/forgego/internal/golangci"
	"github.com/apollogeddon/forgego/internal/taskfile"
	"github.com/apollogeddon/forgego/internal/templates"
)

// Linting sets up golangci-lint, govulncheck and the lefthook git hooks.
type Linting struct{}

func (Linting) Name() string                   { return "linting" }
func (Linting) ShouldRun(cfg config.Init) bool { return cfg.Linting }

func (Linting) Apply(ctx *Context) bool {
	ok := useTool(ctx, templates.Lefthook)
	ctx.Tasks.Add(taskfile.Task{Name: "hooks", Desc: "Install the git hooks", Cmds: []string{ref(templates.Lefthook) + " install"}})

	preCommit, prePush := templates.LefthookWebsitePreCommit, templates.LefthookWebsitePrePush
	if !ctx.Cfg.IsWebsite() {
		preCommit, prePush = templates.LefthookGoPreCommit, templates.LefthookGoPrePush
		ok = lintGo(ctx) && ok
	}
	commitMsg := ""
	if ctx.Cfg.Versioning {
		commitMsg = templates.LefthookCommitMsg
	}
	lefthook := templates.Render(templates.LefthookConfig, map[string]string{
		"PRE_COMMIT": trimNewline(preCommit),
		"PRE_PUSH":   trimNewline(prePush),
		"COMMIT_MSG": commitMsg,
	})
	return CreateFile(ctx, "lefthook.yml", lefthook) && ok
}

func lintGo(ctx *Context) bool {
	ok := useTool(ctx, templates.GolangciLint)
	ok = WriteManaged(ctx, golangci.BasePath, golangci.Base()) && ok
	ok = CreateIfMissing(ctx, golangci.LocalPath, golangci.LocalStarter) && ok

	// .golangci.yml is generated from the base and local configs, so it always refreshes.
	local := golangci.LocalStarter
	if content, err := ctx.FS.ReadFile(ctx.path(golangci.LocalPath)); err == nil {
		local = content
	}
	merged, err := golangci.Render(local)
	if err != nil {
		console.Err("Failed to merge %s into the base config: %v", golangci.LocalPath, err)
		return false
	}
	ok = WriteManaged(ctx, golangci.ConfigPath, merged) && ok

	lint := ref(templates.GolangciLint)
	ctx.Tasks.Add(taskfile.Task{Name: "lint", Desc: "Format, then lint and fix what can be fixed", Cmds: []string{lint + " fmt", lint + " run --fix"}})
	ctx.Tasks.Add(taskfile.Task{Name: "format", Desc: "Format the code", Cmds: []string{lint + " fmt"}})
	return ok
}

func (Linting) Cleanup(ctx *Context) {
	if ctx.Cfg.Linting {
		return
	}
	RemoveTasks(ctx, "hooks", "lint", "format", "LEFTHOOK", "GOLANGCI_LINT")
	RemoveFile(ctx, "lefthook.yml")
	RemoveFile(ctx, golangci.ConfigPath)
	RemoveFile(ctx, golangci.BasePath)
	dropTool(ctx, templates.Lefthook)
	dropTool(ctx, templates.GolangciLint)
}

func trimNewline(s string) string {
	if len(s) > 0 && s[len(s)-1] == '\n' {
		return s[:len(s)-1]
	}
	return s
}
