---
title: Overview
weight: 1
---

Forge.go scaffolds tooling into Go projects: linting, testing, CI/CD, Docker images and Debian packaging. One command, `forgego init`, writes the configuration, the tasks and a GitHub Actions workflow that calls Forge.go's reusable workflows. `forgego sync` keeps the files Forge.go manages up to date afterwards.

## What You Get

| Area | Tools |
| :--- | :--- |
| Tasks | [Task](https://taskfile.dev/) — a `Taskfile.yml` with `lint`, `test`, `build` and friends |
| Linting | [golangci-lint](https://golangci-lint.run/) with a shared base config |
| Security | [govulncheck](https://go.dev/doc/security/vuln/), which reports only the vulnerabilities the code calls |
| Testing | `go test` through [gotestsum](https://github.com/gotestyourself/gotestsum), with coverage and a JUnit report |
| Git hooks | [lefthook](https://lefthook.dev/) — format on commit, lint on push, check commit messages |
| Releases | [release-please](https://github.com/googleapis/release-please) for versions and changelogs, [GoReleaser](https://goreleaser.com/) for binaries and `.deb` packages |
| Websites | [Hugo](https://gohugo.io/) with the [Hextra](https://github.com/imfing/hextra) theme |
| CI/CD | Reusable GitHub Actions workflows for quality, testing, releases, publishing and Docker images |

## Project Modes

| Mode | Flag | What it is |
| :--- | :--- | :--- |
| Backend | `--backend` (default) | A service or command-line tool, built into a binary from `cmd/<name>` and released with GoReleaser |
| Library | `--library` | A Go module others import, published by its release tag |
| Website | `--website` | A static documentation site built with Hugo and deployed to GitHub Pages |

## No Global Installs

Every tool Forge.go sets up is pinned in its own module file under `.forgego/` and run with `go tool -modfile=.forgego/<tool>/go.mod <tool>`. A fresh clone needs only Go and Git: the first run of a task downloads the pinned tool, and no tool's dependencies end up in your project's `go.mod`. See [Configuration]({{< relref "configuration.md#pinned-tools" >}}) for how it works.

## Where Next

- [Getting Started]({{< relref "getting-started.md" >}}) — install Forge.go and scaffold a project.
- [Configuration]({{< relref "configuration.md" >}}) — the managed files and how to change them.
- [Examples]({{< relref "examples.md" >}}) — common configuration and workflow recipes.
- [Workflows]({{< relref "workflows" >}}) — the reusable GitHub Actions workflows.
- [Migrating an Existing Project]({{< relref "migration.md" >}}) — adopting Forge.go in a project that already has tooling.
