#!/usr/bin/env bash
set -euo pipefail

# Scripted security patch set:
# 1) Run strict gosec checks used by CI for cmd/.
# 2) Optionally apply curated minor dependency security updates.
# 3) Optionally create a PR from current branch (requires gh auth).

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GO_BIN="${GO_BIN:-$(command -v go || true)}"

if [[ -z "$GO_BIN" ]]; then
  echo "go binary not found" >&2
  exit 1
fi

APPLY_MINOR=0
CREATE_PR=0
PR_TITLE="Security hygiene patch set"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --apply-minor)
      APPLY_MINOR=1
      shift
      ;;
    --create-pr)
      CREATE_PR=1
      shift
      ;;
    --pr-title)
      PR_TITLE="$2"
      shift 2
      ;;
    *)
      echo "Unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

cd "$ROOT_DIR"
mkdir -p coverage

echo "[1/4] Running strict gosec scan parity for cmd folder"
"$GO_BIN" install github.com/securego/gosec/v2/cmd/gosec@latest
"$($GO_BIN env GOPATH)/bin/gosec" -fmt=json -no-fail -include=G104,G115,G204,G301,G304 ./cmd/... > coverage/gosec.json

python3 - <<'PY'
import json
import sys

with open('coverage/gosec.json', 'r', encoding='utf-8') as f:
    data = json.load(f)

issues = data.get('Issues') or []
target_issues = [
  i for i in issues
  if '/cmd/' in str(i.get('file', '')).replace('\\\\', '/')
]

if target_issues:
    print('Strict-path gosec issues found in cmd folder:')
    for item in target_issues:
        print(f"- {item.get('rule_id')} at {item.get('file')}:{item.get('line')}: {item.get('details')}")
    sys.exit(1)

print('No strict-path gosec issues found in cmd folder')
print(f'Total gosec issues in cmd scan: {len(issues)}')
PY

if [[ "$APPLY_MINOR" -eq 1 ]]; then
  echo "[2/4] Applying curated minor upgrades"
  GOFLAGS=-mod=mod "$GO_BIN" get \
    github.com/aws/aws-sdk-go-v2/config@v1.32.18 \
    github.com/aws/aws-sdk-go-v2/service/sts@v1.42.1 \
    golang.org/x/term@v0.43.0
  GOFLAGS=-mod=mod "$GO_BIN" mod tidy
  GOFLAGS=-mod=mod "$GO_BIN" mod vendor
else
  echo "[2/4] Skipping dependency updates (use --apply-minor to enable)"
fi

echo "[3/4] Running validation tests"
"$GO_BIN" test ./cmd -coverprofile=coverage/cmd.out -covermode=atomic
"$GO_BIN" tool cover -func=coverage/cmd.out | grep '^total:'
"$GO_BIN" test ./pkg/formatter -run '^TestContract_'

if [[ "$CREATE_PR" -eq 1 ]]; then
  echo "[4/4] Creating PR via gh"
  if ! command -v gh >/dev/null 2>&1; then
    echo "gh CLI is required for --create-pr" >&2
    exit 1
  fi
  branch_name="$(git branch --show-current)"
  git push -u origin "$branch_name"
  gh pr create --fill --title "$PR_TITLE"
else
  echo "[4/4] Skipping PR creation (use --create-pr to enable)"
fi

echo "Security patch set completed."
