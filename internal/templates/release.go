package templates

// Goreleaser builds a backend's release binaries. release-please creates the release
// and writes its notes; GoReleaser only adds the binaries and checksums to it.
const Goreleaser = `# Builds and attaches the release binaries. Try it locally with: task release:snapshot
version: 2
project_name: __NAME__

builds:
  - id: __NAME__
    main: ./cmd/__NAME__
    binary: __NAME__
    env:
      - CGO_ENABLED=0
    flags:
      - -trimpath
    ldflags:
      - -s -w -X main.version={{ .Version }}
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]

archives:
  - formats: [tar.gz]
    format_overrides:
      - goos: windows
        formats: [zip]

checksum:
  name_template: checksums.txt

# release-please writes the release notes and creates the release
changelog:
  disable: true
release:
  mode: keep-existing
__NFPMS__`

// GoreleaserNfpms adds the .deb package for --debian.
const GoreleaserNfpms = `
nfpms:
  - package_name: __NAME__
    formats: [deb]
    maintainer: __NAME__ maintainers
    description: __NAME__
    license: MIT
    bindir: /usr/bin
    contents:
      - src: packaging/__NAME__.service
        dst: /lib/systemd/system/__NAME__.service
    scripts:
      postinstall: packaging/postinstall.sh
      preremove: packaging/preremove.sh
`

// SystemdUnit runs the service as its own system user, never root.
const SystemdUnit = `[Unit]
Description=__NAME__
After=network.target

[Service]
Type=simple
ExecStart=/usr/bin/__NAME__
Restart=on-failure
User=__NAME__
Group=__NAME__

[Install]
WantedBy=multi-user.target
`

// Postinstall creates the service user, with no home or login shell, then starts it.
const Postinstall = `#!/bin/sh
set -e
id -u __NAME__ >/dev/null 2>&1 || useradd --system --no-create-home --shell /usr/sbin/nologin --user-group __NAME__
systemctl daemon-reload
systemctl enable __NAME__.service
systemctl restart __NAME__.service
`

// Preremove stops the service before its files go.
const Preremove = `#!/bin/sh
set -e
systemctl disable --now __NAME__.service || true
`

// GoreleaserVersion is the GoReleaser the snapshot task runs. CI pins ~> v2 through
// goreleaser-action.
const GoreleaserVersion = "v2.18.2"

// DockerfileBackend cross-compiles on the build host for each target platform, so a
// multi-platform image needs no emulation, and runs on distroless as a non-root user.
const DockerfileBackend = `# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:__GO_MINOR__ AS build
ARG TARGETOS TARGETARCH VERSION=dev
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags "-s -w -X main.version=$VERSION" -o /out/__NAME__ ./cmd/__NAME__

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/__NAME__ /__NAME__
USER nonroot:nonroot
ENTRYPOINT ["/__NAME__"]
`

// DockerfileWebsite builds the static site once on the build host, then serves it with nginx.
const DockerfileWebsite = `# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:__GO_MINOR__ AS build
WORKDIR /src
COPY . .
RUN go tool -modfile=.forgego/hugo.mod hugo --gc --minify

FROM nginx:stable-alpine
COPY --from=build /src/public /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
`

// Dockerignore keeps build output and git out of the build context.
const Dockerignore = `.git
.github
dist
public
resources/_gen
coverage.out
junit-report.xml
`
