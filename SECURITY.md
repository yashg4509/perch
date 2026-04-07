# Security

## Supported versions

Security fixes are applied to the latest release on the default branch (`main`). There is no long-term support policy for older major versions yet.

## Reporting a vulnerability

Please **do not** open a public issue for security reports.

- Use [GitHub Security Advisories](https://github.com/yashg4509/perch/security/advisories/new) for this repository if you have access, or
- Email the maintainers with enough detail to reproduce the problem (steps, affected version or commit, impact).

We aim to acknowledge reports within a few business days. perch is a local CLI: many issues are limited to the user’s machine, but we still take mis-parsing, credential handling, and supply-chain concerns seriously.

## Scope

In scope: the `perch` binary, this repository’s code, and bundled provider YAML under `providers/`. Out of scope: third-party platforms (Vercel, OpenAI, etc.).

## Threat model and data handling

**Secrets**

- `perch.yaml` must not contain API keys; credentials live under **`~/.perch/credentials`** (JSON, restrictive file permissions). Treat that file like `~/.ssh` — backup-aware, not world-readable.
- Init and detection intentionally **do not read `.env`** files to avoid pulling secrets into logs or generated config.

**Custom provider (`provider: custom`)**

- `status` and `logs` fields are executed via the system shell (`/bin/sh -c` on Unix, `cmd /C` on Windows) with the user’s environment. **Untrusted `perch.yaml` is equivalent to untrusted shell code.** Only use stacks you trust, or run perch from a dedicated user/VM.

**Provider HTTP runtime**

- Outbound API calls use a shared helper that **rejects scheme-relative paths** (`//host/...`) when joining `base_url` and endpoint paths, and **blocks HTTP redirects** that change scheme or host. That reduces SSRF and open-redirect style issues when calling platform APIs.
- Bundled `providers/*.yaml` is part of the supply chain: install perch from official releases or verify commits.

**Agent / JSON output**

- `perch context` is designed to expose topology and health, not raw credentials. If you add fields there, **never** include env var values, tokens, or unredacted connection strings.

**Subprocess execution**

- `pkg/exec` runs named binaries with argv slices (no shell) for normal integrations. Shell execution is limited to the custom-provider path above.

## Automated checks

CI runs **`govulncheck`** (known vulnerable call paths) and **`gosec`** (common Go anti-patterns). `gosec` excludes **G304** (reading paths inside the user’s project) because that is inherent to a CLI that loads `perch.yaml`, `package.json`, and local provider files.

The module pins a **`toolchain`** in `go.mod` so CI can use a Go release with current stdlib security fixes while keeping a minimum language version.

Locally: `make security` (or the commands in [`.cursor/docs/ci.md`](.cursor/docs/ci.md)).

## What is not guaranteed (V1)

- Credentials on disk are **not** encrypted with an OS keychain (future hardening).
- Log lines and API error bodies echoed by platforms could theoretically contain sensitive data; perch truncates some error text but does not redact arbitrary payloads.
