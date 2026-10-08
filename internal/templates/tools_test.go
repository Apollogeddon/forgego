package templates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestToolCopiesMatchTheToolsModules(t *testing.T) {
	for _, tool := range Tools {
		for embedded, source := range map[string]string{tool.ModFile(): "go.mod", tool.SumFile(): "go.sum"} {
			b, err := os.ReadFile(filepath.Join("..", "..", "tools", tool.Name, source))
			if err != nil {
				t.Fatal(err)
			}
			if string(b) != embedded {
				t.Errorf("%s/%s changed: run `go generate ./...`", tool.Name, source)
			}
		}
	}
}
