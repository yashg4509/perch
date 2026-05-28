# Bundled provider definitions

YAML files are grouped **by category** (flat lists were hard to scan). The `name` field inside each file is still the provider id used in `perch.yaml` and detection—**folder names are not provider ids**.

| Directory | Contents |
|-----------|-----------|
| **`hosting/`** | Deployable hosts: Vercel, Netlify, Fly, Railway, Cloudflare, Render, Firebase, Supabase |
| **`data/`** | Databases, cache, object/vector storage: Neon, Postgres, PlanetScale, MySQL, MongoDB, Upstash, Redis, Pinecone, AWS S3, Cloudinary |
| **`saas/`** | Auth, payments, email, SMS: Clerk, Auth0, Stripe, Resend, SendGrid, Twilio |
| **`workflows/`** | Jobs and realtime: Trigger.dev, Inngest, Pusher |
| **`ai/`** | LLM APIs: OpenAI, Anthropic, LangSmith |
| **`observability/`** | Sentry, PostHog, Datadog, Logtail (Better Stack) |

- **`_template.yaml`** — start here when adding a platform; place the new file in the **most fitting** directory (or ask in a PR).
- **`embed.go`** — `//go:embed` globs must list each category; add a new glob if you introduce another top-level folder.

Loaders (`internal/provider`) scan **`providers/` recursively**, so a project override directory can use this layout or a single flat `providers/*.yaml` file.

### Credentials and `.env`

Each provider may declare:

- **`credentials.env_aliases`** — app env var names for secrets. Run **`perch auth sync-env`** to import into `~/.perch/credentials` (one-way; never writes back to `.env`).
- **`project_env_aliases`** / **`service_env_aliases`** — non-secret resource ids (index name, app id, etc.). **`perch status`** and **`perch graph`** read these from the project `.env` at runtime when `perch.yaml` still has placeholders. **`perch config sync-env`** optionally persists them into `perch.yaml`.
- **`api.status_probe`** — `rest` (default): parallel GET on `api.endpoints.status`; `signed` (Pusher); `none` when auth is non-REST (e.g. Atlas digest).
- **`api.status_headers`** — extra headers on status GET (e.g. `anthropic-version`, `X-Pinecone-Api-Version`, `X-GitHub-Api-Version`). These are vendor **API contract versions**, not Perch release dates—check each provider’s docs before changing them.

Per-node escape hatch in `perch.yaml`: `env_project` / `env_service` name a custom `.env` key.
