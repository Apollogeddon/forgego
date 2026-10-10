package config

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The workflows fall back to DefaultGo when a project passes no go_version and has no go.mod,
// so a new project and one without a go.mod build on the same version.
func TestWorkflowsFallBackToDefaultGo(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("..", "..", ".github", "workflows", "*.yml"))
	if err != nil {
		t.Fatal(err)
	}
	fallback := regexp.MustCompile(`else echo "version=(\S+)"`)
	checked := 0
	for _, file := range files {
		b, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		content := string(b)
		setups := strings.Count(content, "uses: actions/setup-go")
		if setups == 0 {
			continue
		}
		checked++
		if n := strings.Count(content, "go-version: ${{ steps.go.outputs.version }}"); n != setups {
			t.Errorf("%s: %d of %d setup-go steps use the resolved version", filepath.Base(file), n, setups)
		}
		matches := fallback.FindAllStringSubmatch(content, -1)
		if len(matches) != setups {
			t.Errorf("%s: %d fallbacks for %d setup-go steps", filepath.Base(file), len(matches), setups)
		}
		for _, m := range matches {
			if m[1] != DefaultGo {
				t.Errorf("%s: falls back to Go %s, not DefaultGo %s", filepath.Base(file), m[1], DefaultGo)
			}
		}
	}
	if checked == 0 {
		t.Fatal("no workflow sets up Go")
	}
}
