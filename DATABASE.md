# Database Setup

This project uses SQLite for persistent data storage. The database is automatically initialized with all necessary tables and migrations on first run.

## Running the Application

### Prerequisites

- Go 1.26.1 or later

### Using the CLI (Recommended)

The CLI provides better control over database and server options:

```bash
# Start the web server (default: localhost:8080)
go run ./cmd/cli serve

# Start on a custom port
go run ./cmd/cli serve --port :3000

# Run database migrations
go run ./cmd/cli migrate

# Reset the database
go run ./cmd/cli reset
```

Or build the CLI binary first:

```bash
go build -o northbasis-cli ./cmd/cli
./northbasis-cli serve
./northbasis-cli migrate
./northbasis-cli reset
```

### Legacy Entry Point

For backward compatibility, you can also start the server directly:

```bash
go run ./cmd/web/main.go
```

Or build and run the binary:

```bash
go build -o web ./cmd/web/main.go
./web
```

The server will start on `http://localhost:8080`.

## Database

### Location

The SQLite database file `northbasis.db` is created in the current working directory when you run the application.

### Automatic Setup

All database tables are created automatically using migrations when the application starts. You don't need to manually run any setup commands.

### Schema

The database includes the following tables:

- **users**: Stores user account information (id, email)
- **users_credentials**: Stores password hashes for authentication
- **sessions**: Stores authenticated session tokens
- **plans**: Stores business plans in JSON format
- **plan_access**: Manages user access to plans (owner, editor, viewer)
- **schema_migrations**: Tracks applied database migrations

### Type Safety

The implementation uses Go's type system and compile-time checks:

- Domain models are defined in `internal/domain/`
- The store interface is defined in `internal/store/memory.go`
- SQLite implementation in `internal/store/sqlite.go` enforces type safety through Go's static typing
- JSON serialization/deserialization for complex objects (plans)

## Development

### Database Inspection

To inspect the database during development, you can use the `sqlite3` CLI:

```bash
sqlite3 northbasis.db
```

Common commands:

```sql
-- List all tables
.tables

-- Show schema for a table
.schema users

-- Query data
SELECT * FROM users;
SELECT * FROM plan_access;
```

### Resetting the Database

To reset the database and start fresh, simply delete the file and restart the application:

```bash
rm northbasis.db
go run ./cmd/web/main.go
```

### Adding New Migrations

To add a new migration:

1. Edit `internal/store/migrations.go`
2. Add a new migration SQL string to the `migrations` array
3. The next application restart will automatically apply the new migration
