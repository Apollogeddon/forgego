package templates

import (
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

var workflowsDir = filepath.Join("..", "..", ".github", "workflows")

type step struct {
	Name string            `yaml:"name"`
	ID   string            `yaml:"id"`
	If   string            `yaml:"if"`
	Uses string            `yaml:"uses"`
	Run  string            `yaml:"run"`
	Env  map[string]string `yaml:"env"`
	With map[string]string `yaml:"with"`
}

type stringList []string

// needs is a job id or a list of them
func (l *stringList) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.ScalarNode {
		*l = []string{n.Value}
		return nil
	}
	var list []string
	err := n.Decode(&list)
	*l = list
	return err
}

type job struct {
	Name        string            `yaml:"name"`
	If          string            `yaml:"if"`
	Uses        string            `yaml:"uses"`
	Needs       stringList        `yaml:"needs"`
	With        map[string]any    `yaml:"with"`
	Secrets     map[string]any    `yaml:"secrets"`
	Permissions map[string]string `yaml:"permissions"`
	Outputs     map[string]string `yaml:"outputs"`
	Strategy    struct {
		Matrix map[string]string `yaml:"matrix"`
	} `yaml:"strategy"`
	Steps []step `yaml:"steps"`
}

type callInput struct {
	Default  any  `yaml:"default"`
	Required bool `yaml:"required"`
}

type workflow struct {
	Name string `yaml:"name"`
	On   struct {
		WorkflowCall *struct {
			Inputs  map[string]callInput      `yaml:"inputs"`
			Outputs map[string]any            `yaml:"outputs"`
			Secrets map[string]map[string]any `yaml:"secrets"`
		} `yaml:"workflow_call"`
		RepositoryDispatch struct {
			Types []string `yaml:"types"`
		} `yaml:"repository_dispatch"`
	} `yaml:"on"`
	Concurrency map[string]any `yaml:"concurrency"`
	Jobs        map[string]job `yaml:"jobs"`
}

func parse(t *testing.T, name, text string) workflow {
	t.Helper()
	var w workflow
	if err := yaml.Unmarshal([]byte(text), &w); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return w
}

func readWorkflow(t *testing.T, name string) workflow {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(workflowsDir, name))
	if err != nil {
		t.Fatal(err)
	}
	return parse(t, name, string(b))
}

func readText(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(workflowsDir, name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// The spec's stages: a run page reads <pipeline> / <stage>, the same words in every forge.
var stages = map[string][]string{
	"library": {"lint", "scan", "test", "test (windows)", "test (macos)", "build"},
	"service": {"lint", "scan", "test", "test (windows)", "test (macos)", "build"},
	"website": {"lint", "scan", "test", "build", "check"},
	"docker":  {"build", "test", "push"},
	"version": {"release"},
	"publish": {"package", "assets", "attest"},
	"pages":   {"deploy", "check"},
	"patch":   {"update"},
	"merge":   {"review", "squash"},
}

var variant = regexp.MustCompile(` \([^)]+\)$`)

func pipelineFiles(t *testing.T) (pipelines, roots []string) {
	t.Helper()
	entries, err := os.ReadDir(workflowsDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			roots = append(roots, e.Name())
		} else {
			pipelines = append(pipelines, e.Name())
		}
	}
	return pipelines, roots
}

func TestThePipelinesAndOneRootWorkflow(t *testing.T) {
	pipelines, roots := pipelineFiles(t)
	var names []string
	for _, p := range pipelines {
		names = append(names, strings.TrimSuffix(p, ".yml"))
	}
	if want := slices.Sorted(maps.Keys(stages)); !slices.Equal(names, want) {
		t.Errorf("pipelines = %v, want %v", names, want)
	}
	if !slices.Equal(roots, []string{".index.yml"}) || readWorkflow(t, ".index.yml").Name != "index" {
		t.Errorf("root workflows = %v, want .index.yml named index", roots)
	}
}

// A pipeline is only ever called, and calls no other: that would add a third level to the run page's names.
// Each job is named with a stage from the spec, its id the stage in kebab case.
func TestPipelinesAreFlatAndNamedByStage(t *testing.T) {
	pipelines, _ := pipelineFiles(t)
	for _, file := range pipelines {
		pipeline := strings.TrimSuffix(file, ".yml")
		text := readText(t, file)
		var on map[string]any
		var raw struct {
			On map[string]any `yaml:"on"`
		}
		if err := yaml.Unmarshal([]byte(text), &raw); err != nil {
			t.Fatal(err)
		}
		on = raw.On
		if len(on) != 1 || on["workflow_call"] == nil {
			t.Errorf("%s: triggered by %v, want workflow_call alone", file, slices.Collect(maps.Keys(on)))
		}
		w := parse(t, file, text)
		if w.Name != pipeline {
			t.Errorf("%s: name = %q", file, w.Name)
		}
		for id, j := range w.Jobs {
			if j.Uses != "" {
				t.Errorf("%s: %s calls %s", file, id, j.Uses)
			}
			stage := variant.ReplaceAllString(j.Name, "")
			if !slices.Contains(stages[pipeline], j.Name) && !slices.Contains(stages[pipeline], stage) {
				t.Errorf("%s: %s is named %q, not a stage of %s", file, id, j.Name, pipeline)
			}
			// a skipped job shows its name unevaluated, so a name never takes a matrix's values
			want := strings.ReplaceAll(strings.NewReplacer("(", "", ")", "").Replace(j.Name), " ", "-")
			if strings.Contains(j.Name, "${{") {
				t.Errorf("%s: %s's name %q is an expression", file, id, j.Name)
			} else if id != want {
				t.Errorf("%s: job id %s, want %s for %q", file, id, want, j.Name)
			}
			for i, s := range j.Steps {
				if s.Name == "" {
					t.Errorf("%s: %s's step %d has no name", file, id, i+1)
				}
			}
		}
	}
}

var levels = map[string]int{"": 0, "none": 0, "read": 1, "write": 2}

// expectWiring checks a root workflow's calls: each passes only inputs the pipeline declares, and all it
// requires; grants every permission the pipeline's jobs ask for, as GitHub refuses the whole run otherwise; and
// reads only outputs the pipeline declares.
func expectWiring(t *testing.T, caller string, w workflow, resolve func(uses string) (string, bool)) {
	t.Helper()
	for id, j := range w.Jobs {
		if j.Uses == "" {
			if id != "changes" && (id != "scorecard" || caller != ".index.yml") {
				t.Errorf("%s: %s is a job of its own; only changes, and forgego's own scorecard, may be", caller, id)
			}
			continue
		}
		file, ok := resolve(j.Uses)
		if !ok {
			t.Errorf("%s: %s calls %s", caller, id, j.Uses)
			continue
		}
		pipeline := strings.TrimSuffix(file, ".yml")
		// the root job's id is the pipeline's name, and a variant names itself in brackets
		if id == pipeline {
			if j.Name != "" {
				t.Errorf("%s: %s is named %q", caller, id, j.Name)
			}
		} else if !regexp.MustCompile(`^` + pipeline + ` \([^)]+\)$`).MatchString(j.Name) {
			t.Errorf("%s: %s calls %s but is named %q", caller, id, file, j.Name)
		}
		called := readWorkflow(t, file)
		declared := called.On.WorkflowCall.Inputs
		for in := range j.With {
			if _, ok := declared[in]; !ok {
				t.Errorf("%s: %s passes %s, which %s doesn't declare", caller, id, in, file)
			}
		}
		for in, spec := range declared {
			if _, ok := j.With[in]; spec.Required && !ok {
				t.Errorf("%s: %s misses %s's required %s", caller, id, file, in)
			}
		}
		for secret := range j.Secrets {
			if _, ok := called.On.WorkflowCall.Secrets[secret]; !ok {
				t.Errorf("%s: %s passes secret %s, which %s doesn't declare", caller, id, secret, file)
			}
		}
		for calledID, calledJob := range called.Jobs {
			for scope, level := range calledJob.Permissions {
				if levels[j.Permissions[scope]] < levels[level] {
					t.Errorf("%s: %s grants %s too little %s for %s", caller, id, file, scope, calledID)
				}
			}
		}
	}
	text, err := yaml.Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range regexp.MustCompile(`needs\.([\w-]+)\.outputs\.(\w+)`).FindAllStringSubmatch(string(text), -1) {
		file, ok := resolve(w.Jobs[m[1]].Uses)
		if !ok {
			continue
		}
		if _, ok := readWorkflow(t, file).On.WorkflowCall.Outputs[m[2]]; !ok {
			t.Errorf("%s reads %s.%s, which %s doesn't declare", caller, m[1], m[2], file)
		}
	}
}

func TestOwnIndexCallsLocalPipelines(t *testing.T) {
	w := readWorkflow(t, ".index.yml")
	expectWiring(t, ".index.yml", w, func(uses string) (string, bool) {
		return strings.CutPrefix(uses, "./.github/workflows/")
	})
	if w.Jobs["changes"].Name != "changes" {
		t.Errorf("changes is named %q", w.Jobs["changes"].Name)
	}
	if w.Jobs["service"].Uses == "" || w.Jobs["publish"].Uses == "" ||
		w.Jobs["website"].With["working_directory"] != "docs" ||
		w.Jobs["pages"].With["artifact_name"] != w.Jobs["website"].With["artifact_name"] {
		t.Errorf("forgego's own CI doesn't check, release and deploy its docs: %+v", w.Jobs)
	}
	if !slices.Equal(w.Jobs["merge"].Needs, []string{"changes", "service", "website"}) {
		t.Errorf("merge needs %v", w.Jobs["merge"].Needs)
	}
	generated := parse(t, "generated", RenderWorkflow(WorkflowOptions{Mode: "backend", Testing: true, Linting: true, Versioning: true}))
	if !maps.Equal(w.Concurrency, generated.Concurrency) {
		t.Errorf("concurrency = %v, want the generated %v", w.Concurrency, generated.Concurrency)
	}
	if w.Jobs["patch"].If != "github.event_name == 'schedule' || github.event_name == 'workflow_dispatch' || github.event_name == 'repository_dispatch'" {
		t.Errorf("patch if = %q", w.Jobs["patch"].If)
	}
}

// Every combination init can write.
func allOptions() map[string]WorkflowOptions {
	all := map[string]WorkflowOptions{}
	for _, mode := range []string{"backend", "library", "website"} {
		for _, docker := range []bool{false, true} {
			for _, tests := range []bool{false, true} {
				for _, linting := range []bool{false, true} {
					for _, versioning := range []bool{false, true} {
						o := WorkflowOptions{Mode: mode, Docker: docker, Testing: tests, Linting: linting, Versioning: versioning}
						name := mode
						for flag, on := range map[string]bool{" docker": docker, " no-testing": !tests, " no-linting": !linting, " no-versioning": !versioning} {
							if on {
								name += flag
							}
						}
						all[name] = o
					}
				}
			}
		}
	}
	return all
}

var atMain = regexp.MustCompile(`^apollogeddon/forgego/\.github/workflows/([\w-]+\.yml)@main$`)

func TestGeneratedIndexCallsForgegosPipelines(t *testing.T) {
	sent := regexp.MustCompile(`DISPATCH_TYPE: ([\w-]+)`).FindStringSubmatch(readText(t, "merge.yml"))
	if sent == nil {
		t.Fatal("merge.yml sends no repository_dispatch with a DISPATCH_TYPE")
	}
	for name, o := range allOptions() {
		w := parse(t, name, RenderWorkflow(o))
		expectWiring(t, name, w, func(uses string) (string, bool) {
			m := atMain.FindStringSubmatch(uses)
			if m == nil {
				return "", false
			}
			return m[1], true
		})
		// a merge made with GITHUB_TOKEN starts no push run, so merge.yml sends a dispatch the root must listen for
		if !slices.Equal(w.On.RepositoryDispatch.Types, []string{sent[1]}) {
			t.Errorf("%s listens for repository_dispatch %v, want [%s]", name, w.On.RepositoryDispatch.Types, sent[1])
		}
		// each run outside a pull request has a group of its own, so a queued run on main is never cancelled
		if w.Concurrency["group"] != "${{ github.workflow }}-${{ github.head_ref || github.run_id }}" || w.Concurrency["cancel-in-progress"] != true {
			t.Errorf("%s: concurrency = %v", name, w.Concurrency)
		}
		if w.Jobs["patch"].If != "${{ github.event_name == 'schedule' || github.event_name == 'workflow_dispatch' || github.event_name == 'repository_dispatch' }}" {
			t.Errorf("%s: patch if = %q", name, w.Jobs["patch"].If)
		}
		for id, j := range w.Jobs {
			// only GitHub Pages and a service's attestations need an OIDC token
			if _, ok := j.Permissions["id-token"]; ok && id != "pages" && (id != "publish" || o.Mode != "backend") {
				t.Errorf("%s: %s asks for id-token", name, id)
			}
			// only the release GitHub App's key, which is empty unless the repository sets it
			if j.Secrets != nil && (id != "version" || !maps.Equal(j.Secrets, map[string]any{"APP_PRIVATE_KEY": "${{ secrets.RELEASE_APP_PRIVATE_KEY }}"})) {
				t.Errorf("%s: %s passes secrets %v", name, id, j.Secrets)
			}
		}
	}
}

// release-please opens its pull request with a GitHub App's token when the repository sets one, as one opened
// with GITHUB_TOKEN runs no checks, and with GITHUB_TOKEN otherwise; and the release pull request's checks run.
func TestVersionReleasesWithAnOptionalAppToken(t *testing.T) {
	w := readWorkflow(t, "version.yml")
	if in := w.On.WorkflowCall.Inputs["app_client_id"]; in.Default != "" || in.Required {
		t.Errorf("app_client_id = %+v", in)
	}
	if secret := w.On.WorkflowCall.Secrets["APP_PRIVATE_KEY"]; secret == nil || secret["required"] != false {
		t.Errorf("APP_PRIVATE_KEY = %v, want it declared and optional", secret)
	}
	release := w.Jobs["release"]
	app := stepNamed(release, "Create the GitHub App's token")
	if app.ID != "app" || app.If != "inputs.app_client_id != ''" || !strings.HasPrefix(app.Uses, "actions/create-github-app-token@") ||
		app.With["client-id"] != "${{ inputs.app_client_id }}" || app.With["private-key"] != "${{ secrets.APP_PRIVATE_KEY }}" {
		t.Errorf("the app token step = %+v", app)
	}
	if token := stepNamed(release, "Release").With["token"]; token != "${{ steps.app.outputs.token || secrets.GITHUB_TOKEN }}" {
		t.Errorf("release-please's token = %q", token)
	}
	generated := parse(t, "backend", RenderWorkflow(WorkflowOptions{Mode: "backend", Versioning: true})).Jobs["version"]
	for name, j := range map[string]job{".index.yml": readWorkflow(t, ".index.yml").Jobs["version"], "generated": generated} {
		if j.With["app_client_id"] != "${{ vars.RELEASE_APP_CLIENT_ID }}" || j.Secrets["APP_PRIVATE_KEY"] != "${{ secrets.RELEASE_APP_PRIVATE_KEY }}" {
			t.Errorf("%s: version = %+v", name, j)
		}
	}
	if strings.Contains(readText(t, ".index.yml"), "release-please--") {
		t.Error(".index.yml still skips the release pull request's checks")
	}
	for name, o := range allOptions() {
		if strings.Contains(RenderWorkflow(o), "release-please--") {
			t.Errorf("%s skips the release pull request's checks", name)
		}
	}
}

func TestGeneratedJobsPerMode(t *testing.T) {
	for _, c := range []struct {
		o    WorkflowOptions
		jobs []string
	}{
		{WorkflowOptions{Mode: "backend", Versioning: true}, []string{"version", "service", "publish", "patch", "merge"}},
		{WorkflowOptions{Mode: "library", Versioning: true}, []string{"version", "library", "publish", "patch", "merge"}},
		{WorkflowOptions{Mode: "website", Versioning: true}, []string{"version", "website", "pages", "patch", "merge"}},
		{WorkflowOptions{Mode: "backend", Docker: true, Versioning: true}, []string{"version", "service", "docker", "publish", "patch", "merge"}},
		{WorkflowOptions{Mode: "backend"}, []string{"service", "patch", "merge"}},
		{WorkflowOptions{Mode: "website"}, []string{"website", "pages", "patch", "merge"}},
	} {
		var w yaml.Node
		if err := yaml.Unmarshal([]byte(RenderWorkflow(c.o)), &w); err != nil {
			t.Fatal(err)
		}
		var ids []string
		root := w.Content[0]
		for i := 0; i < len(root.Content); i += 2 {
			if root.Content[i].Value == "jobs" {
				for k := 0; k < len(root.Content[i+1].Content); k += 2 {
					ids = append(ids, root.Content[i+1].Content[k].Value)
				}
			}
		}
		if !slices.Equal(ids, c.jobs) {
			t.Errorf("%+v: jobs %v, want %v", c.o, ids, c.jobs)
		}
	}
}

func TestGeneratedIndexGatesOnTheChecksAndTheRelease(t *testing.T) {
	jobs := parse(t, "backend", RenderWorkflow(WorkflowOptions{Mode: "backend", Docker: true, Testing: true, Linting: true, Versioning: true})).Jobs
	if !strings.Contains(jobs["service"].If, "github.event_name == 'pull_request' || needs.version.outputs.new_release_published == 'true'") {
		t.Errorf("service if = %q; want it on pull requests and releases", jobs["service"].If)
	}
	if !strings.Contains(jobs["publish"].If, "needs.service.result == 'success'") || jobs["publish"].With["tag_name"] != "${{ needs.version.outputs.tag_name }}" {
		t.Errorf("publish = %+v", jobs["publish"])
	}
	if jobs["docker"].With["push"] != "${{ needs.version.outputs.new_release_published == 'true' }}" {
		t.Errorf("docker push = %v; want it only on a release", jobs["docker"].With["push"])
	}
	if !slices.Equal(jobs["merge"].Needs, []string{"service", "docker"}) ||
		jobs["merge"].With["checks_passed"] != "${{ !contains(needs.*.result, 'failure') && !contains(needs.*.result, 'cancelled') }}" {
		t.Errorf("merge = %+v; want it to wait for every check", jobs["merge"])
	}

	// a library's publish only fetches through the proxy, so its grant stays read-only
	library := parse(t, "library", RenderWorkflow(WorkflowOptions{Mode: "library", Versioning: true})).Jobs["publish"]
	if library.With["package"] != true || library.With["assets"] != false || !maps.Equal(library.Permissions, map[string]string{"contents": "read"}) {
		t.Errorf("library publish = %+v", library)
	}

	disabled := parse(t, "disabled", RenderWorkflow(WorkflowOptions{Mode: "backend"})).Jobs["service"]
	if disabled.With["run_tests"] != false || disabled.With["lint"] != false || disabled.If != "${{ github.event_name != 'schedule' }}" {
		t.Errorf("service = %+v; want the disabled features off, and the checks on every push", disabled)
	}

	unversioned := parse(t, "website", RenderWorkflow(WorkflowOptions{Mode: "website"})).Jobs["pages"]
	if unversioned.If != "${{ github.ref == 'refs/heads/main' && github.event_name != 'pull_request' && github.event.repository.private == false }}" {
		t.Errorf("pages if = %q; want every change to a public repository's main deployed", unversioned.If)
	}
}

// GitHub Free attests builds and serves Pages only for a public repository. A scheduled run has no
// repository.private, and `== false` reads that as private, which always works; `!private` wouldn't.
func TestFreePlanExtrasNeedAPublicRepository(t *testing.T) {
	const public = "github.event.repository.private == false"
	own := readWorkflow(t, ".index.yml").Jobs
	backend := parse(t, "backend", RenderWorkflow(WorkflowOptions{Mode: "backend", Versioning: true})).Jobs
	for name, publish := range map[string]job{".index.yml": own["publish"], "backend": backend["publish"]} {
		if publish.With["attest"] != "${{ "+public+" }}" ||
			publish.Permissions["id-token"] != "write" || publish.Permissions["attestations"] != "write" {
			t.Errorf("%s: publish = %+v; want it to attest a public repository's release", name, publish)
		}
	}
	website := parse(t, "website", RenderWorkflow(WorkflowOptions{Mode: "website", Versioning: true})).Jobs
	for name, pages := range map[string]job{".index.yml": own["pages"], "website": website["pages"]} {
		if !strings.Contains(pages.If, " && "+public) {
			t.Errorf("%s: pages if = %q; want it only from a public repository", name, pages.If)
		}
	}
	for name, o := range allOptions() {
		if text := RenderWorkflow(o); strings.Contains(text, "!github.event.repository.private") {
			t.Errorf("%s reads a missing repository.private as public", name)
		}
	}
	if strings.Contains(readText(t, ".index.yml"), "!github.event.repository.private") {
		t.Error(".index.yml reads a missing repository.private as public")
	}
}

func TestCallsForgego(t *testing.T) {
	for content, want := range map[string]bool{
		"uses: apollogeddon/forgego/.github/workflows/service.yml@main": true,
		"uses: Apollogeddon/forgego/.github/workflows/library.yml@v1":   true,
		"uses: someone/else/.github/workflows/ci.yml@v1":                false,
		"runs-on: ubuntu-latest":                                        false,
	} {
		if got := CallsForgego(content); got != want {
			t.Errorf("CallsForgego(%q) = %v", content, got)
		}
	}
}

func legacyFixture(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "legacy", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestLegacyWorkflowsAreKnown(t *testing.T) {
	const old = ".github/workflows/index.yml"
	for _, name := range []string{"main-backend.yml", "main-library.yml", "main-website.yml", "v1.0.0-backend.yml", "v1.6.1-library.yml"} {
		if !IsLegacyWorkflow(old, legacyFixture(t, name)) {
			t.Errorf("%s isn't known", name)
		}
	}
	// line endings, trailing spaces, blank lines and the Go version an old release copied from go.mod
	text := strings.ReplaceAll(legacyFixture(t, "v1.0.0-backend.yml"), "go_version: '1.27'", "go_version: '1.25.3'")
	if !IsLegacyWorkflow(old, strings.ReplaceAll(text, "\n", "  \r\n\r\n")) {
		t.Error("a reformatted old workflow isn't known")
	}
	changed := strings.Replace(legacyFixture(t, "main-library.yml"), "    with:\n", "    with:\n      runs_on: xerus\n", 1)
	if changed == legacyFixture(t, "main-library.yml") {
		t.Fatal("the fixture has no with:")
	}
	if IsLegacyWorkflow(old, changed) || IsLegacyWorkflow(".github/workflows/build.yml", legacyFixture(t, "main-library.yml")) {
		t.Error("a changed workflow, or one under another name, is known")
	}
	// the new root workflow is never mistaken for an old one
	for name, o := range allOptions() {
		if IsLegacyWorkflow(old, RenderWorkflow(o)) {
			t.Errorf("%s: the new workflow counts as an old one", name)
		}
	}
}

// A module in a subdirectory keeps its release-please config in its own .github/, and
// release-please prefixes its outputs with its path.
func TestVersionUsesWorkingDirectory(t *testing.T) {
	release := readWorkflow(t, "version.yml").Jobs["release"]
	for _, s := range release.Steps {
		if !strings.HasPrefix(s.Uses, "googleapis/release-please-action") {
			continue
		}
		for key, file := range map[string]string{"config-file": "release.json", "manifest-file": ".release.json"} {
			if want := "format('{0}/.github/" + file + "', inputs.working_directory)"; !strings.Contains(s.With[key], want) {
				t.Errorf("%s = %q, want it to contain %s", key, s.With[key], want)
			}
		}
	}
	for _, key := range []string{"release_created", "version", "tag_name"} {
		if want := "format('{0}--" + key + "', inputs.working_directory)"; !strings.Contains(release.Outputs[key], want) {
			t.Errorf("output %s = %q, want it to contain %s", key, release.Outputs[key], want)
		}
	}
	// releases_created is true when any package is released, not just this one,
	// so it only counts for the root package
	_, published, _ := strings.Cut(readText(t, "version.yml"), "new_release_published:")
	published, _, _ = strings.Cut(published, "\n      version:")
	beforeRoot, rootClause, found := strings.Cut(published, "inputs.working_directory == '.' && ")
	if !found || strings.Contains(beforeRoot, "releases_created") ||
		!strings.HasPrefix(rootClause, "(jobs.release.outputs.releases_created") {
		t.Errorf("new_release_published counts releases_created beyond the root package: %s", published)
	}
}

func stepNamed(j job, name string) step {
	for _, s := range j.Steps {
		if s.Name == name {
			return s
		}
	}
	return step{}
}

// CODEOWNERS requests nothing in a private repository on a free plan, so the release pull request and
// Dependabot's ask themselves, and a review request never fails the pipeline.
func TestBotPullRequestsRequestAReview(t *testing.T) {
	release := stepNamed(readWorkflow(t, "version.yml").Jobs["release"], "Request a review")
	if release.If != "steps.release.outputs.pr != ''" || release.Env["REVIEWERS"] != "${{ inputs.reviewers }}" {
		t.Errorf("version.yml's review step = %+v", release)
	}
	review := readWorkflow(t, "merge.yml").Jobs["review"]
	// before the checks, and whatever they found, so a failing update reaches the reviewer too
	if len(review.Needs) != 0 || strings.Contains(review.If, "checks_passed") ||
		!strings.Contains(review.If, "github.event.pull_request.user.login == 'dependabot[bot]'") {
		t.Errorf("merge.yml's review = %+v", review)
	}
	for _, s := range []step{release, stepNamed(review, "Request a review")} {
		if !strings.Contains(s.With["script"], "catch (error)") || !strings.Contains(s.With["script"], "core.warning") {
			t.Error("a review request can fail the pipeline")
		}
	}
}

// The Windows and macOS stages and the native builds are opt-in: their defaults run nothing, and an empty
// matrix would fail every caller's workflow, so the native builds are skipped before their matrix expands.
func TestRunnerStagesDefaultToNothing(t *testing.T) {
	for _, pipeline := range []string{"library.yml", "service.yml"} {
		w := readWorkflow(t, pipeline)
		for _, system := range []string{"windows", "macos"} {
			in := "test_" + system
			if w.On.WorkflowCall.Inputs[in].Default != "" {
				t.Errorf("%s: %s defaults to %v", pipeline, in, w.On.WorkflowCall.Inputs[in].Default)
			}
			j := w.Jobs["test-"+system]
			if j.If != "inputs."+in+" != ''" || !slices.Equal(j.Needs, []string{"lint", "scan"}) {
				t.Errorf("%s: test-%s = if %q, needs %v", pipeline, system, j.If, j.Needs)
			}
		}
	}
	// a generated project spends no Windows or macOS minutes unless it asks; forgego's own CI tests on both
	for name, o := range allOptions() {
		if text := RenderWorkflow(o); strings.Contains(text, "test_windows") || strings.Contains(text, "test_macos") {
			t.Errorf("%s tests on Windows or macOS", name)
		}
	}
	if own := readWorkflow(t, ".index.yml").Jobs["service"].With; own["test_windows"] != "windows-latest" || own["test_macos"] != "macos-latest" {
		t.Errorf("forgego's own service = %v, want Windows and macOS", own)
	}
	publish := readWorkflow(t, "publish.yml")
	if publish.On.WorkflowCall.Inputs["native_builds"].Default != "[]" {
		t.Errorf("native_builds defaults to %v", publish.On.WorkflowCall.Inputs["native_builds"].Default)
	}
	native := publish.Jobs["assets-native"]
	if !strings.Contains(native.If, "inputs.native_builds != '[]'") || !strings.Contains(native.If, "inputs.native_builds != ''") ||
		native.Strategy.Matrix["target"] != "${{ fromJSON(inputs.native_builds) }}" {
		t.Errorf("assets-native = if %q, matrix %v", native.If, native.Strategy.Matrix)
	}
	// a published release may be immutable, so everything is attached before assets publishes the draft
	if assets := publish.Jobs["assets"]; !slices.Equal(assets.Needs, []string{"assets-native"}) ||
		!strings.Contains(assets.If, "needs.assets-native.result != 'failure'") {
		t.Errorf("assets = needs %v, if %q", assets.Needs, assets.If)
	}
}

// publish's package and native builds ask for only what they need; assets and attest ask for nothing and take the
// caller's grant, so a library's caller, which turns both off, never grants contents: write or an OIDC token.
func TestPublishCapsItsPermissions(t *testing.T) {
	jobs := readWorkflow(t, "publish.yml").Jobs
	for _, id := range []string{"package", "assets-native"} {
		if !maps.Equal(jobs[id].Permissions, map[string]string{"contents": "read"}) {
			t.Errorf("%s permissions = %v", id, jobs[id].Permissions)
		}
	}
	for _, id := range []string{"assets", "attest"} {
		if jobs[id].Permissions != nil {
			t.Errorf("%s permissions = %v, want the caller's", id, jobs[id].Permissions)
		}
	}
	if !strings.Contains(jobs["attest"].If, "inputs.attest") || !slices.Equal(jobs["attest"].Needs, []string{"assets"}) {
		t.Errorf("attest = if %q, needs %v", jobs["attest"].If, jobs["attest"].Needs)
	}
	if jobs["package"].If != "inputs.package" || !strings.Contains(jobs["assets"].If, "inputs.assets") {
		t.Errorf("package if = %q, assets if = %q", jobs["package"].If, jobs["assets"].If)
	}
}

// The site checks are opt-in, and check the build the deploy ships.
func TestSiteChecksAreOptIn(t *testing.T) {
	w := readWorkflow(t, "website.yml")
	for name, want := range map[string]any{
		"check_links": false, "check_external_links": false, "link_check_args": "",
		"check_accessibility": false, "accessibility_paths": "", "accessibility_impact": "serious", "base_path": "",
	} {
		if got := w.On.WorkflowCall.Inputs[name].Default; got != want {
			t.Errorf("%s defaults to %v, want %v", name, got, want)
		}
	}
	check := w.Jobs["check"]
	if check.If != "inputs.check_links || inputs.check_accessibility" || !slices.Equal(check.Needs, []string{"build"}) {
		t.Errorf("check = if %q, needs %v", check.If, check.Needs)
	}
}

// Signing is opt-in: by default assets imports no key.
func TestReleaseSigningIsOptIn(t *testing.T) {
	call := readWorkflow(t, "publish.yml").On.WorkflowCall
	if call.Inputs["gpg_sign"].Default != false {
		t.Errorf("gpg_sign = %v", call.Inputs["gpg_sign"])
	}
	for _, secret := range []string{"GPG_PRIVATE_KEY", "GPG_PASSPHRASE"} {
		if required, _ := call.Secrets[secret]["required"].(bool); call.Secrets[secret] == nil || required {
			t.Errorf("secret %s = %v, want it declared and optional", secret, call.Secrets[secret])
		}
	}
	for _, s := range readWorkflow(t, "publish.yml").Jobs["assets"].Steps {
		if (strings.HasPrefix(s.Uses, "crazy-max/ghaction-import-gpg") || strings.Contains(s.Run, "GPG_FINGERPRINT")) &&
			s.If != "inputs.gpg_sign" {
			t.Errorf("assets step %q runs without gpg_sign: if = %q", s.Name, s.If)
		}
	}
}

// A release is built with exactly the GoReleaser the snapshot task tried it with.
func TestReleaseRunsTheSnapshotsGoReleaser(t *testing.T) {
	found := 0
	pipelines, _ := pipelineFiles(t)
	for _, file := range pipelines {
		for _, j := range readWorkflow(t, file).Jobs {
			for _, s := range j.Steps {
				if strings.HasPrefix(s.Uses, "goreleaser/goreleaser-action@") {
					found++
					if file != "publish.yml" || s.With["version"] != GoreleaserVersion {
						t.Errorf("%s: goreleaser-action version = %q, want %q in publish.yml", file, s.With["version"], GoreleaserVersion)
					}
				}
			}
		}
	}
	if found != 1 {
		t.Fatalf("%d goreleaser-action steps, want publish.yml's one", found)
	}
}

// The author alone would merge someone else's push to Dependabot's branch, and the actor alone a pull request
// someone else opened that Dependabot then updated; and nothing merges before the checks pass.
func TestSquashNeedsDependabotAndPassingChecks(t *testing.T) {
	squash := readWorkflow(t, "merge.yml").Jobs["squash"]
	for _, want := range []string{
		"inputs.checks_passed && ", "github.event.pull_request.user.login == 'dependabot[bot]' && ", "github.actor == 'dependabot[bot]' && ",
		"github.event.pull_request.head.repo.full_name == github.repository && ",
	} {
		if !strings.Contains(squash.If, want) || strings.Contains(squash.If, "||") {
			t.Errorf("squash if = %q, want it to need %q", squash.If, want)
		}
	}
	if s := stepNamed(squash, "Enable auto-merge"); s.Env["DISPATCH_TYPE"] != "auto-merged" || !strings.Contains(s.With["script"], "createDispatchEvent") {
		t.Errorf("squash doesn't send the auto-merged dispatch: %+v", s)
	}
}

// Only a minor or patch update merges unseen, and only at the commit whose checks passed.
func TestSquashMergesOnlyMinorUpdatesAtTheCheckedCommit(t *testing.T) {
	squash := readWorkflow(t, "merge.yml").Jobs["squash"]
	if first := squash.Steps[0]; first.ID != "metadata" || !strings.HasPrefix(first.Uses, "dependabot/fetch-metadata@") ||
		first.With["github-token"] != "${{ secrets.GITHUB_TOKEN }}" {
		t.Errorf("squash's first step = %+v, want dependabot/fetch-metadata", first)
	}
	merge := stepNamed(squash, "Enable auto-merge")
	script := merge.With["script"]
	if merge.Env["UPDATE_TYPE"] != "${{ steps.metadata.outputs.update-type }}" {
		t.Errorf("UPDATE_TYPE = %q", merge.Env["UPDATE_TYPE"])
	}
	for _, want := range []string{
		"'version-update:semver-minor', 'version-update:semver-patch'", "core.notice(",
		"sha: pr.head.sha", "expectedHeadOid: $oid", "oid: pr.head.sha",
	} {
		if !strings.Contains(script, want) {
			t.Errorf("the squash script doesn't contain %q", want)
		}
	}
	if notice, merges := strings.Index(script, "core.notice("), strings.Index(script, "enablePullRequestAutoMerge"); notice > merges {
		t.Error("the squash script merges before it checks the update type")
	}
}

// Dependabot's pull request lints against forgego's files refreshed only when Dependabot opened it and made its
// latest change; anyone else's push to it is checked as it is.
func TestLintRefreshesOnlyDependabotsOwnPullRequests(t *testing.T) {
	const dependabot = "github.event.pull_request.user.login == 'dependabot[bot]' && github.actor == 'dependabot[bot]'"
	for _, pipeline := range []string{"library.yml", "service.yml", "website.yml"} {
		lint := readWorkflow(t, pipeline).Jobs["lint"]
		if check := stepNamed(lint, "Check forgego's files are current"); !strings.HasSuffix(check.If, " && !("+dependabot+")") {
			t.Errorf("%s: sync-check if = %q", pipeline, check.If)
		}
		if refresh := stepNamed(lint, "Refresh forgego's files"); !strings.HasSuffix(refresh.If, " && "+dependabot) {
			t.Errorf("%s: sync if = %q", pipeline, refresh.If)
		}
	}
}

// patch refreshes forgego's files through a pull request rather than pushing to main, with the token handed to
// the push alone; vulnerable modules are left to Dependabot security updates.
func TestPatchProposesTheRefreshedFiles(t *testing.T) {
	text := readText(t, "patch.yml")
	update := readWorkflow(t, "patch.yml").Jobs["update"]
	if checkout := stepNamed(update, "Check out"); checkout.With["persist-credentials"] != "false" {
		t.Errorf("checkout = %+v, want no persisted credentials", checkout)
	}
	push := stepNamed(update, "Push forgego's files")
	for _, want := range []string{"branch=forgego/sync", "push --force origin \"HEAD:refs/heads/$branch\"", "http.https://github.com/.extraheader"} {
		if !strings.Contains(push.Run, want) {
			t.Errorf("the push doesn't run %q", want)
		}
	}
	if pr := stepNamed(update, "Open a pull request"); !strings.Contains(pr.With["script"], "title: 'chore(deps): refresh the files forgego manages'") ||
		!strings.Contains(pr.With["script"], "pulls.list(") {
		t.Errorf("the pull request step = %+v", pr)
	}
	if strings.Contains(push.Run, "main") {
		t.Errorf("the push names main: %q", push.Run)
	}
	for _, gone := range []string{"git-auto-commit-action", "govulncheck", "go get"} {
		if strings.Contains(text, gone) {
			t.Errorf("patch.yml still has %q", gone)
		}
	}
	if !maps.Equal(update.Permissions, map[string]string{"contents": "write", "pull-requests": "write"}) {
		t.Errorf("update permissions = %v", update.Permissions)
	}
}

// Dependency review and Scorecard are free only for a public repository, so they run only there.
func TestPublicOnlyExtras(t *testing.T) {
	for _, pipeline := range []string{"library.yml", "service.yml", "website.yml"} {
		review := stepNamed(readWorkflow(t, pipeline).Jobs["scan"], "Review the dependency changes")
		if !strings.HasPrefix(review.Uses, "actions/dependency-review-action@") || review.With["fail-on-severity"] != "high" ||
			review.If != "github.event_name == 'pull_request' && github.event.repository.private == false" {
			t.Errorf("%s: dependency review = %+v", pipeline, review)
		}
	}
	w := readWorkflow(t, ".index.yml")
	scorecard := w.Jobs["scorecard"]
	if !maps.Equal(scorecard.Permissions, map[string]string{"contents": "read", "security-events": "write", "actions": "read"}) {
		t.Errorf("scorecard permissions = %v", scorecard.Permissions)
	}
	if !strings.Contains(scorecard.If, "github.event.repository.private == false") || !strings.Contains(scorecard.If, "github.event_name == 'schedule'") ||
		!strings.Contains(scorecard.If, "github.ref == 'refs/heads/main'") {
		t.Errorf("scorecard if = %q", scorecard.If)
	}
	score := stepNamed(scorecard, "Score with Scorecard")
	if !strings.HasPrefix(score.Uses, "ossf/scorecard-action@") || score.With["publish_results"] != "false" {
		t.Errorf("scorecard step = %+v", score)
	}
	if upload := stepNamed(scorecard, "Upload the results"); !strings.HasPrefix(upload.Uses, "github/codeql-action/upload-sarif@") ||
		upload.With["sarif_file"] != score.With["results_file"] {
		t.Errorf("upload step = %+v", upload)
	}
	// a generated project scores nothing of its own
	for name, o := range allOptions() {
		if strings.Contains(RenderWorkflow(o), "scorecard") {
			t.Errorf("%s runs scorecard", name)
		}
	}
}
