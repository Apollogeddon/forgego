---
title: Configuration
weight: 2
---

This page describes the files Forge.go writes into a project, which of them you can edit, and how each tool is configured.

Forge.go writes three kinds of file:

| Kind | Examples | Re-running `init` |
| :--- | :--- | :--- |
| **Config** | `lefthook.yml`, `.goreleaser.yaml`, `Dockerfile`, `hugo.toml`, `.github/workflows/index.yml`, `.editorconfig`, `.github/dependabot.yml`, `.github/CODEOWNERS` | Left alone; overwritten only with `--force` |
| **Your source** | `go.mod`, `.gitignore`, `.golangci.local.yml`, the starter source and test, `content/` | Created once, never overwritten — even with `--force` |
| **Managed** | Everything under `.forgego/`, and `.golangci.yml` | Always refreshed; `--force` makes no difference |

## Managed files

Managed files belong to Forge.go. Don't edit them: `init` and `sync` overwrite them.

| File | Purpose |
| :--- | :--- |
| `.forgego/<tool>/go.mod` / `go.sum` | The pinned module for each tool — see [Pinned Tools](#pinned-tools). |
| `.forgego/golangci.yml` | Forge.go's base golangci-lint config. |
| `.golangci.yml` | The config golangci-lint reads, generated from the base and `.golangci.local.yml`. |

After upgrading Forge.go, or after editing `.golangci.local.yml`, refresh them:

```bash
task sync
```

`sync` refreshes:

- each `.forgego/<tool>/go.mod` and `go.sum` the project already has — it never adds a tool the project doesn't use;
- `.forgego/golangci.yml` and `.golangci.yml`, when the project has `.forgego/golangci.yml` (linting on, not a website), using `.golangci.local.yml` if it exists;
- the `toolchain` line in `go.mod`, when the pinned tools need a newer Go than the project declares (see below);
- the `FORGEGO` var in `Taskfile.yml`, when it runs a published Forge.go with `go run`: it moves to `go tool -modfile=.forgego/forgego/go.mod forgego`, and `.forgego/forgego/go.mod` pins the version doing the sync. `sync` writes that pin only when it's missing, as Dependabot proposes the next version from then on, and writes its `go.sum` with `go mod tidy`;
- tools pinned by an earlier Forge.go as `.forgego/<tool>.mod` and `.sum`, which it moves to `.forgego/<tool>/` and repoints `Taskfile.yml` and `lefthook.yml` at. Run `task hooks` afterwards, as the installed Git hooks still name the old path.

`task sync-check` (`forgego sync --check`) reports what has drifted without writing anything, and exits `1` if anything is out of date. It runs in the lefthook pre-commit hook and in the CI linting job, so a stale file fails the checks.

## Pinned tools

Each tool lives in its own module, a `go.mod` and `go.sum` under `.forgego/<tool>/`, and runs with:

```bash
go tool -modfile=.forgego/<tool>/go.mod <tool>
```

Keeping one module per tool means no two tools' dependencies are ever resolved together, so a tool can't force another onto an incompatible version — and none of them touch your project's `go.mod` or `go.sum`. `go` downloads and caches a tool the first time it runs.

| Tool | Module | Pinned version | Needs Go | Used for |
| :--- | :--- | :--- | :--- | :--- |
| `task` | `github.com/go-task/task/v3` | `v3.54.0` | 1.26.4 | Running every task (all modes) |
| `lefthook` | `github.com/evilmartians/lefthook/v2` | `v2.2.0` | 1.27.0 | Git hooks (linting on) |
| `golangci-lint` | `github.com/golangci/golangci-lint/v2` | `v2.14.0` | 1.26.0 | Linting and formatting (linting on, not websites) |
| `govulncheck` | `golang.org/x/vuln` | `v1.8.0` | 1.26.0 | Vulnerability checks (not websites) |
| `gotestsum` | `gotest.tools/gotestsum` | `v1.13.0` | 1.26.0 | Tests (testing on, not websites) |
| `hugo` | `github.com/gohugoio/hugo` | `v0.167.0` | 1.27.0 | Building the site (websites) |

These are the versions this release of Forge.go pins; `sync` moves a project to the versions of the Forge.go that runs it.

### The `toolchain` line

Go chooses its toolchain from the project's `go.mod` before it reads a tool's `-modfile`, so a project that declares an older Go than a tool needs couldn't run it. When that happens, `init` and `sync` add or raise the `toolchain` line in `go.mod` to the newest Go the project's tools need:

```text
module github.com/acme/billing-api

go 1.25.0

toolchain go1.27.0
```

The `go` line, which modules that depend on yours see, is left alone.

## Quality and testing

### golangci-lint: linting and formatting

golangci-lint can't extend another config file, so Forge.go merges two files into the `.golangci.yml` that golangci-lint reads:

| File | Owner |
| :--- | :--- |
| `.forgego/golangci.yml` | Forge.go — the base config |
| `.golangci.local.yml` | You — your project's changes |
| `.golangci.yml` | Generated from the two above. **Don't edit.** |

The merge rules are:

- **Mappings** merge key by key.
- **Lists** gain the items from `.golangci.local.yml` that aren't already in the base.
- **Any other value** in `.golangci.local.yml` replaces the base one.

Run `task sync` after editing `.golangci.local.yml` to regenerate `.golangci.yml`; `sync-check` fails until you do.

The base config uses golangci-lint's config `version: "2"`:

- **Linters:** the `standard` set (`errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`) plus `bodyclose`, `errorlint`, `gocritic`, `gosec`, `misspell`, `nilerr`, `revive`, `unconvert`, `unparam`, `usestdlibvars` and `wastedassign`. revive's `exported` rule is off.
- **Exclusions:** generated code is excluded leniently (`generated: lax`), with the `comments`, `common-false-positives` and `std-error-handling` presets; `gosec` and `unparam` don't run on `_test.go` files.
- **Formatters:** `gofumpt` and `goimports`.
- **Limits:** no cap on the number of issues reported per linter or per identical issue, and a five-minute timeout.

### govulncheck: vulnerabilities

`task security` runs `govulncheck ./...`, which reports only the known vulnerabilities your code actually calls. In CI it warns rather than fails, and the `patch` job upgrades the vulnerable modules on `main` — see [Job reference]({{< relref "workflows/reference.md#testingyml" >}}).

### gotestsum: testing

`task test` runs `go test ./...` through gotestsum with atomic coverage, writing `coverage.out` and a JUnit report, `junit-report.xml`. CI runs it with the race detector on. Both reports are in the generated `.gitignore`, along with `dist/`.

### lefthook: Git hooks

`lefthook.yml` runs everything through the pinned tools, so the hooks need nothing on `PATH` and use the same versions as CI. Its first line tells the installed hooks how to run lefthook itself:

```yaml
lefthook: go tool -modfile=.forgego/lefthook/go.mod lefthook
```

| Hook | Runs | Stage |
| :--- | :--- | :--- |
| `format` | `golangci-lint fmt` on the staged `.go` files, re-staging the result | pre-commit |
| `sync-check` | `task sync-check` | pre-commit |
| `lint` | `golangci-lint run`, when the push includes `.go` files | pre-push |
| `conventional-commit` | `task commit-msg` | commit-msg — only with versioning on |

A website has no Go to format or lint: its pre-commit hook runs `sync-check`, and its pre-push hook runs `task build`, which fails on broken templates and links.

Install them with `task hooks`.

## Build and release

### Task: building

A backend's `build` task compiles `./cmd/<name>` into `dist/<name>` (`dist/<name>.exe` on Windows) with `-trimpath`, setting `main.version` from the `VERSION` Taskfile var. The starter `main.go` declares that variable and prints it:

```go
// version is set at build time with -ldflags "-X main.version=...".
var version = "dev"
```

A library's `build` task runs `go build ./...` as a compile check; a library ships as source, so there's nothing else to build.

### GoReleaser: release binaries

A backend gets a `.goreleaser.yaml` that builds `./cmd/<name>` with `CGO_ENABLED=0`, `-trimpath` and `-s -w -X main.version={{ .Version }}` for Linux, macOS and Windows on `amd64` and `arm64`. It packages them as `.tar.gz` (`.zip` on Windows) with a `checksums.txt`.

release-please creates each release and writes its notes, so GoReleaser's changelog is disabled and it uses `release.mode: keep-existing` with `use_existing_draft: true`. release-please creates a backend's release as a draft; GoReleaser attaches the binaries and checksums to it, then publishes it. Files can't be added to a published release once a repository turns on immutable releases, so the release is only published once it's complete. Try it locally with `task release:snapshot`, which runs GoReleaser `v2.18.2` with `--snapshot --clean` and publishes nothing. CI runs the latest GoReleaser `v2`.

### release-please: versioning

`.github/release.json` configures release-please for a Go module. A Go module's versions are its Git tags, so there is no version file to bump:

```json
{
  "packages": {
    ".": {
      "release-type": "go"
    }
  }
}
```

A backend's release CI attaches binaries, so its config makes the release a draft for GoReleaser to publish, and creates the tag straight away, as GitHub doesn't tag a draft until it's published:

```json
{
  "packages": {
    ".": {
      "release-type": "go",
      "draft": true,
      "force-tag-creation": true
    }
  }
}
```

`.github/.release.json` is the release-please manifest, starting at `0.0.0`. The first release is `v0.1.0` if it includes a `feat` commit, or `v0.0.1` if it only has fixes.

### Docker

`--docker` adds a multi-stage `Dockerfile` and a `.dockerignore` that keeps `.git`, `.github`, build output and test reports out of the build context.

- **Backend:** builds on the `golang` image for your `--go` minor version (such as `golang:1.27`) on the build platform and cross-compiles for each target platform with `GOOS`/`GOARCH` and `CGO_ENABLED=0`, so a multi-platform image needs no emulation. The binary runs on `gcr.io/distroless/static-debian12:nonroot` as the non-root `nonroot` user. Pass `--build-arg VERSION=<version>` to set `main.version`; CI passes the release version.
- **Website:** builds the site once with the pinned Hugo on the build host, then serves `public/` with `nginx:stable-alpine` on port 80.

CI builds the image for every configured platform — see [Job reference]({{< relref "workflows/reference.md#dockeryml" >}}).

### Debian packaging

`--debian` adds an `nfpms` section to `.goreleaser.yaml`, so each release also attaches a `.deb`, and three files under `packaging/`:

| File | Purpose |
| :--- | :--- |
| `packaging/<name>.service` | systemd unit, installed to `/lib/systemd/system/`. Runs `/usr/bin/<name>` as its own `<name>` user and group, never root, restarting on failure. |
| `packaging/postinstall.sh` | Creates the `<name>` system user (no home directory, no login shell) if it doesn't exist, then enables and restarts the service. |
| `packaging/preremove.sh` | Disables and stops the service before its files are removed. |

`task release:snapshot` builds the `.deb` into `dist/` alongside the binaries. There's no separate packaging workflow: CI's release job runs GoReleaser, which builds the `.deb` with everything else.

Update `maintainer`, `description` and `license` in the `nfpms` section; Forge.go fills them with placeholders (`license: MIT`).

## Websites

### Hugo: static site

`--website` scaffolds a [Hugo](https://gohugo.io/) site using the [Hextra](https://github.com/imfing/hextra) theme as a Hugo module, so the theme is versioned in `go.mod` like any other dependency. Hugo itself is pinned in `.forgego/hugo/go.mod`.

| File | Purpose |
| :--- | :--- |
| `hugo.toml` | Site config: the Hextra import, a Docs menu entry and search. |
| `content/_index.md` | The landing page (`layout: hextra-home`). |
| `content/docs/_index.md` | The first documentation page. |

`task dev` serves the site with live reload; `task build` builds it into `public/`. Hugo manages the site's `go.mod` itself — don't run `go mod tidy` on it, as nothing imports the theme from Go code and tidying would drop it.

## Repository files

`init` also writes three files for the repository itself. Like the other configs, an existing one is kept unless you pass `--force`.

| File | What it does |
| :--- | :--- |
| `.editorconfig` | LF line endings, UTF-8 and 120 columns; tabs in Go files, `go.mod` and `go.sum`, two spaces elsewhere |
| `.github/dependabot.yml` | Weekly Go module and GitHub Actions updates, plus Docker with `--docker`. Minor and patch updates are grouped into one pull request, and each update waits 3 days after it's published before it's proposed, so a compromised release has time to be caught upstream. Forge.go itself is proposed the same way, through its pin in `.forgego/forgego/go.mod`. The workflow's auto-merge job merges them once CI passes. The other tools in `.forgego/` aren't included: `forgego sync` keeps them in step with Forge.go. |
| `.github/CODEOWNERS` | `* @owner`, so every pull request someone else opens, Dependabot's and release-please's included, requests your review and shows in your review requests. It doesn't block merging. |

The `CODEOWNERS` owner is the account in a `github.com/<owner>/...` [module path]({{< relref "getting-started.md#module-path" >}}). A project whose module path isn't on GitHub gets no `CODEOWNERS`; run `init` again once it has a GitHub remote.
