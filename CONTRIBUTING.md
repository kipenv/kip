# Contributing to kip

Thanks for your interest. kip is a small, focused project — contributions that keep it simple are welcome.

## Prerequisites

- **Go 1.26+**
- **Redis** (running locally or via Docker)
- **Node 22+** (only if working on `web/`)
- **Make**

## Setup

```bash
git clone https://github.com/kipenv/kip.git
cd kip

# Build the CLI
make build

# Run the server in dev mode
make server

# Run tests
make test

# Lint
make lint
```

For the web frontend:

```bash
cd web
npm install
npm run dev
```

## Project Structure

```
cmd/kip/       CLI entrypoint
internal/
  cli/              Cobra command definitions
  crypto/           AES-256-GCM encryption/decryption
  client/           HTTP client for the server API
  config/           Local config + .kip file handling
  scanner/          Security scan (regex + AI)
  envfile/          .env parsing and utilities
server/             API server (Go stdlib net/http)
web/                Astro + Vue landing page + decrypt page
deploy/             Docker, docker-compose, fly.toml
```

## Code Style

- Run `gofmt` on all Go code
- Run `golangci-lint run` before submitting — CI will catch it anyway
- Follow stdlib conventions: small interfaces, error wrapping with `%w`, context propagation
- No global state — use dependency injection

## Submitting a PR

1. Fork the repo
2. Create a branch from `main` (`git checkout -b fix/thing`)
3. Make your changes
4. Run `make test` and `make lint`
5. Open a PR against `main` with a clear description of what and why

Keep PRs small and focused. One thing per PR.

## Good First Issues

Look for issues labeled [`good first issue`](https://github.com/kipenv/kip/labels/good%20first%20issue). These are scoped tasks that don't require deep knowledge of the codebase.

## Security Issues

**Do not open public issues for security vulnerabilities.** See [SECURITY.md](SECURITY.md) for responsible disclosure.
