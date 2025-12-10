# Database Migration & Management Guide - Seaply

Ideally, database changes are managed through standard migrations to ensure consistency across Development, Staging, and Production (VPS) environments. This guide explains the workflow using **Golang Migrate** and **Docker**, including how to recover/initialize from a `db.sql` dump.

## 1. Database Credentials (Local Docker)
- **Container Name**: `sea_postgres`
- **User**: `seaply`
- **Password**: `@Seaply123`
- **Database**: `seaply_db`
- **Port**: `5432`

---

## 2. Converting `db.sql` Dump to Migrations (Schema + Seeds)

If you have a full database dump (`db.sql`) and want to convert it into clean, ordered migration files (separating schema, seeds, and foreign keys), we have a utility script for this.

### Prerequisites
- Python 3.x installed.
- `db.sql` file placed in `Backend/database/`.

### How to Run
1.  Navigate to the database directory:
    ```powershell
    cd Backend/database
    ```
2.  Run the converter script:
    ```powershell
    python split_dump.py
    ```
    *This script performs the following:*
    *   Reads `db.sql` (handling UTF-16/UTF-8 encoding).
    *   Parses tables, indexes, triggers, and data (COPY/INSERT).
    *   Generates sequential migration files in `Backend/database/migrations/`:
        *   `000001_init_setup.up.sql`: Global types, extensions, functions.
        *   `000002_...` to `000045_...`: Individual tables (creation + seed data).
        *   `000046_add_foreign_keys.up.sql`: Adds all Foreign Key constraints at the end (safest approach).

---

## 3. Migration Workflow (Standard)

We use `golang-migrate` to manage schema changes.

### File Structure
Migrations are located in `Backend/database/migrations/`.
- `XXXXXX_name.up.sql`: Applies the change.
- `XXXXXX_name.down.sql`: Reverts the change.

### How to Create a New Migration
When you need to change the DB (e.g., add a column, create a table):

1.  **Create migration files**:
    ```bash
    # Install migrate tool: https://github.com/golang-migrate/migrate/tree/master/cmd/migrate
    migrate create -ext sql -dir database/migrations -seq add_new_feature
    ```
2.  **Edit the files**:
    *   `up.sql`: `ALTER TABLE users ADD COLUMN age INT;`
    *   `down.sql`: `ALTER TABLE users DROP COLUMN age;`

---

## 4. Running Migrations (Local & VPS)

### Prerequisites on VPS
Ensure `golang-migrate` is installed.
```bash
# Example install on Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz
mv migrate /usr/local/bin/
```

### Command to Run Migrations
Run this command from the `Backend` directory (or wherever the `migrations` folder is relative to you).

**Local (Docker):**
```bash
migrate -path database/migrations -database "postgres://seaply:@Seaply123@localhost:5432/seaply_db?sslmode=disable" up
```

**VPS (Production):**
Replace `USER`, `PASSWORD`, `HOST`, `PORT`, and `DB_NAME` with your production values.

```bash
# General Syntax
migrate -path database/migrations -database "postgres://USER:PASSWORD@HOST:PORT/DB_NAME?sslmode=disable" up

# Example (if running inside docker container network, host might be the service name 'postgres')
migrate -path database/migrations -database "postgres://seaply:YOUR_STRONG_PASSWORD@postgres:5432/seaply_db?sslmode=disable" up
```

### Common Commands
*   **Up**: Apply all pending migrations.
    ```bash
    migrate -path database/migrations -database "..." up
    ```
*   **Down**: Revert the last migration step.
    ```bash
    migrate -path database/migrations -database "..." down 1
    ```
*   **Force**: If migration gets stuck (dirty state), force the version to the last successful one (e.g., 45).
    ```bash
    migrate -path database/migrations -database "..." force 45
    ```

---

## 5. Syncing Local Changes to Migration
If you manually modified the DB and want to save it as a migration:
1.  Dump the schema only (`pg_dump -s`).
2.  Compare/Diff with your migration files.
3.  Create a new migration file with the missing SQL commands.

*Note: For Seaply, we strictly follow the migration sequence number (`000001`, `000002`, ...). Always check the last number in `database/migrations` before creating a new one.*
