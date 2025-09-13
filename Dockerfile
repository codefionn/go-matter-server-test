# syntax=docker/dockerfile:1.4

# Builder base with Go toolchain
FROM --platform=${BUILDPLATFORM} docker.io/golang:1.24-alpine AS builder
RUN apk add --no-cache git make gcc musl-dev
WORKDIR /src
ENV CGO_ENABLED=1
COPY ./go.* ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Format check stage
FROM --platform=${BUILDPLATFORM} builder AS format-check
COPY . .
RUN find . -name "*.go" -type f -not -path "./vendor/*" | xargs gofmt -d -s -l | tee /tmp/gofmt.out && \
    test ! -s /tmp/gofmt.out

# Test stage
FROM --platform=${BUILDPLATFORM} builder AS test
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 go test -v ./... -coverprofile=coverage.out -covermode=count

# Unit test stage
FROM --platform=${BUILDPLATFORM} builder AS unit-test
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 go test -v $(go list ./... | grep -v /e2e) -coverprofile=coverage.out -covermode=count

# Integration test stage
FROM --platform=${BUILDPLATFORM} builder AS integration-test
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 go test -v -run TestE2E ./...

# Lint stage
FROM --platform=${BUILDPLATFORM} builder AS lint
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
COPY . .
RUN golangci-lint run

# Nix build stage (if available)
FROM --platform=${BUILDPLATFORM} nixos/nix:latest AS nix-build
WORKDIR /src
COPY . .
RUN nix --extra-experimental-features "nix-command flakes" build

# Development build stage
FROM --platform=${BUILDPLATFORM} builder AS build-dev
ARG TARGETOS
ARG TARGETARCH
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /bin/go-matter-server ./cmd/matter-server

# Release build stage
FROM --platform=${BUILDPLATFORM} builder AS build-release
ARG TARGETOS
ARG TARGETARCH
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/go-matter-server ./cmd/matter-server

# Runtime development stage
FROM docker.io/alpine:latest AS runtime-dev
RUN apk add --no-cache ca-certificates tzdata dbus
WORKDIR /app
COPY --from=build-dev /bin/go-matter-server /usr/local/bin/go-matter-server
EXPOSE 8080
USER 1000:1000
ENTRYPOINT ["/usr/local/bin/go-matter-server"]

# Runtime release stage
FROM docker.io/alpine:latest AS runtime-release
RUN apk add --no-cache ca-certificates tzdata dbus
WORKDIR /app
COPY --from=build-release /bin/go-matter-server /usr/local/bin/go-matter-server
EXPOSE 8080
USER 1000:1000
ENTRYPOINT ["/usr/local/bin/go-matter-server"]

# Default to release runtime
FROM runtime-release