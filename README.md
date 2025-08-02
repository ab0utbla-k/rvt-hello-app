# RVT Hello App

A Go-based REST API application with PostgreSQL database integration.

## Features

- RESTful API with user management
- PostgreSQL database integration
- Configurable database connection pooling

# Local PostgreSQL Setup with Docker Compose

## Prerequisites

- Docker and Docker Compose installed
- `psql` client installed on your machine
- `make` installed

---

## Environment Variables

Create a `.env` file in the project root with the following content:

```env
POSTGRES_DB=hello
POSTGRES_USER=hello
POSTGRES_PASSWORD='s3creTp@$sW0rd'
HELLO_DB_DSN=postgres://hello:s3creTp%40%24sW0rd@localhost:5432/hello?sslmode=disable
```

**Note:** The password is URL-encoded in HELLO_DB_DSN (@ → %40, $ → %24)

## Local Development

### Complete Setup (First Time)

1. **Create environment file:**
   ```bash
   # Copy the .env example above to a new .env file
   ```

2. **Start the full development environment:**
   ```bash
   # This will start PostgreSQL, run migrations, and start the API
   make setup/dev
   ```

### Manual Setup Steps

If you prefer to run commands individually:

```bash
# 1. Start PostgreSQL with Docker Compose
make compose/up

# 2. Run database migrations
make db/migrations/up

# 3. Run the API server
make run/api
```

### Other Useful Commands

```bash
# Connect to the database using psql
make db/psql

# Stop PostgreSQL and remove volumes
make compose/down

# View all available commands
make help
```

The application will start on port 4000 by default.

### Database Commands

```bash
# Connect to PostgreSQL container
docker compose exec postgres psql -U hello -d hello

# Essential psql commands once connected:
\l           # List all databases
\c hello     # Connect to hello database
\dt          # List all tables
\q           # Exit psql
```

### Configuration Options

The application accepts the following command-line flags:

- `-port`: API server port (default: 4000)
- `-env`: Environment (development|staging|production) (default: development)
- `-db-dsn`: PostgreSQL connection string (required)
- `-db-max-open-conns`: Maximum open database connections (default: 25)
- `-db-max-idle-conns`: Maximum idle database connections (default: 25)
- `-db-max-idle-time`: Maximum connection idle time (default: 15m)

## API Endpoints

### User Management

**Create/Update User**
```bash
curl -X PUT http://localhost:4000/hello/username \
  -H "Content-Type: application/json" \
  -d '{"dateOfBirth": "YYYY-MM-DD"}'
```

**Requirements:**
- Username must contain only letters (no numbers or special characters)
- Date of birth must be in YYYY-MM-DD format and before today
- Returns 204 No Content on success

**Example:**
```bash
curl -X PUT http://localhost:4000/hello/john \
  -H "Content-Type: application/json" \
  -d '{"dateOfBirth": "1990-01-15"}'
```

**Get User Birthday Message**
```bash
curl http://localhost:4000/hello/username
```

### Other Endpoints

- Health check and monitoring endpoints available
- Application metrics at `/debug/vars`

## Dependencies

- [httprouter](https://github.com/julienschmidt/httprouter) - HTTP request router
- [lib/pq](https://github.com/lib/pq) - PostgreSQL driver