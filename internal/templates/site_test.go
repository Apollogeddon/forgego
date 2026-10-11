package templates

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Checksums of the vendored FlexSearch, as npm's flexsearch@0.8.143 ships them, so a changed copy is noticed.
var flexSearchSums = map[string]string{
	"flexsearch.bundle.min.js": "433e941a8a573ebb9931fc16fc75266ab6b93f569ac2fb4f3dc66882e0416f4c",
	"LICENSE":                  "1eb85fc97224598dad1852b5d6483bbcf0aa8608790dcc657a5a2a761ae9c8c6",
}

func TestVendoredFlexSearchIsUnchanged(t *testing.T) {
	files := SiteFiles()
	if len(files) != len(flexSearchSums) {
		t.Fatalf("%d site files, want %d", len(files), len(flexSearchSums))
	}
	for _, f := range files {
		sum := sha256.Sum256([]byte(f.Content))
		if got, want := hex.EncodeToString(sum[:]), flexSearchSums[filepath.Base(f.Path)]; got != want {
			t.Errorf("%s has sha256 %s, want %s", f.Path, got, want)
		}
	}
	if header := "FlexSearch.js v" + FlexSearchVersion + " "; !strings.Contains(files[0].Content[:200], header) {
		t.Errorf("the vendored script isn't FlexSearch %s", FlexSearchVersion)
	}
}

// Dependabot bumps Hextra in docs/go.mod; a new Hextra may load a new FlexSearch, so the pins move together.
func TestHextraPinMatchesTheDocsSite(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "docs", "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "github.com/imfing/hextra "+HextraVersion+" ") {
		t.Errorf("docs/go.mod's Hextra isn't %s: check which FlexSearch the new Hextra loads, then move HextraVersion",
			HextraVersion)
	}
}
