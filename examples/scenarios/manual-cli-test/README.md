# manual-cli-test scenario

End-to-end **manual** exercise for `perch graph`, `perch status`, `perch context`, and the **TUI**, with **two local HTTP health checks** in `dev` so both custom nodes can go green without cloud accounts.

## What is in this folder

| File | Role |
|------|------|
| `perch.yaml` | Three environments (`production`, `staging`, `dev`). **dev** uses `custom` + `curl` to `127.0.0.1:18081` and `:18082`. Prod/staging use placeholder Vercel + read-only OpenAI (stubbed in the current milestone). |
| `scripts/local_stack.py` | Tiny dual-port `/health` servers for those curls. |

## One-time setup (from repository root)

```bash
cd /path/to/perch
go build -o perch ./cmd/perch
# or: make build   →  binary at bin/perch
```

Optional while editing provider YAML without rebuilding:

```bash
export PERCH_PROVIDERS_DIR=/path/to/perch/providers
```

## Third-party setup

**None for the recommended path.** Stick to **`--env dev`** plus **`local_stack.py`**: no cloud accounts, no API keys, and **no changes** to `perch.yaml` unless ports **18081** / **18082** are busy and you switch to `PERCH_MANUAL_*_PORT` (then update the dev `curl` lines in `perch.yaml` to match).

**Production / staging** in `perch.yaml` use placeholder Vercel project names and OpenAI as a read-only node. You can run `perch graph` / `perch status` there without signing up anywhere; do **not** expect real deploy health until you replace placeholders and configure credentials per the root README.

## Terminal A — fake “web + API” locally

From **this directory** (`examples/scenarios/manual-cli-test`):

```bash
python3 scripts/local_stack.py
```

Leave it running. You should see listeners on ports **18081** (web) and **18082** (api).

## Terminal B — run perch

Still from **this directory** (`examples/scenarios/manual-cli-test`). The repo root is **three** levels up (`../../../`), not two—`../../perch` would look under `examples/` and fail.

- After **`go build -o perch`** at the root: **`../../../perch`**
- After **`make build`**: **`../../../bin/perch`**

Below, `P` is whichever path exists on your machine; or put `perch` on your `PATH`.

```bash
P=../../../perch          # or: P=../../../bin/perch

# Topology (human and JSON)
"$P" graph
"$P" graph --json

# Dev: both custom nodes should be healthy while local_stack.py runs
"$P" --env dev status
"$P" --env dev status --json

# Merged view for scripts / agents
"$P" --env dev context --json
"$P" --env dev context --for-agent

# Interactive TUI: graph + health colors + footer. Keys: ? palette · E next env · arrows · q quit
"$P"
```

**Expected while local stack is up**

- `"$P" --env dev status --json` includes `web` and `api` with `"healthy": true`.

**Expected if you stop `local_stack.py`**

- Those two nodes should flip to **unhealthy** (curl fails).

**Production / staging**

- `web` is a Vercel deployable placeholder (`YOUR_VERCEL_PROJECT`) — status may show unhealthy until you point at a real project and the runtime is fully wired.
- `api` is OpenAI (read-only) — current milestone uses a **stub** that reports healthy without calling the network.

## Quick sanity checklist

- [ ] `graph --json` lists `web`, `api`, and edge `web -> api`.
- [ ] `dev` + running script → both nodes healthy.
- [ ] `dev` + stopped script → both nodes unhealthy.
- [ ] `context --for-agent` prints a single blob you could paste into an LLM.

## See also

- Parent guide: [../../README.md](../../README.md)
- Init detection only: [../init-signals](../init-signals)
- Single-port example: [../full-stack](../full-stack)
