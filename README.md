# RVT Hello App

A Go-based REST API application with PostgreSQL database integration.

## Features

- RESTful API with user management
- PostgreSQL database integration
- Configurable database connection pooling

## Prerequisites

To run this application locally, you need:

- **Go 1.24.5** or later
- **PostgreSQL** - Must be installed and running locally
- A PostgreSQL database with appropriate permissions

### PostgreSQL Setup

1. Install PostgreSQL on your system
2. Create a database for the application
3. Ensure you have the connection string (DSN) ready

## Local Development

### Environment Setup

Set the required environment variable for database connection:

```bash
export HELLO_DB_DSN="postgres://username:password@localhost/dbname?sslmode=disable"
```

### Running the Application

Use the provided Makefile commands:

```bash
# Run the API server
make run/api

# Connect to the database using psql
make db/psql

# View all available commands
make help
```

The application will start on port 4000 by default.

### Configuration Options

The application accepts the following command-line flags:

- `-port`: API server port (default: 4000)
- `-env`: Environment (development|staging|production) (default: development)
- `-db-dsn`: PostgreSQL connection string (required)
- `-db-max-open-conns`: Maximum open database connections (default: 25)
- `-db-max-idle-conns`: Maximum idle database connections (default: 25)
- `-db-max-idle-time`: Maximum connection idle time (default: 15m)

## API Endpoints

- Health check and monitoring endpoints available
- User management functionality
- Application metrics at `/debug/vars`

## Dependencies

- [httprouter](https://github.com/julienschmidt/httprouter) - HTTP request router
- [lib/pq](https://github.com/lib/pq) - PostgreSQL driver