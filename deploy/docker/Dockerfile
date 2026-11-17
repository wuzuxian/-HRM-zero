# Unified jzero-admin image: Go backend serves the built SPA frontend from a
# single process (see docs/adr/0005-unified-image-go-serves-web.md).
# Build context is the repo root:
#   docker build -f deploy/docker/Dockerfile .
# The frontend is optional at build time; without web/ the image is API-only
# (same behavior as WEB_DIST unset at runtime).

# ---------- frontend ----------
# API_ONLY=true skips the actual build and yields an empty /web-dist (runtime
# static.Handler sees no index.html and stays API-only); the final COPY below
# still resolves. Guard runs before each COPY so skipped stages need no inputs.
# Base images use docker.xuanyuan.run mirror (see xuanyuan/agents.md §2;
# logged in once in ~/.docker/config.json — pull directly, do not re-login).
FROM docker.xuanyuan.run/library/node:20.12.2-bookworm-slim AS web-builder
ARG API_ONLY=false

WORKDIR /web
# npm registry mirror (host uses npmmirror too; registry.npmjs.org times out here)
ENV NPM_CONFIG_REGISTRY=https://registry.npmmirror.com
RUN if [ "$API_ONLY" = "true" ]; then mkdir -p /web-dist && exit 0; fi

COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
COPY web/patches ./patches
COPY web/packages ./packages

RUN if [ "$API_ONLY" = "true" ]; then mkdir -p /web-dist && exit 0; fi \
  && corepack enable && COREPACK_NPM_REGISTRY=https://registry.npmmirror.com corepack prepare pnpm@9.14.2 --activate

RUN --mount=type=cache,id=pnpm,target=/web/.pnpm-store pnpm i --frozen-lockfile

COPY web/ ./
RUN pnpm build && mv dist /web-dist

# ---------- backend ----------
# jzero builder comes from ghcr.io via xuanyuan mirror (see xuanyuan/agents.md
# §2.1: ghcr maps to {PREFIX}-ghcr.xuanyuan.run; exclusive domains never login).
FROM q8j3alf3tyrb8obveu-ghcr.xuanyuan.run/jzero-io/jzero:latest AS server-builder

ARG TARGETARCH
ARG LDFLAGS

ENV GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct

WORKDIR /app

COPY server/ ./

RUN jzero serverless build

RUN --mount=type=cache,target=/go/pkg CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -a -ldflags="$LDFLAGS" -o /dist/app main.go \
    && cp -r etc /dist \
    && mkdir -p /dist/desc && cp -r desc/swagger /dist/desc && cp -r desc/sql_migration /dist/desc \
    && for plugin in $(ls -d plugins/*/); do \
         mkdir -p /dist/$plugin && cp -r $plugin/etc /dist/$plugin; \
         [ -d $plugin/desc/sql_migration ] && mkdir -p /dist/$plugin/desc && cp -r $plugin/desc/sql_migration /dist/$plugin/desc || true; \
       done

# ---------- runtime ----------
FROM docker.xuanyuan.run/library/alpine:3.20

RUN apk add --no-cache tzdata

WORKDIR /dist

COPY --from=server-builder /dist .
COPY --from=web-builder /web-dist ./web

ENV WEB_DIST=/dist/web

EXPOSE 8001

CMD ["./app", "server"]
