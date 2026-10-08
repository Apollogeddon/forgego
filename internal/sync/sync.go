// Package sync refreshes the files forgego manages in a project, or with check reports
// which ones have drifted from the running forgego.
package sync

import (
	"path/filepath"

	"github.com/apollogeddon/forgego/internal/console"
	"github.com/apollogeddon/forgego/internal/fsys"
	"github.com/apollogeddon/forgego/internal/golangci"
	"github.com/apollogeddon/forgego/internal/gomod"
	"github.com/apollogeddon/forgego/internal/templates"
	"github.com/apollogeddon/forgego/internal/version"
	"github.com/apollogeddon/forgego/internal/yamlx"
)

// TaskfileVar is the Taskfile var that runs this forgego version.
const TaskfileVar = "FORGEGO"

type managed struct {
	rel, content string
}

// Expected lists every managed file present in dir with the content this forgego writes.
// Tool modules are only refreshed when the project already has them.
func expected(fs fsys.FS, dir string) ([]managed, error) {
	var files []managed
	for _, tool := range templates.Tools {
		if fs.Exists(filepath.Join(dir, tool.ModPath())) {
			files = append(files, managed{tool.ModPath(), tool.ModFile()}, managed{tool.SumPath(), tool.SumFile()})
		}
	}
	var pinned []templates.Tool
	for _, tool := range templates.Tools {
		if fs.Exists(filepath.Join(dir, tool.ModPath())) {
			pinned = append(pinned, tool)
		}
	}
	if gomodContent, err := fs.ReadFile(filepath.Join(dir, "go.mod")); err == nil {
		updated, _, err := gomod.EnsureToolchain(gomodContent, gomod.ToolsGo(pinned))
		if err != nil {
			return nil, err
		}
		files = append(files, managed{"go.mod", updated})
	}

	local, err := fs.ReadFile(filepath.Join(dir, golangci.LocalPath))
	switch {
	case err == nil:
		config, err := golangci.Render(local)
		if err != nil {
			return nil, err
		}
		files = append(files, managed{golangci.BasePath, golangci.Base()}, managed{golangci.ConfigPath, config})
	case !fsys.IsNotExist(err):
		return nil, err
	}
	return files, nil
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
	taskfile, taskfileDrift, err := forgegoVar(fs, dir)
	if err != nil {
		console.Err("Failed to read Taskfile.yml: %v", err)
		return 1
	}

	if check {
		for _, f := range drifted {
			console.Warn("%s is out of date", f.rel)
		}
		if taskfileDrift {
			console.Warn("Taskfile.yml's %s var doesn't run forgego %s", TaskfileVar, version.Current())
		}
		if n := len(drifted) + boolInt(taskfileDrift); n > 0 {
			console.Err("%d managed file(s) out of date - run `task sync` to refresh", n)
			return 1
		}
		console.OK("All managed files are up to date")
		return 0
	}

	if len(drifted) == 0 && !taskfileDrift {
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
		console.OK("Updated Taskfile.yml's %s var to forgego %s", TaskfileVar, version.Current())
	}
	return 0
}

// forgegoVar returns the Taskfile with its FORGEGO var pointing at this forgego, and
// whether that differs from the file on disk. A Taskfile without the var is left alone.
func forgegoVar(fs fsys.FS, dir string) (string, bool, error) {
	content, err := fs.ReadFile(filepath.Join(dir, "Taskfile.yml"))
	if fsys.IsNotExist(err) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	root, err := yamlx.Parse(content)
	if err != nil {
		return "", false, err
	}
	vars := yamlx.Get(root, "vars")
	if vars == nil {
		return "", false, nil
	}
	current := yamlx.Get(vars, TaskfileVar)
	want := version.RunCommand()
	// a project that runs forgego some other way, such as a local build, keeps it
	if current == nil || current.Value == want || !version.IsPublishedRun(current.Value) {
		return "", false, nil
	}
	current.Value = want
	out, err := yamlx.Encode(root)
	return out, true, err
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
