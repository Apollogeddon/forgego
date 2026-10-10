package core

import (
	"io"
	"slices"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"

	"github.com/apollogeddon/forgego/internal/config"
	"github.com/apollogeddon/forgego/internal/console"
	"github.com/apollogeddon/forgego/internal/fsys"
	"github.com/apollogeddon/forgego/internal/gomod"
	"github.com/apollogeddon/forgego/internal/sync"
	"github.com/apollogeddon/forgego/internal/templates"
	"github.com/apollogeddon/forgego/internal/version"
)

const dir = "/work/billing-api"

// tidied records each module gomod.Tidy was asked to write the go.sum of.
var tidied []string

func TestMain(m *testing.M) {
	console.SetOutput(io.Discard, io.Discard)
	version.Set("v1.2.3")
	// the files are in memory, where go mod tidy can't run
	gomod.Tidy = func(dir, rel string) error {
		tidied = append(tidied, dir+"/"+rel)
		return nil
	}
	m.Run()
}

func run(t *testing.T, fs *fsys.Memory, edit func(*config.Init)) {
	t.Helper()
	cfg := config.Default(dir)
	if edit != nil {
		edit(&cfg)
	}
	if code := Init(cfg, fs); code != 0 {
		t.Fatalf("Init returned %d", code)
	}
}

func read(t *testing.T, fs *fsys.Memory, rel string) string {
	t.Helper()
	content, err := fs.ReadFile(dir + "/" + rel)
	if err != nil {
		t.Fatalf("%s: %v", rel, err)
	}
	return content
}

func files(fs *fsys.Memory) []string {
	var rel []string
	for _, p := range fs.Paths() {
		rel = append(rel, strings.TrimPrefix(p, dir+"/"))
	}
	return rel
}

func assertFiles(t *testing.T, fs *fsys.Memory, want ...string) {
	t.Helper()
	got := files(fs)
	for _, w := range want {
		if !slices.Contains(got, w) {
			t.Errorf("missing %s in %v", w, got)
		}
	}
}

func refuteFiles(t *testing.T, fs *fsys.Memory, unwanted ...string) {
	t.Helper()
	got := files(fs)
	for _, u := range unwanted {
		if slices.Contains(got, u) {
			t.Errorf("unexpected %s", u)
		}
	}
}

func TestBackend(t *testing.T) {
	fs := fsys.NewMemory(nil)
	run(t, fs, nil)
	assertFiles(t, fs,
		"go.mod", "cmd/billing-api/main.go", "cmd/billing-api/main_test.go", ".gitignore",
		".forgego/task/go.mod", ".forgego/task/go.sum", ".forgego/golangci-lint/go.mod", ".forgego/gotestsum/go.mod",
		".forgego/govulncheck/go.mod", ".forgego/lefthook/go.mod", ".forgego/golangci.yml",
		".golangci.yml", ".golangci.local.yml", "lefthook.yml", ".goreleaser.yaml",
		".github/release.json", ".github/.release.json", ".github/workflows/index.yml", "Taskfile.yml",
	)
	refuteFiles(t, fs, ".forgego/hugo/go.mod", "Dockerfile", "hugo.toml")
	if got := read(t, fs, "go.mod"); got != "module billing-api\n\ngo 1.27.0\n" {
		t.Errorf("go.mod = %q", got)
	}
	taskfile := read(t, fs, "Taskfile.yml")
	for _, task := range []string{"lint:", "type:", "test:", "build:", "start:", "security:", "sync:", "sync-check:", "hooks:", "commit-msg:", "release:snapshot:"} {
		if !strings.Contains(taskfile, "  "+task+"\n") {
			t.Errorf("Taskfile.yml has no %s task", task)
		}
	}
	if !strings.Contains(taskfile, "FORGEGO: "+templates.SelfCommand) {
		t.Errorf("Taskfile.yml doesn't run forgego from its pin:\n%s", taskfile)
	}
	if pin := read(t, fs, ".forgego/forgego/go.mod"); !strings.Contains(pin, "\nrequire github.com/apollogeddon/forgego v1.2.3\n") ||
		!strings.Contains(pin, "\ntool github.com/apollogeddon/forgego/cmd/forgego\n") {
		t.Errorf("the pin isn't this forgego:\n%s", pin)
	}
	if !slices.Contains(tidied, dir+"/.forgego/forgego/go.mod") {
		t.Errorf("init didn't write the pin's go.sum: %v", tidied)
	}
}

func TestLibrary(t *testing.T) {
	fs := fsys.NewMemory(nil)
	run(t, fs, func(c *config.Init) { c.Mode = config.Library })
	assertFiles(t, fs, "billingapi.go", "billingapi_test.go")
	refuteFiles(t, fs, "cmd/billing-api/main.go", ".goreleaser.yaml")
	if !strings.HasPrefix(read(t, fs, "billingapi.go"), "// Package billingapi") {
		t.Error("the library starter isn't package billingapi")
	}
}

func TestOnlyABackendsReleaseIsADraftForGoReleaserToPublish(t *testing.T) {
	backend := fsys.NewMemory(nil)
	run(t, backend, nil)
	if !strings.Contains(read(t, backend, ".github/release.json"), `"draft": true`) {
		t.Error("a backend's release isn't a draft, so GoReleaser can't attach files to it in a repository with immutable releases")
	}
	if !strings.Contains(read(t, backend, ".goreleaser.yaml"), "use_existing_draft: true") {
		t.Error("GoReleaser doesn't publish release-please's draft")
	}

	library := fsys.NewMemory(nil)
	run(t, library, func(c *config.Init) { c.Mode = config.Library })
	if strings.Contains(read(t, library, ".github/release.json"), "draft") {
		t.Error("a library's release is a draft, but nothing publishes it")
	}
}

func TestWebsite(t *testing.T) {
	fs := fsys.NewMemory(nil)
	run(t, fs, func(c *config.Init) { c.Mode = config.Website })
	assertFiles(t, fs, "hugo.toml", "content/_index.md", "content/docs/_index.md", ".forgego/hugo/go.mod", "lefthook.yml")
	// a Hugo site has no Go to lint or test
	refuteFiles(t, fs, ".golangci.yml", ".forgego/golangci-lint/go.mod", ".forgego/gotestsum/go.mod", "cmd/billing-api/main.go")
}

func TestModuleComesFromTheGitRemote(t *testing.T) {
	fs := fsys.NewMemory(map[string]string{
		dir + "/.git/config": "[remote \"origin\"]\n\turl = git@github.com:Acme/Billing-API.git\n",
	})
	run(t, fs, nil)
	if !strings.HasPrefix(read(t, fs, "go.mod"), "module github.com/acme/billing-api\n") {
		t.Errorf("go.mod = %q", read(t, fs, "go.mod"))
	}
}

func TestAnExistingGoModNamesTheProject(t *testing.T) {
	fs := fsys.NewMemory(map[string]string{dir + "/go.mod": "module example.com/tools/v2\n\ngo 1.27.1\n"})
	run(t, fs, nil)
	refuteFiles(t, fs, "cmd/tools/main.go", "cmd/tools/main_test.go")
	if !strings.Contains(read(t, fs, "Taskfile.yml"), "./cmd/tools") {
		t.Error("the build task doesn't use the module's name")
	}
	if read(t, fs, "go.mod") != "module example.com/tools/v2\n\ngo 1.27.1\n" {
		t.Error("rewrote an existing go.mod")
	}
}

func TestInvalidCombinationsWriteNothing(t *testing.T) {
	for name, edit := range map[string]func(*config.Init){
		"docker library": func(c *config.Init) { c.Mode, c.Docker = config.Library, true },
		"debian website": func(c *config.Init) { c.Mode, c.Debian = config.Website, true },
		"bad go":         func(c *config.Init) { c.Go = "go1.27" },
	} {
		fs := fsys.NewMemory(nil)
		cfg := config.Default(dir)
		edit(&cfg)
		if code := Init(cfg, fs); code != ExitInvalidConfig {
			t.Errorf("%s: Init returned %d", name, code)
		}
		if len(fs.Paths()) != 0 {
			t.Errorf("%s: wrote %v", name, fs.Paths())
		}
	}
}

func TestDryRunWritesNothing(t *testing.T) {
	fs := fsys.NewMemory(nil)
	run(t, fs, func(c *config.Init) { c.DryRun, c.Docker, c.Debian = true, true, true })
	if len(fs.Paths()) != 0 {
		t.Errorf("dry run wrote %v", fs.Paths())
	}
}

func TestRerunKeepsConfigsUnlessForcedAndSourceAlways(t *testing.T) {
	fs := fsys.NewMemory(map[string]string{
		dir + "/lefthook.yml":            "mine\n",
		dir + "/cmd/billing-api/main.go": "package main // mine\n",
		dir + "/Taskfile.yml":            "version: '3'\ntasks:\n  test:\n    cmds: [make test]\n  deploy:\n    cmds: [./deploy.sh]\n",
	})
	run(t, fs, nil)
	if read(t, fs, "lefthook.yml") != "mine\n" {
		t.Error("overwrote lefthook.yml without --force")
	}
	taskfile := read(t, fs, "Taskfile.yml")
	if !strings.Contains(taskfile, "make test") || !strings.Contains(taskfile, "./deploy.sh") || !strings.Contains(taskfile, "  lint:") {
		t.Errorf("Taskfile.yml should keep the project's tasks and gain the rest:\n%s", taskfile)
	}

	run(t, fs, func(c *config.Init) { c.Force = true })
	if read(t, fs, "lefthook.yml") == "mine\n" {
		t.Error("--force didn't overwrite lefthook.yml")
	}
	if read(t, fs, "cmd/billing-api/main.go") != "package main // mine\n" {
		t.Error("--force overwrote the project's own source")
	}
	taskfile = read(t, fs, "Taskfile.yml")
	if strings.Contains(taskfile, "make test") || !strings.Contains(taskfile, "./deploy.sh") {
		t.Errorf("--force should replace forgego's tasks and keep the others:\n%s", taskfile)
	}
}

func TestDisabledFeaturesAreRemovedOnlyWithForce(t *testing.T) {
	fs := fsys.NewMemory(nil)
	run(t, fs, func(c *config.Init) { c.Docker, c.Debian = true, true })
	assertFiles(t, fs, "Dockerfile", "packaging/billing-api.service", "packaging/postinstall.sh")

	run(t, fs, func(c *config.Init) { c.Linting = false })
	assertFiles(t, fs, "Dockerfile", "lefthook.yml")

	run(t, fs, func(c *config.Init) { c.Linting, c.Testing, c.Force = false, false, true })
	refuteFiles(t, fs, "Dockerfile", "packaging/billing-api.service", "lefthook.yml", ".golangci.yml",
		".forgego/golangci-lint/go.mod", ".forgego/gotestsum/go.mod")
	// the project's own local config is never removed
	assertFiles(t, fs, ".golangci.local.yml")
}

func TestDebianRunsTheServiceAsItsOwnUser(t *testing.T) {
	fs := fsys.NewMemory(nil)
	run(t, fs, func(c *config.Init) { c.Debian = true })
	unit := read(t, fs, "packaging/billing-api.service")
	if !strings.Contains(unit, "User=billing-api") || strings.Contains(unit, "root") {
		t.Errorf("unit:\n%s", unit)
	}
	if !strings.Contains(read(t, fs, "packaging/postinstall.sh"), "useradd --system --no-create-home --shell /usr/sbin/nologin") {
		t.Error("postinstall doesn't create a system user")
	}
	if !strings.Contains(read(t, fs, ".goreleaser.yaml"), "nfpms:") {
		t.Error(".goreleaser.yaml has no .deb")
	}
}

type workflow struct {
	Concurrency map[string]string `yaml:"concurrency"`
	Jobs        map[string]struct {
		Uses        string            `yaml:"uses"`
		Needs       string            `yaml:"needs"`
		Secrets     any               `yaml:"secrets"`
		Permissions map[string]string `yaml:"permissions"`
		With        map[string]any    `yaml:"with"`
	} `yaml:"jobs"`
}

func TestGeneratedWorkflowIsLeastPrivilegeAndNeverCancelsMain(t *testing.T) {
	for name, edit := range map[string]func(*config.Init){
		"backend":        nil,
		"library":        func(c *config.Init) { c.Mode = config.Library },
		"website":        func(c *config.Init) { c.Mode = config.Website },
		"backend docker": func(c *config.Init) { c.Docker = true },
		"website docker": func(c *config.Init) { c.Mode, c.Docker = config.Website, true },
	} {
		fs := fsys.NewMemory(nil)
		run(t, fs, edit)
		var w workflow
		if err := yaml.Unmarshal([]byte(read(t, fs, ".github/workflows/index.yml")), &w); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if !strings.Contains(w.Concurrency["cancel-in-progress"], "refs/heads/main") {
			t.Errorf("%s: cancels runs on main", name)
		}
		for job, j := range w.Jobs {
			if j.Secrets != nil {
				t.Errorf("%s: %s passes secrets", name, job)
			}
			if _, ok := j.Permissions["id-token"]; ok && job != "website" {
				t.Errorf("%s: %s asks for id-token", name, job)
			}
			if !strings.HasPrefix(j.Uses, "apollogeddon/forgego/.github/workflows/") {
				t.Errorf("%s: %s uses %s", name, job, j.Uses)
			}
		}
		if edit != nil && strings.Contains(name, "docker") && w.Jobs["docker"].Needs == "" {
			t.Errorf("%s: no docker job after the pipeline", name)
		}
	}
}

func TestDisabledFeaturesBecomeWorkflowInputs(t *testing.T) {
	fs := fsys.NewMemory(nil)
	run(t, fs, func(c *config.Init) { c.Testing, c.Versioning, c.Linting = false, false, false })
	var w workflow
	if err := yaml.Unmarshal([]byte(read(t, fs, ".github/workflows/index.yml")), &w); err != nil {
		t.Fatal(err)
	}
	with := w.Jobs["service"].With
	if with["run_tests"] != false || with["enable_versioning"] != false || with["lint"] != false || with["go_version"] != nil {
		t.Errorf("with = %v", with)
	}
}

func TestSyncFindsAndFixesDrift(t *testing.T) {
	fs := fsys.NewMemory(nil)
	run(t, fs, nil)
	if code := sync.Run(fs, dir, true); code != 0 {
		t.Fatalf("a fresh project has drift: %d", code)
	}

	_ = fs.WriteFile(dir+"/.forgego/task/go.mod", "stale\n")
	_ = fs.WriteFile(dir+"/.golangci.local.yml", "linters:\n  enable: [wsl_v5]\n")
	version.Set("v1.3.0")
	defer version.Set("v1.2.3")
	if code := sync.Run(fs, dir, true); code != 1 {
		t.Fatalf("sync --check found no drift")
	}
	if read(t, fs, ".forgego/task/go.mod") != "stale\n" {
		t.Error("sync --check wrote a file")
	}

	if code := sync.Run(fs, dir, false); code != 0 {
		t.Fatalf("sync returned %d", code)
	}
	if code := sync.Run(fs, dir, true); code != 0 {
		t.Errorf("drift remains after sync")
	}
	if !strings.Contains(read(t, fs, ".golangci.yml"), "wsl_v5") {
		t.Error(".golangci.yml didn't pick up the local change")
	}
	// Dependabot owns the pinned version, so sync leaves it
	if !strings.Contains(read(t, fs, ".forgego/forgego/go.mod"), "forgego v1.2.3\n") {
		t.Error("sync rewrote the pinned forgego")
	}
}

func TestSyncMovesToolsPinnedAtTheOldPaths(t *testing.T) {
	fs := fsys.NewMemory(nil)
	run(t, fs, nil)
	// an earlier forgego pinned each tool as .forgego/<tool>.mod and .sum
	for _, rel := range []string{"Taskfile.yml", "lefthook.yml"} {
		_ = fs.WriteFile(dir+"/"+rel, strings.NewReplacer(
			"/task/go.mod", "/task.mod", "/lefthook/go.mod", "/lefthook.mod", "/golangci-lint/go.mod", "/golangci-lint.mod",
		).Replace(read(t, fs, rel)))
	}
	for _, tool := range []string{"task", "lefthook", "golangci-lint"} {
		_ = fs.WriteFile(dir+"/.forgego/"+tool+".mod", read(t, fs, ".forgego/"+tool+"/go.mod"))
		_ = fs.WriteFile(dir+"/.forgego/"+tool+".sum", read(t, fs, ".forgego/"+tool+"/go.sum"))
		_ = fs.Remove(dir + "/.forgego/" + tool + "/go.mod")
		_ = fs.Remove(dir + "/.forgego/" + tool + "/go.sum")
	}
	if code := sync.Run(fs, dir, true); code != 1 {
		t.Fatalf("sync --check didn't report the old paths")
	}

	if code := sync.Run(fs, dir, false); code != 0 {
		t.Fatalf("sync returned %d", code)
	}
	if code := sync.Run(fs, dir, true); code != 0 {
		t.Errorf("drift remains after sync")
	}
	assertFiles(t, fs, ".forgego/task/go.mod", ".forgego/task/go.sum", ".forgego/lefthook/go.mod", ".forgego/golangci-lint/go.sum")
	refuteFiles(t, fs, ".forgego/task.mod", ".forgego/task.sum", ".forgego/lefthook.mod", ".forgego/golangci-lint.sum")
	for _, rel := range []string{"Taskfile.yml", "lefthook.yml"} {
		for _, tool := range []string{"task", "lefthook", "golangci-lint"} {
			if content := read(t, fs, rel); strings.Contains(content, ".forgego/"+tool+".mod") {
				t.Errorf("%s still runs %s from its old path:\n%s", rel, tool, content)
			}
		}
	}
}

func TestTheStarterTestOnlyGoesNextToTheStarter(t *testing.T) {
	fs := fsys.NewMemory(map[string]string{dir + "/cmd/billing-api/main.go": "package main\n\nfunc main() {}\n"})
	run(t, fs, nil)
	refuteFiles(t, fs, "cmd/billing-api/main_test.go")

	fs = fsys.NewMemory(map[string]string{dir + "/billingapi.go": "package billingapi\n"})
	run(t, fs, func(c *config.Init) { c.Mode = config.Library })
	refuteFiles(t, fs, "billingapi_test.go")

	// a re-run keeps adding it next to the untouched starter
	fs = fsys.NewMemory(nil)
	run(t, fs, nil)
	_ = fs.Remove(dir + "/cmd/billing-api/main_test.go")
	run(t, fs, nil)
	assertFiles(t, fs, "cmd/billing-api/main_test.go")
}

func TestSyncLeavesAForgegoTheProjectChoseAlone(t *testing.T) {
	fs := fsys.NewMemory(nil)
	run(t, fs, nil)
	taskfile := strings.Replace(read(t, fs, "Taskfile.yml"),
		templates.SelfCommand, "go run ./cmd/forgego", 1)
	_ = fs.Remove(dir + "/.forgego/forgego/go.mod")
	_ = fs.WriteFile(dir+"/Taskfile.yml", taskfile)
	if code := sync.Run(fs, dir, true); code != 0 {
		t.Errorf("a local forgego counted as drift")
	}
}

func TestAnOlderGoGetsAToolchainTheToolsCanRunOn(t *testing.T) {
	fs := fsys.NewMemory(map[string]string{dir + "/go.mod": "module example.com/api\n\ngo 1.25.0\n"})
	run(t, fs, func(c *config.Init) { c.DryRun = true })
	if read(t, fs, "go.mod") != "module example.com/api\n\ngo 1.25.0\n" {
		t.Error("dry run changed go.mod")
	}

	run(t, fs, nil)
	got := read(t, fs, "go.mod")
	if !strings.Contains(got, "\ngo 1.25.0\n") || !strings.Contains(got, "\ntoolchain go1.2") {
		t.Errorf("go.mod should keep go 1.25.0 and gain a toolchain:\n%s", got)
	}
	if code := sync.Run(fs, dir, true); code != 0 {
		t.Error("sync --check disagrees with init")
	}

	_ = fs.WriteFile(dir+"/go.mod", "module example.com/api\n\ngo 1.25.0\n")
	if code := sync.Run(fs, dir, true); code != 1 {
		t.Error("sync --check missed the toolchain the tools need")
	}
}

func TestASubdirectoryOfARepositoryIsAModuleInsideIt(t *testing.T) {
	fs := fsys.NewMemory(map[string]string{
		"/work/.git/config": "[remote \"origin\"]\n\turl = https://github.com/acme/platform.git\n",
	})
	cfg := config.Default("/work/services/Billing")
	if code := Init(cfg, fs); code != 0 {
		t.Fatalf("Init returned %d", code)
	}
	got, _ := fs.ReadFile("/work/services/Billing/go.mod")
	if !strings.HasPrefix(got, "module github.com/acme/platform/services/billing\n") {
		t.Errorf("go.mod = %q", got)
	}
}

func TestSecurityScanningDoesNotNeedLinting(t *testing.T) {
	fs := fsys.NewMemory(nil)
	run(t, fs, func(c *config.Init) { c.Linting = false })
	// CI's patch job runs govulncheck on every Go project
	assertFiles(t, fs, ".forgego/govulncheck/go.mod")
	if !strings.Contains(read(t, fs, "Taskfile.yml"), "  security:") {
		t.Error("no security task without linting")
	}
	refuteFiles(t, fs, ".forgego/golangci-lint/go.mod", ".golangci.yml")
}

func TestAnExistingGoModsGoVersionDrivesCIUnlessGoIsGiven(t *testing.T) {
	fs := fsys.NewMemory(map[string]string{dir + "/go.mod": "module example.com/api\n\ngo 1.25.3\n"})
	run(t, fs, func(c *config.Init) { c.Go = ""; c.Docker = true })
	// CI reads go.mod itself, so the generated workflow names no version
	if strings.Contains(read(t, fs, ".github/workflows/index.yml"), "go_version") {
		t.Error("the generated workflow pins a Go version instead of reading go.mod")
	}
	if !strings.Contains(read(t, fs, "Dockerfile"), "golang:1.25 AS build") {
		t.Error("the Dockerfile doesn't use go.mod's Go")
	}
}

func TestAnExistingProjectGetsNoStarterBesideItsOwnCode(t *testing.T) {
	fs := fsys.NewMemory(map[string]string{
		dir + "/go.mod":     "module github.com/acme/billing-api\n\ngo 1.27.0\n",
		dir + "/billing.go": "package billing\n",
	})
	run(t, fs, func(c *config.Init) { c.Mode = config.Library })
	refuteFiles(t, fs, "billingapi.go", "billingapi_test.go")
	run(t, fs, nil)
	refuteFiles(t, fs, "cmd/billing-api/main.go", "cmd/billing-api/main_test.go")
}

func TestSwitchingAFeatureOffWithForceRemovesItsTasksAndLeavesNoDrift(t *testing.T) {
	fs := fsys.NewMemory(nil)
	run(t, fs, func(c *config.Init) { c.Docker = true })
	run(t, fs, func(c *config.Init) { c.Linting, c.Testing, c.Docker, c.Force = false, false, false, true })
	taskfile := read(t, fs, "Taskfile.yml")
	for _, gone := range []string{"  lint:", "  format:", "  hooks:", "  test:", "docker:build", "GOLANGCI_LINT", "LEFTHOOK", "GOTESTSUM"} {
		if strings.Contains(taskfile, gone) {
			t.Errorf("Taskfile.yml still has %s:\n%s", gone, taskfile)
		}
	}
	if !strings.Contains(taskfile, "  build:") || !strings.Contains(taskfile, "  security:") {
		t.Error("removed tasks that are still in use")
	}
	assertFiles(t, fs, ".golangci.local.yml")
	if code := sync.Run(fs, dir, true); code != 0 {
		t.Error("sync --check reports drift after linting was switched off")
	}
}

func TestDryRunOnlyReportsRealChanges(t *testing.T) {
	fs := fsys.NewMemory(nil)
	run(t, fs, nil)
	var out strings.Builder
	console.SetOutput(&out, &out)
	defer console.SetOutput(io.Discard, io.Discard)
	run(t, fs, func(c *config.Init) { c.DryRun = true })
	if strings.Contains(out.String(), "Would refresh") {
		t.Errorf("an up-to-date project still reports refreshes:\n%s", out.String())
	}
}

func TestRepositoryFiles(t *testing.T) {
	fs := fsys.NewMemory(nil)
	run(t, fs, nil)
	if !strings.Contains(read(t, fs, ".editorconfig"), "[{*.go,go.mod,go.sum,Makefile}]\nindent_style = tab") {
		t.Error(".editorconfig doesn't indent Go with tabs")
	}
	dependabot := read(t, fs, ".github/dependabot.yml")
	for _, want := range []string{`package-ecosystem: "gomod"`, `package-ecosystem: "github-actions"`, `dependency-name: "apollogeddon/forgego"`} {
		if !strings.Contains(dependabot, want) {
			t.Errorf("dependabot.yml has no %s:\n%s", want, dependabot)
		}
	}
	if got := strings.Count(dependabot, "default-days: 3"); got != 3 {
		t.Errorf("%d of 3 updates have the cooldown", got)
	}
	if strings.Contains(dependabot, "docker") {
		t.Error("dependabot.yml updates Docker without --docker")
	}
	// a module path off github.com names no owner
	refuteFiles(t, fs, ".github/CODEOWNERS")

	docker := fsys.NewMemory(nil)
	run(t, docker, func(c *config.Init) { c.Docker = true })
	if !strings.Contains(read(t, docker, ".github/dependabot.yml"), `package-ecosystem: "docker"`) {
		t.Error("dependabot.yml doesn't update Docker with --docker")
	}
}

func TestCodeownersNamesTheGitHubOwner(t *testing.T) {
	fs := fsys.NewMemory(map[string]string{
		dir + "/.git/config": "[remote \"origin\"]\n\turl = git@github.com:Acme/Billing-API.git\n",
	})
	run(t, fs, nil)
	if got := read(t, fs, ".github/CODEOWNERS"); !strings.Contains(got, "* @acme\n") {
		t.Errorf("CODEOWNERS = %q", got)
	}
}

func TestSyncMovesAGoRunForgegoToThePin(t *testing.T) {
	fs := fsys.NewMemory(nil)
	run(t, fs, nil)
	// an earlier forgego ran itself with go run, at the version in the Taskfile
	_ = fs.WriteFile(dir+"/Taskfile.yml", strings.Replace(read(t, fs, "Taskfile.yml"),
		templates.SelfCommand, "go run github.com/apollogeddon/forgego/cmd/forgego@v1.2.0", 1))
	_ = fs.Remove(dir + "/.forgego/forgego/go.mod")
	if code := sync.Run(fs, dir, true); code != 1 {
		t.Fatal("sync --check didn't report the go run forgego")
	}
	if fs.Exists(dir + "/.forgego/forgego/go.mod") {
		t.Error("sync --check pinned forgego")
	}

	version.Set("v1.3.0")
	defer version.Set("v1.2.3")
	if code := sync.Run(fs, dir, false); code != 0 {
		t.Fatalf("sync returned %d", code)
	}
	if !strings.Contains(read(t, fs, "Taskfile.yml"), "FORGEGO: "+templates.SelfCommand) {
		t.Error("the Taskfile still runs forgego with go run")
	}
	// the forgego doing the sync is the one the project moved to
	if !strings.Contains(read(t, fs, ".forgego/forgego/go.mod"), "forgego v1.3.0\n") {
		t.Error("sync didn't pin the forgego it is")
	}
	if code := sync.Run(fs, dir, true); code != 0 {
		t.Error("drift remains after sync")
	}
}

func TestAMissingPinIsDrift(t *testing.T) {
	fs := fsys.NewMemory(nil)
	run(t, fs, nil)
	_ = fs.Remove(dir + "/.forgego/forgego/go.mod")
	if code := sync.Run(fs, dir, true); code != 1 {
		t.Fatal("sync --check missed the missing pin")
	}
	if code := sync.Run(fs, dir, false); code != 0 {
		t.Fatalf("sync returned %d", code)
	}
	assertFiles(t, fs, ".forgego/forgego/go.mod")
}

func TestABuildThatIsntAReleaseRunsForgegoWithGoRun(t *testing.T) {
	version.Set("latest")
	defer version.Set("v1.2.3")
	fs := fsys.NewMemory(nil)
	run(t, fs, nil)
	if !strings.Contains(read(t, fs, "Taskfile.yml"), "FORGEGO: go run github.com/apollogeddon/forgego/cmd/forgego@latest") {
		t.Error("a development build should run forgego with go run")
	}
	refuteFiles(t, fs, ".forgego/forgego/go.mod")
}

func TestDependabotProposesForgegosPin(t *testing.T) {
	fs := fsys.NewMemory(nil)
	run(t, fs, nil)
	var cfg struct {
		Updates []struct {
			Ecosystem string `yaml:"package-ecosystem"`
			Directory string
			Schedule  struct{ Interval string }
			Allow     []struct {
				Name string `yaml:"dependency-name"`
				Type string `yaml:"dependency-type"`
			}
			Cooldown struct {
				Days int `yaml:"default-days"`
			}
		}
	}
	if err := yaml.Unmarshal([]byte(read(t, fs, ".github/dependabot.yml")), &cfg); err != nil {
		t.Fatal(err)
	}
	for _, u := range cfg.Updates {
		if u.Directory != "/.forgego/forgego" {
			continue
		}
		// the same schedule and cooldown as every other update; only forgego is proposed
		if u.Ecosystem != "gomod" || u.Schedule.Interval != "weekly" || u.Cooldown.Days != 3 ||
			len(u.Allow) != 1 || u.Allow[0].Name != "github.com/apollogeddon/forgego" || u.Allow[0].Type != "all" {
			t.Errorf("forgego's pin entry = %+v", u)
		}
		return
	}
	t.Error("Dependabot doesn't watch forgego's pin")
}
