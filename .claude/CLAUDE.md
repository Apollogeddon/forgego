# Claude Code Instructions

General working rules (subagent routing, tool use, response style) live in the user-level `~/.claude/CLAUDE.md`. This file is project-specific.

## Reading the codebase

`internal/` and `cmd/` are ~2.5k lines of Go and the tests ~900: read what you need directly rather than delegating. Use an `Explore` agent only for `docs/` (a separate Hugo module) and the module cache. `go.sum`, the `.forgego/*.sum` and `tools/*/go.sum` files, `dist/`, `coverage.out` and `junit-report.xml` are denied to `Read` in `settings.json` — use `go list -m <module>` to check a resolved version.

A `PostToolUse` hook (`.claude/hooks/hooks.go reads`) reminds you past 10 Read/Grep/Glob calls in a session.

## Quality Control

Every task runs as `go tool -modfile=.forgego/task.mod task <name>`.

- A `PostToolUse` hook (`.claude/hooks/hooks.go lint`) runs `golangci-lint fmt`, then `golangci-lint run --fix` on the package of every `.go` file touched by `Edit`/`Write`. Findings it can't fix are returned to you — fix them before moving on.
- While iterating, run `task test:unit` (`go test -short ./...`: skips the integration test that scaffolds real projects and downloads every tool).
- Before reporting a non-trivial change complete, run `task lint`, `task test` and `task sync-check`. Always run the full `task test` when a change touches what a generated project installs or runs — tool pins, templates, the Taskfile, hooks — as only the integration test catches those.
- After changing a tool version in `tools/<tool>/go.mod`, run `task generate`: forgego embeds a copy of each pinned module, and a test fails while the copies are stale. Dependabot's tool bumps need this too.
- Run `/code-review` (medium+) on non-trivial diffs before considering them done; use `/simplify` as a cleanup pass afterward.

## Architecture

forgego is a project-scaffolding CLI for Go. `cmd/forgego` calls `internal/cli`, which parses argv (stdlib `flag`, explicit `--no-<x>` flags per standard feature, a global `-C DIR`) and dispatches `init`, `sync` and `commit-msg`. `init` resolves a `config.Init` and calls `core.Init`.

`core.Init` detects the project (`internal/project`: the module path from `go.mod`, else the enclosing repository's git remote plus the subdirectory, else the directory name), then runs the ordered **Feature pipeline** in `internal/features` (`Pipeline`): `Base → Linting → Build → Testing → Versioning → Docker → Debian → Workflow`. Each feature implements `Feature` (`ShouldRun`, `Apply`, `Cleanup`) and is mode-aware (`cfg.IsBackend()`/`IsLibrary()`/`IsWebsite()`). Shared helpers in `features/feature.go`: `CreateFile` (configs: only overwritten with `--force`), `CreateIfMissing` (the project's own source and go.mod: never overwritten), `WriteManaged` (`.forgego/` files: always refreshed), `RemoveFile` (only with `--force`). After the pipeline, core raises go.mod's `toolchain` line if the pinned tools need a newer Go (`internal/gomod`) and merges the features' tasks into `Taskfile.yml` (`internal/taskfile`, keeping the project's own tasks unless `--force`).

**Tools.** Each tool a project uses is pinned in its own module file, `.forgego/<tool>.mod` and `.sum`, and run with `go tool -modfile=.forgego/<tool>.mod <tool>`, so no two tools' dependencies are resolved together and none touch the project's `go.mod` (golangci-lint and Hugo can't build from one module). The pins come from `tools/<tool>/go.mod`, separate modules Dependabot updates; `go generate` (`internal/templates/gen`) copies them into `internal/templates/tools/`, which is embedded. `internal/templates/tools.go` lists the tools. lefthook's git hooks name the `go tool` command in `lefthook.yml`, as there's no lefthook on PATH.

**golangci-lint** can't extend a config, so `internal/golangci` merges forgego's base (`internal/templates/configs/golangci.yml`, written to `.forgego/golangci.yml`) with the project's `.golangci.local.yml` into `.golangci.yml`. `forgego sync` (`internal/sync`) refreshes all the managed files, go.mod's toolchain line, and the `FORGEGO` Taskfile var while it points at a published forgego; `--check` reports drift for CI and the pre-commit hook.

Generated file *content* lives in `internal/templates/*.go` as plain strings with `__TOKEN__` placeholders (`templates.Render`): GitHub Actions' `${{ }}` and Task's `{{ }}` both clash with `text/template`. The generated `.github/workflows/index.yml` (`templates.RenderWorkflow`) calls the reusable workflows in `.github/workflows/` at `@main`; this repository's own CI (`.index.yml`) calls them locally.

`internal/fsys`'s `FS` interface (`OS` in production, `Memory` for tests) is what makes `--dry-run` a single code path, and lets `internal/core` test every mode in memory. `internal/integration` is different: it builds forgego, scaffolds real projects, and runs the generated tasks and the installed git hook — the only place that catches a tool pin, task or hook that doesn't work. The website test needs the network (Hextra fetches its search script from a CDN); set `FORGEGO_OFFLINE=1` to skip it.

This repository scaffolds itself: its `.forgego/`, `Taskfile.yml`, `lefthook.yml`, `.golangci*.yml` and `.goreleaser.yaml` came from `forgego init`, with `FORGEGO` pointed at `go run ./cmd/forgego`. The docs site in `docs/` is a `--website` scaffold, a separate module.

## Change checklist

Adding or changing a CLI flag or feature usually touches all of these — check each one before calling it done:

1. `internal/cli/cli.go` — the flag (a `feature` with `--no-<x>` for standard features) and the usage text.
2. `internal/config/config.go` — the `Init` field and `Validate`.
3. `internal/features/<feature>.go` — `ShouldRun`/`Apply`/`Cleanup`; add new features to `Pipeline` in `features/feature.go`, in order.
4. `internal/templates/<name>.go` — generated content; `tools/<tool>/` + `task generate` for a new or changed tool, and add it to `templates.Tools`.
5. Tests: `internal/core/core_test.go` (in memory); `internal/cli/cli_test.go` for flags; `internal/integration` if it changes what gets installed or run; `internal/templates/workflows_test.go` for the generated CI.
6. Docs: `docs/content/docs/getting-started.md` (flag and task reference), `configuration.md`, and `workflows/*.md` for workflow changes.
