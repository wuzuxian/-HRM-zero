#!/usr/bin/env bash
# dev 环境后端重启脚本（e2e 工作流用，见 e2e-module-acceptance skill Step 4）。
#
# 为什么需要脚本而不是手工 kill+nohup：本次 retro 实测发生过
# "旧进程没杀干净 → 新进程迁移后因端口占用静默退出 → DB 新、内存旧"的状态漂移
# （casbin 策略/迁移已更新但 :8001 还在服务旧二进制），多花一轮排查。
# 脚本强制"等死 → 再起 → 探活 → 对账迁移版本"，杜绝该漂移。
#
# 用法：scripts/dev-restart-server.sh [--skip-build]
#   --skip-build  跳过 go build（只重启，改 desc/ 下的种子后必须重建）
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN=/tmp/hrm-server-e2e
PORT=8001
DB="$REPO_ROOT/server/jzero-admin.db"
LOG=/tmp/hrm-server-e2e.log

# 1. 杀旧进程并确认退出（杜绝端口占用导致的静默启动失败）
old_pids=$(pgrep -f "$BIN server" || true)
if [ -n "$old_pids" ]; then
  echo "kill 旧进程: $old_pids"
  kill $old_pids 2>/dev/null || true
  for _ in $(seq 1 20); do
    pgrep -f "$BIN server" >/dev/null || break
    sleep 0.5
  done
  if pgrep -f "$BIN server" >/dev/null; then
    echo "优雅退出超时，强制 kill -9"
    pkill -9 -f "$BIN server" || true
    sleep 1
  fi
fi

# 2. 构建（迁移种子的改动必须进二进制才会被迁移执行器看到）
if [ "${1:-}" != "--skip-build" ]; then
  echo "go build → $BIN"
  (cd "$REPO_ROOT/server" && go build -o "$BIN" .)
fi

# 3. 起新进程（后台，日志 tee 到 /tmp/hrm-server-e2e.log）
(cd "$REPO_ROOT/server" && nohup "$BIN" server > "$LOG" 2>&1 &)

# 4. 探活
for _ in $(seq 1 40); do
  if curl -s -o /dev/null "localhost:$PORT/"; then
    break
  fi
  sleep 0.5
done
if ! curl -s -o /dev/null "localhost:$PORT/"; then
  echo "✗ 服务未起来，最后 20 行日志："
  tail -20 "$LOG"
  exit 1
fi

# 5. 对账：迁移版本与进程 PID（人眼扫一眼即可发现"版本没动/进程是旧的"）
echo "✓ 后端已重启  pid=$(pgrep -f "$BIN server" | head -1)  迁移版本=$(sqlite3 "$DB" 'SELECT version FROM schema_migrations;')  日志=$LOG"
