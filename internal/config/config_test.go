package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
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

// With no go_version, every pipeline sets up the latest patch of the Go release go.mod names,
// the newer of its go and toolchain lines, so a Go security fix reaches CI without a change
// to go.mod. The test runs the resolve step's own script on a few go.mod files.
func TestWorkflowsTakeTheLatestPatchOfGoModsRelease(t *testing.T) {
	// the script runs on Linux and macOS runners, and in Git Bash on Windows ones; a
	// Windows PATH can find WSL's bash first, which can't read the test's paths
	if runtime.GOOS == "windows" {
		t.Skip("the resolve step's script is checked on Unix")
	}
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("no bash")
	}
	b, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "service.yml"))
	if err != nil {
		t.Fatal(err)
	}
	step := regexp.MustCompile(`(?s)- name: Resolve the Go version\n.*?run: \|\n(.*?)\n\s*- name: Set up Go`).FindStringSubmatch(string(b))
	if step == nil {
		t.Fatal("no resolve step in service.yml")
	}
	script := step[1]

	for goMod, want := range map[string]string{
		"module m\n\ngo 1.27.0\n":                           "1.27",
		"module m\n\ngo 1.27\n":                             "1.27",
		"module m\n\ngo 1.26.0\n\ntoolchain go1.27.3\n":     "1.27",
		"module m\n\ngo 1.27.2\n\ntoolchain go1.26.9\n":     "1.27",
		"module m\n\ngo 1.27.0\n\ntoolchain default\n":      "1.27",
		"module m\n\ngo 1.28rc1\n":                          "1.28",
		"module m\n\n// go 1.30\ngo 1.27.0\nrequire (\n)\n": "1.27",
		// no go line: forgego's default
		"module m\n": DefaultGo,
	} {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o600); err != nil {
			t.Fatal(err)
		}
		out := filepath.Join(dir, "output")
		cmd := exec.Command(bash, "-e", "-c", script)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "INPUT=", "GITHUB_OUTPUT="+out)
		if b, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%q: %v: %s", goMod, err, b)
		}
		got, err := os.ReadFile(out)
		if err != nil {
			t.Fatal(err)
		}
		if strings.TrimSpace(string(got)) != "version="+want {
			t.Errorf("%q: resolved %q, want version=%s", goMod, strings.TrimSpace(string(got)), want)
		}
	}

	files, err := filepath.Glob(filepath.Join("..", "..", ".github", "workflows", "*.yml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		b, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		content := string(b)
		setups := strings.Count(content, "uses: actions/setup-go")
		if n := strings.Count(content, "check-latest: true"); n != setups {
			t.Errorf("%s: %d of %d setup-go steps check for the latest patch", filepath.Base(file), n, setups)
		}
		if strings.Contains(content, "go-version-file") {
			t.Errorf("%s: a setup-go step reads go.mod's exact version", filepath.Base(file))
		}
		if n := strings.Count(content, script); n != setups {
			t.Errorf("%s: %d of %d resolve steps run service.yml's script", filepath.Base(file), n, setups)
		}
	}
}
