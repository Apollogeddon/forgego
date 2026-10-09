---
title: Migrating an existing project
linkTitle: Migration
weight: 5
---

This page is for adopting Forge.go in a Go project that already has its own tooling. Adopting it means retiring the tools and configuration it replaces.

Forge.go adds no starter code to an existing project and never changes your source, or your `go.mod`'s module and `go` lines: it only adds a `toolchain` line when the pinned tools need a newer Go than the project declares. It leaves existing config files alone unless you pass `--force`.

## Migration checklist

### 1. Move your golangci-lint config aside

`.golangci.yml` is the one existing file `init` always replaces: it is generated from Forge.go's base config and `.golangci.local.yml`. Keep your settings by renaming your config to `.golangci.local.yml` **before** running `init`:

```bash
git mv .golangci.yml .golangci.local.yml
```

`init` then merges it into the base config instead of creating an empty `.golangci.local.yml`. It must be a golangci-lint `version: "2"` config — run `golangci-lint migrate` on an older one first. Check the [merge rules]({{< relref "configuration.md#golangci-lint-linting-and-formatting" >}}): your lists add to the base lists, so a `linters.enable` list adds to the base linters, and a `linters.default` of your own replaces the base `standard`.

### 2. Run `init`

Preview the changes, then generate the standard configuration:

```bash
forgego init --dry-run
forgego init
```

Existing files are left alone. Pass `--force` to replace your current `lefthook.yml`, `.goreleaser.yaml`, `Dockerfile`, release-please config and `.github/workflows/index.yml` with the Forge.go versions; your source code is never overwritten.

### 3. Retire old tooling

Remove what Forge.go now does, to avoid running two versions of the same check:

- **Makefile:** move the targets Forge.go doesn't cover into `Taskfile.yml` as your own tasks (see [Adding your own tasks]({{< relref "examples.md#adding-your-own-tasks" >}})), then delete the `Makefile`. An existing `Taskfile.yml` keeps its tasks and vars; `init` only adds the ones it's missing.
- **Tools pinned in `go.mod`:** remove `tool` directives (or a `tools.go` file of blank imports) for golangci-lint, govulncheck, gotestsum, lefthook or Task, then run `go mod tidy`. Forge.go pins each in its own module under `.forgego/`, so they no longer need to share your dependency graph.
- **Hand-written Git hooks:** delete the scripts in `.git/hooks/`, and unset `core.hooksPath` if you pointed Git at a hooks directory of your own, then install lefthook's hooks with `task hooks`.
- **Old CI workflows:** delete the workflows the generated `.github/workflows/index.yml` replaces — linting, tests, release and image builds.

### 4. Update the module

```bash
go mod tidy
```

`init` adds a `toolchain` line to `go.mod` if your project declares an older Go than the pinned tools need. Your `go` line is left alone.

### 5. Fix linting errors

The base config is stricter than many existing setups. Let golangci-lint format and fix what it can, then work through the rest:

```bash
task lint
task security
```

## Tool-specific guides

### Hand-written golangci-lint config

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

### Existing GoReleaser config

`init` keeps an existing `.goreleaser.yaml`. Forge.go's release job runs GoReleaser after release-please has created the release as a draft, so make sure yours doesn't try to create one or write its own notes, and that it publishes the draft:

```yaml
changelog:
  disable: true
release:
  mode: keep-existing
  use_existing_draft: true
```

`init` keeps an existing `.github/release.json` too. A project scaffolded by an earlier Forge.go needs `"draft": true` and `"force-tag-creation": true` added to it (see [Configuration]({{< relref "configuration.md#release-please-versioning" >}})) if the repository has immutable releases turned on.

To pick up Forge.go's config instead, run `forgego init --force` (adding `--debian` if you package a `.deb`), then copy back anything specific to your project.

### Existing releases

release-please starts from the version in `.github/.release.json`, which Forge.go creates at `0.0.0`. If the project already has release tags, set it to the latest released version before the first release PR:

```json
{
  ".": "1.4.2"
}
```

Remove any other version-bumping or changelog tooling; releases are driven by Conventional Commits from here on.

### Existing Dockerfile

`--docker` keeps an existing `Dockerfile`. The `docker` CI job builds it for every platform in one job, so a Dockerfile that builds for the target platform under emulation still works but is slower. Forge.go's backend `Dockerfile` cross-compiles on the build platform instead; to compare, run `forgego init --docker` in a scratch directory and diff the result with yours.
