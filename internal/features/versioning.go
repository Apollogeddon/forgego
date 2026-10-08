package features

import (
	"github.com/apollogeddon/forgego/internal/config"
	"github.com/apollogeddon/forgego/internal/sync"
	"github.com/apollogeddon/forgego/internal/taskfile"
	"github.com/apollogeddon/forgego/internal/templates"
)

// Versioning configures release-please, which tags releases from Conventional Commits.
type Versioning struct{}

func (Versioning) Name() string                   { return "versioning" }
func (Versioning) ShouldRun(cfg config.Init) bool { return cfg.Versioning }

func (Versioning) Apply(ctx *Context) bool {
	ok := CreateFile(ctx, ".github/release.json", templates.ReleaseConfig)
	ok = CreateFile(ctx, ".github/.release.json", templates.ReleaseManifest) && ok
	ctx.Tasks.Add(taskfile.Task{
		Name: "commit-msg",
		Desc: "Check a commit message file against Conventional Commits",
		Cmds: []string{"{{." + sync.TaskfileVar + "}} commit-msg {{.CLI_ARGS}}"},
	})
	return ok
}

func (Versioning) Cleanup(ctx *Context) {
	if !ctx.Cfg.Versioning {
		RemoveFile(ctx, ".github/release.json")
		RemoveFile(ctx, ".github/.release.json")
	}
}
