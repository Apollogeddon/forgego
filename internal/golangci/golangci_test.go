package golangci

import (
	"strings"
	"testing"

	"github.com/apollogeddon/forgego/internal/yamlx"
)

func TestRenderOfTheStarterIsTheBaseConfig(t *testing.T) {
	out, err := Render(LocalStarter)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out, header) {
		t.Errorf("missing the generated header:\n%s", out)
	}
	if strings.Contains(out, "forgego owns this file") {
		t.Errorf("kept the base file's own header:\n%s", out)
	}
	for _, want := range []string{"default: standard", "- gosec", "- gofumpt"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q", want)
		}
	}
}

func TestRenderAddsAndOverridesFromTheLocalConfig(t *testing.T) {
	out, err := Render("linters:\n  enable:\n    - wsl_v5\n    - gosec\n  default: all\n")
	if err != nil {
		t.Fatal(err)
	}
	root, err := yamlx.Parse(out)
	if err != nil {
		t.Fatal(err)
	}
	linters := yamlx.Get(root, "linters")
	if got := yamlx.Get(linters, "default").Value; got != "all" {
		t.Errorf("default = %q, want the local value", got)
	}
	enabled := map[string]int{}
	for _, item := range yamlx.Get(linters, "enable").Content {
		enabled[item.Value]++
	}
	if enabled["wsl_v5"] != 1 || enabled["gosec"] != 1 || enabled["revive"] != 1 {
		t.Errorf("enable = %v, want the base list plus wsl_v5, each once", enabled)
	}
}
