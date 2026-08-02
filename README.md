<div align="center">

# kip

**Stop sending `.env` files over Slack.**

Share encrypted secrets with self-destructing links. One command. No accounts. Zero-knowledge.

[![CI](https://github.com/kipenv/kip/actions/workflows/ci.yml/badge.svg)](https://github.com/kipenv/kip/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Self-Hosted](https://img.shields.io/badge/Self--Hosted-Docker-2496ED?logo=docker&logoColor=white)](#self-hosting)

[Install](#install) · [How It Works](#how-it-works) · [Self-Host](#self-hosting) · [Security](#security) · [Roadmap](#roadmap)

</div>

---

## The Problem

Every dev team has done this:

```
👤 hey can you send me the .env for staging?
👤 sure, one sec

# .env.staging
DATABASE_URL=postgresql://admin:s3cret@db.internal:5432/app
STRIPE_SECRET_KEY=sk_live_51N8x...
AWS_SECRET_ACCESS_KEY=wJal...
```

Now those credentials live forever in your Slack history, searchable by anyone
with access to the workspace.

## The Solution

```console
$ kip push .env.staging

Secret shared successfully!
Link: https://kip.example.com/s/h4l1yddyftd8#8tvB2j3oq38kATqid2fjQkakgECPbmcTTuAyGW3yXg3W
Expires: 2026-08-02 16:32 (in 1h) | Reads: 1

Share this link. The key never leaves your machine.
```

The receiver opens the link in their browser — no CLI needed. The secret is
decrypted locally using the key in the `#fragment`, which is never sent to the
server. Once the read limit is reached, the ciphertext is deleted.

> **There is no public instance yet.** kip is self-hosted: point the CLI at your
> own server (see [Self-Hosting](#self-hosting)) or run it locally.

---

## Install

### Go

```bash
go install github.com/kipenv/kip/cmd/kip@latest
```

### From source

```bash
git clone https://github.com/kipenv/kip.git
cd kip
make build        # → bin/kip
```

Homebrew and a `curl | sh` installer are on the [roadmap](#roadmap); they need a
tagged release first.

---

## Quick Start

```bash
# Point the CLI at your server (once)
kip config set --server https://kip.example.com

# Share a secret — expires in 1 hour, one read
kip push .env

# Custom lifetime and read count (--ttl is in seconds)
kip push .env --ttl 86400 --reads 3

# Add a password on top of the link
kip push .env --password "correct horse battery staple"

# Receive it (CLI)
kip pull https://kip.example.com/s/h4l1yddyftd8#8tvB2j3o...

# Receive it (browser) — just open the link, nothing to install

# Print to stdout instead of writing a file
kip pull <link> --stdout

# Kill a link before anyone reads it
kip revoke <link-or-id>
```

---

## How It Works

### 1. Encrypt locally

```
Your machine                          Server
    |                                    |
    |  AES-256-GCM encrypt               |
    |  key = random 32 bytes             |
    |  nonce = random 12 bytes           |
    |                                    |
    |  POST {ciphertext, nonce}     →    |  Store in Redis with TTL
    |                                    |  Return ID: "h4l1yddyftd8"
    |                                    |
    |  Build URL:                        |
    |  <server>/s/h4l1yddyftd8#<key>     |
```

### 2. Share the link

The decryption key lives in the URL `#fragment`. Browsers do not send fragments
to servers, and neither does the CLI. Even if the server is fully compromised,
an attacker holds ciphertext and nothing else.

### 3. Self-destruct

Each read decrements a counter; at zero the server deletes the record. Redis TTL
guarantees deletion even if nobody ever reads it.

---

## Security

kip is built on a **zero-knowledge** architecture. The server never sees your
secrets.

| Property | Implementation |
|---|---|
| **Encryption** | AES-256-GCM (authenticated encryption) |
| **Key derivation** | Argon2id (for password-protected secrets) |
| **Key transport** | URL `#fragment` — never sent over HTTP |
| **Forward secrecy** | Unique random key per share |
| **Tampering detection** | GCM authentication tag |
| **Time-bound** | Redis TTL auto-deletes |
| **Read-bound** | Server deletes after N reads |
| **At rest locally** | Pulled files and CLI config are written `0600` |

### Threat Model

| Threat | Mitigation |
|---|---|
| Server database breach | Only encrypted blobs stored — the key is in the URL fragment |
| Man-in-the-middle | HTTPS required; the fragment never crosses the network |
| Brute force link IDs | IDs are 12-char nanoid, plus per-IP rate limiting |
| Replay attacks | Read counter + TTL — links die after use |
| Malicious server operator | Cannot decrypt; can only delete (acceptable) |
| Browser history / referrer leak | The fragment lands in browser history; the decrypt page loads zero third-party scripts |

For full details, see [SECURITY.md](SECURITY.md).

---

## Features

| Feature | Description |
|---|---|
| **Zero-knowledge encryption** | AES-256-GCM — the server never receives the key |
| **Self-destructing links** | TTL + read limit, configurable per share |
| **Password protection** | Optional Argon2id layer on top of the link |
| **Browser decrypt** | Web Crypto API in the recipient's browser — no CLI required |
| **Revocation** | Kill a link before it is read |
| **Secret scanning** | Regex patterns for AWS, Stripe, GitHub tokens and weak secrets |
| **`.env.example` generation** | Strip values, keep keys — safe to commit |
| **Self-hostable** | Docker Compose, single binary, MIT |

---

## Secret Scanning

Catch dangerous values before they leave your machine. Runs entirely offline —
no network, no API keys.

```console
$ kip scan .env

  OK    DATABASE_URL — no issues detected
  WARN  STRIPE_SECRET_KEY — Stripe Live Key — looks like a Stripe live secret key

✗ 1 warning(s) found in .env
```

## Generate `.env.example`

```console
$ kip generate .env

# Generated by kip generate
# Fill in the values for your environment

DATABASE_URL=<url>
STRIPE_SECRET_KEY=<secret>
```

Write it straight to a file with `-o .env.example`.

---

## Teams

Teams work like game lobbies: create one, share the invite code, and members
join. No sign-up, no roles, no admin panel.

```bash
# Create a team — you pick your own display name
kip team create my-startup --username antonio
# → Invite code: my-startup-a8f3k2

# A teammate joins with the code
kip team join my-startup-a8f3k2 --username maria

kip team ls                       # teams you belong to
kip team members my-startup       # who else is in
kip team leave my-startup         # last one out deletes the team

# Associate a directory with a team (writes .kip)
kip init my-startup
```

**Status:** the API implements team-scoped sharing — send to a whole team or one
member, an inbox, an audit log, and a pinned "official" env per project — and it
is covered by tests. The CLI does not expose those endpoints yet, so today teams
are membership only. Wiring them up is the top [roadmap](#roadmap) item.

---

## Self-Hosting

kip is **free to self-host forever**. MIT. Your data, your servers.

Two Compose files live under `deploy/`:

```bash
# Local / kicking the tyres — publishes :8080 on the host
docker compose -f deploy/docker-compose.yml up -d
kip config set --server http://localhost:8080
```

For a real server use `deploy/docker-compose.prod.yml`: it adds the static web
(landing + browser decrypt page, served by nginx), keeps Redis and the API off
the host network, and persists team state on a `/data` volume. It is written for
a Traefik-style reverse proxy with a **single-domain layout** — `/api` and
`/health` route to the Go server, everything else to the web — so that the links
`kip push` prints open in a browser and the decrypt page talks to the API
same-origin, with no CORS. Routing details are in the header of that file.

```bash
kip config set --server https://kip.example.com
```

No per-seat pricing, no license keys. SQLite for teams (embedded — no extra
service), Redis for secrets with native TTL.

Configuration reference: [deploy/ENV.md](deploy/ENV.md).

---

## Comparison

| | kip | Slack DM | Yopass | 1Password | Doppler | Vault |
|---|:---:|:---:|:---:|:---:|:---:|:---:|
| Native CLI | **Yes** | No | No | Yes | Yes | Yes |
| Zero-knowledge | **Yes** | No | Yes | No | No | No |
| Self-destructing | **Yes** | No | Yes | No | No | No |
| No account for push/pull | **Yes** | — | Yes | No | No | No |
| Self-hostable | **Yes** | No | Yes | No | No | Yes |
| `.env` focused | **Yes** | No | No | No | Yes | No |
| Free | **Yes** | — | Yes | No | No | Yes |

kip sits between "paste it in Slack" and "deploy HashiCorp Vault" — for teams
that want the secret to disappear without standing up infrastructure.

---

## CLI Reference

```
kip push <file>               Encrypt and share via link
  -t, --ttl <seconds>              lifetime, default 3600
  -r, --reads <n>                  reads before deletion, default 1
  -p, --password <password>        add a password layer

kip pull <link>               Download and decrypt
  -o, --output <path>              save under a different name
  -p, --password <password>        password for protected links
      --stdout                     print instead of writing a file

kip revoke <link-or-id>       Delete a secret so it can no longer be read
kip scan [file]               Scan for leaked credentials and weak secrets
kip generate [file]           Generate .env.example (strip values)
  -o, --output <path>              write to a file instead of stdout

kip team create <name>        Create a team, get an invite code
  -u, --username <name>            your display name (required)
kip team join <invite-code>   Join a team
  -u, --username <name>            your display name (required)
kip team ls                   List your teams
kip team members <team>       List members of a team
kip team leave <team>         Leave a team

kip init <team-name>          Associate this directory with a team (.kip)
      --git-exclude                use .git/info/exclude instead of .gitignore

kip config set --server <url> Point the CLI at a kip server
kip config get                Show current configuration

kip --version                 Print version and build commit
```

---

## Architecture

```
┌─────────────┐         ┌──────────────────┐
│   CLI (Go)  │──push──▶│  Server (Go)     │
│             │◀─pull───│  net/http stdlib │
└─────────────┘         │                  │
                        │  ┌─────────┐     │
┌─────────────┐         │  │  Redis  │ TTL │
│  Browser    │──GET───▶│  │ secrets │     │
│  Web Crypto │◀────────│  └─────────┘     │
│  (decrypt)  │         │  ┌─────────┐     │
└─────────────┘         │  │ SQLite  │     │
                        │  │  teams  │     │
                        │  └─────────┘     │
                        └──────────────────┘
```

- **CLI:** Go + Cobra. Single binary, no runtime dependencies.
- **Server:** Go stdlib `net/http` — no web framework. Redis for ephemeral
  secrets (native TTL means expiry is free), SQLite for teams and members.
- **Web:** Astro + Vue. Static landing page; the decrypt page runs AES-256-GCM
  in the browser via `SubtleCrypto`.
- **Crypto:** `crypto/aes` + `crypto/cipher` (Go stdlib) / `SubtleCrypto`
  (browser). Argon2id for password-derived keys, base58 for keys in URLs.

---

## Roadmap

Honest list of what is not there yet:

- **Team sharing in the CLI** — `push --all` / `--to` / `--pin`, `inbox`, `ls`
  and `diff` against a team's pinned env. The API and its tests already exist;
  the commands do not.
- **Tagged releases** — cross-compiled binaries via GoReleaser, then a Homebrew
  formula and a `curl | sh` installer.
- **Optional LLM-assisted scanning** — bring your own endpoint (Ollama, or any
  OpenAI-compatible API) to complement the regex pass.
- **A public instance.** For now: self-host or run locally.

---

## Development

```bash
# Prerequisites: Go 1.26+, Node 22+, Redis (or run without it — the server
# falls back to an in-memory store), Make

make build      # build the CLI to bin/kip
make server     # build and run the server
make test       # go test -race -cover ./...
make lint       # golangci-lint
make dev        # server + redis via docker compose

cd web && npm install && npm run dev   # landing + decrypt page
```

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[MIT](LICENSE) — free forever, for everyone.

---

<div align="center">

Built by [Antonio Vila](https://antoniovila.dev)

</div>
