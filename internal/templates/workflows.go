package templates

import (
	"regexp"
	"strings"
)

// WorkflowRepo is where the pipelines a generated project calls live.
const WorkflowRepo = "apollogeddon/forgego/.github/workflows/"

// IndexWorkflow is the generated root workflow. Its leading dot keeps it apart from the pipelines it calls.
const IndexWorkflow = ".github/workflows/.index.yml"

// WorkflowOptions is what decides a project's generated CI.
type WorkflowOptions struct {
	Mode       string // backend, library or website
	Docker     bool
	Testing    bool
	Linting    bool
	Versioning bool
	Windows    bool // the checks test and build on Windows too; not for a website
	MacOS      bool // and on macOS
}

const workflowHeader = `name: index

on:
  push:
    branches: ["main"]
  pull_request:
    branches: ["main"]
  workflow_dispatch:
  # sent by the merge pipeline, as a merge it makes starts no push run on main
  repository_dispatch:
    types: [auto-merged]
  # weekly, for the patch pipeline; a manual or dispatched run refreshes too
  schedule:
    - cron: '17 3 * * 1'

concurrency:
  # a pull request's newer run replaces its older one; every other run has a group of its own, so a run on
  # main, running or queued, is never cancelled and a release is never created without being published
  group: ${{ github.workflow }}-${{ github.event.pull_request.number || github.run_id }}
  cancel-in-progress: true

# every job asks for what it needs
permissions: {}

jobs:
`

const released = "needs.version.outputs.new_release_published == 'true'"

// public is true only for a public repository: the field is missing on a scheduled run, which then falls back
// to what a private repository on GitHub Free can do.
const public = "github.event.repository.private == false"

// input is one `with:` entry, kept in order.
type input struct{ key, value string }

// pipelineJob is one root job, which calls one pipeline, so a run page reads <pipeline> / <stage>.
type pipelineJob struct {
	id, comment, cond, pipeline string
	needs                       []string
	with, secrets               []input
	permissions                 []string
}

func expr(e string) string { return "${{ " + e + " }}" }

func (j pipelineJob) render(b *strings.Builder) {
	if j.comment != "" {
		b.WriteString("  # " + j.comment + "\n")
	}
	b.WriteString("  " + j.id + ":\n")
	switch len(j.needs) {
	case 0:
	case 1:
		b.WriteString("    needs: " + j.needs[0] + "\n")
	default:
		b.WriteString("    needs: [" + strings.Join(j.needs, ", ") + "]\n")
	}
	if j.cond != "" {
		b.WriteString("    if: " + expr(j.cond) + "\n")
	}
	b.WriteString("    uses: " + WorkflowRepo + j.pipeline + ".yml@main\n")
	if len(j.with) > 0 {
		b.WriteString("    with:\n")
		for _, in := range j.with {
			b.WriteString("      " + in.key + ": " + in.value + "\n")
		}
	}
	if len(j.secrets) > 0 {
		b.WriteString("    secrets:\n")
		for _, in := range j.secrets {
			b.WriteString("      " + in.key + ": " + in.value + "\n")
		}
	}
	b.WriteString("    permissions:\n")
	for _, p := range j.permissions {
		b.WriteString("      " + p + "\n")
	}
}

// RenderWorkflow returns the project's .github/workflows/.index.yml. Pull requests run the checks, main runs
// release-please and, on a release, the checks again before shipping, and a weekly or dispatched run refreshes
// forgego's files.
func RenderWorkflow(o WorkflowOptions) string {
	checks := "service"
	switch o.Mode {
	case "library", "website":
		checks = o.Mode
	}
	var jobs []pipelineJob

	if o.Versioning {
		jobs = append(jobs, pipelineJob{
			id: "version", pipeline: "version",
			comment:     "release-please, on main; a GitHub App's token, when one is set, lets its release pull request run the checks",
			cond:        "github.event_name != 'pull_request' && github.event_name != 'schedule'",
			with:        []input{{"app_client_id", expr("vars.RELEASE_APP_CLIENT_ID")}},
			secrets:     []input{{"APP_PRIVATE_KEY", expr("secrets.RELEASE_APP_PRIVATE_KEY")}},
			permissions: []string{"contents: write", "pull-requests: write"},
		})
	}

	// Disabled standard features become pipeline inputs, so CI doesn't run what the project doesn't have;
	// --no-versioning leaves out the version job, and with it the release checks and publish.
	check := pipelineJob{
		id: checks, pipeline: checks,
		comment:     "the checks, on pull requests and every push to main",
		cond:        "github.event_name != 'schedule'",
		permissions: []string{"contents: read", "pull-requests: write"},
	}
	if o.Versioning {
		check.comment = "a pull request's checks, and on main a release's, before it ships"
		check.needs = []string{"version"}
		check.cond = "!cancelled() && (github.event_name == 'pull_request' || " + released + ")"
	}
	if o.Mode != "website" {
		if !o.Testing {
			check.with = append(check.with, input{"run_tests", "false"})
		}
		if !o.Linting {
			check.with = append(check.with, input{"lint", "false"})
		}
		if o.Windows {
			check.with = append(check.with, input{"test_windows", "windows-latest"})
		}
		if o.MacOS {
			check.with = append(check.with, input{"test_macos", "macos-latest"})
		}
	}
	jobs = append(jobs, check)
	checkJobs := []string{checks}

	afterChecks := "!cancelled() && needs." + checks + ".result == 'success'"
	if o.Docker {
		// it builds on pull requests, and builds and pushes on a release
		docker := pipelineJob{
			id: "docker", pipeline: "docker", needs: []string{checks}, cond: afterChecks,
			permissions: []string{"contents: read", "packages: write"},
		}
		if o.Versioning {
			docker.needs = []string{"version", checks}
			docker.with = []input{{"push", expr(released)}, {"version", expr("needs.version.outputs.version")}}
		}
		jobs = append(jobs, docker)
		checkJobs = append(checkJobs, "docker")
	}

	if o.Versioning && o.Mode != "website" {
		publish := pipelineJob{
			id: "publish", pipeline: "publish", needs: []string{"version", checks},
			cond: afterChecks + " && " + released,
		}
		if o.Mode == "library" {
			// the assets stage, which would need contents: write, is off, so the grant stays read-only
			publish.comment = "the release, through the Go module proxy"
			publish.with = []input{{"tag_name", expr("needs.version.outputs.tag_name")}, {"package", "true"}, {"assets", "false"}}
			publish.permissions = []string{"contents: read"}
		} else {
			// GitHub Free attests only a public repository's builds
			publish.comment = "the release's binaries, built by GoReleaser, attached to the GitHub release and attested"
			publish.with = []input{{"tag_name", expr("needs.version.outputs.tag_name")}, {"attest", expr(public)}}
			publish.permissions = []string{"contents: write", "id-token: write", "attestations: write"}
		}
		jobs = append(jobs, publish)
	}

	if o.Mode == "website" {
		pages := pipelineJob{
			id: "pages", pipeline: "pages",
			comment:     "every change to main, to GitHub Pages, which GitHub Free serves only from a public repository: a paid plan can drop that check",
			needs:       []string{checks},
			cond:        "github.ref == 'refs/heads/main' && github.event_name != 'pull_request' && " + public,
			permissions: []string{"pages: write", "id-token: write"},
		}
		if o.Versioning {
			pages.comment = "a release, to GitHub Pages, which GitHub Free serves only from a public repository: a paid plan can drop that check"
			pages.needs = []string{"version", checks}
			pages.cond = afterChecks + " && " + released + " && " + public
		}
		jobs = append(jobs, pages)
	}

	jobs = append(jobs, pipelineJob{
		id: "patch", pipeline: "patch",
		comment:     "weekly, on a manual run or after a Dependabot merge, forgego's files refreshed through a pull request",
		cond:        "github.event_name == 'schedule' || github.event_name == 'workflow_dispatch' || github.event_name == 'repository_dispatch'",
		permissions: []string{"contents: write", "pull-requests: write"},
	}, pipelineJob{
		id: "merge", pipeline: "merge",
		comment:     "Dependabot's pull requests: a review request, and a squash merge when the checks passed",
		needs:       checkJobs,
		cond:        "!cancelled() && github.event_name == 'pull_request' && github.event.pull_request.user.login == 'dependabot[bot]'",
		with:        []input{{"checks_passed", expr("!contains(needs.*.result, 'failure') && !contains(needs.*.result, 'cancelled')")}},
		permissions: []string{"contents: write", "pull-requests: write"},
	})

	var b strings.Builder
	b.WriteString(workflowHeader)
	for i, j := range jobs {
		if i > 0 {
			b.WriteString("\n")
		}
		j.render(&b)
	}
	return b.String()
}

var forgegoPipeline = regexp.MustCompile(`(?i)apollogeddon/forgego/\.github/workflows/`)

// CallsForgego reports whether a workflow calls forgego's pipelines, rather than being the project's own.
func CallsForgego(content string) bool { return forgegoPipeline.MatchString(content) }
