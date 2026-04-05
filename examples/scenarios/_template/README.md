# New scenario template

Use this checklist when adding `examples/scenarios/<your-scenario>/`.

## Checklist

1. **Purpose** — One sentence: what manual flow does this exercise (init only, full stack, custom nodes, etc.)?
2. **Inputs** — Either a committed `perch.yaml` or only detection files (`package.json`, `vercel.json`, …) for `perch init`.
3. **Environments** — If you use multiple `--env` values, every environment must declare the **same node names** (edges are global in `perch.yaml`).
4. **Secrets** — Only placeholders in committed files; document optional real keys in the parent [examples/README.md](../../README.md#third-party-services-optional).
5. **Ignore generated files** — If the scenario runs `perch init`, add `perch.yaml` to `.gitignore` in that scenario so contributors do not commit local output.
6. **Register** — Add a row to the scenarios table in [examples/README.md](../../README.md).

## Files you might add

| File | When |
|------|------|
| `perch.yaml` | Pre-baked stack for `status` / `graph` / `context` / TUI |
| `package.json` | NPM-based provider detection |
| `vercel.json` | Vercel signal |
| `scripts/*.py` or `*.sh` | Local mocks for `custom` `status` / `logs` |
| `.gitignore` | Ignore `perch.yaml` for init-only folders |
