<br />
<div align="center">
  <a href="https://apollogeddon.github.io/forgego/">
    <img src="docs/static/forgego.svg" alt="Logo" width="100" height="100">
  </a>

  <h3 align="center">Forge.go</h3>

  <p align="center">
    Reusable GitHub Actions workflows and tooling configurations for Go projects
    <br />
    <a href="https://apollogeddon.github.io/forgego/"><strong>Read the docs</strong></a>
    <br />
    <br />
    <a href="https://apollogeddon.github.io/forgego/docs/getting-started/">Getting started</a>
    &middot;
    <a href="https://apollogeddon.github.io/forgego/docs/configuration/">Configuration</a>
    &middot;
    <a href="https://apollogeddon.github.io/forgego/docs/workflows/">Workflows</a>
  </p>
</div>

<br />

Forge.go (`forgego`) is a project-scaffolding CLI for Go. Its `init` command sets up a backend, library or website with a standard toolchain (golangci-lint, govulncheck, gotestsum, lefthook, release-please, GoReleaser) and a GitHub Actions pipeline built from reusable workflows. It keeps your golangci-lint, govulncheck and gotestsum setup in one place: each tool is pinned in its own module, and `forgego sync` brings every project up to the configs and versions of a newer Forge.go.

## Requirements

- [Go](https://go.dev/dl/) 1.26 or later; the pinned tools need Go 1.27, which `go` downloads through the `toolchain` line in `go.mod`
- Git, for the generated lefthook hooks
- Docker, only for the `docker:build` and `docker:run` tasks that `--docker` adds

## Installation

Install it with `go install`:

```bash
go install github.com/apollogeddon/forgego/cmd/forgego@latest
```

Or run it without installing:

```bash
go run github.com/apollogeddon/forgego/cmd/forgego@latest init
```

Each [release](https://github.com/apollogeddon/forgego/releases) also attaches prebuilt binaries for Linux, macOS and Windows.

## Quick start

Run `init` in an existing Go project, or in an empty directory to start a new one:

```bash
forgego init             # service or command-line tool (default)
forgego init --library   # Go module others import
forgego init --website   # static documentation site (Hugo)
```

Then tidy the module and install the Git hooks:

```bash
go mod tidy
go tool -modfile=.forgego/task/go.mod task hooks
```

`init` does the following:

- **Writes the configuration:** `.golangci.yml`, `lefthook.yml`, the release-please config, `.github/workflows/index.yml` and, depending on the mode and flags, `.goreleaser.yaml`, a `Dockerfile`, Debian packaging files or `hugo.toml`.
- **Adds tasks** such as `lint`, `test`, `build` and `security` to `Taskfile.yml`, keeping any task the project already has.
- **Starts a new project** with a `go.mod` and a starter command or package, so the generated tooling works straight away. In an existing project it leaves your code alone.

`init` is safe to re-run: existing files are left alone unless you pass `--force`. Run `forgego init --help` for every flag, or see the [flag reference](https://apollogeddon.github.io/forgego/docs/getting-started/#init-options).

Every generated task runs through a pinned copy of [Task](https://taskfile.dev/): `go tool -modfile=.forgego/task/go.mod task <name>`. Alias that command, or install Task and run `task <name>`.

## The toolchain

| Category | Tool | What it does |
| :--- | :--- | :--- |
| Linting and formatting | [golangci-lint](https://golangci-lint.run/) | Runs a shared set of linters and the gofumpt and goimports formatters in one pass. |
| Security scanning | [govulncheck](https://go.dev/doc/security/vuln/), plus [Gitleaks](https://github.com/gitleaks/gitleaks) and [OSV-Scanner](https://osv.dev/) in CI | Reports the vulnerabilities your code calls, leaked secrets and vulnerable dependencies. |
| Testing | [gotestsum](https://github.com/gotestyourself/gotestsum) | Runs `go test` with readable output, coverage and a JUnit report. |
| Task running | [Task](https://taskfile.dev/) | Runs the project's tasks from `Taskfile.yml`. |
| Git hooks | [lefthook](https://github.com/evilmartians/lefthook) | Formats on commit, lints on push and checks commit messages against Conventional Commits. |
| Releases | [release-please](https://github.com/googleapis/release-please) and [GoReleaser](https://goreleaser.com/) | Versions and changelogs from Conventional Commits, with release binaries and `.deb` packages attached. |
| CI/CD | [GitHub Actions](https://github.com/features/actions) | Reusable workflows for quality checks, tests, releases and deployment. |
| Containers | [Docker Buildx](https://docs.docker.com/build/) | With `--docker`, CI builds the image for `linux/amd64` and `linux/arm64` on every pull request and pushes it to GHCR on release. |
| Websites | [Hugo](https://gohugo.io/) and [Hextra](https://imfing.github.io/hextra/) | With `--website`, a documentation site whose theme is versioned in `go.mod`. |

## Keeping projects up to date

- **One module per tool:** each tool is pinned in its own module file under `.forgego/` and run with `go tool -modfile`, so no two tools' dependencies clash and none of them touch your `go.mod`. A fresh clone needs only Go and Git.
- **Upgrades through `sync`:** to move a project to the tool versions of a newer Forge.go, run that version's `forgego sync`. `task sync-check` reports drift without writing anything, and runs in the generated pre-commit hook and in CI.
- **Security patches:** CI runs govulncheck on every change, and on `main` it upgrades the modules govulncheck finds vulnerable and commits the result.
- **Tested end to end:** before each release, an integration test scaffolds a real project in every mode and runs its generated tasks and Git hooks with every pinned tool.

## Documentation

The full documentation is at [apollogeddon.github.io/forgego](https://apollogeddon.github.io/forgego/):

- [Getting started](https://apollogeddon.github.io/forgego/docs/getting-started/): requirements, CLI flags and generated tasks.
- [Configuration](https://apollogeddon.github.io/forgego/docs/configuration/): the files Forge.go writes and how to change them.
- [Examples](https://apollogeddon.github.io/forgego/docs/examples/): common configuration and workflow recipes.
- [Workflows](https://apollogeddon.github.io/forgego/docs/workflows/): the reusable GitHub Actions workflows and their inputs.
- [Migration](https://apollogeddon.github.io/forgego/docs/migration/): adopting Forge.go in a project that already has tooling.

## Contributing

Pull requests are welcome. The [contributing guide](https://apollogeddon.github.io/forgego/docs/contributing/) covers setting up the repository, the checks to run before opening a pull request, and the commit message format releases are generated from.

## License

Forge.go is released under the [MIT License](LICENSE).
