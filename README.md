<h1 align="center">Fusion</h1>
<p align="center">A lightweight RSS reader.</p>

<p align="center">
  <img src="./assets/article_list_light.png" alt="Article list view" width="48.5%" />&nbsp;
  <img src="./assets/article_detail_light.png" alt="Article detail view" width="48.5%" />
</p>

## Features

- Fast reading workflow: unread tracking, bookmarks, search, and Google Reader-style keyboard shortcuts
- Feed management: RSS/Atom parsing, feed auto-discovery, and group organization
- Fever API compatibility for third-party clients (Reeder, Unread, FeedMe, etc.)
- Responsive web UI with PWA support
- Self-hosting friendly: single binary or Docker deployment
- Built-in i18n: English, Chinese, German, French, Spanish, Russian, Portuguese, Swedish
- No AI features by design: focused, distraction-free RSS reading

## Installation

<details>
  <summary><strong>Option 1 (Recommended): Run pre-built binary from Releases</strong></summary>

Download the binary for your platform from [Releases](https://github.com/antstars/fusion/releases), then run:

```shell
chmod +x fusion
FUSION_PASSWORD="fusion" ./fusion
```

Windows (PowerShell):

```powershell
$env:FUSION_PASSWORD="fusion"
.\fusion.exe
```

Open `http://localhost:8080`.
</details>

<details>
  <summary><strong>Option 2: Run with Docker</strong></summary>

`latest` is the latest release image.

`main` is the latest development build.

```shell
docker run -it -d -p 8080:8080 \
  -v $(pwd)/fusion:/data \
  -e FUSION_PASSWORD="fusion" \
  ghcr.io/0x2e/fusion:latest
```

Open `http://localhost:8080`.

Docker Compose example:

```yaml
services:
  fusion:
    build:
      context: .
    env_file:
      - .env.example
    ports:
      - "127.0.0.1:8080:8080"
    restart: unless-stopped
```

The repository `docker-compose.yml` builds Fusion from source and reads `.env`. Provide PostgreSQL, and optionally Redis, as external services.
</details>

<details>
  <summary><strong>Option 3: Build from source</strong></summary>

See [Contributing](./CONTRIBUTING.md).
</details>

<details>
  <summary><strong>Option 4: One-click deployment</strong></summary>

- [Deploy on Fly.io](./fly.toml)
- [Deploy on Railway](https://railway.com/template/XSPFK0?referralCode=milo) (community maintained)
</details>

## Configuration

Most users only need one setting to get started:

- Set `FUSION_PASSWORD`.
- Configure external PostgreSQL with `FUSION_DATABASE_HOST`, `FUSION_DATABASE_USER`, `FUSION_DATABASE_PASSWORD`, and `FUSION_DATABASE_NAME`.
- For local trusted environments only: set `FUSION_ALLOW_EMPTY_PASSWORD=true` to run without auth when OIDC is also disabled.

Then configure based on your goal:

- Run locally or on a home server
  - Optional: `FUSION_PORT`
- Use Docker Compose
  - Copy `.env.example` to `.env`
  - Configure the structured `FUSION_DATABASE_*` settings for your external PostgreSQL service
  - Optional advanced override: set `FUSION_DATABASE_URL` instead of the structured database fields
  - Enable Redis with `FUSION_REDIS_ENABLED=true` and structured `FUSION_REDIS_*` settings, or use `FUSION_REDIS_URL` as an override
- Expose Fusion behind a reverse proxy
  - Configure: `FUSION_CORS_ALLOWED_ORIGINS`, `FUSION_TRUSTED_PROXIES`
- Use mobile/desktop Fever clients (Reeder, Unread, FeedMe)
  - Configure: `FUSION_FEVER_USERNAME` (default: `fusion`)
  - Guide: [`docs/fever-api.md`](./docs/fever-api.md)
- Use SSO instead of password-only login
  - Configure: `FUSION_OIDC_*`
  - Set `FUSION_OIDC_REDIRECT_URI` to `https://<host>/api/oidc/callback`
  - `https://<host>/oidc/callback` is accepted for compatibility
- Tune feed pull behavior
  - Configure: `FUSION_PULL_INTERVAL`, `FUSION_PULL_TIMEOUT`, `FUSION_PULL_CONCURRENCY`, `FUSION_PULL_MAX_BACKOFF`
  - Optional for private networks: `FUSION_ALLOW_PRIVATE_FEEDS`
- Troubleshoot deployments
  - Configure: `FUSION_LOG_LEVEL`, `FUSION_LOG_FORMAT`
- Tune read cache
  - Configure: `FUSION_REDIS_ENABLED`, `FUSION_REDIS_ADDR`, `FUSION_REDIS_DB`, `FUSION_CACHE_TTL_SECONDS`
  - Keep `FUSION_REDIS_ENABLED=false` to disable Redis caching

For the complete variable reference, see [`.env.example`](./.env.example).

Legacy env names (`PASSWORD`, `PORT`) are still accepted for backward compatibility.

## Documentation

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
