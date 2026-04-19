# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| latest  | Yes       |

Only the latest release receives security fixes. This is a small project — update frequently.

## Security Model

kip uses a **zero-knowledge** architecture. The server never sees decryption keys.

### How it works

1. The CLI encrypts your `.env` file locally with **AES-256-GCM** using a random 32-byte key
2. The encrypted blob and nonce are sent to the server — the key is **never** transmitted
3. The key is encoded in the URL `#fragment` (e.g., `kip.dev/s/abc123#<key>`)
4. The `#fragment` is never sent to the server by HTTP protocol
5. The receiver decrypts locally (CLI or browser via Web Crypto API)
6. After the configured number of reads, the server permanently deletes the encrypted data

Password-protected secrets add an **Argon2id** key derivation layer on top.

## Threat Model

| Threat | Mitigation | Residual Risk |
|--------|------------|---------------|
| **Server database breach** | Only encrypted blobs stored; decryption key lives in URL fragment, never on server | None — attacker gets unusable ciphertext |
| **Man-in-the-middle** | HTTPS required; key in `#fragment` is never sent over the network | Compromised TLS CA could intercept the ciphertext (but not the key) |
| **Brute force link IDs** | IDs use nanoid (12+ chars, URL-safe alphabet) — sufficient entropy | Negligible with rate limiting |
| **Replay attacks** | Read counter + Redis TTL — links self-destruct after use | None |
| **Malicious server operator** | Cannot decrypt — can only delete secrets (acceptable tradeoff) | Denial of service by deletion |

## Reporting Vulnerabilities

**Do not open a public GitHub issue for security vulnerabilities.**

Email: **antonio@antoniovila.dev**

Include:
- Description of the vulnerability
- Steps to reproduce
- Impact assessment
- Suggested fix (if you have one)

I'll acknowledge within 48 hours and aim to release a fix within 7 days for critical issues.

## Scope

### In scope

- Cryptographic implementation (AES-256-GCM, Argon2id, key generation)
- Server API (secret storage, access control, TTL enforcement)
- CLI (key handling, URL construction, local file operations)
- Web decrypt page (Web Crypto API usage, fragment handling)
- Docker/self-host configuration

### Out of scope

- Social engineering (tricking users into sharing links)
- Denial of service attacks against the server
- Vulnerabilities in upstream dependencies (report those upstream, but let me know)
- Physical access to user machines
- Browser extensions or OS-level keyloggers
