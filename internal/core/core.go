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
	"github.com/apollogeddon/forgego/internal/gomod"
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
	if cfg.Go == "" {
		// CI and the Docker image use the project's own Go unless --go says otherwise
		cfg.Go = config.DefaultGo
		if info.GoVersion != "" {
			cfg.Go = info.GoVersion
		}
	}

	ctx := &features.Context{Cfg: cfg, FS: fs, Project: info, Tasks: taskfile.NewBuilder()}
	failed := false
	for _, f := range features.Pipeline {
		if f.ShouldRun(cfg) && !f.Apply(ctx) {
			failed = true
		}
	}

	// cleanup first, so the tasks of a feature switched off leave the Taskfile with its files
	for _, f := range features.Pipeline {
		f.Cleanup(ctx)
	}
	if !ensureToolchain(ctx) {
		failed = true
	}
	if !writeTaskfile(ctx) {
		failed = true
	}

	if failed {
		console.Err("forgego init finished with errors")
		return 1
	}
	console.OK("forgego init complete")
	task := "go tool -modfile=.forgego/task.mod task "
	switch {
	case cfg.IsWebsite():
		// Hugo manages a site's go.mod itself; go mod tidy would drop the theme
		console.Info("Next steps: %sdev", task)
	case cfg.Linting:
		console.Info("Next steps: go mod tidy && %shooks", task)
	default:
		console.Info("Next steps: go mod tidy && %s--list", task)
	}
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

// ensureToolchain raises go.mod's toolchain line when the project's Go is older than a
// pinned tool needs: Go picks the toolchain from go.mod before reading the tool's -modfile.
func ensureToolchain(ctx *features.Context) bool {
	path := filepath.Join(ctx.Cfg.Target, "go.mod")
	content, err := ctx.FS.ReadFile(path)
	if fsys.IsNotExist(err) && ctx.Cfg.DryRun {
		return true // a new go.mod targets --go, which the tools are pinned for
	}
	if err != nil {
		console.Err("Failed to read go.mod: %v", err)
		return false
	}
	need := gomod.ToolsGo(ctx.Tools)
	updated, changed, err := gomod.EnsureToolchain(content, need)
	if err != nil {
		console.Err("Failed to update go.mod: %v", err)
		return false
	}
	if !changed {
		return true
	}
	if ctx.Cfg.DryRun {
		console.Dry("Would set go.mod's toolchain to %s, which the pinned tools need", need)
		return true
	}
	if err := ctx.FS.WriteFile(path, updated); err != nil {
		console.Err("Failed to write go.mod: %v", err)
		return false
	}
	console.OK("Set go.mod's toolchain to %s, which the pinned tools need", need)
	return true
}
