# Database Migrations

This directory contains SQL migration files for database schema changes.

## Migration Files

Migration files follow the naming convention: `YYYYMMDDHHMMSS_description.sql`

Example: `20260126173717_add_webhook_url_to_transcoding_tasks.sql`

## Migration runner

The application uses [Atlas](https://atlasgo.io) for versioned SQL migrations. The runner selects the dialect-specific directory automatically. Run them with:

```bash
task migrate
```

生产环境先执行迁移，再启动 API server 和 MPS Worker。

Development and production use the same versioned Atlas migrations. Schema changes are never inferred from Go model tags.

Existing development databases may set `database.allow_dirty_migrations: true` to establish the initial Atlas baseline over an already-created schema. Production keeps this disabled by default and requires an explicit deployment override after a verified backup.

## Manual Migrations

Place SQLite migrations in `migrations/sqlite` and MySQL migrations in `migrations/mysql`. Atlas applies the SQL files in filename order.

## Migration Strategy

1. **All environments**: Apply versioned Atlas migrations before startup
2. **Schema source**: Keep SQL migrations as the only schema source
3. **Rollback**: Keep rollback scripts for each migration

## Best Practices

- Always test migrations on a development database first
- Keep migrations small and focused
- Document breaking changes
- Regenerate `sqlc` after changing `db/schema.sql` or `db/query`
