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
