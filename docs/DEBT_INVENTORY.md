# Debt inventory — `integrate/platform-and-logs`

Branch: **integrate/platform-and-logs** (~22 commits ahead of `main`). Merges stack/01–07 work: provider probes, auth/env sync, web viz, multi-strategy logs (Vercel/Render/Supabase), `perch logs setup`, full-platform example.

Compared to [.cursor/plans/spec.md](../.cursor/plans/spec.md): core graph/status/context/init/edge/viz/logs-setup exist; most deploy/env/trace/snapshot/LLM/provider-marketplace commands do not. Product surface is **web viz + Bubbletea TUI**, not spec-only TUI.

---

## Cleanup candidates

### Duplicate / superseded PR branches (safe to close after merge)

| Branch | Superseded by |
|--------|----------------|
| `feature/logs-provider-detection` (PR #11) | integrate merge `bb9021e` |
| `feature/logs-setup-command` | integrate merge `5b7e008` |
| `feature/logs-frontend-wiring` | integrate merge `e2d931a` |
| `frontend-detailpanel-api-wiring` | same viz/logs wiring on integrate |
| `stack/01` … `stack/07` | integrate is 22 commits ahead of each; keep for history, do not re-merge |

### Junk / never merge as-is

- **`ee1d7ae` ("placeholder test")** — trivial test shuffle; squash or drop on integrate cleanup.
- **`worktrees/`** — stale local git worktrees (`env-scanning`, `full-platform`); not tracked; delete locally.
- **`examples/private/`** (untracked) — private stress harness; keep out of main unless deliberately productized.
- **`scripts/smoke-stack.sh`** (untracked) — useful locally; commit with integrate or leave untracked; not a merge blocker.
- **Stale topic branches** — `add-providers`, `gui`, `ship-infra`, `env-scanning`, `full-platform-example`, `edge-scm-to-api-only`: audit vs integrate; archive if fully absorbed.

### `stack/03-viz-status-wiring` carried too much

Original stack/03 diff bundles **three stacks in one PR**: provider probes (01), credentials + env sync + auth CLI (02), and viz API + web wiring (03). Makes review/revert painful. On future stacks: **one concern per branch** (status API, viz wiring, auth separately).

---

## Known limitations

| Area | Gap vs spec |
|------|-------------|
| **Deployable status** | `internal/stackstatus/collect.go` returns `status_source: placeholder` for all `deployable: true` nodes — no live Vercel/Render/Supabase host health yet. Web deployments tab and TUI detail show empty `last_deploy` / errors. |
| **Read-only status** | SaaS/AI nodes probe when credentialed; otherwise `app_env` or `unchecked`. No incidents/status-page integration. |
| **`auto_setup` (logs)** | Last-resort in `internal/stacklogs/setup.go`: install CLI via npm/brew/apt, run `auth_cmd` subprocess, persist token from auth file/env. **Not browser OAuth**; headless/CI fails; Supabase/Render need manual token or `perch logs setup`. |
| **Log streaming** | API fetch only (no `--tail` stream); Render/Supabase/Vercel pagination TODOs in `render.go`, `supabase.go`, `vercel.go`. |
| **Web UI** | Tabs: deployments + logs only; **env tab removed** (`8435303`). In-panel token paste for logs connect. |
| **TUI** | Graph + status refresh work; `l` hints only; **`e` / `d` / `t` are roadmap stubs** in `internal/tui/panels.go`. |
| **`perch init`** | Scaffold from files/deps; no resource picker, reauth flow, `perch link`, or dev port scan from spec. |
| **Missing commands** | deploy, rollback, env *, snapshot, trace, incidents, costs, ps, pr, blame, timeline, tail, llm *, provider marketplace, `--for-agent` richness tied to real deploy data. |
| **Agent context** | `perch context` merges graph + status but inherits placeholder deployable health and no log excerpts from unhealthy nodes. |
| **Examples** | `full-platform` scenario is large; `full-stack` path referenced in UI copy may confuse vs `full-platform`. |

---

## Recommended next stack (after integrate → main)

1. **stack/08-deployable-status** — Wire provider `status` endpoints for Vercel/Render/Supabase; populate `last_deploy`, error_rate, recent_errors; unblocks deployments tab + context.
2. **stack/09-tui-logs** — Tail logs in Bubbletea (`l` → stream or poll); reuse `internal/stacklogs` + custom shell paths.
3. **stack/10-env-tab** — `perch env list` (masked) + web env tab + optional `perch config sync-env` polish; restore spec `e` keybinding.
4. **stack/11-deploy-rollback** — `perch deploy` / `perch rollback` via provider CLI/API; TUI `d` panel; auto-snapshot deferred.

Defer: trace, snapshots, LLM costs, provider registry, MCP (spec V2/non-goals).

---

## Files to review first for cleanup

| Priority | Path | Why |
|----------|------|-----|
| 1 | `internal/stackstatus/collect.go` | Central placeholder for deployable health — root cause of empty UI deploy data. |
| 2 | `internal/stacklogs/setup.go`, `auth.go` | auto_setup semantics, token persistence, strategy ordering. |
| 3 | `web/src/components/DetailPanel.jsx` | Logs connect UX, deployment empty states, removed env tab. |
| 4 | `internal/cli/viz.go`, `viz_api_test.go` | HTTP API surface; grows with every feature — split handlers before next stack. |
| 5 | `internal/tui/panels.go`, `model.go` | Stub vs real panels; align keybindings with spec or trim palette text. |
| 6 | `internal/config/envinfer.go`, `envsync.go` | Placeholder + `.env` inference; `local-dev` treated as placeholder — verify example scenarios. |
| 7 | `internal/cli/root.go` | Command tree drift vs spec; document intentional omissions. |
| 8 | `examples/scenarios/full-platform/` | Size vs value; overlap with `full-stack` / `manual-cli-test`. |
| 9 | `providers/hosting/*.yaml` | Status endpoint coverage for stack/08. |
| 10 | Merge commit messages on integrate | Squash "merge: logs …" + fixups into readable history before main. |

---

*Generated on integrate/platform-and-logs. Revisit after merge to main.*
