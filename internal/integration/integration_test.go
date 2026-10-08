// Package integration scaffolds real projects with a built forgego and runs the tasks it
// generates: the only test that catches a task, hook or tool pin that doesn't work.
package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var forgego string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "forgego-bin")
	if err != nil {
		panic(err)
	}
	forgego = filepath.Join(dir, "forgego")
	if out, err := exec.Command("go", "build", "-o", forgego, "../../cmd/forgego").CombinedOutput(); err != nil {
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
	pinned := regexp.MustCompile(`(?m)^  FORGEGO: .*$`).ReplaceAll(b, []byte("  FORGEGO: "+forgego))
	if err := os.WriteFile(taskfile, pinned, 0o600); err != nil {
		t.Fatal(err)
	}
	if args[0] != "--website" {
		sh(t, dir, "go", "mod", "tidy")
	}
	return dir
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
	if out := task(t, dir, "test"); !strings.Contains(out, "DONE 1 tests") {
		t.Errorf("test ran no tests:\n%s", out)
	}
	task(t, dir, "build")
	if out := sh(t, dir, filepath.Join("dist", "demo")); out != "demo dev\n" {
		t.Errorf("the built binary printed %q", out)
	}
	task(t, dir, "sync-check")

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
	task(t, dir, "test")
	task(t, dir, "build")
	task(t, dir, "sync-check")
}

func TestWebsite(t *testing.T) {
	if os.Getenv("FORGEGO_OFFLINE") != "" {
		t.Skip("the theme fetches its search script from a CDN at build time")
	}
	dir := scaffold(t, "--website")
	task(t, dir, "build")
	if _, err := os.Stat(filepath.Join(dir, "public", "docs", "index.html")); err != nil {
		t.Errorf("the site has no docs page: %v", err)
	}
	task(t, dir, "sync-check")
}
