package templates

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

type osvStep struct {
	Name string            `yaml:"name"`
	Run  string            `yaml:"run"`
	Env  map[string]string `yaml:"env"`
	With map[string]string `yaml:"with"`
}

func scanSteps(t *testing.T, pipeline string) map[string]osvStep {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", pipeline))
	if err != nil {
		t.Fatal(err)
	}
	var w struct {
		Jobs map[string]struct {
			Steps []osvStep `yaml:"steps"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal(b, &w); err != nil {
		t.Fatal(err)
	}
	steps := map[string]osvStep{}
	for _, step := range w.Jobs["scan"].Steps {
		steps[step.Name] = step
	}
	return steps
}

func secureSteps(t *testing.T) map[string]osvStep {
	t.Helper()
	return scanSteps(t, "service.yml")
}

func TestOSVScanNeverSwallowsTheResult(t *testing.T) {
	steps := secureSteps(t)
	if run := steps["Scan with OSV"].Run; strings.Contains(run, "|| true") || !strings.Contains(run, "--format json") ||
		!strings.Contains(run, `"$RUNNER_TEMP/osv-scanner/osv-scanner" scan`) {
		t.Errorf("Scan with OSV run = %q", run)
	}
	if got := steps["Report OSV findings"].Env["OSV_ENFORCE"]; got != "${{ github.event_name == 'pull_request' }}" {
		t.Errorf("OSV_ENFORCE = %q", got)
	}
}

// The scanner is a pinned release, checked against its published checksums, and a failed download or check
// fails the step rather than scanning with nothing.
func TestOSVScannerIsPinnedAndVerified(t *testing.T) {
	install := secureSteps(t)["Install OSV-Scanner"]
	if !regexp.MustCompile(`^v\d+\.\d+\.\d+$`).MatchString(install.Env["OSV_SCANNER_VERSION"]) {
		t.Errorf("OSV_SCANNER_VERSION = %q, want a release", install.Env["OSV_SCANNER_VERSION"])
	}
	for _, want := range []string{"curl -fsSLO", "osv-scanner_SHA256SUMS", "sha256sum -c --ignore-missing", "exit 1"} {
		if !strings.Contains(install.Run, want) {
			t.Errorf("Install OSV-Scanner doesn't run %q", want)
		}
	}
	if strings.Contains(install.Run, "latest") || strings.Contains(install.Run, "|| true") {
		t.Errorf("Install OSV-Scanner = %q", install.Run)
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
		for _, name := range []string{"Install OSV-Scanner", "Scan with OSV", "Report OSV findings"} {
			if got[name].Run != want[name].Run || got[name].With["script"] != want[name].With["script"] ||
				got[name].Env["OSV_ENFORCE"] != want[name].Env["OSV_ENFORCE"] ||
				got[name].Env["OSV_SCANNER_VERSION"] != want[name].Env["OSV_SCANNER_VERSION"] {
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

func runOSVReport(t *testing.T, status string, results []map[string]any, enforce bool) map[string][]string {
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
	report := filepath.Join(dir, "osv.json")
	if results != nil {
		b, _ := json.Marshal(map[string]any{"results": results})
		if err := os.WriteFile(report, b, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	cmd := exec.Command(node, "-e", osvHarness, script)
	cmd.Env = append(os.Environ(), "OSV_STATUS="+status, "OSV_REPORT="+report, "GITHUB_WORKSPACE="+dir,
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
		"source":   map[string]any{"path": "/w/go.mod"},
		"packages": []any{finding("GHSA-high", "9.8", ""), finding("GO-2026-0001", "", "")},
	}}
	for enforce, failed := range map[bool]int{true: 1, false: 0} {
		calls := runOSVReport(t, "1", results, enforce)
		if len(calls["setFailed"]) != failed || len(calls["error"]) != 1 {
			t.Errorf("enforce %v: calls = %v", enforce, calls)
		}
		// a Go advisory has no severity, so it's a warning
		if !strings.Contains(strings.Join(calls["warning"], "\n"), "GO-2026-0001") {
			t.Errorf("enforce %v: warnings = %v", enforce, calls["warning"])
		}
	}
}

func TestOSVReportPassesWhenThereIsNothingToReport(t *testing.T) {
	for _, status := range []string{"0", "128"} {
		calls := runOSVReport(t, status, []map[string]any{}, true)
		if len(calls["setFailed"]) != 0 || len(calls["warning"]) != 0 {
			t.Errorf("status %s: calls = %v", status, calls)
		}
	}
}

// A scan that couldn't finish fails a pull request, which can't pass unchecked, and warns elsewhere.
func TestOSVReportFailsAPullRequestOnAnIncompleteScan(t *testing.T) {
	for _, tc := range []struct {
		status  string
		results []map[string]any
	}{
		{"127", []map[string]any{}},
		{"1", nil},
	} {
		for enforce, failed := range map[bool]int{true: 1, false: 0} {
			calls := runOSVReport(t, tc.status, tc.results, enforce)
			if len(calls["setFailed"]) != failed || len(calls["warning"]) != 1-failed {
				t.Errorf("status %s, enforce %v: calls = %v", tc.status, enforce, calls)
			}
		}
	}
}
