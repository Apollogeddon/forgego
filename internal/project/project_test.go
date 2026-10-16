package project

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/apollogeddon/forgego/internal/fsys"
)

func TestRemoteModule(t *testing.T) {
	for url, want := range map[string]string{
		"git@github.com:Acme/API.git":               "github.com/acme/api",
		"https://github.com/acme/api":               "github.com/acme/api",
		"https://user@github.com/acme/api.git":      "github.com/acme/api",
		"ssh://git@github.com/acme/api.git":         "github.com/acme/api",
		"ssh://git@gitlab.example.com:2222/a/b.git": "gitlab.example.com/a/b",
		"ssh://gitlab.example.com:22/a/b":           "gitlab.example.com/a/b",
	} {
		if got := remoteModule("[remote \"origin\"]\n\turl = " + url + "\n"); got != want {
			t.Errorf("%s: got %q, want %q", url, got, want)
		}
	}
}

func TestDetectMakesNamesSafe(t *testing.T) {
	fs := fsys.NewMemory(map[string]string{"/w/go.mod": "module github.com/Acme/BillingAPI/v3\n\ngo 1.25.0\n"})
	info, err := Detect(fs, "/w")
	if err != nil {
		t.Fatal(err)
	}
	if info.Name != "BillingAPI" || info.Slug != "billingapi" || info.Package != "billingapi" || info.GoVersion != "1.25.0" {
		t.Errorf("info = %+v", info)
	}

	info, err = Detect(fsys.NewMemory(nil), "/work/My Project")
	if err != nil {
		t.Fatal(err)
	}
	if info.Module != "my-project" || info.Slug != "my-project" {
		t.Errorf("info = %+v", info)
	}

	info, err = Detect(fsys.NewMemory(nil), "/work/Ünï Café")
	if err != nil {
		t.Fatal(err)
	}
	if info.Slug != "uni-cafe" || info.Package != "unicafe" {
		t.Errorf("accented: info = %+v", info)
	}
}

func TestGitHubOwner(t *testing.T) {
	for module, want := range map[string]string{
		"github.com/acme/api":                  "acme",
		"github.com/acme/api/services/billing": "acme",
		"gitlab.com/acme/api":                  "",
		"billing-api":                          "",
	} {
		if got := (Info{Module: module}).GitHubOwner(); got != want {
			t.Errorf("GitHubOwner(%q) = %q, want %q", module, got, want)
		}
	}
}

func TestRemoteModulePrefersOrigin(t *testing.T) {
	config := "[core]\n\tbare = false\n[submodule \"lib\"]\n\turl = https://github.com/other/lib\n" +
		"[remote \"upstream\"]\n\turl = https://github.com/upstream/api\n" +
		"[remote \"origin\"]\n\turl = git@github.com:acme/api.git\n"
	if got := remoteModule(config); got != "github.com/acme/api" {
		t.Errorf("got %q, want origin's module", got)
	}
	if got := remoteModule("[remote \"fork\"]\n\turl = https://github.com/me/api\n"); got != "github.com/me/api" {
		t.Errorf("without origin: got %q, want the first remote's module", got)
	}
	if got := remoteModule("[submodule \"lib\"]\n\turl = https://github.com/other/lib\n"); got != "" {
		t.Errorf("no remote: got %q", got)
	}
}

// In a worktree or submodule .git is a file naming the git directory; a worktree's config
// is the main repository's.
func TestDetectFollowsAGitFile(t *testing.T) {
	origin := "[remote \"origin\"]\n\turl = https://github.com/o/r\n"
	fs := fsys.NewMemory(map[string]string{
		"/repo/.git/config":                 origin,
		"/repo/.git/worktrees/wt/commondir": "../..\n",
		"/wt/.git":                          "gitdir: /repo/.git/worktrees/wt\n",
		"/repo/.git/modules/sub/config":     "[remote \"origin\"]\n\turl = https://github.com/o/sub\n",
		"/repo/sub/.git":                    "gitdir: ../.git/modules/sub\n",
	})
	for dir, want := range map[string]string{
		"/wt":             "github.com/o/r",
		"/wt/services/db": "github.com/o/r/services/db",
		"/repo/sub":       "github.com/o/sub",
	} {
		info, err := Detect(fs, dir)
		if err != nil {
			t.Fatal(err)
		}
		if info.Module != want {
			t.Errorf("%s: module %q, want %q", dir, info.Module, want)
		}
	}
}

func TestDetectInARealWorktree(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("needs git")
	}
	dir := t.TempDir()
	repo, wt := filepath.Join(dir, "repo"), filepath.Join(dir, "wt")
	for _, args := range [][]string{
		{"init", "-q", repo},
		{"-C", repo, "remote", "add", "origin", "git@github.com:o/r.git"},
		{"-C", repo, "-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "-q", "--allow-empty", "-m", "init"},
		{"-C", repo, "worktree", "add", "-q", wt},
	} {
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	info, err := Detect(fsys.OS{}, wt)
	if err != nil {
		t.Fatal(err)
	}
	if info.Module != "github.com/o/r" {
		t.Errorf("module %q, want github.com/o/r", info.Module)
	}
}

func TestASubdirectoryBecomesValidPathElements(t *testing.T) {
	fs := fsys.NewMemory(map[string]string{"/repo/.git/config": "[remote \"origin\"]\n\turl = git@github.com:acme/api.git\n"})
	for dir, want := range map[string]string{
		"/repo/My Service":            "github.com/acme/api/my-service",
		"/repo/Services/billing_v2.1": "github.com/acme/api/services/billing_v2.1",
		"/repo/.config/Ünï Café":      "github.com/acme/api/.config/uni-cafe",
	} {
		info, err := Detect(fs, dir)
		if err != nil {
			t.Fatalf("%s: %v", dir, err)
		}
		if info.Module != want {
			t.Errorf("%s: module %q, want %q", dir, info.Module, want)
		}
	}
}

func TestDetectRefusesAModulePathItCantMakeValid(t *testing.T) {
	for dir, config := range map[string]string{
		"/repo/con": "[remote \"origin\"]\n\turl = git@github.com:acme/api.git\n",
		"/repo":     "[remote \"origin\"]\n\turl = file:///srv/git/repo.git\n",
	} {
		fs := fsys.NewMemory(map[string]string{"/repo/.git/config": config})
		if info, err := Detect(fs, dir); err == nil {
			t.Errorf("%s: module %q, want an error", dir, info.Module)
		}
	}
}
