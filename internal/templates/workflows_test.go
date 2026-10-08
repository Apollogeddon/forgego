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
