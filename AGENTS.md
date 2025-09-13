# Repository Guidelines

## Project Structure & Module Organization
- `cmd/matter-server/`: CLI entrypoint and main application.
- `internal/`: Go packages (config, logger, mdns, models, server, storage, websocket).
- `config.example.yaml`: Sample configuration.
- `e2e_test.go`: End‑to‑end and API tests.
- `.env.example`: Example environment file. See `ENV_VARIABLES.md` for details.

## Build, Test, and Development Commands
- `make build`: Build the binary to `./go-matter-server`.
- `make run`: Print CLI help (use flags to run), e.g. `go run ./cmd/matter-server --port 5580`.
- `make server-dev`: Run server with debug logging.
- `make test` / `make test-coverage`: Run tests; generate `coverage.html`.
- `make format` / `make lint`: Apply `gofmt`/`goimports` and run `golangci-lint`.
- Nix users: `make nix-dev` (dev shell), `make nix-build`, `make nix-run`. See `NIX_USAGE.md`.

## Coding Style & Naming Conventions
- Go style: tabs for indentation; exported identifiers use `CamelCase`; files use `snake_case.go`.
- Format before commit: `make format` (runs `go fmt` and `goimports`).
- Lint locally: `make lint` (requires `golangci-lint` or use Nix dev shell).
- Packages are lower‑case, short, and without underscores (e.g., `server`, `mdns`).

## Testing Guidelines
- Framework: standard Go `testing` package.
- Test files: `*_test.go`; function names `TestXxx`. E2E lives in `e2e_test.go`.
- Run: `make test` (unit + integration) or `make test-integration` for E2E focus.
- Coverage: `make test-coverage` to create `coverage.html`.

## Commit & Pull Request Guidelines
- Commits: prefer Conventional Commits (e.g., `feat(server): add diagnostics API`, `fix(mdns): handle IPv6`) with focused diffs.
- PRs must include: clear description, rationale, test coverage, and any config docs updates (`ENV_VARIABLES.md`, `README.md`).
- Before opening a PR: `make format lint test` must pass.

## Security & Configuration Tips
- Default ports: WebSocket `5580`, mDNS `5353`. Avoid conflicts (e.g., disable `avahi-daemon` during development).
- Configuration precedence: flags > env vars > config file > defaults. See `ENV_VARIABLES.md`.
- Local dev: `.env` is supported; example in `.env.example`. Storage defaults to `$HOME/.matter_server` unless overridden.
