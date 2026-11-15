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

### GitHub access

`raw.githubusercontent.com` and plain `WebFetch` to github.com are unreliable from this machine; fetch repo file contents with `gh api repos/<owner>/<repo>/contents/<path>` instead.
