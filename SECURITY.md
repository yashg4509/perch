# Security

## Supported versions

Security fixes are applied to the latest release on the default branch (`main`). There is no long-term support policy for older major versions yet.

## Reporting a vulnerability

Please **do not** open a public issue for security reports.

- Use [GitHub Security Advisories](https://github.com/yashg4509/perch/security/advisories/new) for this repository if you have access, or
- Email the maintainers with enough detail to reproduce the problem (steps, affected version or commit, impact).

We aim to acknowledge reports within a few business days. perch is a local CLI: many issues are limited to the user’s machine, but we still take mis-parsing, credential handling, and supply-chain concerns seriously.

## Scope

In scope: the `perch` binary, this repository’s code, and bundled provider YAML under `providers/`. Out of scope: third-party platforms (Vercel, OpenAI, etc.) and user-written `perch.yaml` / shell commands in `custom` nodes—those run with the user’s privileges.
