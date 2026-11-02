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

// A project's pin that raises forgego's, as a Dependabot security update leaves it, is
// kept with its go.sum; any other is replaced with forgego's.
func TestPinnedKeepsAPinTheProjectRaised(t *testing.T) {
	tool := templates.Task
	forgego := tool.ModFile()
	if !strings.Contains(forgego, "golang.org/x/net ") {
		t.Fatal("task's pin no longer requires golang.org/x/net; pick another module to raise")
	}
	raised := raise(t, forgego, "golang.org/x/net", "v0.99.0")
	lowered := raise(t, forgego, "golang.org/x/net", "v0.0.1")
	raisedSum := tool.SumFile() + "golang.org/x/net v0.99.0 h1:x=\ngolang.org/x/net v0.99.0/go.mod h1:x=\n"

	for _, tc := range []struct {
		name     string
		mod, sum string // empty for no file
		want     Pin
	}{
		{"no pin", "", "", Pin{tool, forgego, tool.SumFile()}},
		{"forgego's pin", forgego, tool.SumFile(), Pin{tool, forgego, tool.SumFile()}},
		{"a raised pin", raised, raisedSum, Pin{tool, raised, raisedSum}},
		{"a raised pin without its go.sum", raised, "", Pin{tool, forgego, tool.SumFile()}},
		{"a raised pin whose go.sum lacks the raise", raised, tool.SumFile(), Pin{tool, forgego, tool.SumFile()}},
		{"an older pin", lowered, tool.SumFile() + "golang.org/x/net v0.0.1/go.mod h1:x=\n", Pin{tool, forgego, tool.SumFile()}},
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
