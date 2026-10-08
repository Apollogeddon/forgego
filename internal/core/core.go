// Package core runs `forgego init`: it detects the project, runs the feature pipeline
// and writes the Taskfile the features built up.
package core

import (
	"path/filepath"
	"strings"

	"github.com/apollogeddon/forgego/internal/config"
	"github.com/apollogeddon/forgego/internal/console"
	"github.com/apollogeddon/forgego/internal/features"
	"github.com/apollogeddon/forgego/internal/fsys"
	"github.com/apollogeddon/forgego/internal/project"
	"github.com/apollogeddon/forgego/internal/taskfile"
)

// ExitInvalidConfig is returned when the options can't be combined.
const ExitInvalidConfig = 2

// Init scaffolds cfg.Target and returns the process exit code.
func Init(cfg config.Init, fs fsys.FS) int {
	if errs := cfg.Validate(); len(errs) > 0 {
		for _, e := range errs {
			console.Err("%s", e)
		}
		return ExitInvalidConfig
	}

	header := "forgego init — mode=" + string(cfg.Mode)
	if cfg.DryRun {
		header += " (dry run)"
	}
	console.Info("%s", header)
	if cfg.DryRun {
		console.Warn("DRY RUN MODE — no files will be written")
	}

	var active []string
	for _, f := range features.Pipeline {
		if f.ShouldRun(cfg) {
			active = append(active, f.Name())
		}
	}
	console.Info("Active features: %s", strings.Join(active, ", "))

	info, err := project.Detect(fs, cfg.Target)
	if err != nil {
		console.Err("Failed to read go.mod: %v", err)
		return 1
	}
	console.Info("Module: %s", info.Module)

	ctx := &features.Context{Cfg: cfg, FS: fs, Project: info, Tasks: taskfile.NewBuilder()}
	failed := false
	for _, f := range features.Pipeline {
		if f.ShouldRun(cfg) && !f.Apply(ctx) {
			failed = true
		}
	}

	if !writeTaskfile(ctx) {
		failed = true
	}

	for _, f := range features.Pipeline {
		f.Cleanup(ctx)
	}

	if failed {
		console.Err("forgego init finished with errors")
		return 1
	}
	console.OK("forgego init complete")
	next := "go tool -modfile=.forgego/task.mod task --list"
	if cfg.Linting {
		next = "go tool -modfile=.forgego/task.mod task hooks"
	}
	console.Info("Next steps: go mod tidy && %s", next)
	return 0
}

// writeTaskfile merges the features' tasks into Taskfile.yml, keeping the project's
// own tasks and vars unless --force is set.
func writeTaskfile(ctx *features.Context) bool {
	path := filepath.Join(ctx.Cfg.Target, "Taskfile.yml")
	existing, err := ctx.FS.ReadFile(path)
	if err != nil && !fsys.IsNotExist(err) {
		console.Err("Failed to read Taskfile.yml: %v", err)
		return false
	}
	content, changed, err := ctx.Tasks.Render(existing, ctx.Cfg.Force)
	if err != nil {
		console.Err("Failed to update Taskfile.yml: %v", err)
		return false
	}
	if len(changed) == 0 {
		console.Info("Taskfile.yml already has every task. Skipping.")
		return true
	}
	if ctx.Cfg.DryRun {
		console.Dry("Would update Taskfile.yml: %s", strings.Join(changed, ", "))
		return true
	}
	if err := ctx.FS.WriteFile(path, content); err != nil {
		console.Err("Failed to write Taskfile.yml: %v", err)
		return false
	}
	console.OK("Updated Taskfile.yml: %s", strings.Join(changed, ", "))
	return true
}
