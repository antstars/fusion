# Fusion

Fusion 是一个轻量级、自托管的 RSS 阅读器和聚合器。

[English](./README.md)

<p align="center">
  <img src="./assets/article_list_light.png" alt="Article list view" width="48.5%" />&nbsp;
  <img src="./assets/article_detail_light.png" alt="Article detail view" width="48.5%" />
</p>

## 功能特性

- 高效阅读流程：未读状态、书签、稍后读、搜索、键盘快捷键
- 订阅源管理：RSS/Atom 解析、订阅源自动发现、OPML 导入/导出、分组管理
- Fever API 兼容，可用于 Reeder、Unread、FeedMe 等第三方客户端
- 密码登录，可选 OIDC 单点登录
- 响应式 Web UI，支持 PWA
- 内置多语言界面：英语、中文、德语、法语、西班牙语、俄语、葡萄牙语、瑞典语
- 基于 PostgreSQL 自托管，Redis 可作为可选读取缓存
- 不内置 AI 功能，专注纯粹 RSS 阅读

## 部署

PostgreSQL 是必需数据库。Redis 可选，仅用于读取响应缓存。

### Docker Compose（推荐）

先准备外部 PostgreSQL。需要缓存时，再准备外部 Redis。

复制环境变量示例：

```shell
cp .env.example .env
```

编辑 `.env`，至少设置：

```env
FUSION_DATABASE_URL=postgres://postgres:change-me@192.168.2.1:5432/fusion?sslmode=disable
FUSION_PASSWORD=change-me
```

如果没有 Redis，请删除或注释 `FUSION_REDIS_URL`。如果使用 Redis，请改成你的 Redis 地址：

```env
FUSION_REDIS_URL=redis://:password@192.168.2.1:6379/0
```

启动：

```shell
docker compose up -d --build
```

访问：

```text
http://localhost:8080
```

仓库中的 `docker-compose.yml` 会从源码构建 Fusion，并读取 `.env`。

### 本地二进制运行

从 [Releases](https://github.com/antstars/fusion/releases) 下载对应平台的二进制后，配置 PostgreSQL 并运行。

Linux/macOS：

```shell
chmod +x fusion
FUSION_DATABASE_URL="postgres://postgres:change-me@127.0.0.1:5432/fusion?sslmode=disable" \
FUSION_PASSWORD="fusion" \
./fusion
```

Windows PowerShell：

```powershell
$env:FUSION_DATABASE_URL="postgres://postgres:change-me@127.0.0.1:5432/fusion?sslmode=disable"
$env:FUSION_PASSWORD="fusion"
.\fusion.exe
```

默认访问 `http://localhost:8080`。

### Docker Run

`latest` 是最新发布镜像，`main` 是最新开发构建。

```shell
docker run -it -d -p 8080:8080 \
  -e FUSION_DATABASE_URL="postgres://postgres:change-me@192.168.2.1:5432/fusion?sslmode=disable" \
  -e FUSION_PASSWORD="fusion" \
  ghcr.io/0x2e/fusion:latest
```

访问 `http://localhost:8080`。

### 从源码构建

见 [Contributing](./CONTRIBUTING.md)。

## 配置说明

大多数部署需要：

- `FUSION_DATABASE_URL`：必填，PostgreSQL 连接字符串
- `FUSION_PASSWORD`：必填，除非设置 `FUSION_ALLOW_EMPTY_PASSWORD=true`
- `FUSION_PORT`：可选，默认 `8080`

也可以不用 `FUSION_DATABASE_URL`，改用结构化数据库配置：

```env
FUSION_DATABASE_HOST=192.168.2.1
FUSION_DATABASE_PORT=5432
FUSION_DATABASE_USER=postgres
FUSION_DATABASE_PASSWORD=change-me
FUSION_DATABASE_NAME=fusion
FUSION_DATABASE_SSLMODE=disable
```

可选 Redis 读取缓存：

```env
FUSION_REDIS_URL=redis://:password@192.168.2.1:6379/0
```

Redis 也支持结构化 `FUSION_REDIS_*` 配置。设置 `FUSION_REDIS_URL` 后会自动启用 Redis，并优先于结构化字段。

常用可选配置：

- 反向代理/CORS：`FUSION_CORS_ALLOWED_ORIGINS`、`FUSION_TRUSTED_PROXIES`
- Fever 客户端：`FUSION_FEVER_USERNAME`，默认 `fusion`
- OIDC 单点登录：`FUSION_OIDC_*`，回调路径为 `/api/oidc/callback`
- 订阅源拉取：`FUSION_PULL_INTERVAL`、`FUSION_PULL_TIMEOUT`、`FUSION_PULL_CONCURRENCY`、`FUSION_PULL_MAX_BACKOFF`
- 私有网络订阅源：`FUSION_ALLOW_PRIVATE_FEEDS`
- 登录限流：`FUSION_LOGIN_RATE_LIMIT`、`FUSION_LOGIN_WINDOW`、`FUSION_LOGIN_BLOCK`
- 日志：`FUSION_LOG_LEVEL`、`FUSION_LOG_FORMAT`

完整环境变量见 [`.env.example`](./.env.example)。

当前版本不再支持 SQLite，也不会自动迁移 SQLite 数据。

兼容旧环境变量名：`PASSWORD`、`PORT`。

`FUSION_CORS_ALLOWED_ORIGINS` 是可选项，默认允许所有 Origin，方便本地、内网和简单自托管部署。公网部署如需限制浏览器来源，可显式配置允许的 Origin。

## 文档

- 英文 README：[`README.md`](./README.md)
- API 契约：[`docs/openapi.yaml`](./docs/openapi.yaml)
- Fever API：[`docs/fever-api.md`](./docs/fever-api.md)
- 后端设计：[`docs/backend-design.md`](./docs/backend-design.md)
- 前端设计：[`docs/frontend-design.md`](./docs/frontend-design.md)
- 旧数据库结构参考：[`docs/old-database-schema.md`](./docs/old-database-schema.md)

## 开发

- 要求：Go `1.25+`、Node.js `24+`、pnpm
- 常用命令见 [`scripts.sh`](./scripts.sh)
- 前端 i18n key 检查：`cd frontend && npm run check:i18n`
