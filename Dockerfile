FROM node:24-alpine AS frontend
WORKDIR /src/frontend
RUN corepack enable
COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile
COPY frontend/ ./
ARG FUSION_VERSION=dev
RUN VITE_FUSION_VERSION="${FUSION_VERSION}" pnpm run build

FROM golang:1.26-alpine AS backend
WORKDIR /src
COPY backend/go.mod backend/go.sum ./backend/
RUN cd backend && go mod download
COPY backend/ ./backend/
COPY --from=frontend /src/frontend/dist ./backend/internal/web/dist
RUN cd backend && CGO_ENABLED=0 go build -trimpath -ldflags='-extldflags "-static"' -o /out/fusion ./cmd/fusion

FROM alpine:3.21.0
LABEL org.opencontainers.image.source="https://github.com/antstars/fusion"

RUN addgroup -S fusion && adduser -S -D -H -h /fusion -G fusion fusion && \
    mkdir -p /data && chown -R fusion:fusion /data

WORKDIR /fusion
COPY --from=backend --chown=fusion:fusion --chmod=755 /out/fusion ./fusion
EXPOSE 8080
VOLUME ["/data"]
HEALTHCHECK --interval=10s --timeout=3s --start-period=2s --retries=3 \
  CMD wget -q -O /dev/null http://127.0.0.1:8080/api/oidc/enabled || exit 1
USER fusion
CMD [ "./fusion" ]
