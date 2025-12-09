# Database Migration & Management Guide - Seaply

Ideally, database changes are managed through standard migrations to ensure consistency across Development, Staging, and Production (VPS) environments. This guide explains the workflow using **Golang Migrate** and **Docker**.

## 1. Database Credentials (Local Docker)
- **Container Name**: `gate_postgres`
- **User**: `gate`
- **Password**: `@Gate123`
- **Database**: `gate_db`
- **Port**: `5432`

## 2. Backup / Dump Database
If you need to backup the current database state (data + schema), run this command:

```bash
# From Windows PowerShell in Backend directory
docker exec -e PGPASSWORD=@Gate123 gate_postgres pg_dump -U gate -d gate_db > database/db.sql
```

This creates a full snapshot in `backend/database/db.sql`.

## 3. Migration Workflow (Standard)

We use `golang-migrate` to manage schema changes.

### File Structure
Migrations are located in `Backend/migrations/`.
- `XXXXXX_name.up.sql`: Applies the change (e.g., Create Table).
- `XXXXXX_name.down.sql`: Reverts the change (e.g., Drop Table).

### How to Create a New Migration
When you need to change the DB (e.g., add a column, create a table), **DO NOT edit the database manually**. Instead:

1.  Create a new migration file pair:
    ```bash
    # install migrate tool first if not present
    migrate create -ext sql -dir migrations -seq add_new_feature
    ```
    This creates:
    - `migrations/00000X_add_new_feature.up.sql`
    - `migrations/00000X_add_new_feature.down.sql`

2.  Write your SQL in these files.
    - `up.sql`: `ALTER TABLE users ADD COLUMN age INT;`
    - `down.sql`: `ALTER TABLE users DROP COLUMN age;`

3.  Apply the migration:
    ```bash
    # Using Makefile (if configured) or direct migrate command
    migrate -path migrations -database "postgres://gate:@Gate123@localhost:5432/gate_db?sslmode=disable" up
    ```

## 4. Deploying to VPS (Production)

When you move to a VPS, you don't need to manually copy tables. The migration system handles it.

1.  **Pull Latest Code**: Get the latest code (including `migrations/` folder) on the VPS.
2.  **Run Migrations**: Execute the migrate command against the VPS database.
    ```bash
    # Example command on VPS
    migrate -path ./migrations -database "postgres://user:pass@localhost:5432/dbname?sslmode=disable" up
    ```
    This will check which migrations are already applied (via `schema_migrations` table) and **only apply the new ones**. Data from previous migrations is preserved.

### Ensuring existing data is safe
- **Migrations are additive**: Adding tables/columns doesn't delete data.
- **Avoid Destructive operations**: Be careful with `DROP TABLE` or `DROP COLUMN` in production migrations.
- **Always Backup**: Before running migrations on Prod, run the Dump command (Step 2).

## 5. Syncing Local Changes to Migration
If you have manually modified the DB and want to "save" it as a migration:
1.  Dump the schema only (`pg_dump -s`).
2.  Compare it with your migrations.
3.  Create a new migration file with the missing SQL commands.

*Note: For Seaply, we already have initial migrations (`000001` to `000006`). Always continue the sequence (000007, etc.).*
