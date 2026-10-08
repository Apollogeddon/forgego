package gomod

import (
	"go/version"
	"strings"
	"testing"

	"github.com/apollogeddon/forgego/internal/templates"
)

func TestToolsGoIsTheNewestToolRequirement(t *testing.T) {
	if got := ToolsGo(nil); got != "" {
		t.Errorf("ToolsGo(nil) = %q", got)
	}
	got := ToolsGo(templates.Tools)
	for _, tool := range templates.Tools {
		if version.Compare("go"+tool.GoVersion(), got) > 0 {
			t.Errorf("%s needs %s, newer than %s", tool.Name, tool.GoVersion(), got)
		}
	}
	if want := ToolsGo([]templates.Tool{templates.Gotestsum}); got == want {
		t.Errorf("every tool needs the same Go as gotestsum, %s", got)
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
