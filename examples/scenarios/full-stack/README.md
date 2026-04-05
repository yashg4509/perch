# full-stack scenario

Pre-generated `perch.yaml` with **production**, **staging**, and **dev**. Use this to test `perch status`, `perch graph`, `perch context`, and the root TUI without running `perch init`.

## Local dev health check

The **dev** `web` node runs a shell `status` command that curls `http://127.0.0.1:18080/health`.

In one terminal:

```bash
python3 scripts/dev_health_server.py
```

In another (from this directory):

```bash
/path/to/perch --env dev status --json
```

Without the server, the `web` node reports unhealthy for **dev**; **production** / **staging** behavior is unchanged.

## Full setup

See [../../README.md](../../README.md) (build binary, optional `PERCH_PROVIDERS_DIR`, optional Vercel/OpenAI credentials).
