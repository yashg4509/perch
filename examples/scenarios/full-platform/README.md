# Perch Brief

A small **team knowledge + AI Q&A** app wired to a **20-vendor** stack in [`perch.yaml`](perch.yaml) (Clerk, Neon, Pinecone, Stripe, Pusher, Sentry, PostHog, …). **Perch** is the point: one graph and `perch status` across the whole thing.

**You only need 5 API signups to run the app the first time** (§0A). The other integrations are still in the graph — Perch shows them as **unconfigured** until you add keys (§0B). That’s intentional: you see immediate Perch value on a realistic stack without signing up for everything on day one.

---

## 5-minute path

### 0. Accounts

```text
perch.yaml (20 nodes)          What you need when
─────────────────────          ───────────────────
§0A — 5 vendors + local jobs   First run: sign in, snippets, Ask
§0B — 13 more vendors          Optional: rate limits, live updates, billing, email, observability…
web / api (custom in dev)      No keys — npm run dev
```

#### §0A — Minimum to run the app (`npm run setup` checks these)

| Service | Get keys | Node |
|---------|----------|------|
| [Clerk](https://dashboard.clerk.com) | `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY`, `CLERK_SECRET_KEY` | `auth` |
| [Neon](https://console.neon.tech) | `DATABASE_URL` (`…neon.tech…`) | `postgres` |
| [OpenAI](https://platform.openai.com/api-keys) | `OPENAI_API_KEY` (billing for embeddings) | `openai` |
| [Pinecone](https://app.pinecone.io) | `PINECONE_API_KEY`, `PINECONE_INDEX` | `vectors` |
| [Anthropic](https://console.anthropic.com) | `ANTHROPIC_API_KEY` (or `ANSWER_MODEL=openai`) | `anthropic` |

**Jobs:** `INNGEST_DEV=1` in `.env.example` — local Inngest dev server, no cloud keys (`jobs` node).

#### §0B — Rest of the benchmark (Perch + app when you add keys)

Run `perch graph --env dev` anytime — all nodes appear. Add env vars below + `perch auth sync-env` to move nodes from **unconfigured** → **configured**.

| Node | Provider | Env vars (see `.env.example`) | What it does |
|------|----------|-------------------------------|--------------|
| `web` | custom / vercel | — (dev) · Vercel token for hosted | Next.js UI |
| `api` | custom / railway | — (dev) · Railway token for hosted | Hono API |
| `redis` | upstash | `UPSTASH_REDIS_REST_URL`, `UPSTASH_REDIS_REST_TOKEN` | Ask rate limits |
| `realtime` | pusher | `PUSHER_*`, `NEXT_PUBLIC_PUSHER_*` | Live “snippet ready” |
| `billing` | stripe | `STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET`, `STRIPE_PRICE_ID` | Pro checkout |
| `email` | resend | `RESEND_API_KEY`, `RESEND_FROM_EMAIL` | Weekly digest cron |
| `errors` | sentry | `SENTRY_DSN`, `NEXT_PUBLIC_SENTRY_DSN` | Error tracking |
| `analytics` | posthog | `NEXT_PUBLIC_POSTHOG_KEY` | Product analytics |
| `logs` | logtail | `LOGTAIL_SOURCE_TOKEN` | API logs (Better Stack) |
| `traces` | langsmith | `LANGCHAIN_API_KEY`, `LANGCHAIN_PROJECT` | LLM tracing |
| `media` | cloudinary | `CLOUDINARY_*` | Avatar URLs |
| `queue` | upstash-qstash | `QSTASH_*` signing keys | Webhook verify demo |
| `metrics` | datadog | `DD_API_KEY`, `DD_SITE` | APM (host env) |
| `repo` | github | `GITHUB_TOKEN` (Perch / deploy meta) | Repo linkage in graph |

Hosted envs (`preview` / `staging` / `production`) use the same node names with Vercel/Railway/Neon project ids in `perch.yaml` — fill when you deploy.

### 1. Perch CLI (from perch repo root, once)

```bash
go build -o perch ./cmd/perch
export PATH="$(pwd):$PATH"
```

### 2. App env

```bash
cd examples/scenarios/full-platform
npm run setup    # copies .env + .env.local, checks required keys
# Edit .env until `npm run setup` passes
npm install
```

### 3. Run

```bash
npm run dev      # API :4000 + web :3000 + Inngest dev server
```

Open **http://localhost:3000** → sign in → add a snippet → wait for **Ready** → **Ask**.

### 4. Perch (same folder, ~30 seconds)

```bash
perch auth sync-env
perch config sync-env --env dev   # optional: project ids from .env → perch.yaml
perch graph --json --env dev
perch status --env dev
perch --env dev
```

**What Perch shows:** all **20 nodes** in one graph. `perch status` summarizes then groups: **up** (local health + vendor API), **down**, **skipped** (optional, no keys in `.env`), etc. Vendors you use are probed **in parallel** when you have `auth sync-env` keys — not the same as “app broken.”

---

## Try it — paste one snippet

**Title:** What is Perch Brief  
**Body:** Perch Brief is the full-platform Perch example. Snippets live in Postgres, OpenAI embeds to Pinecone, Anthropic answers Ask. The stack is in perch.yaml: web, api, clerk, inngest, pinecone, and more.

**Ask:** `What stack does this app use?`

More ideas: *What port is the API?* · *How do we ship to prod?* (add your own snippets)

---

## Commands

| Command | What it does |
|---------|----------------|
| `npm run setup` | Create `.env` files + validate required keys |
| `npm run dev` | Start API, web, and Inngest together |
| `npm run dev:api` / `dev:web` / `dev:inngest` | Run one service |
| `perch auth sync-env` | Import `.env` → `~/.perch/credentials` |
| `perch config sync-env --env dev` | Optional: persist inferred project ids into `perch.yaml` |

---

## If something breaks

| Problem | Fix |
|---------|-----|
| `npm run setup` lists missing keys | Fill §0A only; §0B is optional for the app |
| Snippets stuck **Pending** | OpenAI quota; run `npm run dev` (Inngest must be up); check API terminal |
| `ECONNREFUSED` :5432 | Use a Neon `DATABASE_URL`, or `docker compose up -d` + local URL in `.env` |
| `perch: command not found` | Build CLI (step 1) and `export PATH` to repo root |
| Page unstyled | Restart `npm run dev:web` |

---

## What’s in this folder

| Path | Role |
|------|------|
| `apps/web` | Next.js UI (Clerk) |
| `apps/api` | Hono API |
| `perch.yaml` | 20-node graph · `dev` / `preview` / `staging` / `production` |
| `perch.envmap.yaml` | `.env` → `perch.yaml` for `config sync-env` |
| `.env.example` | All env vars (§0A required, §0B optional) |

**Perch maintainers:** regression spec at [`docs/perch-brief-benchmark.md`](../../../docs/perch-brief-benchmark.md).
