---
title: Getting started
weight: 1
---

This page covers installing Forge.go, scaffolding a project, and the reference for the CLI flags and the tasks it generates.

## Requirements

- [Go](https://go.dev/dl/). Forge.go itself builds with Go 1.26 or newer. The pinned tools need Go 1.27, which `go` downloads automatically through the `toolchain` line in `go.mod` (unless you've set `GOTOOLCHAIN=local`).
- Git, for the generated lefthook hooks.
- Docker, only for the `docker:build` and `docker:run` tasks that `--docker` adds.

## Setup

### 1. Install Forge.go

Install Forge.go with `go install`:

```bash
go install github.com/apollogeddon/forgego/cmd/forgego@latest
```

Or run it once without installing:

```bash
go run github.com/apollogeddon/forgego/cmd/forgego@latest init
```

Generated projects don't need Forge.go installed: their `Taskfile.yml` runs the same version with `go run` (see [`sync`](#keeping-up-to-date)).

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
| `--go VERSION` | Target Go version, such as `1.27` or `1.27.1` (default: the existing `go.mod`'s, else `1.27`). |
| `-C DIR`, `--path DIR` | Target directory (default: the current directory). |

`--website` scaffolds a static documentation site built with Hugo and the Hextra theme, which Hugo fetches as a Go module: `hugo.toml`, `content/_index.md` and `content/docs/_index.md`. There is no Go source. `task dev` serves the site locally with live reload, and `task build` writes it to `public/`. In CI, the `website.yml` workflow builds the site and deploys `public/` to GitHub Pages.

Choose at most one mode; passing more than one exits with code `1`. An invalid combination — `--docker` with `--library`, `--debian` outside backend mode, or a `--go` that isn't a Go release — stops `init` before it writes anything, with exit code `2`.

`--go` sets the `go` line of a new `go.mod` (`1.27` becomes `go 1.27.0`), the `go_version` passed to CI, and the `golang` image the `Dockerfile` builds with. Without it, an existing project's `go.mod` decides those, and its `go` line is never changed.

`--no-testing`, `--no-linting` and `--no-versioning` also switch off the matching jobs in the generated CI workflow.

### Keeping up to date

`forgego sync` refreshes the files Forge.go manages; `forgego sync --check` reports drift and exits `1` without writing anything. Both also accept `-C DIR`. See [Configuration]({{< relref "configuration.md#managed-files" >}}).

The generated `Taskfile.yml` runs Forge.go through a `FORGEGO` var, `go tool -modfile=.forgego/forgego/go.mod forgego`, from a pin in `.forgego/forgego/go.mod` set to the version that scaffolded the project. Dependabot proposes each new Forge.go release there daily, and the pull request auto-merges once CI passes:

- a new release usually changes the files Forge.go manages, so on Dependabot's pull requests the CI linting job runs `task sync` instead of `task sync-check`, and checks the refreshed files;
- once the update merges, the testing workflow's `patch` job runs `task sync` on `main` and commits the refreshed files.

A project scaffolded by an earlier Forge.go runs it with `go run …/forgego@<version>`. Run a newer version's `sync` once to move it to the pin:

```bash
go run github.com/apollogeddon/forgego/cmd/forgego@latest sync
```

A `FORGEGO` var that runs Forge.go some other way, such as a local build, is left alone.

### Commit message checks

`forgego commit-msg FILE` checks the first line of a commit message (skipping blank and `#` comment lines) against [Conventional Commits](https://www.conventionalcommits.org/): `<type>(<scope>): <subject>`, with an optional `!` before the colon for a breaking change.

- **Types:** `build`, `chore`, `ci`, `docs`, `feat`, `fix`, `perf`, `refactor`, `revert`, `style`, `test`.
- **Length:** the subject line can be at most 100 characters.
- **Exempt:** messages starting with `Merge `, `Revert "`, `fixup! `, `squash! ` or `amend! `, which Git and other tools write.

With versioning on, lefthook runs it as the `commit-msg` hook.

## Module path

`init` works out the module path in this order:

1. The `module` line of an existing `go.mod`.
2. The first remote in the enclosing Git repository's `.git/config`, plus the subdirectory: `git@github.com:acme/api.git` and `https://github.com/acme/api` both give `github.com/acme/api`, and its `services/billing` directory gives `github.com/acme/api/services/billing`.
3. The directory name.

The project name is the last element of the module path without a `/vN` suffix (`billing-api`). The library starter's package name is that name lowercased with anything but letters and digits removed (`billingapi`), prefixed with `app` if it would start with a digit. The Docker image, the `.deb` package, its systemd unit and its system user use the name lowercased, with anything but letters and digits turned into dashes (`BillingAPI` gives `billingapi`).

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
│   ├── dependabot.yml          # weekly updates with a 3-day cooldown; our own modules daily
│   ├── release.json            # release-please config
│   └── workflows/index.yml     # CI/CD calling the reusable workflows
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

`--library` replaces `cmd/` with `billingapi.go` and `billingapi_test.go` at the root, and has no `.goreleaser.yaml`. `--website` has no Go source, lint config or tests: it adds `hugo.toml`, `content/_index.md`, `content/docs/_index.md` and `.forgego/hugo/go.mod`. `--docker` adds `Dockerfile` and `.dockerignore`, and `--debian` adds a `packaging/` directory.

The starter source is only created in a new project, one without a `go.mod`: an existing project keeps its own code, and the build tasks point at `./cmd/<name>`, so adjust them if your `main` package lives elsewhere. The starter test is only created next to the untouched starter source, never beside your own code.
