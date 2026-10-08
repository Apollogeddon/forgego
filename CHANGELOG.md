# Changelog

## [1.0.1](https://github.com/Apollogeddon/forgego/compare/v1.0.0...v1.0.1) (2026-10-08)


### Bug Fixes

* release a backend as a draft, so GoReleaser can attach its binaries ([9adffad](https://github.com/Apollogeddon/forgego/commit/9adffade9968157dbff4e8c0c88c1c6e5eff78b4))
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
