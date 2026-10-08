// Package features holds the ordered pipeline that `forgego init` runs. Each feature
// decides whether it applies, writes its files, and removes them when switched off.
package features

import (
	"path/filepath"

	"github.com/apollogeddon/forgego/internal/config"
	"github.com/apollogeddon/forgego/internal/console"
	"github.com/apollogeddon/forgego/internal/fsys"
	"github.com/apollogeddon/forgego/internal/project"
	"github.com/apollogeddon/forgego/internal/taskfile"
)

// Context is what every feature reads and writes during one run.
type Context struct {
	Cfg     config.Init
	FS      fsys.FS
	Project project.Info
	Tasks   *taskfile.Builder
}

// Feature is one step of the pipeline.
type Feature interface {
	Name() string
	ShouldRun(cfg config.Init) bool
	Apply(ctx *Context) bool
	Cleanup(ctx *Context)
}

// Pipeline is every feature, in the order they run.
var Pipeline = []Feature{
	Base{}, Linting{}, Build{}, Testing{}, Versioning{}, Docker{}, Debian{}, Workflow{},
}

func (c *Context) path(rel string) string {
	return filepath.Join(c.Cfg.Target, filepath.FromSlash(rel))
}

// CreateFile writes a config file, leaving an existing one alone unless --force is set.
func CreateFile(ctx *Context, rel, content string) bool {
	path := ctx.path(rel)
	exists := ctx.FS.Exists(path)
	if exists && !ctx.Cfg.Force {
		console.Info("%s already exists. Skipping.", rel)
		return true
	}
	if ctx.Cfg.DryRun {
		verb := "create"
		if exists {
			verb = "overwrite"
		}
		console.Dry("Would %s %s", verb, rel)
		return true
	}
	if err := ctx.FS.WriteFile(path, content); err != nil {
		console.Err("Failed to write %s: %v", rel, err)
		return false
	}
	if exists {
		console.OK("Overwrote %s", rel)
	} else {
		console.OK("Created %s", rel)
	}
	return true
}

// CreateIfMissing scaffolds the project's own source once. It never touches the file
// again, even with --force, as it's the user's code, not a config forgego manages.
func CreateIfMissing(ctx *Context, rel, content string) bool {
	path := ctx.path(rel)
	if ctx.FS.Exists(path) {
		return true
	}
	if ctx.Cfg.DryRun {
		console.Dry("Would create %s", rel)
		return true
	}
	if err := ctx.FS.WriteFile(path, content); err != nil {
		console.Err("Failed to write %s: %v", rel, err)
		return false
	}
	console.OK("Created %s", rel)
	return true
}

// WriteManaged refreshes a file forgego owns under .forgego/, ignoring --force.
func WriteManaged(ctx *Context, rel, content string) bool {
	path := ctx.path(rel)
	if ctx.Cfg.DryRun {
		console.Dry("Would refresh managed file %s", rel)
		return true
	}
	if current, err := ctx.FS.ReadFile(path); err == nil && current == content {
		return true
	}
	if err := ctx.FS.WriteFile(path, content); err != nil {
		console.Err("Failed to write %s: %v", rel, err)
		return false
	}
	console.OK("Refreshed %s", rel)
	return true
}

// RemoveFile deletes a file a disabled feature created, but only with --force.
func RemoveFile(ctx *Context, rel string) {
	path := ctx.path(rel)
	if !ctx.FS.Exists(path) {
		return
	}
	if !ctx.Cfg.Force {
		console.Info("Skipping removal of %s (use --force to delete)", rel)
		return
	}
	if ctx.Cfg.DryRun {
		console.Dry("Would remove %s", rel)
		return
	}
	if err := ctx.FS.Remove(path); err != nil {
		console.Warn("Failed to remove %s: %v", rel, err)
		return
	}
	console.OK("Removed %s", rel)
}
