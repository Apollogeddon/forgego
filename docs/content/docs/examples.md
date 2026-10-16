---
title: Examples
weight: 6
---

These recipes cover common changes to a Forge.go project's configuration and to `.github/workflows/.index.yml`, the GitHub Actions workflow that calls Forge.go's pipelines.

## Configuration patterns

### Enabling another linter

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

### Disabling a base linter

A list in `.golangci.local.yml` can only add items, so you can't remove a linter from the base `enable` list. Use golangci-lint's own `disable` list instead:

```yaml
version: "2"

linters:
  disable:
    - unparam
```

### Tuning a linter's settings

Mappings merge key by key, so settings for one linter leave the base settings for the others in place:

```yaml
version: "2"

linters:
  settings:
    gosec:
      excludes:
        - G104
```

### Excluding paths

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

### Adding your own tasks

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

### Building a specific version

A backend's `build` task stamps the `VERSION` var into `main.version`. Override it from the command line:

```bash
task build VERSION=1.4.0
```

### Running the tests with the race detector

CI runs `task test` with `GOFLAGS=-race`. Run it the same way locally — the race detector needs cgo:

```bash
GOFLAGS=-race task test
```

## Workflow patterns

### Monorepos

`forgego init -C services/api` scaffolds a subdirectory, using the repository's remote plus the subdirectory as the module path (`github.com/acme/platform/services/api`). GitHub only runs workflows from the repository root's `.github/workflows/`, so move the generated jobs there and point each at the subdirectory with `working_directory`, given relative to the repository root without a leading `./` or trailing `/`. A second module's jobs name their variant in brackets, so the run page reads `service (api) / test`:

```yaml
jobs:
  version-api:
    name: version (api)
    if: ${{ github.event_name != 'pull_request' && github.event_name != 'schedule' }}
    uses: apollogeddon/forgego/.github/workflows/version.yml@main
    permissions:
      contents: write
      pull-requests: write
    with:
      working_directory: 'services/api'

  service-api:
    name: service (api)
    needs: version-api
    if: ${{ !cancelled() && (github.event_name == 'pull_request' || needs.version-api.outputs.new_release_published == 'true') }}
    uses: apollogeddon/forgego/.github/workflows/service.yml@main
    permissions:
      contents: read
      pull-requests: write
    with:
      working_directory: 'services/api'
```

Pass the same `working_directory` to the module's `publish`, `patch` and `docker` jobs.

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

### Testing across Go versions

Call a check pipeline once more with another `go_version` to run its stages on that Go too:

```yaml
jobs:
  service-go126:
    name: service (go 1.26)
    needs: version
    if: ${{ !cancelled() && (github.event_name == 'pull_request' || needs.version.outputs.new_release_published == 'true') }}
    uses: apollogeddon/forgego/.github/workflows/service.yml@main
    permissions:
      contents: read
      pull-requests: write
    with:
      go_version: '1.26'
      # artifact names must be unique per run
      artifact_name: dist-go1.26
```

Add it to `merge`'s `needs` so a Dependabot merge waits for it too.

The pinned tools need Go 1.27, so on an older `go_version` the `toolchain` line in `go.mod` decides which Go actually runs the tools.

### Custom build steps

The check pipelines' `build` stage runs `task build` by default and uploads `dist/`; pass `build_command` and `artifact_path` to change them:

```yaml
jobs:
  service:
    # ...as generated
    uses: apollogeddon/forgego/.github/workflows/service.yml@main
    permissions:
      contents: read
      pull-requests: write
    with:
      build_command: 'go tool -modfile=.forgego/task/go.mod task build VERSION=ci'
      artifact_path: 'dist'
```

### Releasing without binaries

A service that's only ever deployed from its Docker image doesn't need GoReleaser to attach binaries. Turn GoReleaser off in the `publish` job, which then only publishes release-please's draft:

```yaml
jobs:
  publish:
    # ...as generated
    uses: apollogeddon/forgego/.github/workflows/publish.yml@main
    permissions:
      contents: write
    with:
      tag_name: ${{ needs.version.outputs.tag_name }}
      release_binaries: false
```

Don't do this with `--debian`: the `.deb` is built by the same GoReleaser run.

### Building Docker images for more platforms

With `--docker`, CI builds `linux/amd64` and `linux/arm64`. Add platforms with the `docker` job's `platforms` input:

```yaml
jobs:
  docker:
    needs: [version, service]
    if: ${{ !cancelled() && needs.service.result == 'success' }}
    uses: apollogeddon/forgego/.github/workflows/docker.yml@main
    permissions:
      contents: read
      packages: write
    with:
      push: ${{ needs.version.outputs.new_release_published == 'true' }}
      version: ${{ needs.version.outputs.version }}
      platforms: 'linux/amd64,linux/arm64,linux/arm/v7'
```

The backend `Dockerfile` cross-compiles, so extra platforms cost build time but no emulation. A platform must be one the runtime base image publishes: `gcr.io/distroless/static-debian12` for a backend, `nginx:stable-alpine` for a website.

### Self-hosted runners

Every pipeline takes `runs_on`. Set it on each job from a repository variable to switch runners without editing the workflow:

```yaml
jobs:
  service:
    # ...as generated
    uses: apollogeddon/forgego/.github/workflows/service.yml@main
    permissions:
      contents: read
      pull-requests: write
    with:
      runs_on: ${{ vars.RUNS_ON || 'ubuntu-latest' }}
```
