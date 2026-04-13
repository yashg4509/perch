# perch

Terminal UI and CLI for **multi-service deployment stacks**: one keyboard-driven view of your services and health, plus **`--json`** and **agent-oriented** output for scripts and LLMs.

**License:** [MIT](LICENSE) · **Contributing:** [CONTRIBUTING.md](CONTRIBUTING.md) · **Security:** [SECURITY.md](SECURITY.md) · **Changelog:** [CHANGELOG.md](CHANGELOG.md) · **Code of conduct:** [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)

---

## What it does

Modern apps often span many platforms (hosting, DB, auth, payments, AI APIs). perch gives you a **single binary** that:

- **Humans** — explore the stack interactively (run `perch` with no arguments for the TUI).
- **Automation** — run `perch status`, `perch graph`, or `perch context` with **`--json`** or **`context --for-agent`** so tools and models get a structured picture of the same stack.

There is **no perch cloud service**; the CLI reads your repo config and may call vendor CLIs or APIs when a provider supports it.

---

## Install

**Homebrew** (third-party tap; use **`--HEAD`** until the first stable release):

```bash
brew tap yashg4509/perch
brew install --HEAD perch
```

**Source:** [github.com/yashg4509/perch](https://github.com/yashg4509/perch)

## From source

```bash
go build -o perch ./cmd/perch
./perch --help
```

Requires the Go version in [`go.mod`](go.mod).

---

## How it works

1. **Config** — Looks for **`perch.yaml`** starting in the current directory and walking upward.
2. **Providers** — Loads platform definitions from YAML **embedded in the binary**, or from `./providers` / **`PERCH_PROVIDERS_DIR`** when you are developing providers.
3. **Graph** — Builds a **node + edge** model for the selected **`--env`** (e.g. `production` vs `dev`).
4. **Commands** — Subcommands report **status**, **topology**, or **merged context**; running **`perch` alone** starts the **Bubbletea** TUI.
5. **Execution** — **`custom`** nodes run **shell commands** you declare (for local health checks, etc.); other providers use YAML-described CLI/API patterns as the runtime grows.

---

## Bundled providers

The binary embeds **35** platform YAML files under [`providers/`](providers/) in **category subfolders** ([`providers/README.md`](providers/README.md)), plus root `_template.yaml` for new entries. `perch init` can suggest nodes when it finds matching **config files** or **`package.json` dependencies** (see [docs/providers.md](docs/providers.md) for the full mapping). IDs are lowercase with hyphens where needed (e.g. `upstash-qstash`, `aws-s3`).

| Area | Provider IDs |
|------|----------------|
| **Hosting** | `vercel`, `netlify`, `fly`, `railway`, `cloudflare`, `render`, `firebase`, `supabase` |
| **Data / cache / storage** | `neon`, `postgres`, `planetscale`, `mysql`, `mongodb`, `upstash`, `upstash-qstash`, `redis`, `pinecone`, `aws-s3`, `cloudinary` |
| **Auth / payments / messaging** | `clerk`, `auth0`, `stripe`, `resend`, `sendgrid`, `twilio` |
| **Jobs / realtime** | `trigger`, `inngest`, `pusher` |
| **AI** | `openai`, `anthropic`, `langsmith` |
| **Observability / analytics / logging** | `sentry`, `posthog`, `datadog`, `logtail` |

**Special cases:** `custom` is not a YAML file—you declare `status` / optional `logs` commands in `perch.yaml`. **`next-auth`** is not auto-detected (self-hosted); **`langchain`** maps to **`langsmith`**. Secrets use keys named in each YAML’s `credentials` block under **`~/.perch/credentials`**.

---

## Core ideas

| Concept | Meaning |
|--------|---------|
| **Node** | One service in your stack (e.g. `web`, `api`). |
| **Edge** | Dependency or data flow, e.g. `web -> api` in `perch.yaml`. |
| **Provider** | How perch talks to a platform (Vercel, OpenAI, …), described in YAML under **`providers/`** (nested by category). |
| **Environment** | Named slice of the same node names (`production`, `staging`, `dev`) with different resource IDs or `custom` commands. |

---

## Your stack file (`perch.yaml`)

Committed at the **repo root** (no secrets in this file). Minimal shape:

```yaml
name: my-app
environments:
  production:
    web: { provider: vercel, project: my-app }
    api: { provider: openai }
  dev:
    web:
      provider: custom
      status: "curl -sf http://127.0.0.1:3000/health"
edges:
  - web -> api
```

Every environment must use the **same node names**; **edges** are global across environments.

---

## Commands (overview)

| Command | Role |
|---------|------|
| `perch` | Interactive TUI: **arrow keys** focus nodes, **Enter** detail, **`?`** command palette, **`r`** refresh health, **`l`/`e`/`d`/`t`** logs/env/deploy/timeline hints, **`E`** next environment (if several), **`q`** quit. |
| `perch init` | Scaffold / refresh `perch.yaml` using repo detection heuristics. |
| `perch status` | Per-node health-style signals (`--json` for machines). |
| `perch graph` | Topology (`--json` for machines). |
| `perch edge` | **list** / **add FROM TO** / **remove FROM TO** (aliases **rm**, **delete**) — edits `perch.yaml` edges without hand-editing; **`--dry-run`** prints canonical YAML. |
| `perch context` | Merged topology + status; `--for-agent` for plain text LLM context. |

Common flags: **`--env`**, **`--json`**, **`--no-color`**.

---

## Examples & manual testing

The **[examples/](examples/)** folder has copy-paste scenarios (including a **two-service local health** fixture). Start with **[examples/README.md](examples/README.md)**.

---

## Contributing

See **[CONTRIBUTING.md](CONTRIBUTING.md)** for Go version, **`make test` / `make lint`**, PR expectations, **providers** ([docs/providers.md](docs/providers.md)), and **releases / changelog** notes.

---

## Documentation map

| Where | Who it is for |
|-------|----------------|
| This **README** | Everyone: what perch is, how it fits together, install, config shape. |
| **CONTRIBUTING.md** | Build, test, PRs, providers, changelog, tagging. |
| **docs/** | [Bundled providers & init detection](docs/providers.md), [`providers/README.md`](providers/README.md) layout, and [agent skill](docs/add-perch-provider-skill.md) for drafting new YAML. |
| **examples/README.md** | Trying the CLI safely with sample projects. |
| **CHANGELOG.md** | Release notes. |
| **SECURITY.md** / **CODE_OF_CONDUCT.md** | Reporting issues and community expectations. |

This matches common open-source practice for small tools: **one strong README**, a **docs/** folder for contributor depth, and everything else in code and CI config.
