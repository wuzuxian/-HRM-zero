#!/usr/bin/env python3
"""PreToolUse hook: guard the structure of jzero-generated logic files.

Scope: server/internal/logic/**/*.go (excluding *_test.go).

jzero gen generates one logic file per @handler (snake_case(@handler).go)
defining type <Handler>, New<Handler>(...) and the (l *<Handler>) <Handler>(...)
method. `jzero gen` SKIPS existing logic files and only re-creates them when
deleted — so deleting a logic file and rewriting it under a different name
loses all handwritten logic on the next gen.

A logic file is a CONTRACT file when either
  - its filename maps to a @handler (snake_case(@handler).go), or
  - its content defines New<X>() for some @handler X (covers misnamed files).
Contract files may be modified; their structure MUST NOT change.

Enforced rules:
R1. Write/Edit/MultiEdit on an existing contract logic file must keep every
    logic symbol: New<Handler> constructor and every (recv *T) M(...) method
    present in the old content.
R2. Edit chunks that drop a contract symbol: full-signature edits (the
    `*T) M(` form) and fragment renames (old chunk mentions `M(`/`NewT(`,
    new chunk no longer does) are blocked.
R3. Write creating a NEW logic-dir file whose content has a @handler's logic
    shape (New<X> / (recv *X) / type X struct) under a different filename is
    blocked — the "delete original, rewrite unregistered" pattern.
R4. Write creating a NEW logic-dir file colliding with an existing sibling
    top-level symbol (duplicate declaration) is blocked.

Exit codes: 0 pass, 2 block (deny for PreToolUse).
"""
import json
import os
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]

LOGIC_RE = re.compile(r"/server/internal/logic/.+\.go$")
CONSTRUCTOR_RE = re.compile(r"^func New(\w+)\(", re.M)
METHOD_RE = re.compile(r"^func \(\w+ \*(\w+)\) (\w+)\(", re.M)
TYPE_RE = re.compile(r"^type (\w+) struct", re.M)
TOP_FUNC_RE = re.compile(r"^func (\w+)\(", re.M)

BLOCK_MSG = (
    "拒绝修改 logic 结构: {path}\n"
    "该文件是 jzero 按 @handler「{handler}」生成的 logic 文件，其结构属于生成契约：\n"
    "  文件名 = snake_case(@handler)；类型 <Handler> + New<Handler>() + "
    "(l *<Handler>) <Handler>() 方法。\n"
    "  `jzero gen` 对已存在的 logic 文件跳过不覆盖，但文件被删除后会重新生成空壳，\n"
    "  把手写逻辑全部冲掉，或与新写的文件 duplicate declaration。\n"
    "允许：修改/新增函数体、import、辅助函数（可放同目录伴生文件如 base.go）。\n"
    "禁止：删除或改名该类型/构造器/方法、把文件删掉换名字重写、另写不注册的平行 logic。\n"
    "被移除的契约符号：{missing}\n"
    "如确属重构需要，请先向用户说明并取得确认。"
)

NEW_FILE_MSG = (
    "拒绝新建 logic 文件: {path}\n"
    "原因：{reason}\n"
    "正确做法：\n"
    "1. 实现路由「{handler}」的逻辑 → 写入 server/internal/logic/.../{stem}.go "
    "（文件名 = snake_case(@handler)）；文件已存在则直接修改它，不要另起炉灶；\n"
    "2. 辅助函数 → 放进既有伴生文件（base.go 等）或定义无 New<Handler> 形态的新文件。\n"
    "如确属重构需要，请先向用户说明并取得确认。"
)


def pascal_to_snake(name: str) -> str:
    return re.sub(r"(?<!^)(?=[A-Z])", "_", name).lower()


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


def contract_patterns(content: str, handlers: set):
    """(regex, display) pairs for every logic-shaped symbol in content.

    Matched by shape (receiver name may differ: l/r/x...), not literally.
    A file is contract when it defines New<X> for a @handler X; all its
    logic-shaped symbols then become protected.
    """
    ctor_names = {n for n in CONSTRUCTOR_RE.findall(content) if n in handlers}
    type_names = ctor_names | {t for t in TYPE_RE.findall(content) if t in handlers}
    pats = []
    for name in CONSTRUCTOR_RE.findall(content):
        pats.append((re.compile(rf"^func\s+New{name}\(", re.M), f"func New{name}("))
    for rtype, name in METHOD_RE.findall(content):
        if rtype in type_names or rtype in handlers:
            pats.append(
                (
                    re.compile(rf"^func\s+\(\w+\s+\*{rtype}\)\s+{name}\(", re.M),
                    f"func (recv *{rtype}) {name}(",
                )
            )
    return pats, ctor_names | {rtype for rtype, _ in METHOD_RE.findall(content) if rtype in handlers}


def edit_chunks(data: dict):
    """Yield (old, new) string pairs for Edit / MultiEdit / Write."""
    tool = data.get("tool_name") or ""
    ti = data.get("tool_input") or {}
    if tool == "Edit":
        yield (ti.get("old_string") or "", ti.get("new_string") or "")
    elif tool == "MultiEdit":
        for e in ti.get("edits") or []:
            yield (e.get("old_string") or "", e.get("new_string") or "")
    else:  # Write
        yield ("", ti.get("content") or "")


def fragment_missing(chunks, type_names, old_content):
    """Edit chunks that strip every mention of a contract symbol: partial
    signatures (e.g. old_string='GenerateUser(w http') don't carry the full
    `*T) M(` shape, so also catch bare `M(` / `NewT(` disappearances."""
    missing = []
    for t in type_names:
        for old_str, new_str in chunks:
            pat = re.compile(rf"\bNew{t}\(")
            if old_str and pat.search(old_str) and not pat.search(new_str):
                missing.append(f"func New{t}(")
            for rtype, mname in METHOD_RE.findall(old_content):
                if rtype != t:
                    continue
                pat = re.compile(rf"\b{mname}\(")
                if old_str and pat.search(old_str) and not pat.search(new_str):
                    missing.append(f"(recv *{t}) {mname}(")
    return missing


def main() -> int:
    try:
        data = json.load(sys.stdin)
    except Exception:
        return 0
    ti = data.get("tool_input") or {}
    tool = data.get("tool_name") or ""
    path = ti.get("file_path") or ""
    if not path or tool not in ("Edit", "Write", "MultiEdit"):
        return 0
    norm = os.path.abspath(path).replace(os.sep, "/")
    if not LOGIC_RE.search(norm) or norm.endswith("_test.go"):
        return 0

    handlers = handler_names()
    stem = Path(norm).stem
    handler = "".join(p[:1].upper() + p[1:] for p in stem.split("_") if p)

    exists = os.path.isfile(norm)
    old_content = ""
    if exists:
        try:
            with open(norm, encoding="utf-8", errors="ignore") as fh:
                old_content = fh.read()
        except OSError:
            return 0

    # R1/R2: existing contract logic file — its symbols must survive.
    if exists:
        old_syms, type_names = contract_patterns(old_content, handlers)
        if old_syms and type_names:
            chunks = list(edit_chunks(data))
            missing = []
            if tool == "Write":
                # Whole-file rewrite: every old symbol must exist in new content.
                missing = [disp for pat, disp in old_syms if not pat.search(chunks[0][1])]
            else:
                for old_str, new_str in chunks:
                    for pat, disp in old_syms:
                        if pat.search(old_str) and not pat.search(new_str):
                            missing.append(disp)
                missing += fragment_missing(chunks, type_names, old_content)
            if missing:
                print(
                    BLOCK_MSG.format(path=path, handler=handler, missing=", ".join(sorted(set(missing)))),
                    file=sys.stderr,
                )
                return 2
        return 0

    # R3/R4: creating a new file in a logic dir (Write only).
    new_content = ti.get("content") or ""
    if not new_content:
        return 0

    # R3: defines a @handler's logic shape under a different filename.
    for name in CONSTRUCTOR_RE.findall(new_content):
        if name in handlers and name != handler:
            _block_new_file(path, stem, name, new_content, handlers)
            return 2
    for rtype, _ in METHOD_RE.findall(new_content):
        if rtype in handlers and rtype != handler:
            _block_new_file(path, stem, rtype, new_content, handlers)
            return 2
    for tname in TYPE_RE.findall(new_content):
        if tname in handlers and tname != handler:
            _block_new_file(path, stem, tname, new_content, handlers)
            return 2

    # R4: collides with an existing sibling symbol (duplicate declaration).
    sibling_syms = set()
    try:
        for sib in Path(norm).parent.glob("*.go"):
            if sib.is_file():
                text = sib.read_text(encoding="utf-8", errors="ignore")
                sibling_syms.update(CONSTRUCTOR_RE.findall(text))
                sibling_syms.update(TYPE_RE.findall(text))
                sibling_syms.update(TOP_FUNC_RE.findall(text))
    except OSError:
        return 0
    new_syms = set(CONSTRUCTOR_RE.findall(new_content))
    new_syms |= set(TYPE_RE.findall(new_content))
    new_syms |= set(TOP_FUNC_RE.findall(new_content))
    dupes = sorted(new_syms & sibling_syms)
    if dupes:
        reason = (
            "同包已有同名符号：" + ", ".join(dupes)
            + " —— 与现有文件重复声明，编译必然失败（duplicate declaration）。"
        )
        print(
            NEW_FILE_MSG.format(
                path=path, reason=reason, handler=dupes[0], stem=pascal_to_snake(dupes[0])
            ),
            file=sys.stderr,
        )
        return 2

    return 0


def _block_new_file(path: str, stem: str, name: str, new_content: str, handlers: set):
    reason = (
        f"内容定义了路由「{name}」的 logic 形态（type {name} / New{name}() / "
        f"(l *{name}) 方法），但文件名是 {stem}.go —— 这是「删原 logic 文件后换个名字"
        f"重写且不注册」的形态；下次 `jzero gen` 会重新生成空壳 {pascal_to_snake(name)}.go，"
        "与本文件冲突并把手写逻辑冲掉。"
    )
    print(
        NEW_FILE_MSG.format(
            path=path, reason=reason, handler=name, stem=pascal_to_snake(name)
        ),
        file=sys.stderr,
    )


if __name__ == "__main__":
    sys.exit(main())
