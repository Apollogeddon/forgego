package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func run(t *testing.T, args ...string) (int, string) {
	t.Helper()
	var out bytes.Buffer
	code := Run(args, &out, &out)
	return code, out.String()
}

func TestInitDryRunResolvesFeatureFlags(t *testing.T) {
	dir := t.TempDir()
	for _, tc := range []struct {
		args   []string
		active string
	}{
		{nil, "base, linting, build, testing, versioning, workflows"},
		{[]string{"--no-all"}, "base, build, workflows"},
		{[]string{"--no-all", "--testing"}, "base, build, testing, workflows"},
		{[]string{"--no-linting", "--docker", "--debian"}, "base, build, testing, versioning, docker, debian, workflows"},
		{[]string{"--website"}, "base, linting, build, versioning, workflows"},
	} {
		args := append([]string{"init", "--dry-run", "-C", dir}, tc.args...)
		code, out := run(t, args...)
		if code != 0 {
			t.Fatalf("%v: exit %d\n%s", tc.args, code, out)
		}
		if !strings.Contains(out, "Active features: "+tc.active+"\n") {
			t.Errorf("%v: want active %q in\n%s", tc.args, tc.active, out)
		}
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Errorf("dry run wrote %v", entries)
	}
}

func TestInitRejectsConflictingFlags(t *testing.T) {
	for _, args := range [][]string{
		{"init", "--library", "--website"},
		{"init", "--testing", "--no-testing"},
		{"init", "extra"},
		{"init", "--unknown"},
	} {
		if code, _ := run(t, args...); code != 1 {
			t.Errorf("%v: exit %d, want 1", args, code)
		}
	}
	if code, _ := run(t, "init", "--dry-run", "--library", "--docker", "-C", t.TempDir()); code != 2 {
		t.Errorf("--library --docker: exit %d, want 2", code)
	}
}

func TestCommitMsg(t *testing.T) {
	dir := t.TempDir()
	good, bad := filepath.Join(dir, "good"), filepath.Join(dir, "bad")
	_ = os.WriteFile(good, []byte("feat: add sync\n"), 0o600)
	_ = os.WriteFile(bad, []byte("added sync\n"), 0o600)
	if code, out := run(t, "commit-msg", good); code != 0 {
		t.Errorf("valid message rejected: %s", out)
	}
	if code, _ := run(t, "commit-msg", bad); code != 1 {
		t.Error("invalid message accepted")
	}
}

func TestVersionAndUsage(t *testing.T) {
	if code, out := run(t, "--version"); code != 0 || !strings.HasPrefix(out, "forgego ") {
		t.Errorf("--version: %d %q", code, out)
	}
	if code, _ := run(t); code != 1 {
		t.Error("no command should exit 1")
	}
	if code, _ := run(t, "nope"); code != 1 {
		t.Error("an unknown command should exit 1")
	}
}
