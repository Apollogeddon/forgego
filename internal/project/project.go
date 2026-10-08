// Package project works out the module path and names of the project being scaffolded.
package project

import (
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/mod/modfile"

	"github.com/apollogeddon/forgego/internal/fsys"
)

// Info describes the project forgego is scaffolding into.
type Info struct {
	Module    string // module path, e.g. github.com/acme/billing-api
	Name      string // last element of the module path, without a /vN suffix: billing-api
	Package   string // a valid Go package name for the library starter: billingapi
	HasGoMod  bool
	GoVersion string // the go directive of an existing go.mod
}

var (
	majorSuffix = regexp.MustCompile(`/v[0-9]+$`)
	remoteURL   = regexp.MustCompile(`(?m)^\s*url\s*=\s*(\S+)\s*$`)
	nonIdent    = regexp.MustCompile(`[^a-z0-9]`)
)

// Detect reads go.mod when there is one; otherwise it derives the module path from
// the git remote, falling back to the directory name.
func Detect(fs fsys.FS, dir string) (Info, error) {
	var info Info
	if content, err := fs.ReadFile(filepath.Join(dir, "go.mod")); err == nil {
		f, err := modfile.ParseLax("go.mod", []byte(content), nil)
		if err != nil {
			return info, err
		}
		if f.Module != nil {
			info.Module = f.Module.Mod.Path
		}
		if f.Go != nil {
			info.GoVersion = f.Go.Version
		}
		info.HasGoMod = true
	} else if !fsys.IsNotExist(err) {
		return info, err
	}
	if info.Module == "" {
		info.Module = moduleFromRemote(fs, dir)
	}
	if info.Module == "" {
		info.Module = strings.ToLower(filepath.Base(dir))
	}
	info.Name = path.Base(majorSuffix.ReplaceAllString(info.Module, ""))
	info.Package = PackageName(info.Name)
	return info, nil
}

// moduleFromRemote turns the first remote URL in the repository's .git/config into a
// module path: git@github.com:acme/api.git and https://github.com/acme/api both give
// github.com/acme/api, and the services/billing directory of that repository gives
// github.com/acme/api/services/billing.
func moduleFromRemote(fs fsys.FS, dir string) string {
	root, content := dir, ""
	for {
		if c, err := fs.ReadFile(filepath.Join(root, ".git", "config")); err == nil {
			content = c
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			return ""
		}
		root = parent
	}
	sub, err := filepath.Rel(root, dir)
	if err != nil {
		return ""
	}
	module := remoteModule(content)
	if module == "" || sub == "." {
		return module
	}
	return module + "/" + strings.ToLower(filepath.ToSlash(sub))
}

func remoteModule(content string) string {
	match := remoteURL.FindStringSubmatch(content)
	if match == nil {
		return ""
	}
	url := strings.TrimSuffix(match[1], ".git")
	for _, prefix := range []string{"https://", "http://", "ssh://", "git@"} {
		url = strings.TrimPrefix(url, prefix)
	}
	url = strings.Replace(url, ":", "/", 1)
	if at := strings.Index(url, "@"); at >= 0 {
		url = url[at+1:]
	}
	return strings.ToLower(url)
}

// PackageName makes a valid Go package name from a project name: billing-api gives
// billingapi, and a name starting with a digit gets an app prefix.
func PackageName(name string) string {
	pkg := nonIdent.ReplaceAllString(strings.ToLower(name), "")
	if pkg == "" || (pkg[0] >= '0' && pkg[0] <= '9') {
		pkg = "app" + pkg
	}
	return pkg
}
