#!/usr/bin/env bash
# Poll a local task checklist until a line matches [x] TASK_ID.
# The checklist file is not part of this repository; set PLAN to your private path.
#
#   PLAN=~/notes/perch-plan.md ./scripts/wait-for-plan-task.sh T4-006
set -euo pipefail
task="${1:?usage: PLAN=/path/to/plan.md $0 T4-006}"
plan="${PLAN:?set PLAN to a local checklist file (example: PLAN=~/perch-plan.md)}"
interval="${POLL_SEC:-30}"
while ! grep -qE "\\[x\\].*${task}" "$plan"; do
  echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) waiting for ${task} in ${plan} (sleep ${interval}s)"
  sleep "$interval"
done
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) ${task} is checked in ${plan}"
