---
name: add-perch-provider
description: Researches a cloud or SaaS platform and drafts a new perch provider YAML (CLI commands, REST endpoints, credentials) from vendor docs. Use when adding providers/perch integrations, extending providers/*.yaml, or when the user mentions contributing a provider, platform YAML, or API token mapping for perch.
---

# Add a perch provider (agent workflow)

## Goal

Produce **one new file** `providers/<name>.yaml` that matches the schema used by this repo, without inventing URLs or CLI flags. Every path and command string should be **traceable to an official doc** (or the vendor CLI `--help`).

## Before writing YAML

1. Read [docs/providers.md](./providers.md) and copy [`providers/_template.yaml`](../providers/_template.yaml) → `providers/<name>.yaml`.
2. Skim [`providers/vercel.yaml`](../providers/vercel.yaml) or [`providers/openai.yaml`](../providers/openai.yaml) for shape and tone; use **Runtime** notes in [providers.md](./providers.md) for placeholders and HTTP/CLI wiring.

## Research checklist (use the web or local CLI help)

Gather **evidence**, not guesses:

| Topic | What to capture |
|------|------------------|
| **CLI** | Official install name, exact subcommands for logs, deploy, env list, project list, JSON output flags. |
| **REST** | Base URL, auth scheme (Bearer, API key header), and **verbatim** paths from API reference. |
| **Credentials** | Where users create keys in the dashboard; stable name for `credentials.key` (snake_case). |
| **Semantics** | Whether the service is **deployable** (logs/env/deploy) vs **read-only** (status/usage only); set `deployable` / `llm` accordingly. |

If the public docs disagree with the CLI, **prefer the doc URL you cite** and note the conflict in the PR.

## Output expectations

- Fill `name`, `category`, `deployable` (and `llm: true` only when modeled like OpenAI-style usage).
- Under `cli.commands`, only list commands that exist; use placeholders the runtime already substitutes (see [providers.md](./providers.md)).
- Under `api.endpoints`, use the same string form as existing providers, e.g. `GET /v1/...` with `{placeholders}` matching template rules.
- Under `credentials`, set `key` and a short `prompt` telling humans where to mint the token.

## Validate in-repo

From the repository root:

```bash
make provider-validate
go test ./internal/provider/...
```

Add focused tests with `httptest` and small JSON fixtures when adding HTTP parsing (see `internal/provider/vercel_dispatch_test.go`).

## Optional: example scenario

If the provider should be easy to demo without secrets, add or extend a row under [examples/README.md](../examples/README.md) and a tiny `perch.yaml` snippet (placeholders only, no secrets).

## Copy-paste prompt (for the user)

Use this in a new session with **any** coding agent after opening this repository:

> Follow the instructions in `docs/add-perch-provider-skill.md` (read that file in full). I want a new provider named **`<platform>`**. Search official docs for REST base URL, authentication, and the smallest set of endpoints for **status** (and **logs** if deployable). Search for the official CLI and commands for **status**, **logs**, **deploy**, **list projects**, and **env list**. Draft `providers/<platform>.yaml`, then list which doc URLs you used for each major field. Do not invent paths.

## Using this workflow in your environment (agent-agnostic)

**Works everywhere**

- Paste the prompt above (with `<platform>` filled in).
- Ensure the model **reads** this file: attach it, use your client’s “@ file” / “include file” / “add to context” action, or open `docs/add-perch-provider-skill.md`.

**Optional: register as a reusable project instruction**

| Client / family | Typical approach |
|-----------------|------------------|
| **Cursor** | Project skill dir with this file, or @‑mention `docs/add-perch-provider-skill.md` per task. |
| **Claude Code** | Reference in `CLAUDE.md`: “For new providers, follow `docs/add-perch-provider-skill.md`.” |
| **GitHub Copilot (VS Code)** | Workspace instructions pointing at the same path, or attach when starting provider work. |
| **OpenCode, Codex CLI, other agents** | Use that product’s rules / context list to include `docs/add-perch-provider-skill.md`, or paste the **Copy-paste prompt** plus key sections. |

Example symlink for **Cursor-style** skills (local only; do not commit unless your project allows):

```bash
mkdir -p .cursor/skills/add-perch-provider
ln -sf ../../../docs/add-perch-provider-skill.md .cursor/skills/add-perch-provider/SKILL.md
```

The canonical copy lives under **`docs/`** in this repository.
