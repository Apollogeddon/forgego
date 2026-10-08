---
title: Philosophy & Stack
weight: 6
---

ForgeGo implements an opinionated toolchain designed to prioritise reproducibility and configuration simplicity.

## Go-Native Tooling

Everything a ForgeGo project runs is a Go program, pinned by version and run with `go tool`. A fresh clone needs only Go and Git.

- **One module per tool:** each tool has its own `.forgego/<tool>/go.mod` and `go.sum`, run with `go tool -modfile=.forgego/<tool>/go.mod <tool>`. No two tools' dependencies are resolved together, so they can't conflict, and none of them reach your `go.mod`.
- **No global installs:** the Taskfile, the git hooks and CI all run the same pinned versions, so "works on my machine" and "passes in CI" mean the same thing.
- **Checksummed:** each tool's `.sum` file is verified by `go` like any other module, and cached after the first run.

## Standardisation as a Service

ForgeGo abstracts configuration to prevent "drift" across repositories. Improvements to the toolchain reach every project through `forgego sync`.

- **Managed versions:** a ForgeGo release pins a tested set of tool versions. Syncing with a newer ForgeGo moves a project to its versions together.
- **Managed base configs:** `forgego sync` refreshes the shared golangci-lint config in `.forgego/` and regenerates `.golangci.yml`, and `forgego sync --check` catches a stale file in the pre-commit hook and CI.
- **release-please:** automates the release lifecycle. Version numbers and changelogs are derived from commit history, removing manual intervention from releases.
- **Centralised CI/CD:** reusable GitHub Actions workflows give every project the same security audits, quality gates and delivery patterns.
- **Verified end to end:** every mode is scaffolded into a real project, and its generated tasks and git hooks are run, before a new version is released.

## The Toolchain

| Tool | Why? | Replaces |
| :--- | :--- | :--- |
| **Task** | Project tasks in one YAML file, cross-platform, written in Go. | Makefiles, ad-hoc scripts |
| **golangci-lint** | One run for dozens of linters and the formatters, with a shared config. | Running `gofmt`, `go vet` and staticcheck separately |
| **govulncheck** | Reports only the vulnerabilities your code actually calls. | Manual dependency audits |
| **gotestsum** | Readable `go test` output plus a JUnit report for CI. | Plain `go test` with a separate converter |
| **lefthook** | Fast git hooks configured in one file, run through `go tool`. | Hand-written hook scripts |
| **release-please** | Deterministic versioning and changelog generation. | Manual tagging, manual changelogs |
| **GoReleaser** | Cross-platform release binaries, checksums and `.deb` packages from one config. | Hand-rolled release scripts, separate packaging tools |
| **Hugo** | Fast static sites, with the theme versioned as a Go module. | Hand-maintained documentation sites |
| **distroless** | Minimal, non-root runtime images for static Go binaries. | Full OS base images |
