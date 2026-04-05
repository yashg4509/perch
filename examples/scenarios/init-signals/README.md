# init-signals scenario

**Detection-only:** this folder has `vercel.json` and `package.json` dependencies (`@supabase/supabase-js`, `stripe`, `openai`) so `perch init` can infer providers and **edges**.

`perch.yaml` is **gitignored** here so the repo stays the source of truth for signals; generate it locally:

```bash
/path/to/perch init --name perch-init-demo
```

Then inspect `perch.yaml` and run `perch graph --json` from this directory.

Full instructions: [../../README.md](../../README.md).

## Is a minimal `perch.yaml` “enough”?

**Schema-wise, yes.** Per validation, each environment needs at least one node, and every node needs a `provider`. Optional fields depend on the provider:

| Field | When it matters |
|-------|-----------------|
| `project` | Common placeholder `CHANGE_ME` for deployable hosts (Vercel, Netlify, …) until you set the real project id. |
| `service` | Same idea for Render / Railway / Fly-style providers. |
| `status` / `logs` | Required for `provider: custom` only. |

Read-only style nodes (e.g. OpenAI, Stripe in the spec) often need **no extra YAML keys** until real API wiring lands; the graph still needs a **provider definition** in the registry for `perch graph` / TUI (bundled providers in this repo may be a subset — use `PERCH_PROVIDERS_DIR` or add YAML under `./providers` when you need more platforms).

## Edges: automatic vs manual

**Automatic (today):** `perch init` runs **deterministic inference** from `package.json`:

- `@supabase/supabase-js` → edge from your **primary connector** node to **supabase**  
  - Connector = first **Render / Railway / Fly** node if present, otherwise **Vercel / Netlify / Cloudflare / Firebase**.
- `openai` dependency → connector → **openai** node.
- `stripe` dependency → connector → **stripe** node.

So a Vercel + Supabase app gets `frontend -> db` without hand-writing edges.

**Manual (CLI):**

```bash
perch edge list
perch edge add frontend db
perch edge remove frontend llm    # also: rm, delete, del
perch edge add api db --dry-run   # preview canonical YAML on stdout
```

This rewrites `perch.yaml` in **canonical form** (sorted keys; **comments are not preserved**).

**Manual (YAML):** you can still edit the `edges:` list by hand.

**Product follow-ups** (not implemented yet): an explicit `edges: auto` flag, `perch init --edges-only` to refresh just edges, or `perch graph suggest-edges` to print a diff — say the word if you want one of these.
