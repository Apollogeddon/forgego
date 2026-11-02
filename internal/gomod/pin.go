package gomod

import (
	"crypto/sha256"
	"encoding/hex"
	"go/version"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/semver"

	"github.com/apollogeddon/forgego/v2/internal/fsys"
	"github.com/apollogeddon/forgego/v2/internal/templates"
)

// Pin is a tool's module files as a project keeps them in .forgego/<tool>/.
type Pin struct {
	Tool     templates.Tool
	Mod, Sum string
}

// Go is the Go the pin needs, such as go1.27.0.
func (p Pin) Go() string {
	f, err := modfile.ParseLax(p.Tool.ModPath(), []byte(p.Mod), nil)
	if err != nil || f.Go == nil {
		return "go" + p.Tool.GoVersion()
	}
	return "go" + f.Go.Version
}

// markPrefix starts the line forgego puts first in each pin it writes, followed by a digest
// of the rest, so it can tell a pin it wrote from one the project changed since.
const markPrefix = "// Pinned by forgego, "

// mark gives mod the line that says forgego wrote it.
func mark(mod string) string {
	return markPrefix + digest(mod) + "\n" + mod
}

// unmark returns mod without forgego's line, and whether mod is unchanged since forgego
// wrote it: its line is there and its digest matches the rest.
func unmark(mod string) (string, bool) {
	first, rest, _ := strings.Cut(mod, "\n")
	recorded, ok := strings.CutPrefix(first, markPrefix)
	if !ok {
		return mod, false
	}
	return rest, recorded == digest(rest)
}

func digest(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:8])
}

// Pinned is the pin a project keeps for tool: its own when the project changed it and it
// raises forgego's, else forgego's. A Dependabot security update in the project raises a
// module of the pin past forgego's, and keeping it lets the project patch a tool without
// waiting for a forgego release; once forgego's pin is newer in any module, forgego's
// replaces it. A pin unchanged since forgego wrote it is always replaced, so a forgego
// release that pins a module lower, as one that undoes a bad bump does, reaches the project.
func Pinned(fs fsys.FS, dir string, tool templates.Tool) Pin {
	forgego := Pin{tool, mark(tool.ModFile()), tool.SumFile()}
	mod, err := fs.ReadFile(filepath.Join(dir, tool.ModPath()))
	if err != nil {
		return forgego
	}
	// a pin without forgego's line came from a release before it, and is treated as written
	// by forgego: it can't be told from one the project changed
	body, unchanged := unmark(mod)
	if unchanged || !strings.HasPrefix(mod, markPrefix) {
		return forgego
	}
	sum, err := fs.ReadFile(filepath.Join(dir, tool.SumPath()))
	if err != nil || !Raises(body, tool.ModFile()) || !summed(body, sum) {
		return forgego
	}
	return Pin{tool, mod, sum}
}

// summed reports whether sum holds the go.mod sum of every module mod requires, as go mod
// tidy leaves it; a go.mod raised by hand without it couldn't run.
func summed(mod, sum string) bool {
	f, err := modfile.Parse("go.mod", []byte(mod), nil)
	if err != nil {
		return false
	}
	lines := make(map[string]bool)
	for _, line := range strings.Split(sum, "\n") {
		if fields := strings.Fields(line); len(fields) == 3 {
			lines[fields[0]+" "+fields[1]] = true
		}
	}
	for _, r := range f.Require {
		if !lines[r.Mod.Path+" "+r.Mod.Version+"/go.mod"] {
			return false
		}
	}
	return true
}

// Raises reports whether mod is the same module as want, runs the same tools, needs no
// older Go, and requires the modules both do at the same version or a newer one. A module
// only want requires may be missing: go mod tidy drops one a raised module stopped
// needing. The modules the tools come from, want's direct requirements, can't be.
func Raises(mod, want string) bool {
	m, err := modfile.Parse("go.mod", []byte(mod), nil)
	if err != nil {
		return false
	}
	w, err := modfile.Parse("go.mod", []byte(want), nil)
	if err != nil || m.Module == nil || w.Module == nil || m.Module.Mod.Path != w.Module.Mod.Path {
		return false
	}
	if m.Go == nil || w.Go == nil || version.Compare("go"+m.Go.Version, "go"+w.Go.Version) < 0 {
		return false
	}
	tools := make(map[string]bool, len(m.Tool))
	for _, t := range m.Tool {
		tools[t.Path] = true
	}
	for _, t := range w.Tool {
		if !tools[t.Path] {
			return false
		}
	}
	required := make(map[string]string, len(m.Require))
	for _, r := range m.Require {
		required[r.Mod.Path] = r.Mod.Version
	}
	for _, r := range w.Require {
		v, ok := required[r.Mod.Path]
		if !ok && !r.Indirect || ok && semver.Compare(v, r.Mod.Version) < 0 {
			return false
		}
	}
	return true
}
