# Changelog

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
