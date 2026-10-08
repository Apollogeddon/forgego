package yamlx

import (
	"strings"
	"testing"
)

func mustParse(t *testing.T, s string) string {
	t.Helper()
	root, err := Parse(s)
	if err != nil {
		t.Fatal(err)
	}
	out, err := Encode(root)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestMergeCombinesMappingsAndSequences(t *testing.T) {
	base, _ := Parse("linters:\n  enable: [gosec, revive]\n  default: standard\nrun:\n  timeout: 5m\n")
	local, _ := Parse("linters:\n  enable: [revive, wsl]\n  default: none\n")
	out, err := Encode(Merge(base, local))
	if err != nil {
		t.Fatal(err)
	}
	want := "linters:\n  enable: [gosec, revive, wsl]\n  default: none\nrun:\n  timeout: 5m\n"
	if out != want {
		t.Errorf("got\n%s\nwant\n%s", out, want)
	}
}

func TestSetIfAbsentKeepsExistingUnlessForced(t *testing.T) {
	m, _ := Parse("lint: mine # keep me\n")
	if SetIfAbsent(m, "lint", Str("theirs"), false) {
		t.Error("changed an existing key without force")
	}
	out, _ := Encode(m)
	if !strings.Contains(out, "mine # keep me") {
		t.Errorf("lost the existing value or comment: %s", out)
	}
	SetIfAbsent(m, "lint", Str("theirs"), true)
	if Get(m, "lint").Value != "theirs" {
		t.Error("force didn't replace the key")
	}
}

func TestParseEmptyDocumentIsAnEmptyMapping(t *testing.T) {
	if got := mustParse(t, ""); got != "{}\n" {
		t.Errorf("got %q", got)
	}
}
