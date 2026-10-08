package taskfile

import (
	"strings"
	"testing"
)

func TestRenderNewTaskfile(t *testing.T) {
	b := NewBuilder()
	b.Var("LINT", "go tool -modfile=.forgego/golangci-lint/go.mod golangci-lint")
	b.Add(Task{Name: "lint", Desc: "Lint and fix", Cmds: []string{"{{.LINT}} run --fix"}})
	out, changed, err := b.Render("", false)
	if err != nil {
		t.Fatal(err)
	}
	want := `version: '3'
vars:
  LINT: go tool -modfile=.forgego/golangci-lint/go.mod golangci-lint
tasks:
  lint:
    desc: Lint and fix
    cmds:
      - '{{.LINT}} run --fix'
`
	if out != want {
		t.Errorf("got\n%s\nwant\n%s", out, want)
	}
	if len(changed) != 2 {
		t.Errorf("changed = %v", changed)
	}
}

func TestRenderKeepsTheProjectsOwnTasksUnlessForced(t *testing.T) {
	existing := "version: '3'\ntasks:\n  # my own lint\n  lint:\n    cmds: [make lint]\n  deploy:\n    cmds: [./deploy.sh]\n"
	b := NewBuilder()
	b.Add(Task{Name: "lint", Cmds: []string{"golangci-lint run"}})
	b.Add(Task{Name: "test", Cmds: []string{"go test ./..."}})

	out, changed, err := b.Render(existing, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, keep := range []string{"# my own lint", "make lint", "./deploy.sh", "go test ./..."} {
		if !strings.Contains(out, keep) {
			t.Errorf("missing %q in\n%s", keep, out)
		}
	}
	if len(changed) != 1 || changed[0] != "task test" {
		t.Errorf("changed = %v", changed)
	}

	forced, _, _ := b.Render(existing, true)
	if strings.Contains(forced, "make lint") || !strings.Contains(forced, "./deploy.sh") {
		t.Errorf("force should replace lint and keep deploy:\n%s", forced)
	}
}
