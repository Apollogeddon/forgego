package templates

import (
	"os"
	"path/filepath"
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
			out := RenderWorkflow(WorkflowOptions{Mode: mode, Go: "1.27", Docker: docker})
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
