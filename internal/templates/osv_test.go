package templates

import (
	"encoding/json"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
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
	Name        string            `yaml:"name"`
	RunsOn      string            `yaml:"runs-on"`
	Needs       stringList        `yaml:"needs"`
	Container   map[string]string `yaml:"container"`
	Permissions map[string]string `yaml:"permissions"`
	Defaults    struct {
		Run map[string]string `yaml:"run"`
	} `yaml:"defaults"`
	Steps []osvStep `yaml:"steps"`
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

var checkPipelines = []string{"library.yml", "service.yml", "website.yml"}

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

func jobSteps(t *testing.T, pipeline, id string) map[string]osvStep {
	t.Helper()
	steps := map[string]osvStep{}
	for _, step := range readScanPipeline(t, pipeline).Jobs[id].Steps {
		steps[step.Name] = step
	}
	return steps
}

func scanSteps(t *testing.T, pipeline string) map[string]osvStep {
	t.Helper()
	return jobSteps(t, pipeline, "scan")
}

func osvSteps(t *testing.T) map[string]osvStep {
	t.Helper()
	return jobSteps(t, "service.yml", "scan-osv")
}

// a tag Dependabot's docker ecosystem can bump
var osvImage = regexp.MustCompile(`^ghcr\.io/google/osv-scanner:v\d+\.\d+\.\d+$`)

const osvRun = `/osv-scanner scan source --format=json --output-file="$GITHUB_WORKSPACE/osv-results.json" --recursive --allow-no-lockfiles .`

// OSV-Scanner runs in Google's own image as the job's container, under an image: key so Dependabot's docker
// updates keep the tag current; it exits non-zero on a finding, so the step carries on and the report decides.
func TestOSVScanRunsInItsOwnImage(t *testing.T) {
	for _, pipeline := range checkPipelines {
		if text := readText(t, pipeline); strings.Contains(text, "google/osv-scanner-action") {
			t.Errorf("%s still uses OSV-Scanner's container action", pipeline)
		}
		job := readScanPipeline(t, pipeline).Jobs["scan-osv"]
		if job.Name != "scan (osv)" || job.RunsOn != "${{ inputs.scan_runs_on || inputs.runs_on }}" {
			t.Errorf("%s: scan-osv name %q, runs-on %q", pipeline, job.Name, job.RunsOn)
		}
		if len(job.Container) != 1 || !osvImage.MatchString(job.Container["image"]) {
			t.Errorf("%s: scan-osv container = %v", pipeline, job.Container)
		}
		if !maps.Equal(job.Permissions, map[string]string{"contents": "read"}) {
			t.Errorf("%s: scan-osv permissions = %v", pipeline, job.Permissions)
		}
		if job.Defaults.Run["working-directory"] != "${{ inputs.working_directory }}" {
			t.Errorf("%s: scan-osv runs in %q", pipeline, job.Defaults.Run["working-directory"])
		}
		var names []string
		for _, s := range job.Steps {
			names = append(names, s.Name)
		}
		if want := []string{"Check out", "Scan with OSV", "Report OSV findings"}; !slices.Equal(names, want) {
			t.Errorf("%s: scan-osv steps = %v, want %v", pipeline, names, want)
		}
	}
	steps := osvSteps(t)
	if checkout := steps["Check out"]; checkout.With["persist-credentials"] != "false" || checkout.With["fetch-depth"] != "" {
		t.Errorf("Check out = %+v", checkout)
	}
	scan := steps["Scan with OSV"]
	if scan.Uses != "" || scan.Run != osvRun || scan.ID != "osv" || !scan.ContinueOnError {
		t.Errorf("Scan with OSV = %+v", scan)
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

// The image needs a runner that can run a job container, so its stage can run elsewhere than the rest; by
// default it runs with every other stage. The scan stage itself runs where everything else does, on the runner.
func TestScanCanRunOnAContainerRunner(t *testing.T) {
	for _, pipeline := range checkPipelines {
		w := readScanPipeline(t, pipeline)
		if input, ok := w.On.WorkflowCall.Inputs["scan_runs_on"]; !ok || input.Default != "" || input.Type != "string" {
			t.Errorf("%s: input scan_runs_on = %+v", pipeline, input)
		}
		if _, ok := w.On.WorkflowCall.Inputs["scan_container"]; ok {
			t.Errorf("%s still takes scan_container", pipeline)
		}
		scan := w.Jobs["scan"]
		if scan.RunsOn != "${{ inputs.runs_on }}" || scan.Container != nil {
			t.Errorf("%s: scan runs-on %q, container %v", pipeline, scan.RunsOn, scan.Container)
		}
		if !maps.Equal(scan.Permissions, map[string]string{"contents": "read", "pull-requests": "write"}) {
			t.Errorf("%s: scan permissions = %v", pipeline, scan.Permissions)
		}
		if _, ok := scanSteps(t, pipeline)["Trust the workspace"]; ok {
			t.Errorf("%s: scan still trusts the workspace, as only a job container needed", pipeline)
		}
		// what waits for the scan waits for OSV's too
		for id, j := range w.Jobs {
			if slices.Contains(j.Needs, "scan") != slices.Contains(j.Needs, "scan-osv") {
				t.Errorf("%s: %s needs %v", pipeline, id, j.Needs)
			}
		}
	}
	// a generated project scans where it runs everything else
	for name, o := range allOptions() {
		if text := RenderWorkflow(o); strings.Contains(text, "scan_runs_on") || strings.Contains(text, "scan_container") {
			t.Errorf("%s sets the scan's runner", name)
		}
	}
	// forgego's own private runs scan with OSV on the Kubernetes-mode set
	if strings.Contains(readText(t, ".index.yml"), "scan_container") {
		t.Error("forgego's own .index.yml still passes scan_container")
	}
	for _, id := range []string{"service", "website"} {
		if with := readWorkflow(t, ".index.yml").Jobs[id].With; with["scan_runs_on"] != "${{ github.event.repository.private && 'xerus-docker' || 'ubuntu-latest' }}" {
			t.Errorf("forgego's own %s scans with %v", id, with)
		}
	}
}

// govulncheck, like OSV, fails a pull request and only warns elsewhere.
func TestGovulncheckFailsOnlyAPullRequest(t *testing.T) {
	for _, pipeline := range checkPipelines {
		scan := scanSteps(t, pipeline)["Scan with govulncheck"]
		if scan.Env["ENFORCE"] != "${{ github.event_name == 'pull_request' }}" ||
			!strings.Contains(scan.Run, `if [ "$ENFORCE" = true ]`) || !strings.Contains(scan.Run, "exit 1") {
			t.Errorf("%s: Scan with govulncheck = %+v", pipeline, scan)
		}
	}
}

// Every check pipeline gates on OSV the same way, so a fix to one reaches them all.
func TestCheckPipelinesShareTheOSVScan(t *testing.T) {
	want := readScanPipeline(t, "service.yml").Jobs["scan-osv"]
	for _, pipeline := range []string{"library.yml", "website.yml"} {
		if got := readScanPipeline(t, pipeline).Jobs["scan-osv"]; !reflect.DeepEqual(got, want) {
			t.Errorf("%s: scan-osv differs from service.yml's", pipeline)
		}
	}
}

// runs the report step's github-script body under Node with a stand-in for its core API; file collects the
// annotations' files
const osvHarness = `
const calls = { info: [], warning: [], error: [], setFailed: [], file: [] };
const record = (kind) => (message, props) => {
  calls[kind].push(message);
  if (props?.file !== undefined) calls.file.push(props.file);
};
const core = {
  info: record('info'), warning: record('warning'), error: record('error'), setFailed: record('setFailed'),
  summary: { addHeading() { return this; }, addTable() { return this; }, async write() {} },
};
const body = require('node:fs').readFileSync(process.argv[1], 'utf8');
const AsyncFunction = Object.getPrototypeOf(async () => {}).constructor;
new AsyncFunction('require', 'core', body)(require, core).then(() => console.log(JSON.stringify(calls)));
`

// runs the report with the scan step's outcome and, unless results is nil, a report holding them; a result's
// source path is taken as relative to the workspace and made absolute, as the scanner reports it
func runOSVReport(t *testing.T, outcome string, results []map[string]any, enforce bool) map[string][]string {
	t.Helper()
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("needs Node.js to run the github-script body")
	}
	dir := t.TempDir()
	script := filepath.Join(dir, "report.js")
	if err := os.WriteFile(script, []byte(osvSteps(t)["Report OSV findings"].With["script"]), 0o600); err != nil {
		t.Fatal(err)
	}
	report := filepath.Join(dir, "osv-results.json")
	if results != nil {
		absolute := make([]map[string]any, len(results))
		for i, r := range results {
			absolute[i] = maps.Clone(r)
			absolute[i]["source"] = map[string]any{"path": filepath.Join(dir, filepath.FromSlash(r["source"].(map[string]any)["path"].(string)))}
		}
		b, _ := json.Marshal(map[string]any{"results": absolute})
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
		"source":   map[string]any{"path": "docs/go.mod"},
		"packages": []any{finding("GHSA-high", "9.8", ""), finding("GO-2026-0001", "", "")},
	}}
	for enforce, failed := range map[bool]int{true: 1, false: 0} {
		// the scanner fails on any finding
		calls := runOSVReport(t, "failure", results, enforce)
		if len(calls["setFailed"]) != failed || len(calls["error"]) != 1 {
			t.Errorf("enforce %v: calls = %v", enforce, calls)
		}
		// each finding is annotated on the lockfile, relative to the workspace
		if want := filepath.Join("docs", "go.mod"); !slices.Equal(calls["file"], []string{want, want}) {
			t.Errorf("enforce %v: annotated files = %v, want %s", enforce, calls["file"], want)
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
