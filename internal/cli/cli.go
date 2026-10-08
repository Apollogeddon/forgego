// Package cli parses forgego's command line and dispatches to init, sync and commit-msg.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/apollogeddon/forgego/internal/commitmsg"
	"github.com/apollogeddon/forgego/internal/config"
	"github.com/apollogeddon/forgego/internal/console"
	"github.com/apollogeddon/forgego/internal/core"
	"github.com/apollogeddon/forgego/internal/fsys"
	"github.com/apollogeddon/forgego/internal/sync"
	"github.com/apollogeddon/forgego/internal/version"
)

const usage = `forgego scaffolds tooling into Go projects: linting, testing, CI/CD, Docker and
Debian packaging.

Usage:
  forgego init [options]       Scaffold the current (or target) project
  forgego sync [--check]       Refresh the files forgego manages
  forgego commit-msg FILE      Check a commit message against Conventional Commits
  forgego --version

  -C DIR                       Run as if forgego was started in DIR

Run forgego <command> --help for a command's options.
`

const initUsage = `Usage: forgego init [options]

Scaffold tooling into the current (or target) project. Safe to re-run: existing
files are left alone unless --force is passed.

Modes (default --backend):
  --backend            Service or command-line tool
  --library            Go module others import
  --website            Static documentation site (Hugo)

Standard features (on by default; turn off with --no-<feature> or --no-all):
  --testing            go test through gotestsum
  --linting            golangci-lint and lefthook git hooks
  --versioning         release-please and commit message checks
  --all                Turn every standard feature on; an explicit flag still wins

Optional features:
  --docker             Add a Dockerfile (not available for --library)
  --debian             Add a .deb package with a systemd unit (--backend only)

Options:
  --force              Overwrite existing config files and tasks
  --dry-run            Show what would change without writing
  --go VERSION         Target Go version (default ` + config.DefaultGo + `)
  -C, --path DIR       Target directory (default: current directory)
`

// Run executes forgego with args (without the program name) and returns the exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	console.SetOutput(stdout, stderr)
	// like git -C and go -C: run as if started in another directory
	if len(args) >= 2 && (args[0] == "-C" || args[0] == "--path") {
		if err := os.Chdir(args[1]); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		args = args[2:]
	}
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 1
	}
	switch args[0] {
	case "init":
		return runInit(args[1:], stdout, stderr)
	case "sync":
		return runSync(args[1:], stderr)
	case "commit-msg":
		return runCommitMsg(args[1:], stderr)
	case "-V", "--version", "-version", "version":
		fmt.Fprintf(stdout, "forgego %s\n", version.Current())
		return 0
	case "-h", "--help", "-help", "help":
		fmt.Fprint(stdout, usage)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n%s", args[0], usage)
		return 1
	}
}

// feature is a standard feature with --<name> and --no-<name> flags.
type feature struct {
	on, off bool
}

func (f *feature) register(fs *flag.FlagSet, name string) {
	fs.BoolVar(&f.on, name, false, "")
	fs.BoolVar(&f.off, "no-"+name, false, "")
}

// resolve gives the explicit flag, then --all/--no-all, then on.
func (f feature) resolve(all feature) bool {
	switch {
	case f.on:
		return true
	case f.off:
		return false
	case all.off:
		return false
	default:
		return true
	}
}

func runInit(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() { fmt.Fprint(stdout, initUsage) }

	var backend, library, website, docker, debian, force, dryRun bool
	var all, testing, linting, versioning feature
	goVersion, target := config.DefaultGo, "."
	fs.BoolVar(&backend, "backend", false, "")
	fs.BoolVar(&library, "library", false, "")
	fs.BoolVar(&website, "website", false, "")
	all.register(fs, "all")
	testing.register(fs, "testing")
	linting.register(fs, "linting")
	versioning.register(fs, "versioning")
	fs.BoolVar(&docker, "docker", false, "")
	fs.BoolVar(&debian, "debian", false, "")
	fs.BoolVar(&force, "force", false, "")
	fs.BoolVar(&dryRun, "dry-run", false, "")
	fs.StringVar(&goVersion, "go", goVersion, "")
	fs.StringVar(&target, "path", target, "")
	fs.StringVar(&target, "C", target, "")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 1
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "unexpected argument %q\n", fs.Arg(0))
		return 1
	}
	modes := 0
	for _, set := range []bool{backend, library, website} {
		if set {
			modes++
		}
	}
	if modes > 1 {
		fmt.Fprintln(stderr, "choose one of --backend, --library or --website")
		return 1
	}
	for _, f := range []feature{all, testing, linting, versioning} {
		if f.on && f.off {
			fmt.Fprintln(stderr, "a feature can't be both on and off")
			return 1
		}
	}

	abs, err := filepath.Abs(target)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	cfg := config.Default(abs)
	switch {
	case library:
		cfg.Mode = config.Library
	case website:
		cfg.Mode = config.Website
	}
	cfg.Testing = testing.resolve(all)
	cfg.Linting = linting.resolve(all)
	cfg.Versioning = versioning.resolve(all)
	cfg.Docker, cfg.Debian = docker, debian
	cfg.Force, cfg.DryRun = force, dryRun
	cfg.Go = goVersion
	return core.Init(cfg, fsys.OS{})
}

func runSync(args []string, stderr io.Writer) int {
	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	fs.SetOutput(stderr)
	check := fs.Bool("check", false, "report drift without writing; exit 1 if any")
	target := "."
	fs.StringVar(&target, "path", target, "target directory")
	fs.StringVar(&target, "C", target, "target directory")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 1
	}
	return sync.Run(fsys.OS{}, target, *check)
}

func runCommitMsg(args []string, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: forgego commit-msg FILE")
		return 1
	}
	b, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := commitmsg.Check(string(b)); err != nil {
		console.Err("%s", strings.TrimSpace(err.Error()))
		return 1
	}
	return 0
}
