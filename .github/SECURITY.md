# Security policy

## Supported versions

Only the latest release of Forge.go receives security fixes.

## Reporting a vulnerability

Report security vulnerabilities privately through [GitHub's private vulnerability reporting](https://github.com/Apollogeddon/forgego/security/advisories/new). Don't open a public issue.

You can expect an initial response within a few days. If the issue is confirmed, the fix is released as a patch version and credited in the advisory unless you ask otherwise.

## Automated security tooling

This repository's CI runs on every pull request and again before each release:

- **Gitleaks** scans for committed secrets.
- **govulncheck** reports known vulnerabilities in the Go code Forge.go calls, and fails a pull request that has one.
- **Dependency review** fails a pull request that adds a dependency with a high or critical advisory, while the repository is public.
- **OpenSSF Scorecard** scores the repository's security practices on every push to `main` and weekly, while it's public, and reports what it finds as code scanning alerts.
- **OSV-Scanner**, run in Google's own image, whose tag Dependabot keeps current, reports known vulnerabilities in the Go modules, including the pinned tools under `tools/` and the documentation site. A high or critical one, or a scan that couldn't finish, fails a pull request.

**Dependabot security updates** propose an upgrade for each vulnerable module as soon as an advisory is published.

**Dependabot** also proposes updates to Go modules, the pinned tools, GitHub Actions and OSV-Scanner's image weekly, with a 3-day cooldown after a version is published. The cooldown gives time for a compromised release to be caught and yanked upstream before it is proposed here.
