# Perch Brief — benchmark spec (maintainers)

Reference app: [`examples/scenarios/full-platform/`](../examples/scenarios/full-platform/).

This scenario is the **regression benchmark** for Perch: 18 nodes, 4 environments, committed edges. User-facing setup is only [`README.md`](../examples/scenarios/full-platform/README.md).

## CLI checklist (dev, no deploy)

```bash
cd examples/scenarios/full-platform
perch graph --json --env dev
perch status --env dev --json
perch context --for-agent --env dev
perch --env dev
perch edge list
```

## Node tiers

- **P0:** web, api, postgres, auth, vectors, openai, anthropic, jobs
- **P1:** redis, realtime, billing, email, repo
- **P2:** errors, analytics, logs, queue, media, traces, metrics

## Environments

Same node names in `dev`, `preview`, `staging`, `production`. `dev` uses `custom` curl health for web/api; hosted envs use Vercel/Railway placeholders until filled.

## Edges

```text
web  → api, auth, analytics, errors, repo
api  → postgres, redis, vectors, openai, anthropic, traces,
       billing, media, queue, email, jobs, realtime
jobs → email
```

## Credential keys

See `providers/**/*.yaml` and `perch auth sync-env` from the scenario `.env`.
