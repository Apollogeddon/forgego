// Package version reports which forgego is running, so a project can run the same one.
package version

import (
	"runtime/debug"
	"strings"
)

// Module is forgego's module path.
const Module = "github.com/apollogeddon/forgego"

// override lets tests and the integration test pin a version.
var override string

// Set pins the reported version, for tests.
func Set(v string) { override = v }

// Current is the running forgego's module version, or "latest" for a development build
// that no project could `go run`.
func Current() string {
	if override != "" {
		return override
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "latest"
	}
	v := info.Main.Version
	if v == "" || v == "(devel)" || strings.Contains(v, "+dirty") || strings.HasPrefix(v, "v0.0.0-") {
		return "latest"
	}
	return v
}

// RunCommand runs this forgego version without installing it.
func RunCommand() string {
	return "go run " + Module + "/cmd/forgego@" + Current()
}
