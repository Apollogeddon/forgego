---
title: Philosophy and stack
weight: 3
---

This page explains the choices behind Forge.go's toolchain. Forge.go is opinionated: it favours reproducible builds and a small amount of configuration per project.

## Go-native tooling

Everything a Forge.go project runs is a Go program, pinned by version and run with `go tool`. A fresh clone needs only Go and Git.

- **One module per tool:** each tool has its own `.forgego/<tool>/go.mod` and `go.sum`, run with `go tool -modfile=.forgego/<tool>/go.mod <tool>`. No two tools' dependencies are resolved together, so they can't conflict, and none of them reach your `go.mod`.
- **No global installs:** the Taskfile, the Git hooks and CI all run the same pinned versions, so a check that passes locally runs with the same tools in CI.
- **Checksummed:** `go` verifies each tool against its `go.sum` like any other module, and caches it after the first run.

## Standardisation

Forge.go keeps shared configuration out of each repository, so projects don't drift apart. Changes to the toolchain reach every project through `forgego sync`.

- **Managed versions:** a Forge.go release pins a tested set of tool versions. Syncing with a newer Forge.go moves a project to its versions together.
- **Managed base configs:** `forgego sync` refreshes the shared golangci-lint config in `.forgego/` and regenerates `.golangci.yml`, and `forgego sync --check` catches a stale file in the pre-commit hook and CI.
- **release-please:** version numbers and changelogs come from the commit history, so releases need no manual tagging.
- **Centralised CI/CD:** reusable GitHub Actions workflows give every project the same security audits, quality gates and delivery patterns.
- **Verified end to end:** every mode is scaffolded into a real project, and its generated tasks and Git hooks are run, before a new version is released.
