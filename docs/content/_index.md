---
title: Forge.go
layout: hextra-home
hero:
  badge: Go · go tool · forgego
  title: Quality tooling,<br>zero configuration drift.
  subtitle: Reusable GitHub Actions workflows and tooling configurations for Go projects. Keep your golangci-lint, govulncheck and gotestsum setup in one place, and scaffold projects with one command.
  start: docs/getting-started
  github: https://github.com/apollogeddon/forgego
  terminal:
    - $ go install github.com/apollogeddon/forgego/v2/cmd/forgego@latest
    - $ forgego init
    - ✓ Configuration files created
    - ✓ Taskfile tasks added
    - "  Forge.go ready."
---

## Why Forge.go

{{< fg-section >}}The tooling and CI/CD a Go project needs, in one tool.{{< /fg-section >}}

{{< fg-cards cols="2" >}}
{{< fg-card title="Standardised tooling" icon="lucide-settings" >}}
A shared golangci-lint config gives every project the same checks from day one, and `forgego sync` keeps it current.
{{< /fg-card >}}
{{< fg-card title="Reusable workflows" icon="lucide-workflow" >}}
GitHub Actions workflows for testing, building, releasing and deploying. `init` writes the workflow that calls them.
{{< /fg-card >}}
{{< fg-card title="Project scaffolding" icon="lucide-rocket" >}}
`init` sets up a backend, library or website, with optional Docker and Debian packaging and no boilerplate to copy.
{{< /fg-card >}}
{{< fg-card title="Automated releases" icon="lucide-tag" >}}
release-please derives versions and changelogs from Conventional Commits, and GoReleaser publishes each release.
{{< /fg-card >}}
{{< /fg-cards >}}

## Quick start

{{< fg-section >}}From empty repo to a standardised toolchain in three steps.{{< /fg-section >}}

{{< fg-steps >}}
{{< fg-step n="01" title="Install" >}}
Install Forge.go from GitHub with `go install`.

```bash
go install github.com/apollogeddon/forgego/v2/cmd/forgego@latest
```
{{< /fg-step >}}
{{< fg-step n="02" title="Initialise" >}}
Run the CLI to scaffold configs, tasks and the CI workflow.

```bash
forgego init
```
{{< /fg-step >}}
{{< fg-step n="03" title="Extend" >}}
The generated config merges with the shared one. Add project-specific settings alongside.

```yaml
# .golangci.local.yml
version: "2"
linters:
  enable:
    - goconst
```
{{< /fg-step >}}
{{< /fg-steps >}}

{{< fg-more href="docs/getting-started" label="Full documentation" >}}
