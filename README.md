<h1 align="center">Fusion</h1>
<p align="center">A lightweight, self-hosted RSS reader and aggregator.</p>

<p align="center">
  <a href="./README.zh-CN.md">简体中文</a>
</p>

<p align="center">
  <img src="./assets/article_list_light.png" alt="Article list view" width="48.5%" />&nbsp;
  <img src="./assets/article_detail_light.png" alt="Article detail view" width="48.5%" />
</p>

## Features

- Fast reading workflow: unread state, bookmarks, read later, search, and keyboard shortcuts
- Feed management: RSS/Atom parsing, feed auto-discovery, OPML import/export, and groups
- Fever API compatibility for third-party clients such as Reeder, Unread, and FeedMe
- Password login with optional OIDC SSO
- Responsive web UI with PWA support
- Built-in i18n: English, Chinese, German, French, Spanish, Russian, Portuguese, Swedish
- PostgreSQL-backed self-hosting with optional Redis read cache
- No AI features by design: focused, distraction-free RSS reading

## Deployment

PostgreSQL is required. Redis is optional and only used as a read response cache.

### Docker Compose (Recommended)

Prepare an external PostgreSQL database first. If you want Redis caching, prepare Redis too.

Copy the example environment file:

```shell
cp .env.example .env
```

Edit `.env` and set at least:

```env
FUSION_DATABASE_URL=postgres://postgres:change-me@192.168.2.1:5432/fusion?sslmode=disable
FUSION_PASSWORD=change-me
```

If Redis is not available, remove or comment out `FUSION_REDIS_URL`. If Redis is available, update it for your Redis instance:

```env
FUSION_REDIS_URL=redis://:password@192.168.2.1:6379/0
```

Start Fusion:

```shell
docker compose up -d --build
```

Open:

```text
http://localhost:8080
```

The repository `docker-compose.yml` builds Fusion from source and reads `.env`.

### Pre-built Binary

Download the binary for your platform from [Releases](https://github.com/antstars/fusion/releases), then run it with PostgreSQL configured.

Linux/macOS:

```shell
chmod +x fusion
FUSION_DATABASE_URL="postgres://postgres:change-me@127.0.0.1:5432/fusion?sslmode=disable" \
FUSION_PASSWORD="fusion" \
./fusion
```

Windows PowerShell:

```powershell
$env:FUSION_DATABASE_URL="postgres://postgres:change-me@127.0.0.1:5432/fusion?sslmode=disable"
$env:FUSION_PASSWORD="fusion"
.\fusion.exe
```

Open `http://localhost:8080`.

### Docker Run

Use `latest` for the latest release image, or `main` for the latest development build.

```shell
docker run -it -d -p 8080:8080 \
  -e FUSION_DATABASE_URL="postgres://postgres:change-me@192.168.2.1:5432/fusion?sslmode=disable" \
  -e FUSION_PASSWORD="fusion" \
  ghcr.io/0x2e/fusion:latest
```

Open `http://localhost:8080`.

### Build From Source

See [Contributing](./CONTRIBUTING.md).

## Configuration

Most deployments need:

- `FUSION_DATABASE_URL`: required PostgreSQL connection string
- `FUSION_PASSWORD`: required unless `FUSION_ALLOW_EMPTY_PASSWORD=true`
- `FUSION_PORT`: optional, defaults to `8080`

Structured database settings are also supported instead of `FUSION_DATABASE_URL`:

```env
FUSION_DATABASE_HOST=192.168.2.1
FUSION_DATABASE_PORT=5432
FUSION_DATABASE_USER=postgres
FUSION_DATABASE_PASSWORD=change-me
FUSION_DATABASE_NAME=fusion
FUSION_DATABASE_SSLMODE=disable
```

Optional Redis read cache:

```env
FUSION_REDIS_URL=redis://:password@192.168.2.1:6379/0
```

You can also configure Redis with structured `FUSION_REDIS_*` settings. `FUSION_REDIS_URL` takes precedence and enables Redis automatically when set.

Common optional settings:

- Reverse proxy/CORS: `FUSION_CORS_ALLOWED_ORIGINS`, `FUSION_TRUSTED_PROXIES`
- Fever clients: `FUSION_FEVER_USERNAME` (default: `fusion`)
- OIDC SSO: `FUSION_OIDC_*`, with callback path `/api/oidc/callback`
- Feed pulling: `FUSION_PULL_INTERVAL`, `FUSION_PULL_TIMEOUT`, `FUSION_PULL_CONCURRENCY`, `FUSION_PULL_MAX_BACKOFF`
- Private network feeds: `FUSION_ALLOW_PRIVATE_FEEDS`
- Login rate limiting: `FUSION_LOGIN_RATE_LIMIT`, `FUSION_LOGIN_WINDOW`, `FUSION_LOGIN_BLOCK`
- Logging: `FUSION_LOG_LEVEL`, `FUSION_LOG_FORMAT`

For the complete variable reference, see [`.env.example`](./.env.example).

SQLite is no longer supported, and Fusion does not automatically migrate SQLite data.

Legacy env names (`PASSWORD`, `PORT`) are still accepted for backward compatibility.

## Documentation

- Chinese README: [`README.zh-CN.md`](./README.zh-CN.md)
- API contract (OpenAPI): [`docs/openapi.yaml`](./docs/openapi.yaml)
- Fever API compatibility: [`docs/fever-api.md`](./docs/fever-api.md)
- Backend design: [`docs/backend-design.md`](./docs/backend-design.md)
- Frontend design: [`docs/frontend-design.md`](./docs/frontend-design.md)
- Legacy schema reference (kept for migration work): [`docs/old-database-schema.md`](./docs/old-database-schema.md)

## Development

- Requirements: Go `1.25+`, Node.js `24+`, pnpm
- Helpful commands are in [`scripts.sh`](./scripts.sh)
- Frontend i18n key check: `cd frontend && npm run check:i18n`

Example:

```shell
./scripts.sh build
```

## Contributing

Contributions are welcome. Please read [Contributing Guidelines](./CONTRIBUTING.md) before opening a PR.

## Credits

- Feed parsing powered by [gofeed](https://github.com/mmcdole/gofeed)
