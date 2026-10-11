---
title: Job reference
weight: 1
---

This page describes each reusable workflow: its jobs, inputs and outputs. For the generated workflow and the inputs the orchestrators share, see [Workflows]({{< relref "_index.md" >}}).

## Pipeline overview

Every pipeline follows the same three stages:

1. **Quality and testing** — `testing.yml` runs `quality.yml`, the tests and the build.
2. **Versioning** — `version.yml` runs release-please on the main branch.
3. **Delivery** — `service.yml` (release binaries and `.deb`), `library.yml` (Go module proxy), `website.yml` (GitHub Pages) or `docker.yml` (GHCR) publishes the result.

Dependabot pull requests are auto-merged by `merge.yml` once testing passes, except GitHub Actions updates, which wait for a person to merge them. `review.yml` requests a review on Dependabot's and release-please's pull requests, so they reach your review requests in a private repository too.

`service.yml`, `library.yml` and `website.yml` expose `version.yml`'s `new_release_published`, `version` and `tag_name` as outputs, which the generated `docker` job uses to decide when to push.

## Choosing runners

Every workflow takes a `runs_on` input, default `ubuntu-latest`, and passes it down to each workflow it calls, so every job runs on that runner label. Set it per repository to use self-hosted runners, e.g. from a repository variable: `runs_on: ${{ vars.RUNS_ON || 'ubuntu-latest' }}`.

To test and build on other operating systems as well, list their runners in `extra_runners` (see [More operating systems](#more-operating-systems)), and in `service.yml`'s `native_builds` to build release binaries on them (see [Native release builds](#native-release-builds)).

`docker.yml` builds every platform in one job on `runs_on` too. A runner without a Docker daemon can build and push through a remote BuildKit set with `buildkit_endpoint`. The `tests` step runs with the race detector, which needs cgo, so a self-hosted runner needs a C compiler.

## Checking once per change

By default the checks run on every push and pull request, so a change is checked on its PR, again on `main`, and again around its release. To check each change only on its pull request:

```yaml
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  schedule:
    - cron: '17 3 * * 1'   # weekly: a full check of main

concurrency:
  group: ${{ github.workflow }}-${{ github.head_ref || github.run_id }}
  cancel-in-progress: true

jobs:
  service:
    uses: apollogeddon/forgego/.github/workflows/service.yml@main
    permissions:
      contents: write
      pull-requests: write
    with:
      test_on_push: false        # pushes to main only run release-please
      # test_release_prs: false  # also skip release-please's release PRs
```

- Pull requests run the full checks. A newer push cancels the run it replaces, and `main`'s own runs are never cancelled.
- Pushes to `main` run release-please first. Only when it makes a release does the **`release-testing`** job run the checks and the build, followed by the job that ships what they built: the GoReleaser release, the proxy publish or the Pages deploy.
- Release pull requests (branches starting `release-please--`) run the checks unless `test_release_prs` is `false`. When they do run, they're the one place every change since the last release is checked together.
- Skipped jobs count as passed for required status checks.
- With `test_on_push: false`, turn off **Require branches to be up to date before merging**. Otherwise every PR is checked again before merge.

A `website.yml` caller with `enable_versioning: false` deploys every push, so its pushes are always checked.

In this mode release-please tags the release before the push's checks run. If they fail, the tag and the GitHub release exist with nothing shipped, and need putting right by hand. So keep `test_release_prs: true`: the release PR is then checked with exactly what its merge ships.

## quality.yml

*Security and static analysis.*

1. **`secure`** — Gitleaks secret scan (skip with `enable_secrets: false`) and an OSV-Scanner scan of the source tree, which reports without failing. When the project pins govulncheck, it also runs `govulncheck ./...`, which reports only the vulnerabilities the code calls and warns rather than fails: the `patch` job upgrades them on `main`.
2. **`linting`** — `task sync-check` (when the project pins Task), or `task sync` on Dependabot's pull requests, as a Forge.go update changes the files it manages; then, with `lint` on, `go mod tidy -diff`, `golangci-lint fmt --diff` and `golangci-lint run`.

| Input | Default | Purpose |
| :--- | :--- | :--- |
| `runs_on` | `'ubuntu-latest'` | Runner label |
| `working_directory` | `'.'` | Directory containing `go.mod` |
| `go_version` | `''` | Go version; empty reads the project's `go.mod`, else Go 1.27 |
| `enable_secrets` | `true` | Run the Gitleaks scan |
| `lint` | `true` | Run the `go.mod` tidy check and golangci-lint; a website has no Go to lint |
| `apt_packages` | `''` | apt packages each job installs before compiling, such as a cgo dependency's C libraries |

## testing.yml

*The full QA suite.*

1. Calls → `quality.yml`.
2. **`testing`** — Runs `task test` with `GOFLAGS=-race` (skip with `run_tests: false`) and uploads `coverage.out` and `junit-report.xml` as the `coverage-<artifact_name>` artifact. *(Needs: quality)*
3. **`build`** — Runs `build_command` (default `task build`) and uploads `artifact_path` as the `artifact_name` artifact. *(Needs: quality; runs alongside testing)*
4. **`extra`** — Only with `extra_runners` set: on each runner it lists, such as macOS or Windows, runs the [setup script](#setup-scripts), then `task <extra_test_task>` (with `GOFLAGS=-race` wherever cgo is on, and skipped with `run_tests: false`), then `build_command`, and uploads `artifact_path` as the `<artifact_name>-<name>` artifact. *(Needs: quality; runs alongside testing)*
5. **`patch`** — On `main` with `auto_patch` enabled, first runs `task sync` and commits whatever it refreshes as `chore(deps): refresh the files forgego manages`, which catches up after a Dependabot update of Forge.go merges. It then runs govulncheck, `go get`s the fixed version of each vulnerable module it reports, runs `go mod tidy`, and commits `go.mod` and `go.sum` as `fix(deps): upgrade modules with known vulnerabilities via govulncheck`. Vulnerabilities in the standard library are fixed by a newer Go, not by `go get`, so they're left to you. Needs `contents: write`. *(Needs: quality, testing, build, extra)*

| Input | Default | Purpose |
| :--- | :--- | :--- |
| `runs_on` | `'ubuntu-latest'` | Runner label |
| `working_directory` | `'.'` | Directory containing `go.mod` |
| `go_version` | `''` | Go version; empty reads the project's `go.mod`, else Go 1.27 |
| `enable_secrets` | `true` | Run the Gitleaks scan |
| `lint` | `true` | Run golangci-lint |
| `apt_packages` | `''` | apt packages each job installs before compiling, passed on to `quality.yml` |
| `run_tests` | `true` | Run the tests |
| `auto_patch` | `true` | Run the `patch` job on `main` |
| `artifact_name` | `'dist'` | Name of the build artifact |
| `artifact_path` | `'dist'` | What the build writes, relative to `working_directory` |
| `build_command` | `'go tool -modfile=.forgego/task/go.mod task build'` | The build step |
| `extra_runners` | `'[]'` | JSON list of [more runners](#more-operating-systems) to test and build on |
| `extra_test_task` | `'test'` | The Task task the extra runners test with, such as a shorter `test:short` |
| `setup_script` | `''` | A [setup script](#setup-scripts), relative to `working_directory`, run before compiling on the extra runners and native builds |

### More operating systems

The `testing` and `build` jobs run on `runs_on` alone. To test and build on other operating systems too, list their runners in `extra_runners`, as a JSON string. An entry is a runner label, or an object:

| Key | Purpose |
| :--- | :--- |
| `runs_on` | Runner label, e.g. `macos-latest` |
| `name` | The artifact's suffix; defaults to `runs_on` |
| `env` | Environment variables for the setup script, the tests and the build, e.g. `{"CGO_ENABLED": "0"}`; single-line values |

```yaml
    with:
      extra_runners: >-
        ["macos-latest", {"runs_on": "windows-latest", "name": "windows", "env": {"CGO_ENABLED": "0"}}]
      extra_test_task: 'test:short'
      setup_script: '.github/setup.sh'
```

An entry's `env` applies to its tests and its build alike, so the tests compile the same way as the binary the job builds. Every `run` step uses bash, Git Bash on Windows. `apt_packages` is installed on Linux entries too. An empty list, the default, skips the job, so existing callers run exactly as before.

### Setup scripts

`setup_script` is a script in your repository, run with `bash` on each extra runner and native build before anything compiles. Use it for what a platform needs that `apt_packages` can't install, such as a system library from Homebrew or an installer on Windows. It runs in `working_directory`, after the entry's `env` is set. `RUNNER_OS` (`Linux`, `macOS` or `Windows`) says where it runs, and `GITHUB_TOKEN` holds the job's read-only token, e.g. for the GitHub API's rate limit. It can export variables to later steps through `$GITHUB_ENV` and add to `PATH` through `$GITHUB_PATH`. To run PowerShell on Windows, call it from the script:

```bash
#!/usr/bin/env bash
set -euo pipefail
case "$RUNNER_OS" in
  macOS) brew install pkg-config ;;
  Windows) pwsh -NoProfile -File .github/setup-windows.ps1 ;;
esac
```

## version.yml

*Manages the release lifecycle.*

1. **`release-please`** — On the main branch, opens or updates the release pull request from Conventional Commits, and creates the tag and GitHub release when it merges. It reads `.github/release.json` and `.github/.release.json` under `working_directory`, so each module in a monorepo is released on its own (see [Monorepos]({{< relref "/docs/examples.md#monorepos" >}})). Its outputs are those of the `working_directory` package, not of any other package released in the same run. Needs `contents: write` and `pull-requests: write`.

| Input | Default | Purpose |
| :--- | :--- | :--- |
| `runs_on` | `'ubuntu-latest'` | Runner label |
| `working_directory` | `'.'` | Directory of the project |

| Output | Description |
| :--- | :--- |
| `new_release_published` | Whether a new release was published |
| `version` | Released version, e.g. `1.2.3`; empty when nothing was released |
| `tag_name` | Released tag, e.g. `v1.2.3`; empty when nothing was released |

## merge.yml

*Dependabot auto-merge.*

1. **`auto-merge`** — For pull requests opened by Dependabot, turns on GitHub's auto-merge through the API, so the pull request merges once the required checks pass; one with no checks left to wait for is merged straight away. It needs no `gh` CLI on the runner. Needs `contents: write` and `pull-requests: write`.

Dependabot's GitHub Actions updates (branches starting `dependabot/github_actions/`) are skipped and left for a person to merge: they change workflow files, which the workflow's `GITHUB_TOKEN` can't merge.

| Input | Default | Purpose |
| :--- | :--- | :--- |
| `runs_on` | `'ubuntu-latest'` | Runner label |

## review.yml

*Review requests on bots' pull requests.*

1. **`request`** — Requests a review on a pull request from `reviewers`, or from the repository's owner when `reviewers` is empty and the owner is a user (an organisation names its reviewers through the input). GitHub only requests code owners' reviews in a private repository on a paid plan, so without this, Dependabot's and release-please's pull requests in a private repository on GitHub Free never reach your review requests. A failed request logs a warning and never fails the pipeline. Needs `pull-requests: write`.

`service.yml`, `library.yml` and `website.yml` call it on Dependabot's pull requests as they open, before the checks, so a failing update reaches you too; `version.yml` calls it on the release pull request it opened or updated. Dependabot updates that `merge.yml` merges leave your review requests once merged.

| Input | Default | Purpose |
| :--- | :--- | :--- |
| `runs_on` | `'ubuntu-latest'` | Runner label |
| `pull_request` | `''` | The pull request to request a review on; empty means the one that triggered the run |
| `reviewers` | `''` | Comma-separated logins to request; empty means the repository's owner |

## service.yml

*Orchestrates the full pipeline for backend projects.*

1. Calls → `testing.yml` to validate and build the project.
2. Calls → `merge.yml` to auto-merge Dependabot pull requests once testing passes. *(Needs: testing)*
3. Calls → `version.yml` to trigger a release on the main branch. Skip with `enable_versioning: false`. It still runs when the checks were skipped, but not when they failed. *(Needs: testing)*
4. Calls → `testing.yml` again as **`release-testing`**, only with `test_on_push: false` — see [Checking Once per Change](#checking-once-per-change). *(Needs: version)*
5. **`native-build`** — Only with `native_builds` set, under the same conditions as `release`: on each runner it lists, runs the [setup script](#setup-scripts) and `build_command`, and uploads the files `artifact_path` holds. See [Native release builds](#native-release-builds). Needs `contents: read`. *(Needs: testing, version, release-testing)*
6. **`release`** — On `main`, when a new release was published and the checks passed, runs GoReleaser (latest `v2`, through `goreleaser-action`) with `release --clean`. It attaches the binaries, `checksums.txt` and any `.deb` to the draft release release-please created, then publishes it. With `release_binaries: false` it attaches the native builds and `release_files`, if any, and publishes the draft. Needs `contents: write`. *(Needs: testing, version, release-testing, native-build)*

Takes the [common inputs]({{< relref "_index.md#common-inputs" >}}), plus:

| Input | Default | Purpose |
| :--- | :--- | :--- |
| `release_binaries` | `true` | Build the release binaries with GoReleaser and attach them to the release |
| `artifact_name` | `'dist'` | Name of the build artifact, passed to `testing.yml` |
| `artifact_path` | `'dist'` | What the build writes, relative to `working_directory`; a native build attaches every file at its top level |
| `build_command` | `'go tool -modfile=.forgego/task/go.mod task build'` | The build step, in `testing.yml` and each native build |
| `extra_runners` | `'[]'` | JSON list of [more runners](#more-operating-systems) to test and build on, passed to `testing.yml` |
| `extra_test_task` | `'test'` | The Task task the extra runners test with |
| `setup_script` | `''` | A [setup script](#setup-scripts) for the extra runners and native builds |
| `native_builds` | `'[]'` | JSON list of runners that build release binaries natively, entries as in [`extra_runners`](#more-operating-systems) |
| `release_files` | `''` | Files to attach beside the native builds, one per line, relative to `working_directory`, e.g. `LICENSE`; with `release_binaries: false` |
| `checksums` | `'checksums.txt'` | The checksum file written for the native builds and `release_files`; empty attaches none; with `release_binaries: false` |

### Native release builds

GoReleaser cross-compiles every platform on one Linux runner, which a cgo build can't do for macOS or Windows, and GoReleaser's prebuilt builds are a Pro feature. `native_builds` builds those on their own runners instead:

```yaml
    with:
      native_builds: >-
        [{"runs_on": "macos-latest", "name": "darwin"},
         {"runs_on": "windows-latest", "name": "windows", "env": {"CGO_ENABLED": "0"}}]
      setup_script: '.github/setup.sh'
```

Each entry is built at the release's commit with its whole history, so a build that names itself with `git describe` gets the tag. It runs `build_command` with the entry's `env`, so a `GOARCH` there builds that architecture where the runner can, and uploads every file at the top of `artifact_path`. Name the files for their platform, as they're all attached to one release.

GitHub only accepts a release's files while it is a draft once immutable releases are on, so release-please must create the release as a draft (`"draft": true` and `"force-tag-creation": true` in `.github/release.json`, as `forgego init` writes it). The `release` job waits for every native build, attaches the files to the draft and then publishes it, so a failed build leaves the draft unpublished for a re-run.

- **Without GoReleaser** (`release_binaries: false`), the `release` job attaches the native builds and `release_files` itself with `gh release upload`. With `checksums` set, it also attaches a checksum file of them all in GoReleaser's format: each file's SHA-256, two spaces and its name, sorted by name.
- **With GoReleaser**, it adds the native builds to its own release. The `release` job puts them in `.forgego-native/` in `working_directory`, which it keeps out of git so GoReleaser doesn't count the checkout as dirty. List them in `.goreleaser.yaml`, so they're attached and in `checksums.txt` with GoReleaser's own files; the job fails rather than publish without them. A local `goreleaser release --snapshot` has no `.forgego-native/` and skips them.

```yaml
release:
  extra_files:
    - glob: .forgego-native/*
checksum:
  extra_files:
    - glob: .forgego-native/*
```

## library.yml

*Orchestrates library releases.*

1. Calls → `testing.yml`, `merge.yml`, `version.yml` and `release-testing`, as `service.yml` does.
2. **`publish`** — On `main`, when a new release was published and the checks passed, fetches the new version through `proxy.golang.org` with `go mod download`. A Go module is published by its tag; this makes the proxy and pkg.go.dev pick the release up straight away rather than on first use. From `v2` on, the job fails if the module path doesn't end in the matching major version suffix (such as `/v2`), as Go can't fetch that release. Disable with `publish: false`. Needs only `contents: read`. *(Needs: testing, version, release-testing)*

Takes the [common inputs]({{< relref "_index.md#common-inputs" >}}), plus:

| Input | Default | Purpose |
| :--- | :--- | :--- |
| `publish` | `true` | Ask the Go module proxy for each new release |

## website.yml

*Orchestrates the full pipeline for website projects and deploys to GitHub Pages.*

1. Calls → `testing.yml` to build the site, with `lint` off and `run_tests` and `auto_patch` off by default — a Hugo site has no Go to lint, test or patch. The build (`build_command`, default `task build`) writes `public/`, uploaded as the `artifact_name` artifact.
2. Calls → `merge.yml` to auto-merge Dependabot pull requests once testing passes, and the site checks if they're on. Disable with `auto_merge: false`. *(Needs: testing, site-checks)*
3. Calls → `version.yml` to check if a new release was published. Skip with `enable_versioning: false`. *(Needs: testing)*
4. Calls → `testing.yml` again as **`release-testing`**, only with `test_on_push: false`. *(Needs: version)*
5. **`site-checks`** — Only with `check_links` or `check_accessibility`: checks the built artifact's links and pages. See [Site checks](#site-checks). Needs `contents: read`. *(Needs: testing, release-testing)*
6. **`deploy`** — Downloads the build artifact and deploys it to GitHub Pages, in the `github-pages` environment. Runs on the main branch only and, when versioning is enabled, only when a new release is published, and not when a site check failed. Needs `pages: write` and `id-token: write`. *(Needs: testing, version, release-testing, site-checks)*

Takes the [common inputs]({{< relref "_index.md#common-inputs" >}}) except `lint`, with `run_tests` and `auto_patch` off by default as a Hugo site has no Go to test or patch, plus:

| Input | Default | Purpose |
| :--- | :--- | :--- |
| `artifact_name` | `'dist'` | Name of the site artifact passed to the deploy |
| `build_command` | `'go tool -modfile=.forgego/task/go.mod task build'` | Builds the site into `public/` |
| `auto_merge` | `true` | Auto-merge Dependabot pull requests |
| `check_links` | `false` | Check the built site's links with lychee |
| `link_check_args` | `'--offline'` | More lychee arguments; the default checks internal links only, `''` checks external links too |
| `check_accessibility` | `false` | Check the built site's pages with axe |
| `accessibility_paths` | `'/'` | The pages axe checks, relative to `base_url`, one per line or space-separated |
| `accessibility_args` | `''` | More axe arguments, such as `--tags wcag2a,wcag2aa` or `--disable color-contrast` |
| `base_url` | `'/'` | The site's `baseURL` from its Hugo config, for the site checks |

### Site checks

Both checks run in one `site-checks` job on the artifact the build uploaded, so they check exactly what the deploy would ship, and a failure stops the deploy and the Dependabot merge. Both are off by default.

```yaml
    with:
      check_links: true
      check_accessibility: true
      accessibility_paths: |
        /
        /docs/
        /docs/getting-started/
      base_url: https://owner.github.io/repo/
```

- **`base_url`** must match `baseURL` in `hugo.toml`, as the build's links include its path. The job puts the site under that path, so `/repo/docs/` resolves, and treats links to the site's own address as internal. The default `/` suits the `hugo.toml` that `forgego init --website` writes.
- **`check_links`** runs [lychee](https://github.com/lycheeverse/lychee) over every HTML page, with a directory link needing its `index.html`. By default it runs `--offline`, checking only links within the site, so another site being down can't fail the build. Set `link_check_args: ''` to check external links too, or pass more lychee options there, such as `--exclude <pattern>`. Its report is in the job summary.
- **`check_accessibility`** serves the site on `127.0.0.1` with Python's `http.server`, installs a matching Chrome and ChromeDriver with browser-driver-manager, and runs [axe](https://github.com/dequelabs/axe-core-npm/tree/develop/packages/cli) on each page in `accessibility_paths`. Any violation fails the job; narrow the rules with `accessibility_args`, e.g. `--tags wcag2a,wcag2aa`, or `--disable <rule>` for a theme's known issue. The runner needs `python3`, and on a self-hosted runner the libraries Chrome needs.

## docker.yml

*Builds a Docker image for any number of platforms and publishes it to GitHub Container Registry.*

`forgego init --docker` adds it to your `index.yml` as its own `docker` job after your pipeline job, so projects without Docker don't carry it. It builds on every run and pushes when the pipeline reports a new release. The job needs `packages: write` to push.

1. **`build`** — One job builds every platform. The generated Dockerfiles cross-compile (backend) or build the site (website) on the build host, and their runtime stages run nothing, so no platform needs emulation or a runner of its own. On pull requests the image is built but not pushed, so a broken Dockerfile fails the pull request's checks; make the `docker` job a required status check to stop Dependabot auto-merge on a failing build. On a release it pushes one multi-platform image tagged `X.Y.Z`, `X.Y`, `X` (not for `0.x` versions), `sha-<commit>` and `latest`, with provenance and SBOM attestations.

Builds that aren't pushed use the GitHub Actions cache. Pushed images build without a cache, which another repository's job could have planted layers in. The build passes the release version as the `VERSION` build argument (`dev` when there's none), and any `build_args` after it.

A Dockerfile that compiles in each platform's own image, such as a cgo build against a platform's C libraries, needs the other platforms emulated: set `qemu: true`. Emulated builds are many times slower, so the job's timeout rises from 30 to 90 minutes. With `buildkit_endpoint` set, QEMU isn't set up, as the remote BuildKit runs the build.

`tags` replaces the default tags with your own [metadata-action rules](https://github.com/docker/metadata-action#tags-input), one per line, for example:

```yaml
      tags: |
        type=semver,pattern={{version}},value=${{ needs.service.outputs.tag_name }}
        type=raw,value=latest
        type=sha,format=long,prefix=
```

With `buildkit_endpoint` set, the job builds on that remote BuildKit instead of the runner's own Docker, so self-hosted runners without a Docker daemon can build and push multi-platform images. The registry credentials are written for buildx directly, as there's no `docker login` without a daemon.

| Input | Default | Purpose |
| :--- | :--- | :--- |
| `push` | `false` | Push to GHCR; the generated job sets it for a new release on `main` |
| `version` | `''` | Release version used for the semver tags |
| `image` | `ghcr.io/<owner>/<repo>` | Image name override (lowercased) |
| `platforms` | `linux/amd64,linux/arm64` | Comma-separated platforms, e.g. `linux/amd64,linux/arm64,linux/arm/v7` |
| `working_directory` | `'.'` | The build context |
| `buildkit_endpoint` | `''` | A remote BuildKit, e.g. `tcp://runner.builder:1234`; empty uses the runner's own |
| `dockerfile` | `''` | The Dockerfile, relative to `working_directory`; empty means `Dockerfile` |
| `qemu` | `false` | Emulate other platforms with QEMU, for a Dockerfile that compiles in each platform's own image; the runner's own Docker only |
| `build_args` | `''` | More build arguments, one `NAME=value` per line, beside `VERSION` |
| `tags` | `''` | metadata-action tag rules, one per line, in place of the default tags |
| `runs_on` | `'ubuntu-latest'` | Runner label |
