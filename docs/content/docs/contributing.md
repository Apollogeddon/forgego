---
title: Contributing
weight: 7
---

This page is for people working on Forge.go itself. The repository uses the same toolchain Forge.go scaffolds into other projects: its `.forgego/`, `Taskfile.yml`, `lefthook.yml`, `.golangci*.yml` and `.goreleaser.yaml` came from `forgego init`.

## Development setup

```bash
git clone https://github.com/apollogeddon/forgego
cd forgego
go tool -modfile=.forgego/task/go.mod task hooks
```

As in any Forge.go project, every task runs as `go tool -modfile=.forgego/task/go.mod task <name>`; alias it or install Task to type `task <name>`.

In this repository the `FORGEGO` Taskfile var is `go run ./cmd/forgego`, so `task sync`, `task sync-check` and the commit message hook run the Forge.go in your working tree rather than a published release.

## Quality checks

| Task | What it runs |
| :--- | :--- |
| `task lint` | `golangci-lint fmt`, then `golangci-lint run --fix` |
| `task test` | The full suite through gotestsum, including the slow integration test |
| `task test:unit` | `go test -short ./...` — skips the integration test |
| `task security` | govulncheck over the module |
| `task build` | Builds `dist/forgego` |
| `task sync-check` | Checks this repository's own managed files are current |
| `task generate` | Copies the pinned tool modules from `tools/` into the embedded templates |

Run `lint`, `test` and `sync-check` before opening a pull request.

> [!NOTE]
> The integration test builds Forge.go, scaffolds a real project in each mode, and runs its generated tasks and Git hooks, downloading every pinned tool. It is the only test that catches a tool pin, task or hook that doesn't work, so run the full `task test` for changes to templates, tool pins, the Taskfile or hooks. The website test needs network access for the theme's search script; set `FORGEGO_OFFLINE=1` to skip it.

## Tool versions

Each tool a generated project pins comes from its own module under `tools/<tool>/`, which Forge.go embeds. After changing a version there, run:

```bash
task generate
```

A test fails while the embedded copies are out of date. Dependabot's tool bumps (`fix(tools): …`) need the same: run `task generate` on the pull request and push the result before its checks pass.

## Documentation

This site lives in `docs/` as its own Go module, scaffolded with `forgego init --website`:

```bash
cd docs
go tool -modfile=.forgego/task/go.mod task dev    # live preview
```

`task build` builds it into `docs/public/`. Hextra fetches its search script from a CDN during the build, so building the site needs network access.

## Conventional Commits

Commit messages follow the [Conventional Commits](https://www.conventionalcommits.org/) specification, checked by `forgego commit-msg` in the commit-msg hook. release-please reads them to decide the next version and write the changelog.

### Commit types

1. **Features** (`feat`) — Triggers a **minor** release.
   Example: `feat: add a systemd unit for --debian packages`

2. **Fixes** (`fix`) — Triggers a **patch** release.
   Example: `fix: keep the theme in a website's go.mod`

3. **Maintenance** (`chore`) — Does **not** trigger a release.
   Example: `chore: tidy the Taskfile`

The hook also accepts `build`, `ci`, `docs`, `perf`, `refactor`, `revert`, `style` and `test`.

> [!IMPORTANT]
> Include `BREAKING CHANGE:` in the footer or a `!` after the type/scope (e.g., `feat!: rename the sync command`) to trigger a **major** release.
