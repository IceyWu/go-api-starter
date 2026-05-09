# Database Migrations

This directory contains SQL migration files for database schema changes.

## Migration Files

Migration files follow the naming convention: `YYYYMMDDHHMMSS_description.sql`

Example: `20260126173717_add_webhook_url_to_transcoding_tasks.sql`

## Auto Migration

The application uses GORM's AutoMigrate feature to automatically create and update database tables based on model definitions. This is configured in `cmd/server/main.go`.

## Manual Migrations

For complex schema changes that cannot be handled by AutoMigrate, place SQL migration files in this directory. These should be executed manually or through a migration tool.

## Migration Strategy

1. **Development**: Use AutoMigrate for rapid development
2. **Production**: Use versioned SQL migrations for controlled schema changes
3. **Rollback**: Keep rollback scripts for each migration

## Best Practices

- Always test migrations on a development database first
- Keep migrations small and focused
- Document breaking changes
- Maintain backward compatibility when possible
