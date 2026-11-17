# AGENTS.md

Guidance for AI coding agents working in this repo.

## Agent skills

### Issue tracker

GitHub Issues on `wuzuxian/jzero-admin` (origin), via the `gh` CLI. See `docs/agents/issue-tracker.md`.

### Triage labels

Five canonical triage roles, each label string equal to its role name. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: root `GLOSSARY.md` + `docs/adr/` at the repo root. See `docs/agents/domain.md`.

## Environment

### Shell working directory

Do not `cd` persistently in Bash calls — the primary working directory follows the last `cd`, and the PreToolUse hooks resolve their scripts relative to it, so hook invocations fail and every Bash/Write call gets blocked. Run everything from the repo root with absolute paths or `pnpm --dir <pkg>`; use `scripts/dev-restart-server.sh` instead of manual `kill`+`nohup` for the backend.

### Check commands

Run from the repo root, never bare `vue-tsc`/`eslint` (they drown in node_modules noise):

- `pnpm --dir web typecheck` — Vue/TS types
- `pnpm --dir web guard:template` — bans `as unknown as` casts on template value bindings (see `web/scripts/check-template-contracts.mjs` header for the incident behind it)
- `pnpm --dir web lint` — eslint
- `pnpm --dir e2e exec playwright test tests/bugs tests/smoke` — frontend regression (needs the dev server on :9527; helpers in `e2e/helpers/auth.ts`, page inventory in `e2e/helpers/hrm-pages.ts`)

### GitHub access

`raw.githubusercontent.com` and plain `WebFetch` to github.com are unreliable from this machine; fetch repo file contents with `gh api repos/<owner>/<repo>/contents/<path>` instead.
