# Contributing to perch

Thanks for helping improve perch.

Start with the **[README](README.md)** for what perch does and how config, providers, and commands fit together.

**Note:** The `.cursor/` directory is **gitignored** (local Cursor/IDE state). If an older clone still has `.cursor` tracked in git, run `git rm -r --cached .cursor` once after pulling.

## Quick start

1. Install [Go](https://go.dev/dl/) (use the version in `go.mod`, currently 1.24.x).
2. Clone the repository and run from the module root:

   ```bash
   make test
   make lint
   ```

3. Build the binary:

   ```bash
   go build -o perch ./cmd/perch
   ```

## Development guidelines

- **Tests:** Prefer test-driven changes: extend or add `_test.go` files, then implement until `go test ./...` is green. CI must pass before merge.
- **Formatting:** Match `gofmt` / CI (see `.github/workflows/ci.yml`).
- **Scope:** Keep pull requests focused—one feature or fix per PR when practical.
- **Providers:** See **[docs/providers.md](docs/providers.md)** and the agent workflow in **[docs/add-perch-provider-skill.md](docs/add-perch-provider-skill.md)** (any coding agent: Cursor, Claude Code, Copilot, OpenCode, Codex, …).

## Where to change what

| Goal | Start here |
|------|------------|
| New CLI flag or command | `internal/cli/` |
| `perch.yaml` schema or validation | `internal/config/` |
| New platform YAML | `providers/`, then `make provider-validate` |
| Provider HTTP / template behavior | `internal/provider/` (see **Runtime** in [docs/providers.md](docs/providers.md)) |
| TUI layout or keys | `internal/tui/` |
| Init detection heuristics | `internal/detect/`, `internal/scaffold/` |

## Project layout (high level)

| Path | Role |
|------|------|
| `cmd/perch` | CLI entrypoint |
| `internal/` | Application code (not a stable public library API) |
| `pkg/exec` | Shared process runner (intended for reuse) |
| `providers/` | Provider YAML definitions |
| `docs/` | Contributor-facing notes (providers, agent skill) |
| `examples/` | Manual test fixtures—see [examples/README.md](examples/README.md) |

## Manual testing

Use the [examples/](examples/) scenarios to exercise `perch init`, `perch status`, `perch graph`, `perch context`, and the interactive TUI without touching your real projects.

## Changelog & releases

We follow **[Keep a Changelog](https://keepachangelog.com/en/1.1.0/)** in [CHANGELOG.md](CHANGELOG.md). For **user-visible** changes, add a bullet under **`[Unreleased]`** (the PR template reminds you).

**Automation options** (when you outgrow manual edits): generate notes with **[git-cliff](https://github.com/orhun/git-cliff)** from [Conventional Commits](https://www.conventionalcommits.org/), or use **[release-please](https://github.com/googleapis/release-please)** on GitHub. Hand-editing `[Unreleased]` is fine until release cadence picks up.

**Tagging a release** (maintainers): confirm `go test ./...`, finalize `CHANGELOG.md` (rename `[Unreleased]` to a dated version), create an annotated tag, then run [GoReleaser](https://goreleaser.com/) per [`.goreleaser.yaml`](.goreleaser.yaml) (e.g. `goreleaser release --clean` with `GITHUB_TOKEN` for assets and the Homebrew tap).

## Private planning helper (optional)

If you keep a **personal** task checklist outside this repo, you can poll until a line is checked with:

```bash
PLAN=~/path/to/your-plan.md ./scripts/wait-for-plan-task.sh T4-006
```

Nothing under a fixed `maintainer/` path is shipped in the repository anymore.

## Community standards

- [Code of Conduct](CODE_OF_CONDUCT.md)
- [Security policy](SECURITY.md)

## License

By contributing, you agree that your contributions are licensed under the [MIT License](LICENSE).
