# Changelog

## [2.1.0](https://github.com/Apollogeddon/forgego/compare/v2.0.2...v2.1.0) (2026-10-23)


### Features

* **ci:** make the pipelines work on free private and public repositories, and harden them ([#59](https://github.com/Apollogeddon/forgego/issues/59)) ([1cb87ab](https://github.com/Apollogeddon/forgego/commit/1cb87abd37427ea00cbd6bc4113b726b19752ce0))


### Bug Fixes

* **ci:** scan (osv) on github's ubuntu-latest in private repositories, whose runs_on may not run it ([#60](https://github.com/Apollogeddon/forgego/issues/60)) ([628ccfb](https://github.com/Apollogeddon/forgego/commit/628ccfb3b3ddfca3db95221c67a0efbcd1972286))
* **deps:** bump golang.org/x/text ([#57](https://github.com/Apollogeddon/forgego/issues/57)) ([66d6f6d](https://github.com/Apollogeddon/forgego/commit/66d6f6d7657d2cdd0a4841b4c7b6bafb591cedba))

## [2.0.2](https://github.com/Apollogeddon/forgego/compare/v2.0.1...v2.0.2) (2026-10-16)


### Bug Fixes

* drop sync's migrations for v1.0 tool-pin paths and go run FORGEGO Taskfile vars ([#54](https://github.com/Apollogeddon/forgego/issues/54)) ([6504184](https://github.com/Apollogeddon/forgego/commit/65041845e0a5a98b7819083f4b6e265b451a8933))

## [2.0.1](https://github.com/Apollogeddon/forgego/compare/v2.0.0...v2.0.1) (2026-10-16)


### Bug Fixes

* use the /v2 module path so forgego v2 can be installed and pinned ([#52](https://github.com/Apollogeddon/forgego/issues/52)) ([99eb491](https://github.com/Apollogeddon/forgego/commit/99eb49107e918cf5183960a2a7c1c1809a539d55))

## [2.0.0](https://github.com/Apollogeddon/forgego/compare/v1.6.1...v2.0.0) (2026-10-16)


### ⚠ BREAKING CHANGES

* the old pipelines are gone; call the split pipelines from a root .index.yml.

### Features

* **ci:** run main's pipeline after a dependabot auto-merge, through a repository_dispatch ([#47](https://github.com/Apollogeddon/forgego/issues/47)) ([cb9376b](https://github.com/Apollogeddon/forgego/commit/cb9376b2aae4f8f93829599bee877a3aa6e15215))
* split the pipelines by role, with a root .index.yml and two-level job names ([#51](https://github.com/Apollogeddon/forgego/issues/51)) ([8445804](https://github.com/Apollogeddon/forgego/commit/84458040a72506d291d683f616b6c7670ded0d1a))


### Bug Fixes

* **deps:** bump golang.org/x/mod ([#46](https://github.com/Apollogeddon/forgego/issues/46)) ([d6c6c81](https://github.com/Apollogeddon/forgego/commit/d6c6c816fcac4cda3a50c0237e46733284a0c4e9))
* make each subdirectory a valid module path element and refuse a path that isn't valid ([#50](https://github.com/Apollogeddon/forgego/issues/50)) ([774868b](https://github.com/Apollogeddon/forgego/commit/774868b7c52c817a7c43fcb9f26a87c79432c078))
* review 4 findings (worktree module detection, GoReleaser pin, OSV gate, auto-merge author) ([#48](https://github.com/Apollogeddon/forgego/issues/48)) ([6667cbc](https://github.com/Apollogeddon/forgego/commit/6667cbc22f78534049759e8c5763035d1cfef802))

## [1.6.1](https://github.com/Apollogeddon/forgego/compare/v1.6.0...v1.6.1) (2026-10-11)


### Bug Fixes

* **website:** ship Hextra's search script so sites build offline, and move to download-artifact v8 ([720c550](https://github.com/Apollogeddon/forgego/commit/720c550f569f1d99c0be5f701faec4c3fc0dd09b))

## [1.6.0](https://github.com/Apollogeddon/forgego/compare/v1.5.0...v1.6.0) (2026-10-11)


### Features

* check a website's links and accessibility before it deploys ([#43](https://github.com/Apollogeddon/forgego/issues/43)) ([48ad78b](https://github.com/Apollogeddon/forgego/commit/48ad78bc27961d46c32e8fe920bf4cf76422f449))
* run tests and native builds on more than one operating system ([1461af3](https://github.com/Apollogeddon/forgego/commit/1461af332de56a244cfe7618c1ed570723bd6b87))
* run tests and native builds on more than one operating system ([a6fd15a](https://github.com/Apollogeddon/forgego/commit/a6fd15aed679de5ad1db8b4c2b1fdc7d0b18e08e))
* sign release checksums with GPG and let callers run the checks themselves ([#42](https://github.com/Apollogeddon/forgego/issues/42)) ([4967974](https://github.com/Apollogeddon/forgego/commit/496797477a42fd7c5f57b49ce19d096ae01fcde6))


### Bug Fixes

* leave the extra and native-build jobs' names to GitHub, which shows a skipped one plainly ([41cd292](https://github.com/Apollogeddon/forgego/commit/41cd292437c1c0ee1d42b116a69dac30f980cb75))
* leave the extra and native-build jobs' names to GitHub, which shows a skipped one plainly ([7358ccc](https://github.com/Apollogeddon/forgego/commit/7358ccc59bf0bf2900598f74448b571807efbee8))
* tidy forgego's pin in its own directory ([c8184dc](https://github.com/Apollogeddon/forgego/commit/c8184dcc43eeb0484a34318321918870d0effc35))
* tidy forgego's pin in its own directory ([10a72e2](https://github.com/Apollogeddon/forgego/commit/10a72e254666249345833ef2b82fea2c3026cad2))

## [1.5.0](https://github.com/Apollogeddon/forgego/compare/v1.4.0...v1.5.0) (2026-10-10)


### Features

* pin forgego in .forgego/forgego/go.mod so Dependabot proposes each release ([#36](https://github.com/Apollogeddon/forgego/issues/36)) ([10cf252](https://github.com/Apollogeddon/forgego/commit/10cf2524e567ba23128a734c6ed129b10c0111e2))

## [1.4.0](https://github.com/Apollogeddon/forgego/compare/v1.3.0...v1.4.0) (2026-10-10)


### Features

* **ci:** fall back to Go 1.27 without a go.mod, and leave the version to go.mod ([28f830e](https://github.com/Apollogeddon/forgego/commit/28f830e5a6b9f6033deb3417484face998d29261))
* **ci:** fall back to Go 1.27 without a go.mod, and leave the version to go.mod ([4477ef9](https://github.com/Apollogeddon/forgego/commit/4477ef992e29e52578bbc984d7ba28e1d1ecf3eb))

## [1.3.0](https://github.com/Apollogeddon/forgego/compare/v1.2.0...v1.3.0) (2026-10-10)


### Features

* **ci:** request a review on Dependabot's and release-please's pull requests ([9ea4986](https://github.com/Apollogeddon/forgego/commit/9ea498640b316b0beadbc3f657511c92edcb9c2a))
* **ci:** request a review on Dependabot's and release-please's pull requests ([ceda08b](https://github.com/Apollogeddon/forgego/commit/ceda08bcddf0d433c80d770fb95c5111a316cb61))

## [1.2.0](https://github.com/Apollogeddon/forgego/compare/v1.1.1...v1.2.0) (2026-10-10)


### Features

* write repo files on init, and merge Dependabot PRs where auto-merge isn't offered ([#29](https://github.com/Apollogeddon/forgego/issues/29)) ([fd8f756](https://github.com/Apollogeddon/forgego/commit/fd8f756f1b422687508cddac301c76c192707aa5))

## [1.1.1](https://github.com/Apollogeddon/forgego/compare/v1.1.0...v1.1.1) (2026-10-10)


### Bug Fixes

* **ci:** merge Dependabot PRs to branches without protection rules ([53b1183](https://github.com/Apollogeddon/forgego/commit/53b11835ec4b15a7b0666b617d04a1f57dfe2020))
* **ci:** merge Dependabot PRs to branches without protection rules ([840fe2e](https://github.com/Apollogeddon/forgego/commit/840fe2ee63a754376dcb03c8fa9042783b452680))

## [1.1.0](https://github.com/Apollogeddon/forgego/compare/v1.0.6...v1.1.0) (2026-10-09)


### Features

* **ci:** add a system_packages input for cgo projects ([7aae37e](https://github.com/Apollogeddon/forgego/commit/7aae37e2d0a52c056a2e3cb3eddc702c7a950eb2))
* **ci:** add a system_packages input for cgo projects ([2e0ca19](https://github.com/Apollogeddon/forgego/commit/2e0ca19d5489e67b854916da623a5ee94e55ce14))

## [1.0.6](https://github.com/Apollogeddon/forgego/compare/v1.0.5...v1.0.6) (2026-10-09)


### Bug Fixes

* **ci:** keep counting releases_created for a root package, where release_created can be unset ([68ea434](https://github.com/Apollogeddon/forgego/commit/68ea434b243c14e1b81769475fc1f0acd923f088))
* **ci:** use working_directory in version.yml ([99cb192](https://github.com/Apollogeddon/forgego/commit/99cb19261b5969e66ccf36751a97b872b889915f))

## [1.0.5](https://github.com/Apollogeddon/forgego/compare/v1.0.4...v1.0.5) (2026-10-08)


### Bug Fixes

* **ci:** leave Dependabot's GitHub Actions updates for a person to merge ([e1ac352](https://github.com/Apollogeddon/forgego/commit/e1ac352bc01943019d1e413f88cf51e2b0ceb33f))

## [1.0.4](https://github.com/Apollogeddon/forgego/compare/v1.0.3...v1.0.4) (2026-10-08)


### Bug Fixes

* **deps:** upgrade gotestsum's x/mod and x/text past known vulnerabilities ([da702eb](https://github.com/Apollogeddon/forgego/commit/da702eb765990ae33fcd6cca1f2814a4281f8c19))

## [1.0.3](https://github.com/Apollogeddon/forgego/compare/v1.0.2...v1.0.3) (2026-10-08)


### Bug Fixes

* auto-merge a Dependabot PR that only waits on checks nobody requires ([4fc82b3](https://github.com/Apollogeddon/forgego/commit/4fc82b34f63ec7dcbb186403b09d9f30e5f60743))

## [1.0.2](https://github.com/Apollogeddon/forgego/compare/v1.0.1...v1.0.2) (2026-10-08)


### Bug Fixes

* auto-merge Dependabot PRs through the API, so a runner needs no gh CLI ([d68b82b](https://github.com/Apollogeddon/forgego/commit/d68b82b975a632cdf21115274163dd045680f58c))

## [1.0.1](https://github.com/Apollogeddon/forgego/compare/v1.0.0...v1.0.1) (2026-10-08)


### Bug Fixes

* release a backend as a draft, so GoReleaser can attach its binaries ([862a0fd](https://github.com/Apollogeddon/forgego/commit/862a0fd202f74ed89e2dc87054c7ed5e4d81fb7e))

## 1.0.0 (2026-10-08)


### Features

* a type task, and run_tests and auto_patch inputs on website.yml ([25705a7](https://github.com/Apollogeddon/forgego/commit/25705a7c1771c56b2efba42c32d038cf353b30f9))
* give a project on an older Go a toolchain its tools can run on ([61af1b9](https://github.com/Apollogeddon/forgego/commit/61af1b9000a3d983ff7545d5ee8012469d418dba))
* name a module inside a repository by its subdirectory, and take a global -C ([50e4d89](https://github.com/Apollogeddon/forgego/commit/50e4d89c8cd74ce55678ebf7eec97fc22ffaa718))
* reusable workflows for services, libraries, websites and images ([e1ece77](https://github.com/Apollogeddon/forgego/commit/e1ece77119240e69278d844d92f2ea943cfaffd2))
* scaffold backend, library and website projects with forgego init ([81c8f1d](https://github.com/Apollogeddon/forgego/commit/81c8f1d5f58974a38349a9eaf9a9d322e890b562))
* scaffold Go projects with forgego init ([eadc721](https://github.com/Apollogeddon/forgego/commit/eadc7213e7afcfdefafc48490bd6b02ee90172b4))


### Bug Fixes

* address the code review of the scaffolder ([62348b2](https://github.com/Apollogeddon/forgego/commit/62348b29a1eb5362e7f07da993b5a4bdb27df894))
* install a .deb without systemd, and refuse a v2+ library release Go can't fetch ([aef076c](https://github.com/Apollogeddon/forgego/commit/aef076ceca62552e967ca94c9cbf3f8132482b05))
* keep CI working for a project scaffolded with --no-linting ([b29799f](https://github.com/Apollogeddon/forgego/commit/b29799f55928d1651a2eabbf11eae8e5cf7b382d))
* pin each tool in .forgego/&lt;tool&gt;/go.mod so secret scanners skip its go.sum ([d290f25](https://github.com/Apollogeddon/forgego/commit/d290f25078b6e6d646556a793bcd37de2a603e41))
* run the git hooks through the pinned lefthook ([d8e178f](https://github.com/Apollogeddon/forgego/commit/d8e178f95be9ab9b5d0f416d8e71278d1d92f802))
* use gitleaks' [allowlist] table, which the 8.24 in gitleaks-action reads ([bce5407](https://github.com/Apollogeddon/forgego/commit/bce54075b6c7b4ff2cadd7887015c7423dca883d))
