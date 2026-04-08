# Production Deploy Checklist

## Predeploy

1. Pull the latest `main`.
2. Confirm the working tree is clean.
3. Review `.env.production` or runtime environment variables.
4. Run the predeploy check script.
5. Build the Linux binary.
6. Run startup healthcheck after migrations.

## Commands

```bash
git pull --ff-only origin main
bash scripts/predeploy-check.sh
go test ./...
go build -ldflags="-s -w" -trimpath -o loan-app-linux
./loan-app-linux migrate
./loan-app-linux healthcheck
./loan-app-linux
```

## Critical Files

- `config/database.go`
- `main.go`
- `startup/health.go`
- `models/user.go`
- `models/loan_file.go`
- `models/schema_migration.go`
- `migrations/migrations.go`
- `handlers/auth.go`
- `handlers/helpers.go`
- `handlers/api.go`
- `handlers/forms.go`

## Database Requirements

- The app user must be able to create and alter tables.
- The database must contain the `schema_migrations` table after the first migration run.

## Verification After Deploy

1. Login works.
2. Loan list loads.
3. Step 1 to Step 7 can open.
4. File upload and download work.
5. Session timeout and revoke-all still work.
6. Startup healthcheck passes before the app starts serving traffic.
