package templates

import (
	"encoding/json"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

type osvStep struct {
	Name            string            `yaml:"name"`
	ID              string            `yaml:"id"`
	Uses            string            `yaml:"uses"`
	Run             string            `yaml:"run"`
	ContinueOnError bool              `yaml:"continue-on-error"`
	Env             map[string]string `yaml:"env"`
	With            map[string]string `yaml:"with"`
}

type scanJob struct {
	RunsOn      string            `yaml:"runs-on"`
	Container   string            `yaml:"container"`
	Permissions map[string]string `yaml:"permissions"`
	Steps       []osvStep         `yaml:"steps"`
}

type scanPipeline struct {
	On struct {
		WorkflowCall struct {
			Inputs map[string]struct {
				Default any    `yaml:"default"`
				Type    string `yaml:"type"`
			} `yaml:"inputs"`
		} `yaml:"workflow_call"`
	} `yaml:"on"`
	Jobs map[string]scanJob `yaml:"jobs"`
}

func readScanPipeline(t *testing.T, pipeline string) scanPipeline {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", pipeline))
	if err != nil {
		t.Fatal(err)
	}
	var w scanPipeline
	if err := yaml.Unmarshal(b, &w); err != nil {
		t.Fatal(err)
	}
	return w
}

func scanSteps(t *testing.T, pipeline string) map[string]osvStep {
	t.Helper()
	steps := map[string]osvStep{}
	for _, step := range readScanPipeline(t, pipeline).Jobs["scan"].Steps {
		steps[step.Name] = step
	}
	return steps
}

func secureSteps(t *testing.T) map[string]osvStep {
	t.Helper()
	return scanSteps(t, "service.yml")
}

const osvAction = "google/osv-scanner-action/osv-scanner-action@b1d267aea5211fa648d9d442fa64b90dbdcbdd36 # v2.7.0"

// OSV-Scanner is Google's own action, pinned so Dependabot keeps it current; it exits non-zero on a finding,
// so the step carries on and the report decides.
func TestOSVScanUsesItsOwnAction(t *testing.T) {
	steps := secureSteps(t)
	scan := steps["Scan with OSV"]
	text := readText(t, "service.yml")
	if scan.Uses == "" || !strings.Contains(text, "uses: "+osvAction+"\n") || scan.ID != "osv" || !scan.ContinueOnError {
		t.Errorf("Scan with OSV = %+v", scan)
	}
	want := "--format=json\n--output=osv-results.json\n--recursive\n./${{ inputs.working_directory }}"
	if scan.With["scan-args"] != want {
		t.Errorf("scan-args = %q, want %q", scan.With["scan-args"], want)
	}
	if _, ok := steps["Install OSV-Scanner"]; ok {
		t.Error("the scan still installs OSV-Scanner itself")
	}
	report := steps["Report OSV findings"]
	for name, want := range map[string]string{
		"OSV_ENFORCE": "${{ github.event_name == 'pull_request' }}",
		"OSV_OUTCOME": "${{ steps.osv.outcome }}",
		"OSV_REPORT":  "osv-results.json",
	} {
		if report.Env[name] != want {
			t.Errorf("%s = %q, want %q", name, report.Env[name], want)
		}
	}
}

// The container action needs a Linux runner that can run one, so the scan stage can run elsewhere than the rest,
// in a job container if the runner set needs one; by default it runs with every other stage. A job container
// runs git as root over the runner user's checkout, so the workspace is trusted before Gitleaks reads it.
func TestScanCanRunOnAContainerRunner(t *testing.T) {
	for _, pipeline := range []string{"library.yml", "service.yml", "website.yml"} {
		w := readScanPipeline(t, pipeline)
		for _, in := range []string{"scan_runs_on", "scan_container"} {
			if input, ok := w.On.WorkflowCall.Inputs[in]; !ok || input.Default != "" || input.Type != "string" {
				t.Errorf("%s: input %s = %+v", pipeline, in, input)
			}
		}
		scan := w.Jobs["scan"]
		if scan.RunsOn != "${{ inputs.scan_runs_on || inputs.runs_on }}" || scan.Container != "${{ inputs.scan_container }}" {
			t.Errorf("%s: scan runs-on %q, container %q", pipeline, scan.RunsOn, scan.Container)
		}
		if !maps.Equal(scan.Permissions, map[string]string{"contents": "read", "pull-requests": "write"}) {
			t.Errorf("%s: scan permissions = %v", pipeline, scan.Permissions)
		}
		if len(scan.Steps) < 2 || scan.Steps[0].Name != "Check out" || scan.Steps[1].Name != "Trust the workspace" ||
			scan.Steps[1].Run != `git config --global --add safe.directory "$GITHUB_WORKSPACE"` {
			t.Errorf("%s: the workspace isn't trusted right after the checkout", pipeline)
		}
	}
	// a generated project scans where it runs everything else
	for name, o := range allOptions() {
		if text := RenderWorkflow(o); strings.Contains(text, "scan_runs_on") || strings.Contains(text, "scan_container") {
			t.Errorf("%s sets the scan's runner", name)
		}
	}
	// forgego's own private runs scan on the Kubernetes-mode set, in a container with git
	for _, id := range []string{"service", "website"} {
		with := readWorkflow(t, ".index.yml").Jobs[id].With
		if with["scan_runs_on"] != "${{ github.event.repository.private && 'xerus-docker' || 'ubuntu-latest' }}" ||
			with["scan_container"] != "${{ github.event.repository.private && 'buildpack-deps:trixie-scm' || '' }}" {
			t.Errorf("forgego's own %s scans with %v", id, with)
		}
	}
}

// govulncheck, like OSV, fails a pull request and only warns elsewhere.
func TestGovulncheckFailsOnlyAPullRequest(t *testing.T) {
	for _, pipeline := range []string{"library.yml", "service.yml", "website.yml"} {
		scan := scanSteps(t, pipeline)["Scan with govulncheck"]
		if scan.Env["ENFORCE"] != "${{ github.event_name == 'pull_request' }}" ||
			!strings.Contains(scan.Run, `if [ "$ENFORCE" = true ]`) || !strings.Contains(scan.Run, "exit 1") {
			t.Errorf("%s: Scan with govulncheck = %+v", pipeline, scan)
		}
	}
}

// Every check pipeline gates on OSV the same way, so a fix to one reaches them all.
func TestCheckPipelinesShareTheOSVScan(t *testing.T) {
	want := secureSteps(t)
	for _, pipeline := range []string{"library.yml", "website.yml"} {
		got := scanSteps(t, pipeline)
		for _, name := range []string{"Trust the workspace", "Scan with OSV", "Report OSV findings"} {
			if !reflect.DeepEqual(got[name], want[name]) {
				t.Errorf("%s: %q differs from service.yml's", pipeline, name)
			}
		}
	}
}

// runs the report step's github-script body under Node with a stand-in for its core API
const osvHarness = `
const calls = { info: [], warning: [], error: [], setFailed: [] };
const record = (kind) => (message) => calls[kind].push(message);
const core = {
  info: record('info'), warning: record('warning'), error: record('error'), setFailed: record('setFailed'),
  summary: { addHeading() { return this; }, addTable() { return this; }, async write() {} },
};
const body = require('node:fs').readFileSync(process.argv[1], 'utf8');
const AsyncFunction = Object.getPrototypeOf(async () => {}).constructor;
new AsyncFunction('require', 'core', body)(require, core).then(() => console.log(JSON.stringify(calls)));
`

// runs the report with the scan step's outcome and, unless results is nil, a report holding them
func runOSVReport(t *testing.T, outcome string, results []map[string]any, enforce bool) map[string][]string {
	t.Helper()
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("needs Node.js to run the github-script body")
	}
	dir := t.TempDir()
	script := filepath.Join(dir, "report.js")
	if err := os.WriteFile(script, []byte(secureSteps(t)["Report OSV findings"].With["script"]), 0o600); err != nil {
		t.Fatal(err)
	}
	report := filepath.Join(dir, "osv-results.json")
	if results != nil {
		b, _ := json.Marshal(map[string]any{"results": results})
		if err := os.WriteFile(report, b, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	cmd := exec.Command(node, "-e", osvHarness, script)
	cmd.Env = append(os.Environ(), "OSV_OUTCOME="+outcome, "OSV_REPORT="+filepath.Base(report), "GITHUB_WORKSPACE="+dir,
		"OSV_ENFORCE="+map[bool]string{true: "true", false: "false"}[enforce])
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("report: %v", err)
	}
	var calls map[string][]string
	if err := json.Unmarshal(out, &calls); err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	return calls
}

func finding(id, maxSeverity, rating string) map[string]any {
	db := map[string]any{}
	if rating != "" {
		db["severity"] = rating
	}
	return map[string]any{
		"package":         map[string]any{"name": "example.com/" + id, "version": "1.0.0"},
		"vulnerabilities": []any{map[string]any{"id": id, "summary": id, "database_specific": db}},
		"groups":          []any{map[string]any{"ids": []string{id}, "max_severity": maxSeverity}},
	}
}

func TestOSVReportFailsOnlyAPullRequestOnAHighFinding(t *testing.T) {
	results := []map[string]any{{
		"source":   map[string]any{"path": "/github/workspace/go.mod"},
		"packages": []any{finding("GHSA-high", "9.8", ""), finding("GO-2026-0001", "", "")},
	}}
	for enforce, failed := range map[bool]int{true: 1, false: 0} {
		// the action fails on any finding
		calls := runOSVReport(t, "failure", results, enforce)
		if len(calls["setFailed"]) != failed || len(calls["error"]) != 1 {
			t.Errorf("enforce %v: calls = %v", enforce, calls)
		}
		// a Go advisory has no severity, so it's a warning
		if !strings.Contains(strings.Join(calls["warning"], "\n"), "GO-2026-0001") {
			t.Errorf("enforce %v: warnings = %v", enforce, calls["warning"])
		}
	}
}

// A scan that succeeded without a report had nothing to scan; one with a report found nothing.
func TestOSVReportPassesWhenThereIsNothingToReport(t *testing.T) {
	for _, results := range [][]map[string]any{nil, {}} {
		calls := runOSVReport(t, "success", results, true)
		if len(calls["setFailed"]) != 0 || len(calls["warning"]) != 0 || len(calls["info"]) != 1 {
			t.Errorf("report %v: calls = %v", results, calls)
		}
	}
}

// A scan that failed without a report couldn't finish: that fails a pull request, which can't pass unchecked,
// and warns elsewhere.
func TestOSVReportFailsAPullRequestOnAnIncompleteScan(t *testing.T) {
	for enforce, failed := range map[bool]int{true: 1, false: 0} {
		calls := runOSVReport(t, "failure", nil, enforce)
		if len(calls["setFailed"]) != failed || len(calls["warning"]) != 1-failed {
			t.Errorf("enforce %v: calls = %v", enforce, calls)
		}
	}
}
