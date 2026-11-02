package gomod

import (
	"go/version"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/apollogeddon/forgego/v2/internal/templates"
)

func TestToolsGoIsTheNewestToolRequirement(t *testing.T) {
	if got := ToolsGo(nil); got != "" {
		t.Errorf("ToolsGo(nil) = %q", got)
	}
	var pins []Pin
	for _, tool := range templates.Tools {
		pins = append(pins, Pin{tool, tool.ModFile(), tool.SumFile()})
	}
	got := ToolsGo(pins)
	for _, tool := range templates.Tools {
		if version.Compare("go"+tool.GoVersion(), got) > 0 {
			t.Errorf("%s needs %s, newer than %s", tool.Name, tool.GoVersion(), got)
		}
	}
	gotestsum := templates.Gotestsum
	if want := ToolsGo([]Pin{{gotestsum, gotestsum.ModFile(), gotestsum.SumFile()}}); got == want {
		t.Errorf("every tool needs the same Go as gotestsum, %s", got)
	}
	// a pin the project raised to a newer Go needs that Go
	raised := strings.Replace(gotestsum.ModFile(), "go "+gotestsum.GoVersion(), "go 1.99.0", 1)
	if got := ToolsGo([]Pin{{gotestsum, raised, ""}}); got != "go1.99.0" {
		t.Errorf("a raised pin's Go = %s, want go1.99.0", got)
	}
}

func TestEnsureToolchain(t *testing.T) {
	for _, tc := range []struct {
		name, gomod, need string
		changed           bool
		want              string
	}{
		{"older go line", "module x\n\ngo 1.26.0\n", "go1.27.0", true, "toolchain go1.27.0"},
		{"go line reaches it", "module x\n\ngo 1.27.1\n", "go1.27.0", false, ""},
		{"toolchain reaches it", "module x\n\ngo 1.25.0\n\ntoolchain go1.27.3\n", "go1.27.0", false, ""},
		{"older toolchain", "module x\n\ngo 1.25.0\n\ntoolchain go1.26.1\n", "go1.27.0", true, "toolchain go1.27.0"},
		{"no tools", "module x\n\ngo 1.20\n", "", false, ""},
	} {
		out, changed, err := EnsureToolchain(tc.gomod, tc.need)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if changed != tc.changed {
			t.Errorf("%s: changed = %v", tc.name, changed)
		}
		if tc.changed && (!strings.Contains(out, tc.want) || !strings.Contains(out, "\ngo 1.2")) {
			t.Errorf("%s: got\n%s", tc.name, out)
		}
		if !tc.changed && out != tc.gomod {
			t.Errorf("%s: rewrote an unchanged go.mod", tc.name)
		}
	}
}

func TestTidyKeepsTheProjectOutOfTheSelfPin(t *testing.T) {
	if testing.Short() {
		t.Skip("downloads a published forgego")
	}
	t.Setenv("GOFLAGS", "")
	dir := t.TempDir()
	files := map[string]string{
		"go.mod":     "module example.com/proj\n\ngo 1.26.0\n\nrequire github.com/google/go-cmp v0.7.0\n",
		"main.go":    "package main\n\nimport (\n\t\"example.com/proj/sub\"\n\t\"github.com/google/go-cmp/cmp\"\n)\n\nfunc main() { _ = cmp.Diff(sub.X, 2) }\n",
		"sub/sub.go": "package sub\n\nconst X = 1\n",
		// a v1 pin, as Tidy needs a release published at the path it requires
		templates.SelfModPath: strings.ReplaceAll(templates.SelfModFile("v1.5.0"), "forgego/v2", "forgego"),
	}
	for rel, content := range files {
		path := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := Tidy(dir, templates.SelfModPath); err != nil {
		t.Fatal(err)
	}
	pin, err := os.ReadFile(filepath.Join(dir, templates.SelfModPath))
	if err != nil {
		t.Fatal(err)
	}
	for _, leaked := range []string{"example.com/proj", "go-cmp"} {
		if strings.Contains(string(pin), leaked) {
			t.Errorf("the pin requires the project's %s:\n%s", leaked, pin)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, templates.SelfSumPath)); err != nil {
		t.Errorf("no go.sum beside the pin: %v", err)
	}
	cmd := exec.Command("go", "tool", "-modfile="+templates.SelfModPath, "forgego", "--version")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Errorf("running the pinned forgego: %v\n%s", err, out)
	}
}
