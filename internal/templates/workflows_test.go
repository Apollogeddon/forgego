package templates

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

type callerJob struct {
	Uses string         `yaml:"uses"`
	With map[string]any `yaml:"with"`
}

type reusable struct {
	On struct {
		WorkflowCall struct {
			Inputs  map[string]any `yaml:"inputs"`
			Outputs map[string]any `yaml:"outputs"`
		} `yaml:"workflow_call"`
	} `yaml:"on"`
}

func readReusable(t *testing.T, name string) reusable {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", name))
	if err != nil {
		t.Fatal(err)
	}
	var w reusable
	if err := yaml.Unmarshal(b, &w); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return w
}

// The generated CI calls forgego's workflows remotely, where actionlint can't check its inputs.
func TestGeneratedWorkflowsOnlyPassDeclaredInputs(t *testing.T) {
	for _, mode := range []string{"backend", "library", "website"} {
		for _, docker := range []bool{false, true} {
			if docker && mode == "library" {
				continue
			}
			out := RenderWorkflow(WorkflowOptions{Mode: mode, Docker: docker})
			var w struct {
				Jobs map[string]callerJob `yaml:"jobs"`
			}
			if err := yaml.Unmarshal([]byte(out), &w); err != nil {
				t.Fatalf("%s: %v\n%s", mode, err, out)
			}
			for name, job := range w.Jobs {
				file := strings.TrimSuffix(strings.TrimPrefix(job.Uses, WorkflowRepo), "@main")
				declared := readReusable(t, file).On.WorkflowCall
				for input := range job.With {
					if _, ok := declared.Inputs[input]; !ok {
						t.Errorf("%s: job %s passes %s, which %s doesn't declare", mode, name, input, file)
					}
				}
			}
		}
	}
}

// The docker job reads the pipeline's release outputs.
func TestPipelinesExposeReleaseOutputs(t *testing.T) {
	for _, file := range []string{"service.yml", "library.yml", "website.yml"} {
		outputs := readReusable(t, file).On.WorkflowCall.Outputs
		for _, want := range []string{"new_release_published", "version", "tag_name"} {
			if _, ok := outputs[want]; !ok {
				t.Errorf("%s has no %s output", file, want)
			}
		}
	}
}

// A module in a subdirectory keeps its release-please config in its own .github/, and
// release-please prefixes its outputs with its path.
func TestVersionUsesWorkingDirectory(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "version.yml"))
	if err != nil {
		t.Fatal(err)
	}
	var w struct {
		Jobs map[string]struct {
			Outputs map[string]string `yaml:"outputs"`
			Steps   []struct {
				Uses string            `yaml:"uses"`
				With map[string]string `yaml:"with"`
			} `yaml:"steps"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal(b, &w); err != nil {
		t.Fatal(err)
	}
	job := w.Jobs["release-please"]
	for _, step := range job.Steps {
		if !strings.HasPrefix(step.Uses, "googleapis/release-please-action") {
			continue
		}
		for key, file := range map[string]string{"config-file": "release.json", "manifest-file": ".release.json"} {
			if want := "format('{0}/.github/" + file + "', inputs.working_directory)"; !strings.Contains(step.With[key], want) {
				t.Errorf("%s = %q, want it to contain %s", key, step.With[key], want)
			}
		}
	}
	for _, key := range []string{"release_created", "version", "tag_name"} {
		if want := "format('{0}--" + key + "', inputs.working_directory)"; !strings.Contains(job.Outputs[key], want) {
			t.Errorf("output %s = %q, want it to contain %s", key, job.Outputs[key], want)
		}
	}
	// releases_created is true when any package is released, not just this one,
	// so it only counts for the root package
	_, published, _ := strings.Cut(string(b), "new_release_published:")
	published, _, _ = strings.Cut(published, "\n      version:")
	beforeRoot, rootClause, found := strings.Cut(published, "inputs.working_directory == '.' && ")
	if !found || strings.Contains(beforeRoot, "releases_created") ||
		!strings.HasPrefix(rootClause, "(jobs.release-please.outputs.releases_created") {
		t.Errorf("new_release_published counts releases_created beyond the root package: %s", published)
	}
}

type reviewJob struct {
	If    string         `yaml:"if"`
	Uses  string         `yaml:"uses"`
	Needs []string       `yaml:"needs"`
	With  map[string]any `yaml:"with"`
}

func readJobs(t *testing.T, name string) map[string]reviewJob {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", name))
	if err != nil {
		t.Fatal(err)
	}
	var w struct {
		Jobs map[string]reviewJob `yaml:"jobs"`
	}
	if err := yaml.Unmarshal(b, &w); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return w.Jobs
}

// CODEOWNERS requests nothing in a private repository on a free plan, so each pipeline asks
// itself, and before the checks, so a failing update reaches the reviewer too.
func TestBotPullRequestsRequestAReview(t *testing.T) {
	for _, name := range []string{"library.yml", "service.yml", "website.yml"} {
		jobs := readJobs(t, name)
		review := jobs["review"]
		if review.Uses != "./.github/workflows/review.yml" || len(review.Needs) != 0 ||
			!strings.Contains(review.If, "github.event.pull_request.user.login == 'dependabot[bot]'") {
			t.Errorf("%s: review = %+v", name, review)
		}
		if review.With["reviewers"] != "${{ inputs.reviewers }}" || jobs["version"].With["reviewers"] != "${{ inputs.reviewers }}" {
			t.Errorf("%s: reviewers aren't passed on", name)
		}
	}
	review := readJobs(t, "version.yml")["review"]
	if len(review.Needs) != 1 || review.Needs[0] != "release-please" ||
		review.With["pull_request"] != "${{ needs.release-please.outputs.pr_number }}" {
		t.Errorf("version.yml: review = %+v", review)
	}
	b, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "review.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "catch (error)") || !strings.Contains(string(b), "core.warning") {
		t.Error("review.yml can fail the pipeline")
	}
}

// actionlint checks a local call's inputs too, but only where it's installed.
func TestLocalCallsOnlyPassDeclaredInputs(t *testing.T) {
	entries, err := os.ReadDir(filepath.Join("..", "..", ".github", "workflows"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		b, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		var w struct {
			Jobs map[string]callerJob `yaml:"jobs"`
		}
		if err := yaml.Unmarshal(b, &w); err != nil {
			t.Fatalf("%s: %v", entry.Name(), err)
		}
		for name, job := range w.Jobs {
			file, ok := strings.CutPrefix(job.Uses, "./.github/workflows/")
			if !ok {
				continue
			}
			declared := readReusable(t, file).On.WorkflowCall.Inputs
			for input := range job.With {
				if _, ok := declared[input]; !ok {
					t.Errorf("%s: job %s passes %s, which %s doesn't declare", entry.Name(), name, input, file)
				}
			}
		}
	}
}

type matrixJob struct {
	If       string   `yaml:"if"`
	Needs    []string `yaml:"needs"`
	Strategy struct {
		Matrix map[string]string `yaml:"matrix"`
	} `yaml:"strategy"`
}

func readMatrixJobs(t *testing.T, name string) map[string]matrixJob {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", name))
	if err != nil {
		t.Fatal(err)
	}
	var w struct {
		Jobs map[string]matrixJob `yaml:"jobs"`
	}
	if err := yaml.Unmarshal(b, &w); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return w.Jobs
}

// The extra runners and native builds are opt-in: their defaults run nothing, and an empty
// matrix would fail every caller's workflow, so each job is skipped before its matrix expands.
func TestRunnerListsDefaultToNothing(t *testing.T) {
	for _, c := range []struct{ file, job, input string }{
		{"testing.yml", "extra", "extra_runners"},
		{"service.yml", "native-build", "native_builds"},
	} {
		input, _ := readReusable(t, c.file).On.WorkflowCall.Inputs[c.input].(map[string]any)
		if input["default"] != "[]" {
			t.Errorf("%s: %s defaults to %v, want []", c.file, c.input, input["default"])
		}
		job := readMatrixJobs(t, c.file)[c.job]
		if !strings.Contains(job.If, "inputs."+c.input+" != '[]'") || !strings.Contains(job.If, "inputs."+c.input+" != ''") {
			t.Errorf("%s: %s runs with an empty %s: if = %q", c.file, c.job, c.input, job.If)
		}
		if job.Strategy.Matrix["target"] != "${{ fromJSON(inputs."+c.input+") }}" {
			t.Errorf("%s: %s matrix = %v", c.file, c.job, job.Strategy.Matrix)
		}
	}
}

// A published release may be immutable, so everything is attached before the release job
// publishes the draft, and the patch job waits for every runner's tests.
func TestEverythingWaitsForTheRunnerLists(t *testing.T) {
	if needs := readMatrixJobs(t, "testing.yml")["patch"].Needs; !slices.Contains(needs, "extra") {
		t.Errorf("testing.yml: patch needs %v, want extra too", needs)
	}
	if needs := readMatrixJobs(t, "service.yml")["release"].Needs; !slices.Contains(needs, "native-build") {
		t.Errorf("service.yml: release needs %v, want native-build too", needs)
	}
	jobs := readJobs(t, "service.yml")
	for _, job := range []string{"testing", "release-testing"} {
		for _, input := range []string{"extra_runners", "extra_test_task", "setup_script", "artifact_name", "artifact_path", "build_command"} {
			if jobs[job].With[input] != "${{ inputs."+input+" }}" {
				t.Errorf("service.yml: %s doesn't pass %s on", job, input)
			}
		}
	}
}

// Signing and skipping the checks are opt-in: by default the release job imports no key and
// service.yml runs its own checks.
func TestReleaseSigningIsOptIn(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "service.yml"))
	if err != nil {
		t.Fatal(err)
	}
	var w struct {
		On struct {
			WorkflowCall struct {
				Inputs  map[string]map[string]any `yaml:"inputs"`
				Secrets map[string]map[string]any `yaml:"secrets"`
			} `yaml:"workflow_call"`
		} `yaml:"on"`
		Jobs map[string]struct {
			If    string `yaml:"if"`
			Steps []struct {
				If   string `yaml:"if"`
				Uses string `yaml:"uses"`
				Run  string `yaml:"run"`
			} `yaml:"steps"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal(b, &w); err != nil {
		t.Fatal(err)
	}
	call := w.On.WorkflowCall
	if call.Inputs["gpg_sign"]["default"] != false || call.Inputs["run_checks"]["default"] != true {
		t.Errorf("gpg_sign = %v, run_checks = %v", call.Inputs["gpg_sign"], call.Inputs["run_checks"])
	}
	for _, secret := range []string{"GPG_PRIVATE_KEY", "GPG_PASSPHRASE"} {
		if required, _ := call.Secrets[secret]["required"].(bool); call.Secrets[secret] == nil || required {
			t.Errorf("secret %s = %v, want it declared and optional", secret, call.Secrets[secret])
		}
	}
	for _, step := range w.Jobs["release"].Steps {
		if (strings.HasPrefix(step.Uses, "crazy-max/ghaction-import-gpg") || strings.Contains(step.Run, "GPG_FINGERPRINT")) &&
			step.If != "inputs.gpg_sign" {
			t.Errorf("release step %q runs without gpg_sign: if = %q", step.Uses+step.Run, step.If)
		}
	}
	for _, job := range []string{"testing", "release-testing"} {
		if !strings.Contains(w.Jobs[job].If, "inputs.run_checks") {
			t.Errorf("%s runs with run_checks off: if = %q", job, w.Jobs[job].If)
		}
	}
}
