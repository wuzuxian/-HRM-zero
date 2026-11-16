#!/usr/bin/env python3
"""PreToolUse hook: deny Bash commands that delete/rename/overwrite
jzero-generated logic files.

`jzero gen` skips existing logic files but re-creates an empty skeleton once
the file is deleted — so `rm`/`mv` on a logic file destroys all handwritten
logic on the next gen. A file is a CONTRACT file when its filename maps to a
@handler (snake_case(@handler).go) OR its content defines New<X>() for a
@handler X (covers misnamed files like trial_handler.go). This catches
rm/mv/git rm/git mv/rmdir/cp-onto/dd-of/truncate and output redirects aimed at
such files, including single-file globs. Helper files (base.go, errors.go,
*_test.go, ...) are left alone.

Exit codes: 0 pass, 2 block (deny for PreToolUse).
"""
import glob as globmod
import json
import re
import shlex
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]

LOGIC_PATH_RE = re.compile(r"[\w./\"'*-]*internal/logic/[\w./\"'*-]+")

DELETE_VERBS = re.compile(r"\b(rm|git\s+rm|mv|git\s+mv|rmdir|truncate)\b")
DD_OF_RE = re.compile(r"\bdd\b[^|]*\bof=")
REDIRECT_RE = re.compile(r">>?\s*['\"]?([\w./-]*internal/logic/[\w./-]+\.go)")
TEE_RE = re.compile(r"\btee\b")

BLOCK_MSG = (
    "拒绝删除/移动/覆写 logic 文件（命令引用了生成 logic 契约文件且含删除/移动/覆写动作）: {cmd}\n"
    "以下文件对应 .api 中的 @handler，属于 jzero 生成契约：{targets}\n"
    "`jzero gen` 对已存在的 logic 文件跳过不覆盖，但删除后会重新生成空壳，\n"
    "把手写逻辑全部冲掉。\n"
    "正确做法：\n"
    "1. 修改逻辑 → 用 Edit 直接改该文件（函数体、import 可随意改，"
    "但不得删除/改名 <Handler> 类型、New<Handler>、方法签名）；\n"
    "2. 确需删除路由 → 先改 server/desc/ 下的 .api 定义再 `jzero gen`，"
    "并连同 handler/routes 一致收敛；\n"
    "3. 如确属重构需要，请先向用户说明并取得确认。"
)


def pascal(stem: str) -> str:
    return "".join(p[:1].upper() + p[1:] for p in stem.split("_") if p)


def handler_names() -> set:
    names = set()
    api_root = ROOT / "server" / "desc" / "api"
    if not api_root.is_dir():
        return names
    for fh in api_root.rglob("*.api"):
        try:
            names.update(re.findall(r"@handler\s+(\w+)", fh.read_text(encoding="utf-8", errors="ignore")))
        except OSError:
            continue
    return names


def _file_is_contract(path: Path, handlers: set) -> bool:
    """Filename maps to a @handler, or content defines New<X> for a @handler X."""
    if path.suffix != ".go" or path.name.endswith("_test.go"):
        return False
    if pascal(path.stem) in handlers:
        return True
    try:
        text = path.read_text(encoding="utf-8", errors="ignore")[:4096]
    except OSError:
        return False
    ctor = re.findall(r"^func New(\w+)\(", text, re.M)
    return any(n in handlers for n in ctor)


def cp_writes_contract(cmd: str, handlers: set) -> bool:
    """`cp` with a contract file as DESTINATION is destructive; as SOURCE it is
    a read-only backup and must pass."""
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
            if len(operands) >= 2:
                dest = operands[-1].strip("\"'")
                full_dest = Path(ROOT) / dest if not dest.startswith("/") else Path(dest)
                if full_dest.is_file() and _file_is_contract(full_dest, handlers):
                    return True
                if full_dest.is_dir():
                    for s in operands[:-1]:
                        name = Path(s.strip("\"'")).name
                        if name.endswith(".go") and pascal(Path(name).stem) in handlers:
                            return True  # writes a contract file (new or overwrite)
            i = j
        else:
            i += 1
    return False


def contract_targets(cmd: str, handlers: set):
    """Logic paths in cmd that are contract files, or dirs containing contract files."""
    targets = []
    for raw in LOGIC_PATH_RE.findall(cmd):
        p = raw.strip("\"'")
        full = Path(ROOT) / p if not p.startswith("/") else Path(p)
        if any(ch in p for ch in "*?[{"):
            # glob pattern: expand against the repo
            matches = globmod.glob(str(full)) or globmod.glob(p)
            if any(_file_is_contract(Path(m), handlers) for m in matches):
                targets.append(p)
            continue
        if full.is_file():
            if _file_is_contract(full, handlers):
                targets.append(p)
        elif full.is_dir():
            try:
                has_contract = any(
                    _file_is_contract(f, handlers) for f in full.glob("*.go")
                )
            except OSError:
                has_contract = False
            if has_contract:
                targets.append(p)
    return targets


def main() -> int:
    try:
        data = json.load(sys.stdin)
    except Exception:
        return 0
    cmd = (data.get("tool_input") or {}).get("command") or ""
    if not cmd:
        return 0
    # Pure generator invocations are the sanctioned writer of logic files.
    if re.search(r"\bjzero\s+gen\b", cmd) and not re.search(
        r"\b(rm|mv|rmdir|cp|truncate|dd)\b", cmd
    ):
        return 0
    handlers = handler_names()
    if not handlers:
        return 0

    targets = contract_targets(cmd, handlers)
    if not targets:
        return 0

    blocked = (
        bool(DELETE_VERBS.search(cmd))
        or bool(REDIRECT_RE.search(cmd))
        or bool(DD_OF_RE.search(cmd))
        or (bool(TEE_RE.search(cmd)) and "internal/logic/" in cmd)
        or cp_writes_contract(cmd, handlers)
    )
    if blocked:
        print(BLOCK_MSG.format(cmd=cmd[:300], targets=", ".join(targets)), file=sys.stderr)
        return 2
    return 0


if __name__ == "__main__":
    sys.exit(main())
