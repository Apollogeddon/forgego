---
title: Forge.go
layout: hextra-home
hero:
  badge: Go 1.27 · go tool · forgego
  title: Quality tooling,<br>zero configuration drift.
  subtitle: Reusable GitHub Actions workflows and tooling configurations for Go projects. Keep your golangci-lint, test and release setup in one place, and scaffold projects with one command.
  start: docs/getting-started
  github: https://github.com/apollogeddon/forgego
  terminal:
    - $ go run github.com/apollogeddon/forgego/cmd/forgego@latest init
    - ✓ Configuration files created
    - ✓ forgego init complete
    - "  Next steps: go mod tidy && task hooks"
---

## Why Forge.go

{{< fg-section >}}The tooling and CI/CD a Go project needs, in one command.{{< /fg-section >}}

{{< fg-cards cols="2" >}}
{{< fg-card title="Standardised tooling" icon="adjustments" >}}
A shared golangci-lint config, govulncheck, gotestsum and lefthook Git hooks give every project the same checks from day one.
{{< /fg-card >}}
{{< fg-card title="Pinned, Go-native tools" icon="cube" >}}
Each tool is pinned in its own module under `.forgego/` and run with `go tool`. Nothing to install globally, and nothing added to your `go.mod`.
{{< /fg-card >}}
{{< fg-card title="Reusable workflows" icon="refresh" >}}
Reusable GitHub Actions workflows for testing, releasing, publishing and deploying. `init` writes the workflow that calls them.
{{< /fg-card >}}
{{< fg-card title="Automated releases" icon="tag" >}}
release-please versions and tags releases from Conventional Commits; GoReleaser attaches the binaries and any `.deb` package.
{{< /fg-card >}}
{{< /fg-cards >}}

## Quick start

{{< fg-section >}}From empty repo to a standardised toolchain in three steps.{{< /fg-section >}}

{{< fg-steps >}}
{{< fg-step n="01" title="Initialise" >}}
Run `init` in an existing Go project, or in an empty directory to start a new one.

```bash
go run github.com/apollogeddon/forgego/cmd/forgego@latest init
```
{{< /fg-step >}}
{{< fg-step n="02" title="Install" >}}
Tidy the module and install the Git hooks.

```bash
go mod tidy
go tool -modfile=.forgego/task/go.mod task hooks
```
{{< /fg-step >}}
{{< fg-step n="03" title="Extend" >}}
Put your lint changes in `.golangci.local.yml`, then regenerate `.golangci.yml`.

```bash
go tool -modfile=.forgego/task/go.mod task sync
```
{{< /fg-step >}}
{{< /fg-steps >}}

{{< fg-more href="docs/getting-started" label="Full documentation" >}}

## The toolchain

{{< fg-section >}}Go-native tools, each pinned in its own module and run with `go tool`.{{< /fg-section >}}

{{< fg-cards cols="4" >}}
{{< fg-tool name="Task" lang="Go" >}}Replaces Makefiles and ad-hoc scripts{{< /fg-tool >}}
{{< fg-tool name="golangci-lint" lang="Go" >}}Replaces running gofmt, go vet and staticcheck separately{{< /fg-tool >}}
{{< fg-tool name="govulncheck" lang="Go" >}}Replaces manual dependency audits{{< /fg-tool >}}
{{< fg-tool name="gotestsum" lang="Go" >}}Replaces plain go test plus a JUnit converter{{< /fg-tool >}}
{{< fg-tool name="lefthook" lang="Go" >}}Replaces hand-written Git hooks{{< /fg-tool >}}
{{< fg-tool name="release-please" lang="JS" >}}Replaces manual tagging and changelogs{{< /fg-tool >}}
{{< fg-tool name="GoReleaser" lang="Go" >}}Replaces hand-rolled release build scripts{{< /fg-tool >}}
{{< fg-tool name="Hugo" lang="Go" >}}Replaces hand-maintained documentation sites{{< /fg-tool >}}
{{< /fg-cards >}}
