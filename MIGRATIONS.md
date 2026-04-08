# Database Migrations

This project uses versioned migrations via `migrations/` and the `schema_migrations` table.

## Commands

Run only migrations and exit:

```bash
./loan-app-linux migrate
```

Run the application normally:

```bash
./loan-app-linux
```

On startup, the app checks and applies pending migrations automatically.

## Tracking Table

- `schema_migrations`

## Adding a New Migration

1. Add a new entry in `migrations/migrations.go`.
2. Use an ordered version such as `2026040803`.
3. Do not edit migrations that were already deployed. Add a new one instead.
