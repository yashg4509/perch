# Adding and extending providers

Providers are **one YAML file per platform** under [`providers/`](../providers/), grouped into **category subfolders** (see [`providers/README.md`](../providers/README.md)). The runtime substitutes placeholders and can call HTTP APIs or shell out to CLIs.

## Bundled standard providers

The repository ships **35** platform definitions (every `**/*.yaml` under `providers/` except [`_template.yaml`](../providers/_template.yaml) at the package root). They are **embedded in the binary** via [`providers/embed.go`](../providers/embed.go) unless you override with a local `./providers` directory or **`PERCH_PROVIDERS_DIR`**. [`LoadRegistry`](../internal/provider/registry.go) walks **`providers/` recursively**, so overrides can be flat files, nested, or mixed.

Each file defines at minimum: `name`, `category`, `deployable`, `api` (`base_url`, `auth_header`, `endpoints.status`), and `credentials` (`key`, `prompt`). Deployable providers also require `cli` (`binary`, `commands.status`). Some vendors need **extra headers or auth** at runtime (for example Anthropic’s API version header, Pinecone’s API version header, Redis Cloud’s second secret key, AWS SigV4); those limits are called out in YAML comments or `credentials.prompt` where relevant.

## Init detection (`perch init`)

Detection is **committed files only** (no `.env` scanning):

| Source | File / signal | Maps to provider id(s) |
|--------|----------------|-------------------------|
| Config files | [`internal/detect/configfiles.go`](../internal/detect/configfiles.go) | `vercel.json` → `vercel`, `netlify.toml` → `netlify`, `fly.toml` → `fly`, `railway.toml` / `railway.json` → `railway`, `wrangler.toml` → `cloudflare`, `render.yaml` → `render`, `firebase.json` → `firebase`, `prisma/schema.prisma` → `postgres` / `mysql` / `mongodb`, `docker-compose.yml` → `custom` |
| Dependencies | [`internal/detect/packagejson.go`](../internal/detect/packagejson.go) | NPM package names → provider ids (e.g. `@supabase/supabase-js` → `supabase`, `stripe` → `stripe`, `langchain` → **`langsmith`**) |

**Not detected:** `next-auth` — use explicit SaaS SDKs (`@clerk/nextjs`, `@auth0/auth0-react`, …) if you want auth nodes in the graph.

Scaffold logic in [`internal/scaffold/`](../internal/scaffold/) assigns **node names** (e.g. `frontend` for Vercel-class hosting) and writes **`project`** or **`service`** placeholders for deployable platforms. Some **read-only** platforms need a resource id in URLs; those get tailored `project:` hints in generated YAML (see `nodeYAML` in [`internal/scaffold/init.go`](../internal/scaffold/init.go)).

## Agent-assisted workflow

To use a **coding agent** (Cursor, Claude Code, Copilot, OpenCode, Codex, etc.) to gather official CLI/API details and draft YAML from vendor docs, follow **[add-perch-provider-skill.md](./add-perch-provider-skill.md)** (copy-paste prompt, validation commands, agent-agnostic setup).

## Steps

1. Copy [`providers/_template.yaml`](../providers/_template.yaml) to `providers/<category>/<name>.yaml` (pick `hosting`, `data`, `saas`, `workflows`, `ai`, or `observability`—see [`providers/README.md`](../providers/README.md)) and add a matching `//go:embed` line in [`embed.go`](../providers/embed.go) if you create a **new** top-level category folder. Fill in `name`, `category`, `deployable`, `cli` / `api`, and `credentials` as needed.
2. Run validation:

   ```bash
   make provider-validate
   ```

3. Add tests: prefer `httptest.Server` and JSON fixtures under `internal/provider/testdata/` (see `vercel_dispatch_test.go`).
4. Update [examples/](../examples/) if the provider should appear in manual scenarios.

## Provider runtime (for implementers)

1. **YAML** — `internal/provider` parses into `Spec`; stricter rules live in `ParseProviderYAML` / `validateSpec`.
2. **Execution** — Use `pkg/exec.Run` for CLI templates from `Spec.CLI.Commands` after `SubstitutePlaceholders` with node fields + credentials.
3. **HTTP** — Use `DoGETJSON` (or the same substitution + `ApplyAuthHeader` pattern) for API endpoints; add non-GET verbs when a provider needs them.
4. **Tests** — Prefer `httptest.Server` + fixture JSON; keep `ReadOnlyStub` for pure unit cases.

Placeholder syntax (`{project}`, `{token}`, auth headers, endpoint string forms) is implemented in `internal/provider`; match patterns used in [`providers/hosting/vercel.yaml`](../providers/hosting/vercel.yaml) and [`providers/ai/openai.yaml`](../providers/ai/openai.yaml).

## Init detection note (auth)

`package.json` detection does **not** map `next-auth` to a provider: NextAuth is self-hosted and has no single vendor API. Infer auth nodes from concrete SaaS SDKs instead (for example `@clerk/nextjs`, `@auth0/auth0-react`). The `langchain` package is mapped to the **LangSmith** provider id (`langsmith`) for observability APIs (see **Init detection** above).

## Concepts

Product-level ideas (nodes, edges, environments) are in the [root README](../README.md). For YAML shape, use bundled providers and `internal/providerspec` tests as the source of truth.
