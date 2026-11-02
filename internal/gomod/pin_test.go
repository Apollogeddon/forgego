package gomod

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/apollogeddon/forgego/v2/internal/fsys"
	"github.com/apollogeddon/forgego/v2/internal/templates"
)

const pinWant = `module github.com/apollogeddon/forgego/tools/task

go 1.26.4

tool github.com/go-task/task/v3/cmd/task

require (
	github.com/go-task/task/v3 v3.49.0
	golang.org/x/net v0.59.0 // indirect
)
`

func TestRaises(t *testing.T) {
	for _, tc := range []struct {
		name, mod string
		raises    bool
	}{
		{"the same pin", pinWant, true},
		{"an indirect module raised", strings.Replace(pinWant, "x/net v0.59.0", "x/net v0.62.0", 1), true},
		{"the tool raised", strings.Replace(pinWant, "v3 v3.49.0", "v3 v3.50.1", 1), true},
		{"a module added", strings.Replace(pinWant, ")\n", "\tgolang.org/x/text v0.43.0 // indirect\n)\n", 1), true},
		{"a newer Go", strings.Replace(pinWant, "go 1.26.4", "go 1.27.0", 1), true},
		{"a module older", strings.Replace(pinWant, "x/net v0.59.0", "x/net v0.58.0", 1), false},
		{"an indirect module dropped, as tidy does", strings.Replace(pinWant, "\tgolang.org/x/net v0.59.0 // indirect\n", "", 1), true},
		{"the tool's module dropped", strings.Replace(pinWant, "\tgithub.com/go-task/task/v3 v3.49.0\n", "", 1), false},
		{"an older Go", strings.Replace(pinWant, "go 1.26.4", "go 1.26.0", 1), false},
		{"another tool", strings.Replace(pinWant, "task/v3/cmd/task", "task/v3/cmd/other", 1), false},
		{"another module", strings.Replace(pinWant, "tools/task", "tools/other", 1), false},
		{"not a go.mod", "module", false},
	} {
		if got := Raises(tc.mod, pinWant); got != tc.raises {
			t.Errorf("%s: Raises = %v, want %v", tc.name, got, tc.raises)
		}
	}
}

// A pin unchanged since forgego wrote it is forgego's, and is replaced even when it's
// newer than forgego's pin now, as after a release that undoes a bad bump. One the project
// changed since, as a Dependabot security update does, is kept with its go.sum while it
// raises forgego's; any other is replaced with forgego's.
func TestPinnedKeepsAPinTheProjectRaised(t *testing.T) {
	tool := templates.Task
	plain := tool.ModFile()
	forgego := Pin{tool, mark(plain), tool.SumFile()}
	if !strings.Contains(plain, "golang.org/x/net ") {
		t.Fatal("task's pin no longer requires golang.org/x/net; pick another module to raise")
	}
	raised := raise(t, plain, "golang.org/x/net", "v0.99.0")
	lowered := raise(t, plain, "golang.org/x/net", "v0.0.1")
	raisedSum := tool.SumFile() + "golang.org/x/net v0.99.0 h1:x=\ngolang.org/x/net v0.99.0/go.mod h1:x=\n"
	// the project changed a pin forgego wrote: its line stays, and no longer matches
	edited := func(mod string) string { return markPrefix + digest(plain) + "\n" + mod }

	for _, tc := range []struct {
		name     string
		mod, sum string // empty for no file
		want     Pin
	}{
		{"no pin", "", "", forgego},
		{"forgego's pin", mark(plain), tool.SumFile(), forgego},
		{"a raised pin", edited(raised), raisedSum, Pin{tool, edited(raised), raisedSum}},
		{"a raised pin without its go.sum", edited(raised), "", forgego},
		{"a raised pin whose go.sum lacks the raise", edited(raised), tool.SumFile(), forgego},
		{"an older pin", edited(lowered), tool.SumFile() + "golang.org/x/net v0.0.1/go.mod h1:x=\n", forgego},
		{"a newer pin forgego wrote", mark(raised), raisedSum, forgego},
		{"a raised pin from before forgego marked its pins", raised, raisedSum, forgego},
	} {
		fs := fsys.NewMemory(nil)
		if tc.mod != "" {
			if err := fs.WriteFile(filepath.Join("/p", tool.ModPath()), tc.mod); err != nil {
				t.Fatal(err)
			}
		}
		if tc.sum != "" {
			if err := fs.WriteFile(filepath.Join("/p", tool.SumPath()), tc.sum); err != nil {
				t.Fatal(err)
			}
		}
		if got := Pinned(fs, "/p", tool); got != tc.want {
			t.Errorf("%s: kept the wrong pin", tc.name)
		}
	}
}

func TestUnmark(t *testing.T) {
	mod := templates.Task.ModFile()
	if body, unchanged := unmark(mark(mod)); body != mod || !unchanged {
		t.Errorf("a marked pin reads as changed, or loses its body")
	}
	if body, unchanged := unmark(mod); body != mod || unchanged {
		t.Errorf("a pin without forgego's line reads as forgego's")
	}
	if _, unchanged := unmark(strings.Replace(mark(mod), "go ", "go  ", 1)); unchanged {
		t.Errorf("a changed pin reads as unchanged")
	}
}

// raise sets module's version in a pin's require block.
func raise(t *testing.T, mod, module, version string) string {
	t.Helper()
	for _, line := range strings.Split(mod, "\n") {
		if fields := strings.Fields(line); len(fields) >= 2 && fields[0] == module {
			return strings.Replace(mod, line, strings.Replace(line, fields[1], version, 1), 1)
		}
	}
	t.Fatalf("%s isn't in the pin", module)
	return ""
}
