#!/usr/bin/env python3
"""PreToolUse hook: deny Bash commands that modify jzero/goctl generated files.

Catches shell-level bypasses (sed -i, > redirect, tee, patch, git checkout
--theirs on generated paths, perl -pi, ed, etc.) that the Edit/Write guard
cannot see. Detection: the command string references a protected path AND
contains a mutating verb.

Exit codes: 0 pass, 2 block (deny for PreToolUse).
"""
import json
import re
import shlex
import sys

# Protected generated-file paths (substring match on the command string).
PROTECTED_PATTERNS = [
    r"internal/handler/route2code\.go",
    r"internal/handler/routes\.go",
    r"internal/handler/v1/hrm/[^\"' ]*_compact\.go",
    r"internal/types/v1/hrm/[^\"' ]*types\.go",
    r"internal/model/[A-Za-z0-9_]+/hrm_[A-Za-z0-9_]+_model_gen\.go",
    r"internal/model/model\.go",
    r"router/elegant/(routes|imports|transform)\.ts",
    r"typings/elegant-router\.d\.ts",
    r"desc/swagger/swagger\.json",
]

# Shell verbs that (can) write to their operands.
MUTATING = [
    r"\bsed\b[^|]*-i",
    r"\bperl\b[^|]*-i",
    r"\bed\b\s",
    r">\s*['\"]?[^\s|]*\.go['\"]?",
    r">>?\s*['\"]?[^\s|]*\.ts['\"]?",
    r"\btee\b",
    r"\bsponge\b",
    r"\bpatch\b",
    r"\bgit\b\s+checkout\s+(-b\s+\S+\s+)?(--theirs|--ours)",
    r"\bgit\b\s+checkout\s+(-b\s+\S+\s+)?\S*--\s+\S*",
    r"\bgit\b\s+restore\b(?!\s+(?:--staged\b|-S\b))",
    r"\bdd\b[^|]*\bof=",
    r"\btruncate\b",
    r"\bmv\b",
    r"\bcat\b[^|>]*>",
    r"\brm\b",
]

BLOCK_MSG = (
    "拒绝通过 Bash 修改生成文件（命令引用了受保护路径且含写入动词）: {cmd}\n"
    "route2code.go / routes.go / *_compact.go / *_model_gen.go / model.go / "
    "router/elegant/* / elegant-router.d.ts / swagger.json 是 jzero/goctl/"
    "elegant-router 的生成物。\n"
    "正确做法：修改 server/desc/ 下 .api/.sql 或 web/src/views 目录结构后，"
    "运行 `jzero gen`（后端）或 `timeout 45 pnpm dev`（前端路由）重新生成；"
    "合并冲突涉及生成文件时，解冲突后运行生成器收敛，不要手写合并结果。"
)


def cp_dest_protected(cmd: str) -> bool:
    """`cp` with a protected generated file as DESTINATION is destructive;
    as SOURCE (backup) it is read-only and must pass."""
    try:
        tokens = shlex.split(cmd)
    except ValueError:
        return True  # unparseable → be safe
    i, n = 0, len(tokens)
    while i < n:
        if tokens[i] == "cp" or tokens[i].endswith("/cp"):
            operands, j = [], i + 1
            while j < n and tokens[j] not in ("|", ";", "&&", "||", "&"):
                if not tokens[j].startswith("-"):
                    operands.append(tokens[j])
                j += 1
            if operands and any(re.search(p, operands[-1]) for p in PROTECTED_PATTERNS):
                return True
            i = j
        else:
            i += 1
    return False


def main() -> int:
    try:
        data = json.load(sys.stdin)
    except Exception:
        return 0
    cmd = (data.get("tool_input") or {}).get("command") or ""
    if not cmd:
        return 0
    # Skip pure generator invocations — jzero gen may legitimately rewrite
    # protected files; that is the sanctioned path.
    if re.search(r"\bjzero\s+gen\b", cmd) and not re.search(
        r"\b(sed|perl|patch|tee|mv|cp|rm|truncate|dd)\b", cmd
    ):
        return 0
    protected = any(re.search(p, cmd) for p in PROTECTED_PATTERNS)
    if not protected:
        return 0
    mutating = any(re.search(m, cmd) for m in MUTATING) or cp_dest_protected(cmd)
    if mutating:
        print(BLOCK_MSG.format(cmd=cmd[:300]), file=sys.stderr)
        return 2
    return 0


if __name__ == "__main__":
    sys.exit(main())
