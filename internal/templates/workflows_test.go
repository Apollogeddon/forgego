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
	// releases_created is true when any package is released, not just this one
	if strings.Contains(string(b), "releases_created") {
		t.Error("version.yml reads releases_created")
	}
}
