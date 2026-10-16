// Package sync refreshes the files forgego manages in a project, or with check reports
// which ones have drifted from the running forgego.
package sync

import (
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"

	"github.com/apollogeddon/forgego/v2/internal/console"
	"github.com/apollogeddon/forgego/v2/internal/fsys"
	"github.com/apollogeddon/forgego/v2/internal/golangci"
	"github.com/apollogeddon/forgego/v2/internal/gomod"
	"github.com/apollogeddon/forgego/v2/internal/templates"
	"github.com/apollogeddon/forgego/v2/internal/version"
	"github.com/apollogeddon/forgego/v2/internal/yamlx"
)

// TaskfileVar is the Taskfile var that runs this forgego version.
const TaskfileVar = "FORGEGO"

const (
	dependabotPath = ".github/dependabot.yml"
	legacyAllow    = `dependency-name: "` + version.LegacyModule + `"`
	currentAllow   = `dependency-name: "` + version.Module + `"`
)

type managed struct {
	rel, content string
}

// Expected lists every managed file present in dir with the content this forgego writes.
// Tool modules are only refreshed when the project already has them, at either path.
func expected(fs fsys.FS, dir string) ([]managed, error) {
	var files []managed
	var pinned []templates.Tool
	for _, tool := range templates.Tools {
		if fs.Exists(filepath.Join(dir, tool.ModPath())) || fs.Exists(filepath.Join(dir, tool.LegacyModPath())) {
			pinned = append(pinned, tool)
			files = append(files, managed{tool.ModPath(), tool.ModFile()}, managed{tool.SumPath(), tool.SumFile()})
		}
	}
	if content, err := fs.ReadFile(filepath.Join(dir, "lefthook.yml")); err == nil {
		files = append(files, managed{"lefthook.yml", movePins(content)})
	} else if !fsys.IsNotExist(err) {
		return nil, err
	}
	if gomodContent, err := fs.ReadFile(filepath.Join(dir, "go.mod")); err == nil {
		updated, _, err := gomod.EnsureToolchain(gomodContent, gomod.ToolsGo(pinned))
		if err != nil {
			return nil, err
		}
		files = append(files, managed{"go.mod", updated})
	}
	// Dependabot proposes forgego by its module path, which v2 changed
	if content, err := fs.ReadFile(filepath.Join(dir, dependabotPath)); err == nil && strings.Contains(content, legacyAllow) {
		files = append(files, managed{dependabotPath, strings.ReplaceAll(content, legacyAllow, currentAllow)})
	} else if err != nil && !fsys.IsNotExist(err) {
		return nil, err
	}
	// a website is the project that pins Hugo
	hugo := templates.Hugo
	if fs.Exists(filepath.Join(dir, hugo.ModPath())) || fs.Exists(filepath.Join(dir, hugo.LegacyModPath())) {
		for _, f := range templates.SiteFiles() {
			files = append(files, managed{f.Path, f.Content})
		}
	}

	// The project lints while forgego's base config is there; .golangci.local.yml is the
	// project's own and stays when linting is switched off.
	if !fs.Exists(filepath.Join(dir, golangci.BasePath)) {
		return files, nil
	}
	local, err := fs.ReadFile(filepath.Join(dir, golangci.LocalPath))
	if fsys.IsNotExist(err) {
		local, err = golangci.LocalStarter, nil
	}
	if err != nil {
		return nil, err
	}
	config, err := golangci.Render(local)
	if err != nil {
		return nil, err
	}
	return append(files, managed{golangci.BasePath, golangci.Base()}, managed{golangci.ConfigPath, config}), nil
}

// Run refreshes the managed files in dir, or with check only reports drift. It returns
// the process exit code.
func Run(fs fsys.FS, dir string, check bool) int {
	files, err := expected(fs, dir)
	if err != nil {
		console.Err("Failed to read the project's config: %v", err)
		return 1
	}
	var drifted []managed
	for _, f := range files {
		current, err := fs.ReadFile(filepath.Join(dir, f.rel))
		if err != nil || current != f.content {
			drifted = append(drifted, f)
		}
	}
	taskfile, taskfileDrift, err := updateTaskfile(fs, dir)
	if err != nil {
		console.Err("Failed to read Taskfile.yml: %v", err)
		return 1
	}
	var legacy []string
	for _, tool := range templates.Tools {
		for _, rel := range []string{tool.LegacyModPath(), tool.LegacySumPath()} {
			if fs.Exists(filepath.Join(dir, rel)) {
				legacy = append(legacy, rel)
			}
		}
	}

	if check {
		missingPin := 0
		for _, f := range drifted {
			console.Warn("%s is out of date", f.rel)
		}
		if taskfileDrift {
			console.Warn("Taskfile.yml is out of date: its tool or %s vars don't match forgego %s", TaskfileVar, version.Current())
		}
		if selfPinned(fs, dir) && !fs.Exists(filepath.Join(dir, templates.SelfModPath)) {
			console.Warn("%s is missing", templates.SelfModPath)
			missingPin = 1
		}
		if version.Released() && legacyPin(fs, dir) {
			console.Warn("%s pins forgego at %s, which v2 moved to %s", templates.SelfModPath, version.LegacyModule, version.Module)
			missingPin = 1
		}
		for _, rel := range legacy {
			console.Warn("%s has moved under .forgego/<tool>/", rel)
		}
		if n := len(drifted) + boolInt(taskfileDrift) + len(legacy) + missingPin; n > 0 {
			console.Err("%d managed file(s) out of date - run `task sync` to refresh", n)
			return 1
		}
		console.OK("All managed files are up to date")
		return 0
	}

	// a Taskfile moving to the pin, or one already on it, needs the pin to exist
	if (taskfileDrift && strings.Contains(taskfile, templates.SelfCommand)) || selfPinned(fs, dir) {
		if !PinSelf(fs, dir, false) {
			return 1
		}
	}
	if len(drifted) == 0 && !taskfileDrift && len(legacy) == 0 {
		console.Info("All managed files already up to date")
		return 0
	}
	for _, f := range drifted {
		if err := fs.WriteFile(filepath.Join(dir, f.rel), f.content); err != nil {
			console.Err("Failed to refresh %s: %v", f.rel, err)
			return 1
		}
		console.OK("Refreshed %s", f.rel)
	}
	if taskfileDrift {
		if err := fs.WriteFile(filepath.Join(dir, "Taskfile.yml"), taskfile); err != nil {
			console.Err("Failed to update Taskfile.yml: %v", err)
			return 1
		}
		console.OK("Updated Taskfile.yml for forgego %s", version.Current())
	}
	for _, rel := range legacy {
		if err := fs.Remove(filepath.Join(dir, rel)); err != nil {
			console.Err("Failed to remove %s: %v", rel, err)
			return 1
		}
		console.OK("Removed %s, now under .forgego/<tool>/", rel)
		if rel == templates.Lefthook.LegacyModPath() {
			// installed git hooks name the command they run lefthook with
			console.Info("Run `task hooks` to reinstall the git hooks at lefthook's new path")
		}
	}
	return 0
}

// movePins points every command that runs a tool from an earlier forgego's
// .forgego/<tool>.mod at .forgego/<tool>/go.mod.
func movePins(content string) string {
	for _, tool := range templates.Tools {
		content = strings.ReplaceAll(content, "-modfile="+tool.LegacyModPath(), "-modfile="+tool.ModPath())
	}
	return content
}

// RunCommand is how the Taskfile runs forgego: from its pin in .forgego/forgego/ for a
// release, else with go run, as a build that isn't one has no version to pin.
func RunCommand() string {
	if version.Released() {
		return templates.SelfCommand
	}
	return version.RunCommand()
}

// PinSelf pins this forgego in dir when nothing pins it yet, then writes the pin's go.sum
// if it's missing. Dependabot owns the pinned version from then on, so an existing pin
// is only rewritten when it pins v1's module path, which no v2 release can be required at.
func PinSelf(fs fsys.FS, dir string, dryRun bool) bool {
	if !version.Released() {
		return true
	}
	mod := filepath.Join(dir, templates.SelfModPath)
	if legacyPin(fs, dir) {
		if dryRun {
			console.Dry("Would move %s to forgego %s at %s", templates.SelfModPath, version.Current(), version.Module)
			return true
		}
		// the old go.sum holds v1's checksums, so it goes and Tidy writes the new one
		if err := fs.Remove(mod); err != nil {
			console.Err("Failed to remove %s: %v", templates.SelfModPath, err)
			return false
		}
		if fs.Exists(filepath.Join(dir, templates.SelfSumPath)) {
			if err := fs.Remove(filepath.Join(dir, templates.SelfSumPath)); err != nil {
				console.Err("Failed to remove %s: %v", templates.SelfSumPath, err)
				return false
			}
		}
	}
	if !fs.Exists(mod) {
		if dryRun {
			console.Dry("Would pin forgego %s in %s", version.Current(), templates.SelfModPath)
			return true
		}
		if err := fs.WriteFile(mod, templates.SelfModFile(version.Current())); err != nil {
			console.Err("Failed to write %s: %v", templates.SelfModPath, err)
			return false
		}
		console.OK("Pinned forgego %s in %s", version.Current(), templates.SelfModPath)
	}
	if fs.Exists(filepath.Join(dir, templates.SelfSumPath)) {
		return true
	}
	if dryRun {
		console.Dry("Would run go mod tidy in %s", filepath.Dir(templates.SelfModPath))
		return true
	}
	if err := gomod.Tidy(dir, templates.SelfModPath); err != nil {
		console.Err("Failed to write %s: %v", templates.SelfSumPath, err)
		return false
	}
	console.OK("Wrote %s", templates.SelfSumPath)
	return true
}

// legacyPin reports whether dir's pin requires forgego at v1's module path.
func legacyPin(fs fsys.FS, dir string) bool {
	content, err := fs.ReadFile(filepath.Join(dir, templates.SelfModPath))
	if err != nil {
		return false
	}
	f, err := modfile.ParseLax(templates.SelfModPath, []byte(content), nil)
	if err != nil {
		return false
	}
	for _, r := range f.Require {
		if r.Mod.Path == version.LegacyModule {
			return true
		}
	}
	return false
}

// selfPinned reports whether the Taskfile runs forgego from its pin.
func selfPinned(fs fsys.FS, dir string) bool {
	content, err := fs.ReadFile(filepath.Join(dir, "Taskfile.yml"))
	return err == nil && strings.Contains(content, templates.SelfCommand)
}

// updateTaskfile returns the Taskfile with its tool vars on the current pin paths and its
// FORGEGO var pointing at this forgego, and whether that differs from the file on disk.
// A project that runs forgego some other way, such as a local build, keeps it.
func updateTaskfile(fs fsys.FS, dir string) (string, bool, error) {
	original, err := fs.ReadFile(filepath.Join(dir, "Taskfile.yml"))
	if fsys.IsNotExist(err) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	content := movePins(original)
	root, err := yamlx.Parse(content)
	if err != nil {
		return "", false, err
	}
	if vars := yamlx.Get(root, "vars"); vars != nil {
		current := yamlx.Get(vars, TaskfileVar)
		want := RunCommand()
		if current != nil && current.Value != want && version.IsPublishedRun(current.Value) {
			current.Value = want
			if content, err = yamlx.Encode(root); err != nil {
				return "", false, err
			}
		}
	}
	return content, content != original, nil
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
