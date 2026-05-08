# Fusion

Fusion 是一个轻量级、自托管的 RSS 阅读器和聚合器。

## 功能特性

- 高效阅读流程：未读状态、书签、搜索、键盘快捷键
- 订阅源管理：RSS/Atom 解析、订阅源自动发现、分组管理
- Fever API 兼容，可用于 Reeder、Unread、FeedMe 等客户端
- 响应式 Web UI，支持 PWA
- Docker Compose 从源码构建 Fusion，并通过 `.env` 连接外部服务
- PostgreSQL 是必需数据库，Redis 是可选只读缓存

## Docker Compose 部署

先准备外部 PostgreSQL。需要缓存时，再准备外部 Redis。

复制环境变量示例：

```shell
cp .env.example .env
```

编辑 `.env`，至少设置：

```env
FUSION_DATABASE_HOST=192.168.2.6
FUSION_DATABASE_PORT=5432
FUSION_DATABASE_USER=postgres
FUSION_DATABASE_PASSWORD=change-me
FUSION_DATABASE_NAME=fusion
FUSION_DATABASE_SSLMODE=disable
FUSION_PASSWORD=change-me
```

也可以使用高级覆盖项 `FUSION_DATABASE_URL`，此时会优先生效。

启动：

```shell
docker compose up -d --build
```

访问：

```text
http://localhost:12180
```

## 配置说明

PostgreSQL 必填。推荐使用结构化配置：

```env
FUSION_DATABASE_HOST=192.168.2.6
FUSION_DATABASE_PORT=5432
FUSION_DATABASE_USER=postgres
FUSION_DATABASE_PASSWORD=change-me
FUSION_DATABASE_NAME=fusion
FUSION_DATABASE_SSLMODE=disable
FUSION_DATABASE_MAX_OPEN_CONNS=64
FUSION_DATABASE_MAX_IDLE_CONNS=32
FUSION_DATABASE_CONN_MAX_LIFETIME_MINUTES=30
FUSION_DATABASE_CONN_MAX_IDLE_TIME_MINUTES=10
```

Redis 可选。默认关闭：

```env
FUSION_REDIS_ENABLED=false
```

启用 Redis 时填写外部 Redis 配置：

```env
FUSION_REDIS_ENABLED=true
FUSION_REDIS_ADDR=192.168.2.6:6379
FUSION_REDIS_PASSWORD=
FUSION_REDIS_DB=15
FUSION_CACHE_TTL_SECONDS=600
FUSION_REDIS_POOL_SIZE=80
FUSION_REDIS_MIN_IDLE_CONNS=16
FUSION_REDIS_DIAL_TIMEOUT_SECONDS=2
FUSION_REDIS_READ_TIMEOUT_SECONDS=2
FUSION_REDIS_WRITE_TIMEOUT_SECONDS=2
FUSION_REDIS_POOL_TIMEOUT_SECONDS=4
FUSION_REDIS_SCAN_COUNT=500
```

`FUSION_REDIS_URL` 是高级覆盖项；设置后会启用 Redis，并优先于结构化 Redis 字段。

当前版本不再支持 SQLite，也不会自动迁移 SQLite 数据。

## 本地二进制运行

从 Releases 下载对应平台的二进制后运行：

```shell
FUSION_DATABASE_HOST="127.0.0.1" \
FUSION_DATABASE_USER="postgres" \
FUSION_DATABASE_PASSWORD="change-me" \
FUSION_DATABASE_NAME="fusion" \
FUSION_PASSWORD="fusion" \
./fusion
```

Windows PowerShell：

```powershell
$env:FUSION_DATABASE_HOST="127.0.0.1"
$env:FUSION_DATABASE_USER="postgres"
$env:FUSION_DATABASE_PASSWORD="change-me"
$env:FUSION_DATABASE_NAME="fusion"
$env:FUSION_PASSWORD="fusion"
.\fusion.exe
```

默认访问：

```text
http://localhost:8080
```

## 文档

- 英文 README：[`README.md`](./README.md)
- API 契约：[`docs/openapi.yaml`](./docs/openapi.yaml)
- Fever API：[`docs/fever-api.md`](./docs/fever-api.md)
- 后端设计：[`docs/backend-design.md`](./docs/backend-design.md)
- 前端设计：[`docs/frontend-design.md`](./docs/frontend-design.md)
