// Package version reports which forgego is running, so a project can run the same one.
package version

import (
	"runtime/debug"
	"strings"
)

// Module is forgego's module path.
const Module = "github.com/apollogeddon/forgego"

// override lets tests pin a version.
var override string

// injected is set by release builds:
// -ldflags "-X github.com/apollogeddon/forgego/internal/version.injected=v1.2.3".
var injected string

// Set pins the reported version, for tests.
func Set(v string) { override = v }

// Current is the running forgego's module version, or "latest" for a development build
// that no project could `go run`.
func Current() string {
	if override != "" {
		return override
	}
	if injected != "" {
		return injected
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

// runPrefix starts every command that runs a published forgego.
const runPrefix = "go run " + Module + "/cmd/forgego@"

// RunCommand runs this forgego version without installing it.
func RunCommand() string {
	return runPrefix + Current()
}

// IsPublishedRun reports whether command runs a published forgego, rather than one the
// project chose itself, such as a local build.
func IsPublishedRun(command string) bool {
	return strings.HasPrefix(command, runPrefix)
}
