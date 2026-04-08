#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

required_files=(
  "config/database.go"
  "main.go"
  "startup/health.go"
  "models/user.go"
  "models/loan_file.go"
  "models/schema_migration.go"
  "migrations/migrations.go"
  "handlers/auth.go"
  "handlers/helpers.go"
  "handlers/api.go"
  "handlers/forms.go"
)

echo "[predeploy] checking required files..."
for file in "${required_files[@]}"; do
  if [[ ! -f "$file" ]]; then
    echo "[predeploy] missing required file: $file" >&2
    exit 1
  fi
done

echo "[predeploy] checking git status..."
if [[ -n "$(git status --short)" ]]; then
  echo "[predeploy] working tree is not clean" >&2
  git status --short
  exit 1
fi

echo "[predeploy] checking migrations entrypoint..."
if ! grep -q "migrations.Run" main.go; then
  echo "[predeploy] main.go does not run migrations" >&2
  exit 1
fi

echo "[predeploy] checking startup verification entrypoint..."
if ! grep -q "startup.Verify" main.go; then
  echo "[predeploy] main.go does not run startup verification" >&2
  exit 1
fi

echo "[predeploy] checking session fields..."
if ! grep -q "SessionLastActivityAt" models/user.go; then
  echo "[predeploy] models/user.go is missing SessionLastActivityAt" >&2
  exit 1
fi
if ! grep -q "SessionRevokedAt" models/user.go; then
  echo "[predeploy] models/user.go is missing SessionRevokedAt" >&2
  exit 1
fi

echo "[predeploy] checking loan file metadata model..."
if ! grep -q "type LoanFile struct" models/loan_file.go; then
  echo "[predeploy] models/loan_file.go is missing LoanFile struct" >&2
  exit 1
fi

echo "[predeploy] predeploy checks passed"
