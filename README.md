<br />
<div align="center">
  <a href="https://apollogeddon.github.io/forgego/">
    <img src="docs/static/forgego.svg" alt="Logo" width="100" height="100">
  </a>

  <h3 align="center">Forge.go</h3>

  <p align="center">
    DevOps Support and Quality Control for modern Go projects
    <br />
    <a href="https://apollogeddon.github.io/forgego/"><strong>Explore the docs</strong></a>
    <br />
    <br />
    <a href="https://apollogeddon.github.io/forgego/docs/getting-started/">Getting Started</a>
    &middot;
    <a href="https://apollogeddon.github.io/forgego/docs/configuration/">Configuration</a>
    &middot;
    <a href="https://apollogeddon.github.io/forgego/docs/workflows/">Workflows</a>
  </p>
</div>

<br />

## Installation

Run forgego without installing it:

```bash
go run github.com/apollogeddon/forgego/cmd/forgego@latest init
```

Or install it:

```bash
go install github.com/apollogeddon/forgego/cmd/forgego@latest
```

Each release also attaches prebuilt binaries for Linux, macOS and Windows.

## Getting Started

To set up your project with the recommended configs, tasks, and CI workflows, use the `init` command.

```bash
forgego init [options]
```

By default, this sets up a **Go Backend/Service**. You can specify other modes:

* `--backend` (Default) for services and command-line tools.
* `--library` for Go modules others import.
* `--website` for static documentation sites (Hugo).

This command will:

* **Scaffold Configs:** Create `.golangci.yml`, `lefthook.yml`, `.goreleaser.yaml`, the release-please config, and others depending on the mode (e.g., `Dockerfile`, the Debian packaging files, or `hugo.toml`).
* **Inject Tasks:** Add `lint`, `test`, `build`, `security` and more to the project's `Taskfile.yml`, keeping any task it already has.
* **Standardise:** Create `go.mod` and a starter command or package so the generated tooling works immediately, and leave existing files alone unless you pass `--force`.

Every task runs with `go tool -modfile=.forgego/task/go.mod task <name>`; alias it, or install [Task](https://taskfile.dev/). Run `forgego init --help` for the full flag reference.

## Standardised Stack

Forge.go enforces a standardised stack designed for performance and reliability:

| Category | Tool | Description |
| :--- | :--- | :--- |
| **Linting & Formatting** | [golangci-lint](https://golangci-lint.run/) | Runs dozens of linters and the gofumpt and goimports formatters in one pass. |
| **Security Scanning** | [govulncheck](https://go.dev/doc/security/vuln/) + [Gitleaks](https://github.com/gitleaks/gitleaks) + [OSV-Scanner](https://osv.dev/) | Reports only the vulnerabilities your code calls, plus secret detection and dependency scanning. |
| **Testing** | [gotestsum](https://github.com/gotestyourself/gotestsum) | Runs `go test` with readable output, coverage and a JUnit report. |
| **Task Running** | [Task](https://taskfile.dev/) | Runs the project's tasks from `Taskfile.yml`. |
| **Git Hooks** | [Lefthook](https://github.com/evilmartians/lefthook) | Fast Git hooks manager, with a built-in Conventional Commits check. |
| **Releases** | [Release Please](https://github.com/googleapis/release-please) + [GoReleaser](https://goreleaser.com/) | Automated versioning and changelogs, with release binaries and `.deb` packages attached. |
| **CI/CD** | [GitHub Actions](https://github.com/features/actions) | Reusable workflows for Testing, Quality, and Releases. |
| **Containers** | [Docker Buildx](https://docs.docker.com/build/) | With `--docker`, CI cross-compiles the image for `linux/amd64` and `linux/arm64` (configurable via the `docker` job's `platforms` input) on every PR and pushes it to GHCR on release. |
| **Websites** | [Hugo](https://gohugo.io/) + [Hextra](https://imfing.github.io/hextra/) | With `--website`, a documentation site whose theme is versioned in `go.mod`. |

## Tooling & Versioning Strategy

Forge.go takes an opinionated, batteries-included approach to tooling.

* **Managed Versions:** Each tool is pinned in its own module file under `.forgego/` and run with `go tool -modfile`, so no two tools' dependencies clash and none touch your `go.mod`.
* **Simplified Upgrades:** To upgrade your linter or test runner, upgrade forgego and run `task sync` to refresh `.forgego/` (`task sync-check` reports drift without writing, and is already wired into the generated pre-commit hook and CI).
* **Security First:** Security scanning is integrated into the standard workflow, and CI upgrades the modules govulncheck finds vulnerable on `main`.
* **Stability:** Every mode is verified end-to-end (the generated tasks and git hooks, with every pinned tool) before a new version is released.
