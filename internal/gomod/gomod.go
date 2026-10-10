// Package gomod keeps a project's go.mod able to run the tools forgego pins.
package gomod

import (
	"fmt"
	"go/version"
	"os/exec"
	"strings"

	"golang.org/x/mod/modfile"

	"github.com/apollogeddon/forgego/internal/templates"
)

// ToolsGo is the newest Go any of tools needs, such as go1.27.0; empty for none.
func ToolsGo(tools []templates.Tool) string {
	newest := ""
	for _, tool := range tools {
		if v := "go" + tool.GoVersion(); version.Compare(v, newest) > 0 {
			newest = v
		}
	}
	return newest
}

// EnsureToolchain raises go.mod's toolchain line to need when neither its go nor its
// toolchain line already reaches it. Go picks the toolchain from the project's go.mod
// before it reads a tool's -modfile, so without this a project on an older Go can't run
// its tools. The go line, which modules that depend on the project see, is left alone.
// It returns the new content and whether it changed.
func EnsureToolchain(content, need string) (string, bool, error) {
	if need == "" {
		return content, false, nil
	}
	f, err := modfile.Parse("go.mod", []byte(content), nil)
	if err != nil {
		return "", false, err
	}
	effective := ""
	if f.Go != nil {
		effective = "go" + f.Go.Version
	}
	if f.Toolchain != nil && version.Compare(f.Toolchain.Name, effective) > 0 {
		effective = f.Toolchain.Name
	}
	if version.Compare(effective, need) >= 0 {
		return content, false, nil
	}
	if err := f.AddToolchainStmt(need); err != nil {
		return "", false, err
	}
	out, err := f.Format()
	if err != nil {
		return "", false, err
	}
	return string(out), true, nil
}

// Tidy runs `go mod tidy` on the module file at rel inside dir, which writes its go.sum.
// It's a variable so tests, whose files live in memory, can stand in for it.
var Tidy = func(dir, rel string) error {
	cmd := exec.Command("go", "mod", "tidy", "-modfile="+rel) //nolint:gosec // rel is forgego's own pin path
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
