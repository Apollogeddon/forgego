---
title: ForgeGo
layout: hextra-home
---

{{< hextra/hero-badge >}}
  Go 1.27 · go tool · forgego
{{< /hextra/hero-badge >}}

{{< hextra/hero-headline >}}
  Quality tooling, zero configuration drift.
{{< /hextra/hero-headline >}}

{{< hextra/hero-subtitle >}}
  Reusable GitHub Actions workflows and tooling configurations for Go projects. Centralise your golangci-lint, test and release setup and scaffold projects with a single command.
{{< /hextra/hero-subtitle >}}

{{< hextra/hero-button text="Get Started" link="docs/getting-started" >}}

## Why ForgeGo

{{< hextra/feature-grid cols="2" >}}
  {{< hextra/feature-card
    title="Standardised Tooling"
    icon="adjustments"
    subtitle="A shared golangci-lint config, govulncheck, gotestsum and lefthook git hooks give every project the same checks from day one."
  >}}
  {{< hextra/feature-card
    title="Pinned, Go-Native Tools"
    icon="cube"
    subtitle="Each tool is pinned in its own module file under .forgego/ and run with go tool. Nothing to install globally, and nothing added to your go.mod."
  >}}
  {{< hextra/feature-card
    title="Reusable Workflows"
    icon="refresh"
    subtitle="Modular GitHub Actions workflows for testing, releasing, publishing and deploying, with no workflow YAML to write by hand."
    link="docs/workflows"
  >}}
  {{< hextra/feature-card
    title="Automated Releases"
    icon="tag"
    subtitle="release-please versions and tags releases from Conventional Commits; GoReleaser attaches the binaries and any .deb package."
  >}}
{{< /hextra/feature-grid >}}

## Quick Start

### 1. Initialise

Run `init` in an existing Go project, or in an empty directory to start a new one:

```bash
go run github.com/apollogeddon/forgego/cmd/forgego@latest init
```

### 2. Install

Tidy the module and install the git hooks:

```bash
go mod tidy
go tool -modfile=.forgego/task.mod task hooks
```

### 3. Extend

Put your project's lint changes in `.golangci.local.yml`, then regenerate `.golangci.yml`:

```bash
go tool -modfile=.forgego/task.mod task sync
```

## The Toolchain

| Tool | Replaces |
| :--- | :--- |
| **Task** | Makefiles, ad-hoc scripts |
| **golangci-lint** | Running `gofmt`, `go vet` and staticcheck separately |
| **govulncheck** | Manual dependency audits |
| **gotestsum** | Plain `go test` plus a separate JUnit converter |
| **lefthook** | Hand-written git hooks |
| **release-please** | Manual tagging and changelogs |
| **GoReleaser** | Hand-rolled release build scripts |
| **Hugo** | Hand-maintained documentation sites |
