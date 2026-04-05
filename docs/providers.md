# Adding and extending providers

Providers are **one YAML file per platform** under [`providers/`](../providers/). The runtime substitutes placeholders and can call HTTP APIs or shell out to CLIs.

## Agent-assisted workflow

To use a **coding agent** (Cursor, Claude Code, Copilot, OpenCode, Codex, etc.) to gather official CLI/API details and draft YAML from vendor docs, follow **[add-perch-provider-skill.md](./add-perch-provider-skill.md)** (copy-paste prompt, validation commands, agent-agnostic setup).

## Steps

1. Copy [`providers/_template.yaml`](../providers/_template.yaml) to `providers/<name>.yaml` and fill in `name`, `category`, `deployable`, `cli` / `api`, and `credentials` as needed.
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

Placeholder syntax (`{project}`, `{token}`, auth headers, endpoint string forms) is implemented in `internal/provider`; match patterns used in [`providers/vercel.yaml`](../providers/vercel.yaml) and [`providers/openai.yaml`](../providers/openai.yaml).

## Concepts

Product-level ideas (nodes, edges, environments) are in the [root README](../README.md). For YAML shape, use bundled providers and `internal/providerspec` tests as the source of truth.
