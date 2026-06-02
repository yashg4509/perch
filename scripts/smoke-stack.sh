#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "==> make build"
make build
PERCH="$ROOT/perch"

if [[ -d "$ROOT/examples/scenarios/full-stack" ]]; then
  SCENARIO="$ROOT/examples/scenarios/full-stack"
elif [[ -d "$ROOT/examples/scenarios/init-signals" ]]; then
  SCENARIO="$ROOT/examples/scenarios/init-signals"
else
  echo "no scenario directory found" >&2
  exit 1
fi

echo "==> perch graph --json (${SCENARIO#$ROOT/})"
( cd "$SCENARIO" && "$PERCH" graph --json >/dev/null )

echo "==> perch status --json"
( cd "$SCENARIO" && "$PERCH" status --json >/dev/null )

echo "==> perch auth sync-env --help"
"$PERCH" auth sync-env --help >/dev/null

echo "==> perch logs setup --help"
"$PERCH" logs setup --help >/dev/null

VIZ_PORT=3132
VIZ_PID=""
cleanup() {
  if [[ -n "${VIZ_PID:-}" ]] && kill -0 "$VIZ_PID" 2>/dev/null; then
    kill "$VIZ_PID" 2>/dev/null || true
    wait "$VIZ_PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT

echo "==> perch viz + curl /api/*"
( cd "$SCENARIO" && "$PERCH" viz --port "$VIZ_PORT" ) &
VIZ_PID=$!
BASE="http://127.0.0.1:${VIZ_PORT}"
ready=0
for _ in $(seq 1 60); do
  if curl -sf "${BASE}/api/graph" >/dev/null 2>&1; then ready=1; break; fi
  if ! kill -0 "$VIZ_PID" 2>/dev/null; then wait "$VIZ_PID" 2>/dev/null || true; exit 1; fi
  sleep 0.25
done
[[ "$ready" -eq 1 ]] || { echo "viz timeout" >&2; exit 1; }
curl -sf "${BASE}/api/graph" >/dev/null
curl -sf "${BASE}/api/status" >/dev/null
curl -sf "${BASE}/api/logs?env=dev&node=web" >/dev/null

echo "PASS"
