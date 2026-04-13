#!/usr/bin/env bash
# Tag and push a version so .github/workflows/release.yml runs GoReleaser (GitHub
# release artifacts + Homebrew tap update).
#
# Prerequisites:
#   - Repo secret HOMEBREW_TAP_GITHUB_TOKEN on GitHub (PAT with push to homebrew-perch).
#   - Remote yashg4509/homebrew-perch exists (GoReleaser publishes the Formula there).
#
# Usage:
#   ./scripts/release.sh v0.1.0              # tag + push (triggers CI release workflow)
#   ./scripts/release.sh --local v0.1.0      # run goreleaser on this machine instead
#   ./scripts/release.sh --dry-run v0.1.0    # show what would run

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

DRY_RUN=0
LOCAL_GORELEASER=0
TAG=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run) DRY_RUN=1 ;;
    --local) LOCAL_GORELEASER=1 ;;
    -h|--help)
      sed -n '1,20p' "$0"
      exit 0
      ;;
    -*)
      echo "unknown option: $1" >&2
      exit 1
      ;;
    *)
      if [[ -n "$TAG" ]]; then
        echo "extra argument: $1" >&2
        exit 1
      fi
      TAG="$1"
      ;;
  esac
  shift
done

if [[ -z "$TAG" ]]; then
  echo "usage: $0 [--dry-run] [--local] <tag>   e.g. v0.1.0" >&2
  exit 1
fi

if [[ ! "$TAG" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.-]+)?(\+[a-zA-Z0-9.-]+)?$ ]]; then
  echo "tag should look like v1.2.3 (semver with v prefix)" >&2
  exit 1
fi

if [[ "$DRY_RUN" -eq 0 ]] && [[ -n "$(git status --porcelain 2>/dev/null)" ]]; then
  echo "working tree is not clean; commit or stash before releasing" >&2
  exit 1
fi

run() {
  if [[ "$DRY_RUN" -eq 1 ]]; then
    printf '[dry-run]'; printf ' %q' "$@"; echo
  else
    "$@"
  fi
}

if [[ "$LOCAL_GORELEASER" -eq 1 ]]; then
  if [[ -z "${GITHUB_TOKEN:-}" ]]; then
    echo "GITHUB_TOKEN must be set for local goreleaser (release + assets)" >&2
    exit 1
  fi
  if [[ -z "${HOMEBREW_GITHUB_API_TOKEN:-}" ]]; then
    echo "HOMEBREW_GITHUB_API_TOKEN must be set to push the Homebrew tap (same as repo secret HOMEBREW_TAP_GITHUB_TOKEN)" >&2
    exit 1
  fi
  command -v goreleaser >/dev/null 2>&1 || {
    echo "install goreleaser: https://goreleaser.com/install/" >&2
    exit 1
  }
  run make web-build
  run go test ./...
  run git tag -a "$TAG" -m "Release $TAG"
  run goreleaser release --clean
  echo "Done: tagged $TAG and ran goreleaser locally."
  exit 0
fi

run make web-build
run go test ./...

if git rev-parse "$TAG" >/dev/null 2>&1; then
  echo "tag $TAG already exists" >&2
  exit 1
fi

run git tag -a "$TAG" -m "Release $TAG"
run git push origin "$TAG"

echo "Pushed $TAG — GitHub Actions will run GoReleaser if the workflow and secrets are configured."
