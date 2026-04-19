# Environment Variables

All configuration is via CLI flags. Environment variables are used in Docker Compose for convenience.

## Server

| Variable | Flag | Default | Description |
|---|---|---|---|
| — | `-addr` | `:8080` | Listen address (host:port) |
| — | `-redis` | _(empty)_ | Redis URL. If empty, uses in-memory store (dev mode) |
| — | `-db` | `~/.local/share/kip/kip.db` | SQLite database path for teams |
| — | `-rate-limit` | `30` | Max requests per minute per IP |

## Docker Compose (production)

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | Host port to expose |
| `RATE_LIMIT` | `30` | Max requests per minute per IP |

## Redis

The server uses Redis for ephemeral secret storage with native TTL for auto-expiration. Any Redis 6+ instance works.

Connection string format: `redis://[:password@]host:port[/db]`

## SQLite

Teams and members are stored in SQLite. The database file is created automatically. No configuration needed unless you want a custom path.

## Example: Production with custom settings

```bash
PORT=443 RATE_LIMIT=60 docker compose -f deploy/docker-compose.yml up -d
```
