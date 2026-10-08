---
title: Forge.go
layout: hextra-home
hero:
  badge: Go 1.27 · go tool · forgego
  title: Quality tooling,<br>zero configuration drift.
  subtitle: Reusable GitHub Actions workflows and tooling configurations for Go projects. Centralise your golangci-lint, test and release setup and scaffold projects with a single command.
  start: docs/getting-started
  github: https://github.com/apollogeddon/forgego
---

## Why Forge.go

{{< fg-cards cols="2" >}}
{{< fg-card title="Standardised Tooling" icon="adjustments" >}}
A shared golangci-lint config, govulncheck, gotestsum and lefthook git hooks give every project the same checks from day one.
{{< /fg-card >}}
{{< fg-card title="Pinned, Go-Native Tools" icon="cube" >}}
Each tool is pinned in its own module under `.forgego/` and run with `go tool`. Nothing to install globally, and nothing added to your `go.mod`.
{{< /fg-card >}}
{{< fg-card title="Reusable Workflows" icon="refresh" >}}
Modular GitHub Actions workflows for testing, releasing, publishing and deploying, with no workflow YAML to write by hand.
{{< /fg-card >}}
{{< fg-card title="Automated Releases" icon="tag" >}}
release-please versions and tags releases from Conventional Commits; GoReleaser attaches the binaries and any `.deb` package.
{{< /fg-card >}}
{{< /fg-cards >}}

## Quick Start

{{< fg-cards cols="3" >}}
{{< fg-card title="01 · Initialise" >}}
Run `init` in an existing Go project, or in an empty directory to start a new one.

```bash
go run github.com/apollogeddon/forgego/cmd/forgego@latest init
```
{{< /fg-card >}}
{{< fg-card title="02 · Install" >}}
Tidy the module and install the git hooks.

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

## The Toolchain

{{< fg-cards cols="4" >}}
{{< fg-card title="Task" >}}Replaces Makefiles, ad-hoc scripts{{< /fg-card >}}
{{< fg-card title="golangci-lint" >}}Replaces running gofmt, go vet and staticcheck separately{{< /fg-card >}}
{{< fg-card title="govulncheck" >}}Replaces manual dependency audits{{< /fg-card >}}
{{< fg-card title="gotestsum" >}}Replaces plain go test plus a JUnit converter{{< /fg-card >}}
{{< fg-card title="lefthook" >}}Replaces hand-written git hooks{{< /fg-card >}}
{{< fg-card title="release-please" >}}Replaces manual tagging and changelogs{{< /fg-card >}}
{{< fg-card title="GoReleaser" >}}Replaces hand-rolled release build scripts{{< /fg-card >}}
{{< fg-card title="Hugo" >}}Replaces hand-maintained documentation sites{{< /fg-card >}}
{{< /fg-cards >}}
