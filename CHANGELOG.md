# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **`docs/`** — public contributor docs: [docs/providers.md](docs/providers.md), [docs/add-perch-provider-skill.md](docs/add-perch-provider-skill.md) (agent-agnostic provider workflow).
- Open-source baseline: `LICENSE` (MIT), `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `SECURITY.md`, GitHub issue/PR templates.
- `examples/` manual test workspace with scenarios for `perch init`, `status`, `graph`, `context`, and the root TUI.
- Root `.gitignore` (including `.cursor/`). `scripts/wait-for-plan-task.sh` polls a **local** plan file when `PLAN` is set.

### Changed

- **Documentation:** expanded **README** (architecture summary, core concepts, commands); **CONTRIBUTING** is the maintainer entrypoint for changelog and GoReleaser tagging. Internal **maintainer/** tree removed from the repository—keep private specs/plans locally if you still use them.
- **Tests:** CI/release/distribution acceptance tests assert on **`.github/workflows/ci.yml`**, **`.goreleaser.yaml`**, **README**, and the Homebrew formula instead of separate maintainer markdown runbooks.

### Removed

- **`maintainer/`** directory (spec, plan, CI/release/distribution prose, etc.). Contributor-facing material lives under **`docs/`**; release mechanics are described in **CONTRIBUTING** and [.goreleaser.yaml](.goreleaser.yaml).
