---
title: Forge.go
layout: hextra-home
hero:
  badge: Go 1.27 · go tool · forgego
  title: Quality tooling,<br>zero configuration drift.
  subtitle: Reusable GitHub Actions workflows and tooling configurations for Go projects. Keep your golangci-lint, test and release setup in one place, and scaffold projects with one command.
  start: docs/getting-started
  github: https://github.com/apollogeddon/forgego
---

## Why Forge.go

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

{{< fg-cards cols="3" >}}
{{< fg-card title="01 · Initialise" >}}
Run `init` in an existing Go project, or in an empty directory to start a new one.

```bash
go run github.com/apollogeddon/forgego/cmd/forgego@latest init
```
{{< /fg-card >}}
{{< fg-card title="02 · Install" >}}
Tidy the module and install the Git hooks.

```bash
go mod tidy
go tool -modfile=.forgego/task/go.mod task hooks
```
{{< /fg-card >}}
{{< fg-card title="03 · Extend" >}}
Put your lint changes in `.golangci.local.yml`, then regenerate `.golangci.yml`.

```bash
go tool -modfile=.forgego/task/go.mod task sync
```
{{< /fg-card >}}
{{< /fg-cards >}}

## The toolchain

{{< fg-cards cols="4" >}}
{{< fg-card title="Task" >}}Replaces Makefiles and ad-hoc scripts{{< /fg-card >}}
{{< fg-card title="golangci-lint" >}}Replaces running gofmt, go vet and staticcheck separately{{< /fg-card >}}
{{< fg-card title="govulncheck" >}}Replaces manual dependency audits{{< /fg-card >}}
{{< fg-card title="gotestsum" >}}Replaces plain go test plus a JUnit converter{{< /fg-card >}}
{{< fg-card title="lefthook" >}}Replaces hand-written Git hooks{{< /fg-card >}}
{{< fg-card title="release-please" >}}Replaces manual tagging and changelogs{{< /fg-card >}}
{{< fg-card title="GoReleaser" >}}Replaces hand-rolled release build scripts{{< /fg-card >}}
{{< fg-card title="Hugo" >}}Replaces hand-maintained documentation sites{{< /fg-card >}}
{{< /fg-cards >}}
