<div align="center">

# kip

**Stop sending `.env` files over Slack.**

Share encrypted secrets with self-destructing links. One command. No accounts. Zero-knowledge.

[![CI](https://github.com/kipenv/kip/actions/workflows/ci.yml/badge.svg)](https://github.com/kipenv/kip/actions/workflows/ci.yml)
[![Release](https://github.com/kipenv/kip/actions/workflows/release.yml/badge.svg)](https://github.com/kipenv/kip/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Self-Hosted](https://img.shields.io/badge/Self--Hosted-Docker-2496ED?logo=docker&logoColor=white)](#self-hosting)

[Website](https://kipenv.dev) · [Install](#install) · [How It Works](#how-it-works) · [Self-Host](#self-hosting) · [Security](#security)

<!-- TODO: Replace with actual GIF once recorded with VHS -->
<img src="assets/demo.gif" alt="kip demo — push, share, pull" width="600" />

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

Now those credentials live forever in your Slack history, searchable by anyone with access to the workspace.

## The Solution

```bash
$ kip push .env.staging

  Encrypted with AES-256-GCM
  Link: https://kipenv.dev/s/x8f9k2#3Kj8mN...
  Expires: 1 hour | Reads: 1

  Share this link. It self-destructs after being read.
```

The receiver opens the link in their browser — no CLI needed. The secret is decrypted locally using the key in the `#fragment` (never sent to the server). After one read, the link is permanently destroyed.

---

## Install

### Homebrew

```bash
brew install kip
```

### Go

```bash
go install github.com/kipenv/kip@latest
```

### curl

```bash
curl -fsSL https://kipenv.dev/install.sh | sh
```

### From source

```bash
git clone https://github.com/kipenv/kip.git
cd kip
make build
```

---

## Quick Start

```bash
# Share a secret (expires in 1 hour, 1 read)
kip push .env

# Custom expiration and reads
kip push .env --expires 24h --reads 3

# Password-protect
kip push .env --password

# Receive (CLI)
kip pull https://kipenv.dev/s/x8f9k2#3Kj8mN...

# Receive (browser) — just open the link, no install needed
```

---

## How It Works

### 1. Encrypt locally

```
Your machine                          Server
    |                                    |
    |  AES-256-GCM encrypt              |
    |  key = random 32 bytes            |
    |  nonce = random 12 bytes          |
    |                                    |
    |  POST {ciphertext, nonce}    →    |  Store in Redis with TTL
    |                                    |  Return ID: "x8f9k2"
    |                                    |
    |  Build URL:                        |
    |  kipenv.dev/s/x8f9k2#<key>      |
```

### 2. Share the link

The decryption key is in the URL `#fragment` — this part is **never sent to the server** by HTTP protocol. Even if the server is compromised, attackers only see encrypted blobs.

### 3. Self-destruct

After the configured number of reads (default: 1), the server **permanently deletes** the encrypted data from Redis. The TTL also guarantees deletion even if nobody reads it.

---

## Security

kip is built on a **zero-knowledge** architecture. The server never sees your secrets.

| Property | Implementation |
|---|---|
| **Encryption** | AES-256-GCM (authenticated encryption) |
| **Key derivation** | Argon2id (for password-protected secrets) |
| **Key transport** | URL `#fragment` — never sent over HTTP |
| **Forward secrecy** | Unique random key per share |
| **Tampering detection** | GCM authentication tag |
| **Time-bound** | Redis TTL auto-deletes |
| **Read-bound** | Server deletes after N reads |

### Threat Model

| Threat | Mitigation |
|---|---|
| Server database breach | Only encrypted blobs stored — key is in URL fragment |
| Man-in-the-middle | HTTPS required; key in fragment never sent over network |
| Brute force link IDs | IDs are sufficiently random (nanoid, 12+ chars) |
| Replay attacks | Read counter + TTL — links die after use |
| Malicious server operator | Can't decrypt; can only delete (acceptable) |

For full details, see [SECURITY.md](SECURITY.md).

---

## Features

| Feature | Description |
|---|---|
| **Zero-knowledge encryption** | AES-256-GCM — server never sees the key |
| **Self-destructing links** | TTL + read limit — configurable per share |
| **Password protection** | Optional Argon2id-derived password layer |
| **Browser decrypt** | Receivers don't need the CLI — Web Crypto API |
| **Teams** | Lightweight team sharing — like game lobbies, no admin panels |
| **AI security scan** | Regex + optional LLM to detect dangerous keys before sharing |
| **Self-hostable** | Docker Compose — one command, your infrastructure |
| **Generate .env.example** | Strip values, keep keys — safe to commit |

---

## Teams

Teams in kip work like game lobbies — create, join with a code, share. No sign-up, no admin panels, no roles.

```bash
# Create a team
kip team create my-startup

# Share the invite code with your teammate
# → Invite code: my-startup-a8f3k2

# They join
kip team join my-startup-a8f3k2

# Link your project directory
kip init

# Push to the whole team
kip push .env --all

# Or to a specific person
kip push .env --to juan

# Pin the "official" env for the project
kip push .env --pin

# Pull the latest pinned env (no URL needed)
kip pull

# Check what changed
kip diff
```

---

## Self-Hosting

kip is **free to self-host forever**. MIT License. Your data, your servers.

```yaml
# docker-compose.yml
services:
  kip:
    image: ghcr.io/kipenv/kip-server:latest
    ports:
      - "8080:8080"
    environment:
      - REDIS_URL=redis://redis:6379
    depends_on:
      - redis

  redis:
    image: redis:7-alpine
    volumes:
      - redis-data:/data

volumes:
  redis-data:
```

```bash
docker compose up -d
kip config set server http://localhost:8080
```

That's it. No per-seat pricing. No license keys. SQLite for teams, Redis for secrets.

---

## AI Security Scan

Catch dangerous keys before they leave your machine.

```bash
# Built-in regex patterns (always free, no network)
kip scan .env

  WARN  AWS_SECRET_ACCESS_KEY — live AWS credential detected
  WARN  STRIPE_SECRET_KEY — live Stripe key (sk_live_*)
  WARN  JWT_SECRET — weak secret (< 32 chars)
  OK    DATABASE_URL — no issues
```

### Three tiers

| Tier | How it works | Cost |
|---|---|---|
| **Regex** | Built-in patterns for AWS, Stripe, GitHub tokens, weak secrets | Free |
| **BYO AI** | Connect your own Ollama, OpenAI, Anthropic, or any compatible API | Free (you pay your provider) |
| **Included AI** | Zero-config scan on Pro plan | Pro |

```bash
# Use local Ollama
kip config set ai.provider ollama
kip config set ai.url http://localhost:11434
kip config set ai.model llama3.1:8b
```

---

## Comparison

| | kip | Slack DM | Yopass | 1Password | Doppler | Vault |
|---|:---:|:---:|:---:|:---:|:---:|:---:|
| Native CLI | **Yes** | No | No | Yes | Yes | Yes |
| Zero-knowledge | **Yes** | No | Yes | No | No | No |
| Self-destructing | **Yes** | No | Yes | No | No | No |
| No account needed | **Yes** | - | Yes | No | No | No |
| Self-hostable | **Yes** | No | Yes | No | No | Yes |
| Teams | **Yes** | - | No | Yes | Yes | Yes |
| .env focused | **Yes** | No | No | No | Yes | No |
| AI scan | **Yes** | No | No | No | No | No |
| Free | **Yes** | - | Yes | No | No | No |
kip sits between "paste it in Slack" and "deploy HashiCorp Vault". It's the tool for teams that want security without the overhead.

---

## CLI Reference

```
kip push <file> [flags]       Encrypt and share via link
  --expires <duration>              10m, 1h, 24h (default: 1h)
  --reads <n>                       Max reads before destruction (default: 1)
  --password                        Prompt for additional password
  --all                             Share with entire team
  --to <user>                       Share with specific team member
  --pin                             Pin as official team env

kip pull [url]                Download and decrypt
  --output <path>                   Save with different filename
  --stdout                          Print to console instead of saving
  --file <name>                     Pull specific file from multi-file share

kip generate [file]           Generate .env.example (strip values)
kip scan [file]               Scan for exposed/dangerous keys
kip diff                      Compare local .env with team's pinned version
kip inbox                     See pending shares across all teams
kip ls                        List available envs in current team

kip team create <name>        Create team, get invite code
kip team join <code>          Join with invite code
kip team ls                   List your teams
kip team leave [team]         Leave a team
kip team members [team]       List team members

kip init                      Link current directory to a team
kip use <team>                Switch active team
kip config set <key> <value>  Configure settings
```

---

## Architecture

```
┌─────────────┐         ┌──────────────────┐
│   CLI (Go)  │──push──▶│  Server (Go)     │
│             │◀─pull───│  net/http stdlib  │
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

- **CLI:** Go + Cobra. Single binary, zero runtime dependencies.
- **Server:** Go stdlib `net/http`. ~5 endpoints. Redis for ephemeral secrets (native TTL = free auto-destruction). SQLite for teams/members.
- **Web:** Astro + Vue. Static landing page (SSG). Decrypt page uses Web Crypto API for client-side AES-256-GCM decryption.
- **Crypto:** `crypto/aes` + `crypto/cipher` (Go stdlib) / `SubtleCrypto` (browser). Argon2id for password-based key derivation.

---

## Development

```bash
# Prerequisites: Go 1.26+, Node 22+, Redis, Make

# Build CLI
make build

# Run server (dev mode)
make server

# Run web (dev mode)
cd web && npm run dev

# Run tests
make test

# Lint
make lint
```

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

---

## License

[MIT](LICENSE) — free forever, for everyone.

---

<div align="center">

**[kipenv.dev](https://kipenv.dev)**

Built by [Antonio Vila](https://antoniovila.dev)

</div>
