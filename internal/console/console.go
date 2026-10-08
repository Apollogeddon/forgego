// Package console prints forgego's progress in one consistent style.
package console

import (
	"fmt"
	"io"
	"os"
)

var (
	stdout io.Writer = os.Stdout
	stderr io.Writer = os.Stderr
)

// SetOutput redirects console output, for tests.
func SetOutput(out, errOut io.Writer) {
	stdout, stderr = out, errOut
}

func OK(format string, args ...any)   { fmt.Fprintf(stdout, "✅ "+format+"\n", args...) }
func Warn(format string, args ...any) { fmt.Fprintf(stdout, "⚠️  "+format+"\n", args...) }
func Err(format string, args ...any)  { fmt.Fprintf(stderr, "❌ "+format+"\n", args...) }
func Dry(format string, args ...any)  { fmt.Fprintf(stdout, "[DryRun] "+format+"\n", args...) }
func Info(format string, args ...any) { fmt.Fprintf(stdout, format+"\n", args...) }
