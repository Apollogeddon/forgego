package templates

import "strings"

// WorkflowRepo is where the reusable workflows a generated project calls live.
const WorkflowRepo = "apollogeddon/forgego/.github/workflows/"

// WorkflowOptions is what decides a project's generated CI.
type WorkflowOptions struct {
	Mode       string // backend, library or website
	Go         string
	Docker     bool
	Testing    bool
	Versioning bool
}

const workflowHeader = `name: CI

on:
  push:
    branches: ["main"]
  pull_request:
    branches: ["main"]

concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  # never cancel a run on main mid-release, or the release is created but never published
  cancel-in-progress: ${{ github.ref != 'refs/heads/main' }}

jobs:
`

const dockerJob = `
  docker:
    needs: __PIPELINE__
    uses: ` + WorkflowRepo + `docker.yml@main
    permissions:
      contents: read
      packages: write
    with:
      push: ${{ github.ref == 'refs/heads/main' && needs.__PIPELINE__.outputs.new_release_published == 'true' }}
      version: ${{ needs.__PIPELINE__.outputs.version }}
`

// RenderWorkflow returns the project's .github/workflows/index.yml. Docker is its own job
// after the pipeline, so a project without it carries no permanently skipped job.
func RenderWorkflow(o WorkflowOptions) string {
	job, workflow := "service", "service.yml"
	permissions := []string{"contents: write", "pull-requests: write"}
	switch o.Mode {
	case "library":
		job, workflow = "library", "library.yml"
	case "website":
		job, workflow = "website", "website.yml"
		permissions = []string{"contents: write", "pages: write", "id-token: write", "pull-requests: write"}
	}

	var b strings.Builder
	b.WriteString(workflowHeader)
	b.WriteString("  " + job + ":\n")
	b.WriteString("    uses: " + WorkflowRepo + workflow + "@main\n")
	b.WriteString("    permissions:\n")
	for _, p := range permissions {
		b.WriteString("      " + p + "\n")
	}
	b.WriteString("    with:\n")
	b.WriteString("      go_version: '" + o.Go + "'\n")
	// Disabled standard features become pipeline inputs, so CI doesn't run what the project doesn't have.
	if !o.Testing && o.Mode != "website" {
		b.WriteString("      run_tests: false\n")
	}
	if !o.Versioning {
		b.WriteString("      enable_versioning: false\n")
	}
	if o.Mode != "website" {
		b.WriteString("      auto_patch: true\n")
	}
	if o.Docker {
		b.WriteString(strings.ReplaceAll(dockerJob, "__PIPELINE__", job))
	}
	return b.String()
}
