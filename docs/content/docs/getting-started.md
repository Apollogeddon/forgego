---
title: Getting started
weight: 1
---

This page covers installing Forge.go, scaffolding a project, and the reference for the CLI flags and the tasks it generates.

## Requirements

- [Go](https://go.dev/dl/). When the pinned tools need a newer Go than the project declares, `go` downloads it automatically through the `toolchain` line in `go.mod` (unless you've set `GOTOOLCHAIN=local`).
- Git, for the generated lefthook hooks.
- Docker, only for the `docker:build` and `docker:run` tasks that `--docker` adds.

## Setup

### 1. Install Forge.go

Install Forge.go with `go install`:

```bash
go install github.com/apollogeddon/forgego/v2/cmd/forgego@latest
```

Or run it once without installing:

```bash
go run github.com/apollogeddon/forgego/v2/cmd/forgego@latest init
```

Generated projects don't need Forge.go installed: their `Taskfile.yml` runs the version pinned in `.forgego/forgego/go.mod` with `go tool` (see [`sync`](#keeping-up-to-date)).

### 2. Initialise a project

Run `init` in an existing project, or in an empty directory to start a new one:

```bash
forgego init            # Go service or command-line tool (default)
forgego init --library  # Go module others import
forgego init --website  # static documentation site (Hugo)
```

Then tidy the module and install the Git hooks:

```bash
go mod tidy
go tool -modfile=.forgego/task/go.mod task hooks
```

Skip `go mod tidy` for a website: Hugo manages a site's `go.mod` itself, and tidying would drop the theme. Run `go tool -modfile=.forgego/task/go.mod task dev` to preview the site instead.

`init` is safe to re-run. It creates missing files and tasks, and leaves anything that already exists alone.

### 3. Overwrite existing files (optional)

Pass `--force` to overwrite existing config files, tasks and Taskfile vars with the Forge.go defaults:

```bash
forgego init --force
```

`--force` never touches your own source code: `go.mod`, `.gitignore`, `.golangci.local.yml` and the starter source and tests are only ever created when missing. Use `--dry-run` first to see exactly what would change.

`--force` is also what lets `init` delete files: turning a feature off with `--no-<feature>` (or dropping `--docker` or `--debian`) removes the files that feature created only when `--force` is passed. Without it, `init` reports the files it is skipping.

> [!WARNING]
> Repeat the optional flags when re-running with `--force`. `forgego init --force` in a project scaffolded with `--docker` deletes its `Dockerfile` and `.dockerignore`, because `--docker` wasn't passed this time.

## Running tasks

Every task runs through the pinned copy of [Task](https://taskfile.dev/):

```bash
go tool -modfile=.forgego/task/go.mod task <name>
```

To shorten it, alias it, or [install Task](https://taskfile.dev/installation/) and run `task <name>`. Both read the same `Taskfile.yml`:

```bash
alias task='go tool -modfile=.forgego/task/go.mod task'
```

The rest of this site writes `task <name>`. `task --list` shows every task with its description.

## Generated tasks

Forge.go adds these tasks to `Taskfile.yml`. Tools are referenced through Taskfile vars (`{{.GOLANGCI_LINT}}` and so on), each set to `go tool -modfile=.forgego/<tool>/go.mod <tool>`.

| Task | Command | Added when |
| :--- | :--- | :--- |
| `sync` | `forgego sync` | always |
| `sync-check` | `forgego sync --check` | always |
| `hooks` | `lefthook install` | linting on |
| `lint` | `golangci-lint fmt`, then `golangci-lint run --fix` | linting on, not `--website` |
| `format` | `golangci-lint fmt` | linting on, not `--website` |
| `security` | `govulncheck ./...` | not `--website` |
| `type` | `go vet ./...` | not `--website` |
| `test` | `gotestsum --junitfile junit-report.xml -- -coverprofile=coverage.out -covermode=atomic ./...` | testing on, not `--website` |
| `commit-msg` | `forgego commit-msg {{.CLI_ARGS}}` | versioning on |
| `build` | `go build -trimpath -ldflags "-X main.version={{.VERSION}}" -o dist/<name>{{exeExt}} ./cmd/<name>` | `--backend` |
| `build` | `go build ./...` | `--library` |
| `build` | `hugo --gc --minify` | `--website` |
| `start` | `go run ./cmd/<name> {{.CLI_ARGS}}` | `--backend` |
| `release:snapshot` | GoReleaser `release --snapshot --clean`: every release binary (and the `.deb`) into `dist/`, without publishing | `--backend` |
| `dev` | `hugo server` | `--website` |
| `docker:build` / `docker:run` | `docker build -t <image> .` / `docker run --rm <image>` (`-p 8080:80` for websites), where `<image>` is the lowercased project name (see [Module path](#module-path)) | `--docker` |

A backend also gets a `VERSION` var, `dev` by default, which `build` stamps into `main.version`. `{{exeExt}}` adds `.exe` on Windows. Pass arguments to `start` after `--`: `task start -- --help`.

Existing tasks and vars with the same name are kept unless you pass `--force`. Tasks you add yourself are never touched. With `--force`, turning a feature off also removes the tasks and vars it added, along with its files.

## CLI options

```text
forgego init [options]       Scaffold the current (or target) project
forgego sync [--check]       Refresh the files forgego manages
forgego commit-msg FILE      Check a commit message against Conventional Commits
forgego --version

-C DIR                       Run as if forgego was started in DIR
```

`-C DIR` (or `--path DIR`) before the command works like `git -C` and `go -C`.

### `init` options

| Option | Description |
| :--- | :--- |
| `--backend` | Service or command-line tool (default). |
| `--library` | Go module others import. |
| `--website` | Static documentation site built with [Hugo](https://gohugo.io/). |
| `--testing` / `--no-testing` | `go test` through gotestsum (default: on). |
| `--linting` / `--no-linting` | golangci-lint and lefthook Git hooks (default: on). |
| `--versioning` / `--no-versioning` | release-please and commit message checks (default: on). |
| `--all` / `--no-all` | Turn every standard feature on or off at once; an explicit flag such as `--testing` still wins. |
| `--docker` | Add a `Dockerfile` and a Docker CI job (not available for `--library`). |
| `--debian` | Add a `.deb` package with a systemd unit (`--backend` only). |
| `--force` | Overwrite existing config files and tasks. |
| `--dry-run` | Show what would change without writing anything. |
| `--go VERSION` | Target Go version, a minor or a patch release (default: the existing `go.mod`'s, else Forge.go's default). |
| `-C DIR`, `--path DIR` | Target directory (default: the current directory). |

`--website` scaffolds a static documentation site built with Hugo and the Hextra theme, which Hugo fetches as a Go module: `hugo.toml`, `content/_index.md` and `content/docs/_index.md`, plus the theme's search script in `assets/forgego/flexsearch/`, so the site builds without network access. There is no Go source. `task dev` serves the site locally with live reload, and `task build` writes it to `public/`. In CI, the `website` pipeline builds the site and `pages` deploys `public/` to GitHub Pages.

Choose at most one mode. An invalid combination — more than one mode, a feature both on and off, `--docker` with `--library`, `--debian` outside backend mode, or a `--go` that isn't a Go release — stops `init` before it writes anything, with exit code `2`. An option or argument `init` doesn't know exits with code `1`.

`--go` sets the `go` line of a new `go.mod` (a minor release gets a `.0` patch) and the `golang` image the `Dockerfile` builds with. CI reads its Go version from `go.mod`, so the generated workflow passes no `go_version`. Without `--go`, an existing project's `go.mod` decides both. `init` never changes an existing `go.mod`'s `go` line, so to move such a project to another Go, edit `go.mod` (CI follows it) and pass the same `--go` with `--force` to rewrite the `Dockerfile`.

`--no-testing`, `--no-linting` and `--no-versioning` also switch off the matching stages and jobs in the generated `.github/workflows/.index.yml`: see [Generated workflow]({{< relref "workflows/_index.md#generated-workflow" >}}). If `.github/workflows/index.yml` calls Forge.go's pipelines, `init` removes it: see [An existing index.yml]({{< relref "workflows/_index.md#an-existing-indexyml" >}}).

### Keeping up to date

`forgego sync` refreshes the files Forge.go manages; `forgego sync --check` reports drift and exits `1` without writing anything. Both also accept `-C DIR`. See [Configuration]({{< relref "configuration.md#managed-files" >}}).

The generated `Taskfile.yml` runs Forge.go through a `FORGEGO` var, `go tool -modfile=.forgego/forgego/go.mod forgego`, from a pin in `.forgego/forgego/go.mod` set to the version that scaffolded the project. Dependabot proposes each new Forge.go release there, like any other module update, and the pull request auto-merges once CI passes:

- a new release usually changes the files Forge.go manages, so on Dependabot's pull requests the `lint` stage runs `task sync` instead of `task sync-check`, and checks the refreshed files;
- once the update merges, the `merge` pipeline's `auto-merged` dispatch runs the `patch` pipeline, which runs `task sync` on `main` and commits the refreshed files, so `task sync-check` keeps passing on other pull requests.

`sync` never changes the `FORGEGO` var, so one that runs Forge.go some other way, such as a local build, is left alone.

#### A pin without `/v2`

Forge.go is published as `github.com/apollogeddon/forgego/v2`. If `.forgego/forgego/go.mod` requires `github.com/apollogeddon/forgego` without `/v2`, Dependabot can't propose a release there. Fix it once with `sync`:

```bash
go run github.com/apollogeddon/forgego/v2/cmd/forgego@latest sync
task sync-check
```

`sync` rewrites `.forgego/forgego/go.mod` to require `github.com/apollogeddon/forgego/v2` at the version it is, with its `tool` line naming `github.com/apollogeddon/forgego/v2/cmd/forgego`, replaces its `go.sum`, and refreshes the other managed files. It also points the Forge.go entry in `.github/dependabot.yml` at the `/v2` module. Commit the result; Dependabot proposes each release from then on.

Two other layouts need changing by hand first, as `sync` doesn't change them:

- **Tool pins at `.forgego/<tool>.mod` and `.sum`**: move each pair to `.forgego/<tool>/go.mod` and `go.sum`, change every `-modfile=.forgego/<tool>.mod` in `Taskfile.yml` and `lefthook.yml` to `-modfile=.forgego/<tool>/go.mod`, and run `task hooks` to reinstall the Git hooks at the new path.
- **A `FORGEGO` var that runs `go run github.com/apollogeddon/forgego/cmd/forgego@<version>`**: set it to `go tool -modfile=.forgego/forgego/go.mod forgego`. The `sync` above then pins Forge.go in `.forgego/forgego/go.mod`. For Dependabot to propose each release, add the `/.forgego/forgego` entry from a new project's `.github/dependabot.yml`.

### Commit message checks

`forgego commit-msg FILE` checks the first line of a commit message (skipping blank and `#` comment lines) against [Conventional Commits](https://www.conventionalcommits.org/): `<type>(<scope>): <subject>`, with an optional `!` before the colon for a breaking change.

- **Types:** `build`, `chore`, `ci`, `docs`, `feat`, `fix`, `perf`, `refactor`, `revert`, `style`, `test`.
- **Length:** the subject line can be at most 100 characters.
- **Exempt:** messages starting with `Merge `, `Revert "`, `fixup! `, `squash! ` or `amend! `, which Git and other tools write.

With versioning on, lefthook runs it as the `commit-msg` hook.

## Module path

`init` works out the module path in this order:

1. The `module` line of an existing `go.mod`.
2. The enclosing Git repository's `origin` remote (else its first remote), plus the subdirectory. A worktree or submodule works too, as `init` follows its `.git` file to the repository's config: `git@github.com:acme/api.git` and `https://github.com/acme/api` both give `github.com/acme/api`, and its `services/billing` directory gives `github.com/acme/api/services/billing`. Each subdirectory is lowercased, and one that isn't a valid module path element is turned into one as the directory name is below (`My Service` gives `my-service`).
3. The directory name, lowercased with anything but letters and digits turned into dashes (`My Project` gives `my-project`).

If the path it works out still isn't a valid module path, as from a `file://` remote, `init` stops with exit code `1` before it writes anything; run `go mod init <module path>` first.

The project name is the last element of the module path without a `/vN` suffix (`billing-api`). Accents come off letters first (`Café` gives `cafe`). The library starter's package name is that name lowercased with anything but letters and digits removed (`billingapi`), prefixed with `app` if it would start with a digit. The Docker image, the `.deb` package, its systemd unit and its system user use the name lowercased, with anything but letters and digits turned into dashes (`BillingAPI` gives `billingapi`).

## Project structure

A default `forgego init` (backend) in a project named `billing-api`, with a `github.com` module path, produces:

```text
.
├── .forgego/
│   ├── golangci.yml            # managed base config — refreshed by `forgego sync`
│   ├── forgego/                # Forge.go's own pin, which Dependabot updates
│   ├── golangci-lint/          # pinned tools (go.mod, go.sum) — refreshed by `forgego sync`
│   ├── gotestsum/
│   ├── govulncheck/
│   ├── lefthook/
│   └── task/
├── .github/
│   ├── .release.json           # release-please manifest
│   ├── CODEOWNERS              # requests your review on others' pull requests
│   ├── dependabot.yml          # weekly updates with a 3-day cooldown
│   ├── release.json            # release-please config
│   └── workflows/.index.yml    # CI/CD calling Forge.go's pipelines
├── cmd/billing-api/
│   ├── main.go
│   └── main_test.go
├── .editorconfig
├── .gitignore
├── .golangci.local.yml         # your changes to the lint config
├── .golangci.yml               # generated from the two above — don't edit
├── .goreleaser.yaml
├── go.mod
├── lefthook.yml
└── Taskfile.yml
```

`--library` replaces `cmd/` with `billingapi.go` and `billingapi_test.go` at the root, and has no `.goreleaser.yaml`. `--website` has no Go source, lint config or tests: it adds `hugo.toml`, `content/_index.md`, `content/docs/_index.md`, `assets/forgego/flexsearch/` and `.forgego/hugo/go.mod`. `--docker` adds `Dockerfile` and `.dockerignore`, and `--debian` adds a `packaging/` directory.

The starter source is only created in a new project, one without a `go.mod`: an existing project keeps its own code, and the build tasks point at `./cmd/<name>`, so adjust them if your `main` package lives elsewhere. The starter test is only created next to the untouched starter source, never beside your own code.
