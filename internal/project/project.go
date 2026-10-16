// Package project works out the module path and names of the project being scaffolded.
package project

import (
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/apollogeddon/forgego/internal/fsys"
)

// Info describes the project forgego is scaffolding into.
type Info struct {
	Module    string // module path, e.g. github.com/acme/billing-api
	Name      string // last element of the module path, without a /vN suffix: billing-api
	Slug      string // Name made safe for an image tag, a systemd unit and a user: billing-api
	Package   string // a valid Go package name for the library starter: billingapi
	HasGoMod  bool
	GoVersion string // the go directive of an existing go.mod
}

var (
	majorSuffix = regexp.MustCompile(`/v[0-9]+$`)
	remoteURL   = regexp.MustCompile(`^\s*url\s*=\s*(\S+)\s*$`)
	// [remote "origin"] gives remote "origin"
	sectionHeader = regexp.MustCompile(`^\s*\[\s*([^\]]*?)\s*\]`)
	nonIdent      = regexp.MustCompile(`[^a-z0-9]`)
	nonSlug       = regexp.MustCompile(`[^a-z0-9]+`)
	sshPort       = regexp.MustCompile(`^([^/]+):[0-9]+/`)
)

// Detect reads go.mod when there is one; otherwise it derives the module path from
// the git remote, falling back to the directory name.
func Detect(fs fsys.FS, dir string) (Info, error) {
	var info Info
	if content, err := fs.ReadFile(filepath.Join(dir, "go.mod")); err == nil {
		f, err := modfile.ParseLax("go.mod", []byte(content), nil)
		if err != nil {
			return info, fmt.Errorf("read go.mod: %w", err)
		}
		if f.Module != nil {
			info.Module = f.Module.Mod.Path
		}
		if f.Go != nil {
			info.GoVersion = f.Go.Version
		}
		info.HasGoMod = true
	} else if !fsys.IsNotExist(err) {
		return info, fmt.Errorf("read go.mod: %w", err)
	}
	if info.Module == "" {
		info.Module = moduleFromRemote(fs, dir)
		if info.Module == "" {
			info.Module = Slug(filepath.Base(dir))
		}
		// go.mod is written from this path, and the go command can't read one with an invalid path
		if err := module.CheckImportPath(info.Module); err != nil {
			return info, fmt.Errorf("no valid module path from the git remote or the directory name (%w): "+
				"run `go mod init <module path>` first", err)
		}
	}
	info.Name = path.Base(majorSuffix.ReplaceAllString(info.Module, ""))
	info.Slug = Slug(info.Name)
	info.Package = PackageName(info.Name)
	return info, nil
}

// moduleFromRemote turns the repository's origin remote (else its first remote) into a module
// path: git@github.com:acme/api.git and https://github.com/acme/api both give
// github.com/acme/api, and the services/billing directory of that repository gives
// github.com/acme/api/services/billing.
func moduleFromRemote(fs fsys.FS, dir string) string {
	root := dir
	content, ok := gitConfig(fs, root)
	for !ok {
		parent := filepath.Dir(root)
		if parent == root {
			return ""
		}
		root = parent
		content, ok = gitConfig(fs, root)
	}
	sub, err := filepath.Rel(root, dir)
	if err != nil {
		return ""
	}
	base := remoteModule(content)
	if base == "" || sub == "." {
		return base
	}
	elems := strings.Split(filepath.ToSlash(sub), "/")
	for i, elem := range elems {
		elems[i] = pathElement(elem)
	}
	return base + "/" + strings.Join(elems, "/")
}

// pathElement makes a directory name a module path element: one that already is one is only
// lowercased, and any other is slugged as the directory-name fallback is, so My Service gives my-service.
func pathElement(name string) string {
	lower := strings.ToLower(name)
	if module.CheckImportPath(lower) == nil {
		return lower
	}
	return Slug(name)
}

// gitConfig reads the config of the repository whose working tree is root. Its .git is the
// git directory, or in a worktree or submodule a file pointing at it; a worktree's git
// directory names the main repository's, which holds the config, in commondir.
func gitConfig(fs fsys.FS, root string) (string, bool) {
	dotGit := filepath.Join(root, ".git")
	if content, err := fs.ReadFile(filepath.Join(dotGit, "config")); err == nil {
		return content, true
	}
	pointer, err := fs.ReadFile(dotGit)
	if err != nil {
		return "", false
	}
	gitDir, ok := strings.CutPrefix(strings.TrimSpace(pointer), "gitdir:")
	if !ok {
		return "", false
	}
	gitDir = resolve(root, strings.TrimSpace(gitDir))
	if common, err := fs.ReadFile(filepath.Join(gitDir, "commondir")); err == nil {
		gitDir = resolve(gitDir, strings.TrimSpace(common))
	}
	content, err := fs.ReadFile(filepath.Join(gitDir, "config"))
	return content, err == nil
}

func resolve(base, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(base, path)
}

// remoteURLOf gives origin's URL, else the first remote's; a [submodule] or other section's
// url isn't the repository's own.
func remoteURLOf(content string) string {
	first, section := "", ""
	for line := range strings.SplitSeq(content, "\n") {
		if match := sectionHeader.FindStringSubmatch(line); match != nil {
			section = match[1]
			continue
		}
		remote, ok := strings.CutPrefix(section, "remote ")
		if !ok {
			continue
		}
		if match := remoteURL.FindStringSubmatch(line); match != nil {
			if strings.Trim(remote, `"`) == "origin" {
				return match[1]
			}
			if first == "" {
				first = match[1]
			}
		}
	}
	return first
}

func remoteModule(content string) string {
	found := remoteURLOf(content)
	if found == "" {
		return ""
	}
	url := strings.TrimSuffix(found, ".git")
	scheme := ""
	if i := strings.Index(url, "://"); i >= 0 {
		scheme, url = url[:i], url[i+3:]
	}
	if at := strings.Index(url, "@"); at >= 0 && at < strings.IndexAny(url+"/", ":/") {
		url = url[at+1:]
	}
	if scheme == "" {
		// scp-like git@host:owner/repo
		url = strings.Replace(url, ":", "/", 1)
	} else {
		// ssh://host:2222/owner/repo: the port isn't part of the module path
		url = sshPort.ReplaceAllString(url, "$1/")
	}
	return strings.ToLower(url)
}

// GitHubOwner is the account the module lives under on GitHub, or "" when it isn't
// a github.com module path.
func (i Info) GitHubOwner() string {
	rest, ok := strings.CutPrefix(i.Module, "github.com/")
	if !ok {
		return ""
	}
	owner, _, _ := strings.Cut(rest, "/")
	return owner
}

// PackageName makes a valid Go package name from a project name: billing-api gives
// billingapi, and a name starting with a digit gets an app prefix.
func PackageName(name string) string {
	pkg := nonIdent.ReplaceAllString(strings.ToLower(unaccented(name)), "")
	if pkg == "" || (pkg[0] >= '0' && pkg[0] <= '9') {
		pkg = "app" + pkg
	}
	return pkg
}

// Slug makes a name safe for a Docker image tag, a systemd unit and a Debian user:
// lowercase letters, digits and single dashes. BillingAPI gives billingapi, and
// "my project" gives my-project.
func Slug(name string) string {
	slug := strings.Trim(nonSlug.ReplaceAllString(strings.ToLower(unaccented(name)), "-"), "-")
	if slug == "" {
		return "app"
	}
	return slug
}

// unaccented takes the accents off letters, so they keep their place in a name: Ünï gives Uni.
func unaccented(name string) string {
	plain, _, err := transform.String(transform.Chain(norm.NFKD, runes.Remove(runes.In(unicode.Mn))), name)
	if err != nil {
		return name
	}
	return plain
}
