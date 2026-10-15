package templates

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
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

func secureSteps(t *testing.T) map[string]osvStep {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "quality.yml"))
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
	for _, step := range w.Jobs["secure"].Steps {
		steps[step.Name] = step
	}
	return steps
}

func TestOSVScanNeverSwallowsTheResult(t *testing.T) {
	steps := secureSteps(t)
	if run := steps["OSV Scan"].Run; strings.Contains(run, "|| true") || !strings.Contains(run, "--format json") {
		t.Errorf("OSV Scan run = %q", run)
	}
	if got := steps["Report OSV findings"].Env["OSV_ENFORCE"]; got != "${{ github.event_name == 'pull_request' }}" {
		t.Errorf("OSV_ENFORCE = %q", got)
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
	for _, tc := range []struct {
		status  string
		results []map[string]any
		warned  bool
	}{
		{"0", []map[string]any{}, false},
		{"128", []map[string]any{}, false},
		{"127", []map[string]any{}, true},
		{"1", nil, true},
	} {
		calls := runOSVReport(t, tc.status, tc.results, true)
		if len(calls["setFailed"]) != 0 || (len(calls["warning"]) > 0) != tc.warned {
			t.Errorf("status %s: calls = %v", tc.status, calls)
		}
	}
}
