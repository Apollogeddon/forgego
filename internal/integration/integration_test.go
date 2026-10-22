// Package integration scaffolds real projects with a built forgego and runs the tasks it
// generates: the only test that catches a task, hook or tool pin that doesn't work.
package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

var forgego string

// exe is what the go command adds to a binary's name on Windows.
var exe = map[bool]string{true: ".exe"}[runtime.GOOS == "windows"]

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "forgego-bin")
	if err != nil {
		panic(err)
	}
	forgego = filepath.Join(dir, "forgego"+exe)
	// without VCS stamping it's a development build: under /v2, CI's merge commit would
	// stamp a pseudo-version that forgego would try to pin, though no proxy can fetch it
	if out, err := exec.Command("go", "build", "-buildvcs=false", "-o", forgego, "../../cmd/forgego").CombinedOutput(); err != nil {
		panic(string(out))
	}
	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

// scaffold runs forgego init in a fresh git repository and tidies the module.
func scaffold(t *testing.T, args ...string) string {
	t.Helper()
	if testing.Short() {
		t.Skip("scaffolds real projects and downloads every tool")
	}
	dir := t.TempDir()
	sh(t, dir, "git", "init", "-q")
	sh(t, dir, "git", "remote", "add", "origin", "https://github.com/acme/demo")
	sh(t, dir, forgego, append([]string{"init"}, args...)...)
	// run the forgego under test rather than a published one, as a project can
	taskfile := filepath.Join(dir, "Taskfile.yml")
	b, err := os.ReadFile(taskfile)
	if err != nil {
		t.Fatal(err)
	}
	// Task's shell reads a backslash as an escape, and Windows takes either separator
	pinned := regexp.MustCompile(`(?m)^  FORGEGO: .*$`).ReplaceAll(b, []byte("  FORGEGO: "+filepath.ToSlash(forgego)))
	if err := os.WriteFile(taskfile, pinned, 0o600); err != nil {
		t.Fatal(err)
	}
	if args[0] != "--website" {
		sh(t, dir, "go", "mod", "tidy")
	}
	return dir
}

// checkStages runs what the generated .index.yml's check pipeline runs in its lint, test and build stages,
// after checking the pipeline it calls is one this repository serves at @main.
func checkStages(t *testing.T, dir, pipeline string, golang bool) {
	t.Helper()
	index, err := os.ReadFile(filepath.Join(dir, ".github", "workflows", ".index.yml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range regexp.MustCompile(`apollogeddon/forgego/\.github/workflows/([\w-]+\.yml)@main`).FindAllStringSubmatch(string(index), -1) {
		if _, err := os.Stat(filepath.Join("..", "..", ".github", "workflows", m[1])); err != nil {
			t.Errorf(".index.yml calls %s, which this repository doesn't have", m[1])
		}
	}
	if !strings.Contains(string(index), "/"+pipeline+".yml@main") {
		t.Errorf(".index.yml doesn't call %s.yml", pipeline)
	}
	task(t, dir, "sync-check")
	if golang {
		sh(t, dir, "go", "mod", "tidy", "-diff")
		sh(t, dir, "go", "tool", "-modfile=.forgego/golangci-lint/go.mod", "golangci-lint", "fmt", "--diff")
		sh(t, dir, "go", "tool", "-modfile=.forgego/golangci-lint/go.mod", "golangci-lint", "run")
	}
}

// task runs a generated task, with forgego itself pointed at the build under test.
func task(t *testing.T, dir string, args ...string) string {
	t.Helper()
	return sh(t, dir, "go", append([]string{"tool", "-modfile=.forgego/task/go.mod", "task"}, args...)...)
}

func sh(t *testing.T, dir, name string, args ...string) string {
	t.Helper()
	out, err := command(dir, name, args...).CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s: %v\n%s", name, strings.Join(args, " "), err, out)
	}
	return string(out)
}

func command(dir, name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	// CI tests with GOFLAGS=-race, which would rebuild every tool with the race detector
	cmd.Env = append(os.Environ(), "GOFLAGS=")
	return cmd
}

func TestBackend(t *testing.T) {
	dir := scaffold(t, "--docker", "--debian")
	task(t, dir, "hooks")
	task(t, dir, "lint")
	checkStages(t, dir, "service", true)
	if out := task(t, dir, "test"); !strings.Contains(out, "DONE 1 tests") {
		t.Errorf("test ran no tests:\n%s", out)
	}
	task(t, dir, "build")
	if out := sh(t, dir, filepath.Join(dir, "dist", "demo"+exe)); strings.TrimSpace(out) != "demo dev" {
		t.Errorf("the built binary printed %q", out)
	}

	// the installed git hook itself runs the pinned lefthook, which finds no lefthook on PATH
	msgs := t.TempDir()
	good, bad := filepath.Join(msgs, "good"), filepath.Join(msgs, "bad")
	_ = os.WriteFile(good, []byte("feat: scaffold\n"), 0o600)
	_ = os.WriteFile(bad, []byte("scaffolded it\n"), 0o600)
	hook := filepath.Join(dir, ".git", "hooks", "commit-msg")
	sh(t, dir, "sh", hook, good)
	if out, err := command(dir, "sh", hook, bad).CombinedOutput(); err == nil {
		t.Errorf("the commit-msg hook accepted a bad message:\n%s", out)
	}
}

func TestLibrary(t *testing.T) {
	dir := scaffold(t, "--library")
	task(t, dir, "lint")
	checkStages(t, dir, "library", true)
	task(t, dir, "test")
	task(t, dir, "build")
}

// An index.yml an earlier forgego wrote would run beside .index.yml, calling pipelines that are gone.
func TestAnOldGeneratedWorkflowGoes(t *testing.T) {
	if testing.Short() {
		t.Skip("builds forgego")
	}
	legacy, err := os.ReadFile(filepath.Join("..", "templates", "testdata", "legacy", "main-library.yml"))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	old := filepath.Join(dir, ".github", "workflows", "index.yml")
	if err := os.MkdirAll(filepath.Dir(old), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(old, legacy, 0o600); err != nil {
		t.Fatal(err)
	}
	sh(t, dir, "git", "init", "-q")
	out := sh(t, dir, forgego, "init", "--library")
	if _, err := os.Stat(old); !os.IsNotExist(err) || !strings.Contains(out, "Removed .github/workflows/index.yml") {
		t.Errorf("the old index.yml is still there: %v\n%s", err, out)
	}
	if _, err := os.Stat(filepath.Join(dir, ".github", "workflows", ".index.yml")); err != nil {
		t.Error(err)
	}
}

func TestWebsite(t *testing.T) {
	dir := scaffold(t, "--website")
	// Hugo refuses every remote resource, so the build proves the site needs no CDN
	cmd := command(dir, "go", "tool", "-modfile=.forgego/task/go.mod", "task", "build")
	cmd.Env = append(cmd.Env, "HUGO_SECURITY_HTTP_URLS=none")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("the site doesn't build without remote resources: %v\n%s", err, out)
	}
	if _, err := os.Stat(filepath.Join(dir, "public", "docs", "index.html")); err != nil {
		t.Errorf("the site has no docs page: %v", err)
	}
	checkStages(t, dir, "website", false)
}
