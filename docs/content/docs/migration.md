---
title: Migrating an Existing Project
linkTitle: Migration
weight: 5
---

Adopting ForgeGo in an existing Go project reduces configuration overhead, but means retiring the tooling it replaces. ForgeGo adds no starter code to an existing project and never changes your source, or your `go.mod`'s module and `go` lines: it only adds a `toolchain` line when the pinned tools need a newer Go than the project declares. It leaves existing config files alone unless you pass `--force`.

## Migration Checklist

### 1. Move Your golangci-lint Config Aside

`.golangci.yml` is the one existing file `init` always replaces: it is generated from ForgeGo's base config and `.golangci.local.yml`. Keep your settings by renaming your config to `.golangci.local.yml` **before** running `init`:

```bash
git mv .golangci.yml .golangci.local.yml
```

`init` then merges it into the base config instead of creating an empty `.golangci.local.yml`. It must be a golangci-lint `version: "2"` config — run `golangci-lint migrate` on an older one first. Check the [merge rules]({{< relref "configuration.md#golangci-lint--linting--formatting" >}}): your lists add to the base lists, so a `linters.enable` list adds to the base linters, and a `linters.default` of your own replaces the base `standard`.

### 2. Run Initialisation

Preview the changes, then generate the standard configuration:

```bash
forgego init --dry-run
forgego init
```

Existing files are left alone. Pass `--force` to replace your current `lefthook.yml`, `.goreleaser.yaml`, `Dockerfile`, release-please config and `.github/workflows/index.yml` with the ForgeGo versions; your source code is never overwritten.

### 3. Retire Old Tooling

Remove what ForgeGo now does, to avoid running two versions of the same check:

- **Makefile:** move the targets ForgeGo doesn't cover into `Taskfile.yml` as your own tasks (see [Adding Your Own Tasks]({{< relref "examples.md#adding-your-own-tasks" >}})), then delete the `Makefile`. An existing `Taskfile.yml` keeps its tasks and vars; `init` only adds the ones it's missing.
- **Tools pinned in `go.mod`:** remove `tool` directives (or a `tools.go` file of blank imports) for golangci-lint, govulncheck, gotestsum, lefthook or Task, then run `go mod tidy`. ForgeGo pins each in its own module under `.forgego/`, so they no longer need to share your dependency graph.
- **Hand-written git hooks:** delete the scripts in `.git/hooks/`, and unset `core.hooksPath` if you pointed Git at a hooks directory of your own, then install lefthook's hooks with `task hooks`.
- **Old CI workflows:** delete the workflows the generated `.github/workflows/index.yml` replaces — linting, tests, release and image builds.

### 4. Update the Module

```bash
go mod tidy
```

`init` adds a `toolchain` line to `go.mod` if your project declares an older Go than the pinned tools need. Your `go` line is left alone.

### 5. Fix Linting Errors

The base config is stricter than many existing setups. Let golangci-lint format and fix what it can, then work through the rest:

```bash
task lint
task security
```

## Tool-Specific Guides

### Hand-Written golangci-lint Config

If you'd rather adopt the base config as it is, delete your `.golangci.yml` instead of renaming it, run `init`, and add back only the settings you still need to `.golangci.local.yml`.

> [!TIP]
> For a large codebase with many initial findings, add a temporary exclusion rule for the noisiest paths to `.golangci.local.yml` and remove it as you fix them:
>
> ```yaml
> linters:
>   exclusions:
>     rules:
>       - path: internal/legacy/
>         linters: [gosec, gocritic]
> ```

### Existing GoReleaser Config

`init` keeps an existing `.goreleaser.yaml`. ForgeGo's release job runs GoReleaser after release-please has created the release, so make sure yours doesn't try to create one or write its own notes:

```yaml
changelog:
  disable: true
release:
  mode: keep-existing
```

To pick up ForgeGo's config instead, run `forgego init --force` (adding `--debian` if you package a `.deb`), then copy back anything specific to your project.

### Existing Releases

release-please starts from the version in `.github/.release.json`, which ForgeGo creates at `0.0.0`. If the project already has release tags, set it to the latest released version before the first release PR:

```json
{
  ".": "1.4.2"
}
```

Remove any other version-bumping or changelog tooling; releases are driven by Conventional Commits from here on.

### Existing Dockerfile

`--docker` keeps an existing `Dockerfile`. The `docker` CI job builds it for every platform in one job, so a Dockerfile that builds for the target platform under emulation still works but is slower. ForgeGo's backend `Dockerfile` cross-compiles on the build platform instead; to compare, run `forgego init --docker` in a scratch directory and diff the result with yours.
