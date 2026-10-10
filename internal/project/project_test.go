package project

import (
	"testing"

	"github.com/apollogeddon/forgego/internal/fsys"
)

func TestRemoteModule(t *testing.T) {
	for url, want := range map[string]string{
		"git@github.com:Acme/API.git":               "github.com/acme/api",
		"https://github.com/acme/api":               "github.com/acme/api",
		"https://user@github.com/acme/api.git":      "github.com/acme/api",
		"ssh://git@github.com/acme/api.git":         "github.com/acme/api",
		"ssh://git@gitlab.example.com:2222/a/b.git": "gitlab.example.com/a/b",
		"ssh://gitlab.example.com:22/a/b":           "gitlab.example.com/a/b",
	} {
		if got := remoteModule("[remote \"origin\"]\n\turl = " + url + "\n"); got != want {
			t.Errorf("%s: got %q, want %q", url, got, want)
		}
	}
}

func TestDetectMakesNamesSafe(t *testing.T) {
	fs := fsys.NewMemory(map[string]string{"/w/go.mod": "module github.com/Acme/BillingAPI/v3\n\ngo 1.25.0\n"})
	info, err := Detect(fs, "/w")
	if err != nil {
		t.Fatal(err)
	}
	if info.Name != "BillingAPI" || info.Slug != "billingapi" || info.Package != "billingapi" || info.GoVersion != "1.25.0" {
		t.Errorf("info = %+v", info)
	}

	info, err = Detect(fsys.NewMemory(nil), "/work/My Project")
	if err != nil {
		t.Fatal(err)
	}
	if info.Module != "my-project" || info.Slug != "my-project" {
		t.Errorf("info = %+v", info)
	}
}

func TestGitHubOwner(t *testing.T) {
	for module, want := range map[string]string{
		"github.com/acme/api":                  "acme",
		"github.com/acme/api/services/billing": "acme",
		"gitlab.com/acme/api":                  "",
		"billing-api":                          "",
	} {
		if got := (Info{Module: module}).GitHubOwner(); got != want {
			t.Errorf("GitHubOwner(%q) = %q, want %q", module, got, want)
		}
	}
}
