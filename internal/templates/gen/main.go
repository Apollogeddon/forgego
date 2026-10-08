// Command gen copies each pinned tool module from tools/<name>/ into the embedded
// templates as <name>.mod and <name>.sum: a go.mod would make its directory a module,
// which go:embed can't reach. Run it with `go generate ./...` after a tool version
// changes; a test fails while the copies are out of date.
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if err := run(filepath.Join("..", "..", "tools"), "tools"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0o755); err != nil { //nolint:gosec // committed to the repository
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		for from, to := range map[string]string{"go.mod": ".mod", "go.sum": ".sum"} {
			b, err := os.ReadFile(filepath.Join(src, entry.Name(), from))
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(dst, entry.Name()+to), b, 0o644); err != nil { //nolint:gosec // committed to the repository
				return err
			}
		}
	}
	return nil
}
