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
FUSION_DATABASE_URL=postgres://fusion:change-me@postgres.example.com:5432/fusion?sslmode=disable
FUSION_PASSWORD=change-me
```

启动：

```shell
docker compose up -d --build
```

访问：

```text
http://localhost:12180
```

## 配置说明

PostgreSQL 必填：

```env
FUSION_DATABASE_URL=postgres://fusion:change-me@postgres.example.com:5432/fusion?sslmode=disable
```

Redis 可选。留空表示禁用缓存：

```env
FUSION_REDIS_URL=
FUSION_CACHE_TTL_SECONDS=120
```

如果启用 Redis，请填写外部 Redis 地址：

```env
FUSION_REDIS_URL=redis://redis.example.com:6379/0
```

当前版本不再支持 SQLite，也不会自动迁移 SQLite 数据。

## 本地二进制运行

从 Releases 下载对应平台的二进制后运行：

```shell
FUSION_DATABASE_URL="postgres://fusion:change-me@localhost:5432/fusion?sslmode=disable" \
FUSION_PASSWORD="fusion" \
./fusion
```

Windows PowerShell：

```powershell
$env:FUSION_DATABASE_URL="postgres://fusion:change-me@localhost:5432/fusion?sslmode=disable"
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
