---
title: Pipeline reference
weight: 1
---

This page describes each pipeline: its stages, inputs and outputs. For the generated `.index.yml`, when each pipeline runs, and the permissions they need, see [Workflows]({{< relref "_index.md" >}}).

## Choosing runners

Every pipeline takes a `runs_on` input, default `ubuntu-latest`, and runs each of its stages on that runner label. Set it per repository to use self-hosted runners, e.g. from a repository variable: `runs_on: ${{ vars.RUNS_ON || 'ubuntu-latest' }}`.

To test and build on Windows and macOS as well, set `library.yml`'s or `service.yml`'s `test_windows` and `test_macos` (see [Windows and macOS](#windows-and-macos)), and `publish.yml`'s `native_builds` to build release binaries on other runners (see [Native release builds](#native-release-builds)).

`docker.yml` builds every platform in one job on `runs_on` too. A runner without a Docker daemon can build and push through a remote BuildKit set with `buildkit_endpoint`. The `test` stage runs with the race detector, which needs cgo, so a self-hosted runner needs a C compiler.

## Check pipelines: library.yml, service.yml, website.yml

*A project's checks, on pull requests and before a release ships.*

1. **`lint`** — `task sync-check` (when the project pins Task), or `task sync` on Dependabot's pull requests, as a Forge.go update changes the files it manages; then, with `lint` on, `go mod tidy -diff`, `golangci-lint fmt --diff` and `golangci-lint run`. A website has no Go to lint, so its `lint` stage is the sync check alone.
2. **`scan`** — Gitleaks secret scan (skip with `enable_secrets: false`) and an OSV-Scanner scan of the source tree, with a pinned OSV-Scanner release checked against the checksums it was published with; a failed download or check fails the stage. On a pull request, a high or critical finding (a CVSS score of 7 or more, or GitHub's rating when an advisory has no score) fails the stage, so a change can't add one. Elsewhere, such as on a release, it's an error annotation that doesn't fail the stage, so releases still run while a Dependabot security update fixes it. Lower findings, and advisories with no severity at all, as Go's own vulnerability database gives, are warning annotations; every finding is listed in the job summary. If the scanner can't finish, for example because OSV's API is unreachable, the stage fails on a pull request, which mustn't pass unchecked, and warns elsewhere. To accept a finding you can't fix yet, list it in an [`osv-scanner.toml`](https://google.github.io/osv-scanner/configuration/) beside `go.mod`:

   ```toml
   [[IgnoredVulns]]
   id = "GO-2026-0000"
   ignoreUntil = 2026-12-31
   reason = "No fixed version yet; the vulnerable function isn't called"
   ```

   On a public repository's pull request, it also runs GitHub's [dependency review](https://github.com/actions/dependency-review-action), which fails on a dependency the pull request adds or changes with a high or critical advisory. A private repository needs GitHub Code Security for it, so there it's skipped, and OSV-Scanner covers it.

   When the project pins govulncheck, it also runs `govulncheck ./...`, which reports only the vulnerabilities the code calls. Like OSV-Scanner, it fails a pull request and only warns elsewhere, so a release still ships: [Dependabot security updates]({{< relref "_index.md#vulnerable-dependencies" >}}) upgrade them.
3. **`test`** — `task test` with `GOFLAGS=-race` (skip with `run_tests: false`), uploading `coverage.out` and `junit-report.xml` as the `coverage-<artifact_name>` artifact. *(Needs: lint, scan)*
4. **`test (windows)`**, **`test (macos)`** — `library.yml` and `service.yml`, each when asked: see [Windows and macOS](#windows-and-macos). *(Needs: lint, scan)*
5. **`build`** — `build_command` (default `task build`), uploading `artifact_path` (a website's `public/`) as the `artifact_name` artifact. *(Needs: lint, scan; runs alongside the tests)*
6. **`check`** — `website.yml` only, with `check_links` or `check_accessibility`: see [Site checks](#site-checks). *(Needs: build)*

Every stage asks for `contents: read`; `scan` also asks for `pull-requests: write`, for Gitleaks.

| Input | Default | Pipelines | Purpose |
| :--- | :--- | :--- | :--- |
| `runs_on` | `'ubuntu-latest'` | all | Runner label |
| `working_directory` | `'.'` | all | Directory containing `go.mod` |
| `go_version` | `''` | all | Go version; empty reads the project's `go.mod`, else Forge.go's default |
| `enable_secrets` | `true` | all | Run the Gitleaks scan |
| `run_tests` | `true` (`false` for `website.yml`) | all | Run the `test` stages — `forgego init --no-testing` sets it to `false` |
| `artifact_name` | `'dist'` | all | Name of the build artifact |
| `build_command` | `'go tool -modfile=.forgego/task/go.mod task build'` | all | The build stage's command |
| `lint` | `true` | `library`, `service` | Run golangci-lint and the `go mod tidy` check — `forgego init --no-linting` sets it to `false` |
| `apt_packages` | `''` | `library`, `service` | apt packages to install before anything compiles, space-separated: the C libraries a cgo dependency needs, such as `libfuse-dev`. Ubuntu runners only |
| `artifact_path` | `'dist'` | `library`, `service` | What the build writes, relative to `working_directory` |
| `test_windows`, `test_macos` | `''` | `library`, `service` | Test and build on Windows or macOS too; see below |
| `extra_test_task` | `'test'` | `library`, `service` | The Task task the Windows and macOS stages test with, such as a shorter `test:short` |
| `setup_script` | `''` | `library`, `service` | A [setup script](#setup-scripts), run before compiling on Windows and macOS |
| `base_path`, `check_links`, `check_external_links`, `link_check_args`, `check_accessibility`, `accessibility_paths`, `accessibility_impact` | | `website` | See [Site checks](#site-checks) |

### Windows and macOS

The `test` and `build` stages run on `runs_on` alone. To test and build on Windows or macOS too, set `test_windows` or `test_macos` to a runner label, or to a JSON object:

| Key | Purpose |
| :--- | :--- |
| `runs_on` | Runner label; defaults to `windows-latest` or `macos-latest` |
| `env` | Environment variables for the setup script, the tests and the build, e.g. `{"CGO_ENABLED": "0"}`; single-line values |

```yaml
    with:
      test_macos: macos-latest
      test_windows: '{"runs_on": "windows-latest", "env": {"CGO_ENABLED": "0"}}'
      extra_test_task: 'test:short'
      setup_script: '.github/setup.sh'
```

Each runs the [setup script](#setup-scripts), then `task <extra_test_task>` (with `GOFLAGS=-race` wherever cgo is on, and skipped with `run_tests: false`), then `build_command`, and uploads `artifact_path` as the `<artifact_name>-windows` or `<artifact_name>-macos` artifact. The `env` applies to the tests and the build alike, so the tests compile the same way as the binary the stage builds. Every `run` step uses bash, Git Bash on Windows. Empty, the default, skips the stage, so the generated `.index.yml` spends no Windows or macOS minutes, which a private repository pays for at a higher rate; set them on its check job to test there. Forge.go's own CI sets both.

### Setup scripts

`setup_script` is a script in your repository, run with `bash` on Windows and macOS and in each native build before anything compiles. Use it for what a platform needs that `apt_packages` can't install, such as a system library from Homebrew or an installer on Windows. It runs in `working_directory`, after the `env` is set. `RUNNER_OS` (`Linux`, `macOS` or `Windows`) says where it runs, and `GITHUB_TOKEN` holds the job's read-only token, e.g. for the GitHub API's rate limit. It can export variables to later steps through `$GITHUB_ENV` and add to `PATH` through `$GITHUB_PATH`. To run PowerShell on Windows, call it from the script:

```bash
#!/usr/bin/env bash
set -euo pipefail
case "$RUNNER_OS" in
  macOS) brew install pkg-config ;;
  Windows) pwsh -NoProfile -File .github/setup-windows.ps1 ;;
esac
```

### Site checks

Both checks run in one `check` stage on the artifact the build uploaded, so they check exactly what the deploy would ship, and a failure stops the deploy and the Dependabot merge. Both are off by default. The inputs match forgejs's website pipeline.

```yaml
    with:
      check_links: true
      check_accessibility: true
      base_path: /repo/
```

| Input | Default | Purpose |
| :--- | :--- | :--- |
| `check_links` | `false` | Check the built site's links with lychee; internal links only unless `check_external_links` |
| `check_external_links` | `false` | With `check_links`, check links to other sites too |
| `link_check_args` | `''` | Extra lychee arguments, such as `--include-fragments` or `--exclude <pattern>` |
| `check_accessibility` | `false` | Check the built site against WCAG 2.2 A and AA with axe, in a headless Chromium |
| `accessibility_paths` | `''` | Comma- or newline-separated pages to check, relative to `base_path`; empty checks every page in the sitemap, else the home page |
| `accessibility_impact` | `'serious'` | The lowest axe impact that fails the check (`minor`, `moderate`, `serious` or `critical`); lower ones are warnings |
| `base_path` | `''` | The path the site is served under, such as `/repo/` for a GitHub Pages project site; empty means `/` |

- **`base_path`** must be the path in `baseURL` in `hugo.toml`, such as `/repo/` for `https://owner.github.io/repo/`, as the build's links include it. The stage serves the site under that path. The default suits the `hugo.toml` that `forgego init --website` writes.
- **`check_links`** runs [lychee](https://github.com/lycheeverse/lychee) over every HTML page, with a directory link needing its `index.html`. Hugo writes many links with the site's full address; the stage reads that address from the site's `sitemap.xml` and checks those links in the build as well. It runs `--offline` unless `check_external_links` is on, so another site being down can't fail the build. Pass more lychee options in `link_check_args`, or add a `lychee.toml` or `.lycheeignore` to the repository. Its report is in the job summary.
- **`check_accessibility`** serves the site locally and runs [axe](https://github.com/dequelabs/axe-core) in Playwright's headless Chromium on each page in `accessibility_paths`, or every page in the sitemap Hugo writes. Violations at `accessibility_impact` or above fail the stage, and lower ones are reported as warnings, with a table in the job summary. Some themes fail axe's `color-contrast` or `target-size` rules, which are `serious`: fix the theme, or set `accessibility_impact: critical` until you do. A self-hosted runner needs passwordless `sudo`, or Chromium's system libraries already installed.

## version.yml

*Manages the release lifecycle.*

1. **`release`** — On the main branch, opens or updates the release pull request from Conventional Commits, and creates the tag and the GitHub release when it merges. It reads `.github/release.json` and `.github/.release.json` under `working_directory`, so each module in a monorepo is released on its own (see [Monorepos]({{< relref "/docs/examples.md#monorepos" >}})). Its outputs are those of the `working_directory` package, not of any other package released in the same run. It then requests a review on the release pull request from `reviewers`, or from the repository's owner when `reviewers` is empty and the owner is a user, as `CODEOWNERS` requests none in a private repository on GitHub Free; a failed request warns and never fails the pipeline. Needs `contents: write` and `pull-requests: write`.

| Input | Default | Purpose |
| :--- | :--- | :--- |
| `runs_on` | `'ubuntu-latest'` | Runner label |
| `working_directory` | `'.'` | Directory of the project |
| `reviewers` | `''` | Comma-separated logins to request a review of the release pull request from; empty means the repository's owner |
| `app_client_id` | `''` | A GitHub App's client ID: release-please opens its pull request with the app's token, so the pull request runs the checks; empty uses `GITHUB_TOKEN` |

And one secret, optional:

| Secret | Purpose |
| :--- | :--- |
| `APP_PRIVATE_KEY` | The GitHub App's private key; with `app_client_id`. See [Release pull requests that run the checks]({{< relref "_index.md#release-pull-requests-that-run-the-checks" >}}) |

| Output | Description |
| :--- | :--- |
| `new_release_published` | Whether a new release was published |
| `version` | Released version, e.g. `X.Y.Z`; empty when nothing was released |
| `tag_name` | Released tag, e.g. `vX.Y.Z`; empty when nothing was released |

### Checking a release

The generated `.index.yml` checks each change on its pull request, and on `main` runs `version` first. Only when release-please makes a release does it run the checks again and then the job that ships what they built: `publish`, the `docker` push or the Pages deploy. Release pull requests (branches starting `release-please--`) run the checks too, once something starts them: one opened with `GITHUB_TOKEN` starts no workflow, so set up a [release app]({{< relref "_index.md#release-pull-requests-that-run-the-checks" >}}) or push to the branch. They're the one place every change since the last release is checked together. Skipped jobs count as passed for required status checks, and with checks only on pull requests, turn off **Require branches to be up to date before merging**, or every PR is checked again before merge.

release-please tags the release before the release's checks run. If they fail, the tag and the draft release exist with nothing shipped, and need putting right by hand, which is why release pull requests are checked with exactly what their merge ships.

## publish.yml

*Ships a release.*

1. **`package`** — With `package: true`, fetches `tag_name` through `proxy.golang.org` with `go mod download`. A Go module is published by its tag; this makes the proxy and pkg.go.dev pick the release up straight away rather than on first use. From `v2` on, the stage fails if the module path doesn't end in the matching major version suffix (such as `/v2`), as Go can't fetch that release. Needs `contents: read`.
2. **`assets (native)`** — Only with `native_builds` set: on each runner it lists, runs the [setup script](#setup-scripts) and `build_command`, and uploads the files `artifact_path` holds. See [Native release builds](#native-release-builds). Needs `contents: read`.
3. **`assets`** — With `assets` on (the default), runs GoReleaser through `goreleaser-action`, pinned to the version the generated `release:snapshot` task runs, with `release --clean`. It attaches the binaries, `checksums.txt` and any `.deb` to the draft release release-please created, then publishes it. With `gpg_sign: true` it first imports the signing key for GoReleaser. With `release_binaries: false` it attaches the native builds and `release_files`, if any, and publishes the draft. It asks for no permissions of its own and takes the caller's grant, which needs `contents: write`. *(Needs: assets (native))*
4. **`attest`** — With `attest: true`, downloads the release's `checksums` file and attests the build provenance of every file it lists with [attest-build-provenance](https://github.com/actions/attest-build-provenance), so a download can be checked with `gh attestation verify <file> --repo <owner>/<repo>`. GitHub Free attests only a public repository's builds, so the generated job passes `attest: ${{ github.event.repository.private == false }}`. Like `assets`, it takes the caller's grant, which needs `id-token: write` and `attestations: write`. *(Needs: assets)*

| Input | Default | Purpose |
| :--- | :--- | :--- |
| `runs_on` | `'ubuntu-latest'` | Runner label |
| `working_directory` | `'.'` | Directory containing `go.mod` |
| `go_version` | `''` | Go version; empty reads the project's `go.mod`, else Forge.go's default |
| `tag_name` | `''` | The release, as `version.yml`'s `tag_name`; `package` and the attachments without GoReleaser need it |
| `package` | `false` | Fetch the release through the Go module proxy; a library's caller turns it on |
| `assets` | `true` | Attach the release's files and publish the draft; a library's caller turns it off |
| `release_binaries` | `true` | Build the release binaries with GoReleaser and attach them to the release |
| `build_command` | `'go tool -modfile=.forgego/task/go.mod task build'` | Each native build's command |
| `artifact_path` | `'dist'` | What a native build writes, relative to `working_directory`; every file at its top level is attached |
| `apt_packages` | `''` | apt packages a native build on Ubuntu installs before compiling |
| `setup_script` | `''` | A [setup script](#setup-scripts) for the native builds |
| `native_builds` | `'[]'` | JSON list of runners that build release binaries natively; see below |
| `release_files` | `''` | Files to attach beside the native builds, one per line, relative to `working_directory`, e.g. `LICENSE`; with `release_binaries: false` |
| `checksums` | `'checksums.txt'` | The checksum file written for the native builds and `release_files`, empty attaching none, with `release_binaries: false`; and the checksum file `attest` reads, which with GoReleaser is its `checksum.name_template` (`checksums.txt` as `forgego init` writes it) |
| `attest` | `false` | Attest the build provenance of the release's files; needs a public repository on GitHub Free |
| `gpg_sign` | `false` | Import the `GPG_PRIVATE_KEY` secret and give GoReleaser its fingerprint as `GPG_FINGERPRINT`; see [Signed releases](#signed-releases) |

And these secrets, both optional:

| Secret | Purpose |
| :--- | :--- |
| `GPG_PRIVATE_KEY` | The ASCII-armored private key GoReleaser signs with; with `gpg_sign: true` |
| `GPG_PASSPHRASE` | The key's passphrase, if it has one |

To act once a release is out, such as rebuilding a site that lists its files, add a job that `needs` the `publish` job: a job that needs a reusable workflow waits for every job in it.

### Native release builds

GoReleaser cross-compiles every platform on one Linux runner, which a cgo build can't do for macOS or Windows, and GoReleaser's prebuilt builds are a Pro feature. `native_builds` builds those on their own runners instead. An entry is a runner label, or an object with `runs_on` and `env`; GitHub names each build `assets (native)` followed by its entry's values:

```yaml
    with:
      native_builds: >-
        [{"runs_on": "macos-latest", "name": "darwin"},
         {"runs_on": "windows-latest", "name": "windows", "env": {"CGO_ENABLED": "0"}}]
      setup_script: '.github/setup.sh'
```

Each entry is built at the release's commit with its whole history, so a build that names itself with `git describe` gets the tag. It runs `build_command` with the entry's `env`, so a `GOARCH` there builds that architecture where the runner can, and uploads every file at the top of `artifact_path`. Name the files for their platform, as they're all attached to one release.

GitHub only accepts a release's files while it is a draft once immutable releases are on, so release-please must create the release as a draft (`"draft": true` and `"force-tag-creation": true` in `.github/release.json`, as `forgego init` writes it). The `assets` stage waits for every native build, attaches the files to the draft and then publishes it, so a failed build leaves the draft unpublished for a re-run.

- **Without GoReleaser** (`release_binaries: false`), `assets` attaches the native builds and `release_files` itself with `gh release upload`. With `checksums` set, it also attaches a checksum file of them all in GoReleaser's format: each file's SHA-256, two spaces and its name, sorted by name.
- **With GoReleaser**, it adds the native builds to its own release. `assets` puts them in `.forgego-native/` in `working_directory`, which it keeps out of git so GoReleaser doesn't count the checkout as dirty. List them in `.goreleaser.yaml`, so they're attached and in `checksums.txt` with GoReleaser's own files; the stage fails rather than publish without them. A local `goreleaser release --snapshot` has no `.forgego-native/` and skips them.

```yaml
release:
  extra_files:
    - glob: .forgego-native/*
checksum:
  extra_files:
    - glob: .forgego-native/*
```

### Signed releases

Some consumers want a release's checksum file signed, such as a provider registry that checks a detached signature of it. With `gpg_sign: true`, `assets` imports `GPG_PRIVATE_KEY` (with `GPG_PASSPHRASE`) into the runner's keyring through [ghaction-import-gpg](https://github.com/crazy-max/ghaction-import-gpg) and sets `GPG_FINGERPRINT` to its fingerprint for GoReleaser. It fails without the key, or with `release_binaries: false`, as the signing itself is GoReleaser's. Pass the secrets by name:

```yaml
jobs:
  publish:
    needs: [version, service]
    if: ${{ !cancelled() && needs.service.result == 'success' && needs.version.outputs.new_release_published == 'true' }}
    uses: apollogeddon/forgego/.github/workflows/publish.yml@main
    permissions:
      contents: write
    with:
      tag_name: ${{ needs.version.outputs.tag_name }}
      gpg_sign: true
    secrets:
      GPG_PRIVATE_KEY: ${{ secrets.GPG_PRIVATE_KEY }}
      GPG_PASSPHRASE: ${{ secrets.GPG_PASSPHRASE }}
```

Then sign the checksum file in `.goreleaser.yaml`, which attaches `<checksum file>.sig` beside it:

```yaml
signs:
  - artifacts: checksum
    args:
      - --batch
      - --local-user
      - '{{ .Env.GPG_FINGERPRINT }}'
      - --output
      - ${signature}
      - --detach-sign
      - ${artifact}
```

`goreleaser release --snapshot` skips signing, so local snapshots need no key.

### Releasing after your own checks

A project whose release must wait for checks the check pipelines can't run, such as acceptance tests against a service started in the job, calls a check pipeline and its own checks as root jobs beside the generated `version` job, and `publish` after them all:

```yaml
jobs:
  service:
    uses: apollogeddon/forgego/.github/workflows/service.yml@main
    permissions:
      contents: read
      pull-requests: write
  acceptance:
    needs: service
    uses: ./.github/workflows/acceptance.yml
  publish:
    needs: [version, acceptance]
    if: ${{ !cancelled() && needs.acceptance.result == 'success' && needs.version.outputs.new_release_published == 'true' }}
    uses: apollogeddon/forgego/.github/workflows/publish.yml@main
    permissions:
      contents: write
    with:
      tag_name: ${{ needs.version.outputs.tag_name }}
```

## pages.yml

*Deploys a built website to GitHub Pages.*

1. **`deploy`** — Downloads the `artifact_name` artifact `website.yml` uploaded earlier in the run and deploys it, in the `github-pages` environment. One deploy runs at a time, in the `pages-deploy` concurrency group. Needs `pages: write` and `id-token: write`. GitHub Free serves Pages only from a public repository, so the generated `pages` job runs only there; on a paid plan, drop the `github.event.repository.private == false` check from its `if`.

| Input | Default | Purpose |
| :--- | :--- | :--- |
| `runs_on` | `'ubuntu-latest'` | Runner label |
| `artifact_name` | `'dist'` | The build artifact to deploy, as `website.yml`'s |

## docker.yml

*Builds a Docker image for any number of platforms and publishes it to GitHub Container Registry.*

`forgego init --docker` adds it to your `.index.yml` as its own `docker` job after the checks, so projects without Docker don't carry it. It builds on pull requests and pushes on a release.

1. **`build`** — Without `push`: one job builds every platform without pushing, so a broken Dockerfile fails the pull request's checks, and `merge` waits for it. The generated Dockerfiles cross-compile (backend) or build the site (website) on the build host, and their runtime stages run nothing, so no platform needs emulation or a runner of its own. Builds that aren't pushed use the GitHub Actions cache. Needs `contents: read`.
2. **`push`** — With `push`: builds every platform again without a cache, which another repository's job could have planted layers in, and pushes one multi-platform image tagged `X.Y.Z`, `X.Y`, `X` (not for `0.x` versions), `sha-<commit>` and `latest`, with provenance and SBOM attestations. Needs `packages: write`.

The build passes the release version as the `VERSION` build argument (`dev` when there's none), and any `build_args` after it.

A Dockerfile that compiles in each platform's own image, such as a cgo build against a platform's C libraries, needs the other platforms emulated: set `qemu: true`. Emulated builds are many times slower, so the stages' timeout rises from 30 to 90 minutes. With `buildkit_endpoint` set, QEMU isn't set up, as the remote BuildKit runs the build.

`tags` replaces the default tags with your own [metadata-action rules](https://github.com/docker/metadata-action#tags-input), one per line, for example:

```yaml
      tags: |
        type=semver,pattern={{version}},value=${{ needs.version.outputs.tag_name }}
        type=raw,value=latest
        type=sha,format=long,prefix=
```

With `buildkit_endpoint` set, the stages build on that remote BuildKit instead of the runner's own Docker, so self-hosted runners without a Docker daemon can build and push multi-platform images. The registry credentials are written for buildx directly, as there's no `docker login` without a daemon.

| Input | Default | Purpose |
| :--- | :--- | :--- |
| `push` | `false` | Build and push in the `push` stage; the generated job sets it for a new release |
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

## patch.yml

*Keeps the files Forge.go manages current.*

1. **`update`** — On `main`, runs `task sync`, which catches up after a Dependabot update of Forge.go merges. If it changed anything, it commits the result as `chore(deps): refresh the files forgego manages` to the `forgego/sync` branch (`forgego/sync-<working_directory>` for a module in a subdirectory), force-pushes that branch, and opens a pull request for it unless one is open already. It never pushes to `main`, and the checkout keeps no credentials: the token is handed to the push alone. Needs `contents: write` and `pull-requests: write`.

The generated `.index.yml` runs it weekly, on a manual run, and on the `auto-merged` dispatch `merge.yml` sends after a Dependabot merge, so a Forge.go update's refreshed files are proposed straight away and `task sync-check` passes again once that pull request merges. Its pull request is opened with `GITHUB_TOKEN`, so its checks don't start on their own: close and reopen it to run them.

Earlier versions also upgraded the modules govulncheck found vulnerable and committed them to `main`. [Dependabot security updates]({{< relref "_index.md#vulnerable-dependencies" >}}) do that now, as pull requests.

| Input | Default | Purpose |
| :--- | :--- | :--- |
| `runs_on` | `'ubuntu-latest'` | Runner label |
| `working_directory` | `'.'` | Directory containing `go.mod` |
| `go_version` | `''` | Go version; empty reads the project's `go.mod`, else Forge.go's default |
| `apt_packages` | `''` | apt packages to install before compiling |

## merge.yml

*Dependabot's pull requests.*

1. **`review`** — On a pull request Dependabot opened, requests a review from `reviewers`, or from the repository's owner when `reviewers` is empty and the owner is a user (an organisation names its reviewers through the input), whether or not the checks passed, so a failing update reaches you too. `CODEOWNERS` requests none in a private repository on GitHub Free. A failed request warns and never fails the pipeline. Needs `pull-requests: write`.
2. **`squash`** — When `checks_passed`, and Dependabot both opened the pull request, from a branch of this repository, and made its latest change (someone else's push to its branch waits for a person), reads the update's type with [fetch-metadata](https://github.com/dependabot/fetch-metadata). A minor or patch update gets GitHub's auto-merge, turned on through the API, so the pull request squash-merges once the required checks pass; one with no checks left to wait for is merged straight away. Either way it merges only the commit the checks ran on, so a later push waits for its own checks. A major update, or one fetch-metadata can't classify, is left with a notice for a person to merge. It needs no `gh` CLI on the runner. Once the pull request has merged, it sends the `auto-merged` `repository_dispatch` event, so `.index.yml` runs on `main`: a merge made with `GITHUB_TOKEN` starts no push run. When it turned on auto-merge rather than merging, it waits up to 12 minutes for the merge before sending it. See [Secrets]({{< relref "_index.md#secrets" >}}). Needs `contents: write` and `pull-requests: write`.

Dependabot's GitHub Actions updates (branches starting `dependabot/github_actions/`) are never squashed and are left for a person to merge: they change workflow files, which the workflow's `GITHUB_TOKEN` can't merge.

| Input | Default | Purpose |
| :--- | :--- | :--- |
| `runs_on` | `'ubuntu-latest'` | Runner label |
| `reviewers` | `''` | Comma-separated logins to request a review from; empty means the repository's owner |
| `checks_passed` | `false` | Whether the pull request's checks passed in this run, e.g. `${{ !contains(needs.*.result, 'failure') && !contains(needs.*.result, 'cancelled') }}` from a job that needs them all |
