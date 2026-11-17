#!/usr/bin/env bash
# 构建统一 Docker 镜像（后端 + 前端单进程镜像，见 docs/adr/0005-unified-image-go-serves-web.md）。
#
# Build context 是仓库根目录，Dockerfile 在 deploy/docker/Dockerfile；
# 构建/运行分离：本脚本只 build，运行用 deploy/docker/docker-compose.yaml。
#
# 用法：scripts/build-docker-image.sh [--image <name>] [--tag <tag>] [--push] [--api-only]
#   --image    镜像名，默认 jzero-admin
#   --tag      额外 tag；默认自动打 latest + 当前 git short sha 双 tag
#   --push     构建完成后 docker push 所有打出的 tag（默认不 push）
#   --api-only 跳过前端阶段，只构建纯 API 镜像（无 WEB_DIST，行为同旧 server 镜像）
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="jzero-admin"
EXTRA_TAG=""
PUSH=false
API_ONLY=false

while [ $# -gt 0 ]; do
  case "$1" in
    --image) IMAGE="$2"; shift 2 ;;
    --tag) EXTRA_TAG="$2"; shift 2 ;;
    --push) PUSH=true; shift ;;
    --api-only) API_ONLY=true; shift ;;
    *) echo "未知参数: $1" >&2; exit 1 ;;
  esac
done

GIT_SHA="$(git -C "$REPO_ROOT" rev-parse --short HEAD)"
TAGS=("$IMAGE:latest" "$IMAGE:$GIT_SHA")
[ -n "$EXTRA_TAG" ] && TAGS+=("$EXTRA_TAG")

BUILD_ARGS=(--file "$REPO_ROOT/deploy/docker/Dockerfile")
if $API_ONLY; then
  # 跳过前端构建：Dockerfile 前端阶段产出空 /web-dist，
  # 运行时 static.Handler 见不到 index.html 自动退化为纯 API 服务
  BUILD_ARGS+=(--build-arg API_ONLY=true)
fi

# 组装 -t 参数
T_ARGS=()
for t in "${TAGS[@]}"; do T_ARGS+=(-t "$t"); done

echo "docker build ${TAGS[*]}${API_ONLY:+ (api-only, 无前端)}"
# shellcheck disable=SC2086
docker build "${T_ARGS[@]}" "${BUILD_ARGS[@]}" "$REPO_ROOT"

if $PUSH; then
  for t in "${TAGS[@]}"; do
    docker push "$t"
  done
  echo "✓ 已推送 ${TAGS[*]}"
else
  echo "✓ 已构建 ${TAGS[*]}"
fi
