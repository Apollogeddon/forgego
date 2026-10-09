---
title: Examples
weight: 3
---

## Configuration Patterns

### Enabling Another Linter

Add linters to `.golangci.local.yml`. Lists gain items, so these are enabled on top of the base set:

```yaml
version: "2"

linters:
  enable:
    - wsl_v5
    - goconst
```

Then regenerate `.golangci.yml`:

```bash
task sync
```

### Disabling a Base Linter

A list in `.golangci.local.yml` can only add items, so you can't remove a linter from the base `enable` list. Use golangci-lint's own `disable` list instead:

```yaml
version: "2"

linters:
  disable:
    - unparam
```

### Tuning a Linter's Settings

Mappings merge key by key, so settings for one linter leave the base settings for the others in place:

```yaml
version: "2"

linters:
  settings:
    gosec:
      excludes:
        - G104
```

### Excluding Paths

Add paths and extra exclusion rules; both are lists, so the base exclusions still apply:

```yaml
version: "2"

linters:
  exclusions:
    paths:
      - internal/legacy
    rules:
      - path: scripts/
        linters:
          - gosec
```

### Adding Your Own Tasks

Add tasks alongside the generated ones in `Taskfile.yml`. Forge.go never removes or changes a task it didn't create, and keeps its own tasks as you've edited them unless you run `init --force`, which also removes the tasks of a feature you've switched off:

```yaml
tasks:
  generate:
    desc: Regenerate code
    cmds:
      - go generate ./...
  check:
    desc: Lint, then test
    cmds:
      - task: lint
      - task: test
```

### Building a Specific Version

A backend's `build` task stamps the `VERSION` var into `main.version`. Override it from the command line:

```bash
task build VERSION=1.4.0
```

### Running the Tests With the Race Detector

CI runs `task test` with `GOFLAGS=-race`. Run it the same way locally — the race detector needs cgo:

```bash
GOFLAGS=-race task test
```

## Workflow Patterns

### Monorepo Execution

`forgego init -C services/api` scaffolds a subdirectory, using the repository's remote plus the subdirectory as the module path (`github.com/acme/platform/services/api`). GitHub only runs workflows from the repository root's `.github/workflows/`, so move the generated job there and point it at the subdirectory with `working_directory`, given relative to the repository root without a leading `./` or trailing `/`:

```yaml
jobs:
  api:
    uses: apollogeddon/forgego/.github/workflows/service.yml@main
    permissions:
      contents: write
      pull-requests: write
    with:
      working_directory: 'services/api'
      go_version: '1.27'
```

`version.yml` reads the module's own `services/api/.github/release.json` and `.release.json`, but release-please keys a package by its path from the repository root, so change the `"."` key in both to the subdirectory. Go also wants a subdirectory module's tags prefixed with its path (`services/api/v1.2.3`), so set that as the package's `component`, with `/` as the separator:

```json
{
  "packages": {
    "services/api": {
      "release-type": "go",
      "component": "services/api",
      "tag-separator": "/"
    }
  }
}
```

```json
{
  "services/api": "0.0.0"
}
```

release-please then only counts commits under `services/api`, and each module in the repository gets its own release PR and tags.

### Testing Across Go Versions

Call `testing.yml` from a matrix to run the quality, test and build jobs on several Go versions:

```yaml
jobs:
  test:
    strategy:
      matrix:
        go: ['1.26', '1.27']
    uses: apollogeddon/forgego/.github/workflows/testing.yml@main
    permissions:
      contents: write
    with:
      go_version: ${{ matrix.go }}
      # artifact names must be unique per run, and security patching should only run once
      artifact_name: dist-go${{ matrix.go }}
      auto_patch: false
```

The pinned tools need Go 1.27, so on an older `go_version` the `toolchain` line in `go.mod` decides which Go actually runs the tools.

### Custom Build Steps

`testing.yml` runs `task build` by default and uploads `dist/`; pass `build_command` and `artifact_path` to change them:

```yaml
jobs:
  test:
    uses: apollogeddon/forgego/.github/workflows/testing.yml@main
    permissions:
      contents: write
    with:
      build_command: 'go tool -modfile=.forgego/task/go.mod task build VERSION=ci'
      artifact_path: 'dist'
```

### Releasing Without Binaries

A service that's only ever deployed from its Docker image doesn't need GoReleaser to attach binaries. Turn the release job off:

```yaml
jobs:
  service:
    uses: apollogeddon/forgego/.github/workflows/service.yml@main
    permissions:
      contents: write
      pull-requests: write
    with:
      go_version: '1.27'
      release_binaries: false
```

Don't do this with `--debian`: the `.deb` is built by the same release job.

### Building Docker Images for More Platforms

With `--docker`, CI builds `linux/amd64` and `linux/arm64`. Add platforms with the `docker` job's `platforms` input:

```yaml
jobs:
  docker:
    needs: service
    uses: apollogeddon/forgego/.github/workflows/docker.yml@main
    permissions:
      contents: read
      packages: write
    with:
      push: ${{ github.ref == 'refs/heads/main' && needs.service.outputs.new_release_published == 'true' }}
      version: ${{ needs.service.outputs.version }}
      platforms: 'linux/amd64,linux/arm64,linux/arm/v7'
```

The backend `Dockerfile` cross-compiles, so extra platforms cost build time but no emulation. A platform must be one the runtime base image publishes: `gcr.io/distroless/static-debian12` for a backend, `nginx:stable-alpine` for a website.

### Self-Hosted Runners

Every workflow takes `runs_on`. Set it from a repository variable to switch runners without editing the workflow:

```yaml
jobs:
  service:
    uses: apollogeddon/forgego/.github/workflows/service.yml@main
    permissions:
      contents: write
      pull-requests: write
    with:
      runs_on: ${{ vars.RUNS_ON || 'ubuntu-latest' }}
```
