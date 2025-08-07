# RVT Hello App

A Go-based "Hello World" REST API that manages users and provides birthday messages.

## Prerequisites

- Docker and Docker Compose
- Go 1.24.5+
- `make`

## Quick Start

1. **Create environment file:**
   ```bash
   # Create .env file with database configuration
   POSTGRES_DB=hello
   POSTGRES_USER=hello
   POSTGRES_PASSWORD='s3creTp@$sW0rd'
   HELLO_DB_DSN=postgres://hello:s3creTp%40%24sW0rd@localhost:5432/hello?sslmode=disable
   ```

2. **Start the application:**
   ```bash
   make setup/dev
   ```

The app runs on port 4000. Database migrations run automatically on startup.

## Development Commands

```bash
# Run tests
make test

# Run only fast tests  
make test/short

# View application logs
make logs

# Stop and cleanup
make compose/down
```


## API Usage

**Save/Update User:**
```bash
curl -X PUT http://localhost:4000/hello/john \
  -H "Content-Type: application/json" \
  -d '{"dateOfBirth": "1990-01-15"}'
```
Returns: `204 No Content`

**Get Birthday Message:**
```bash
curl http://localhost:4000/hello/john
```
Returns: `{"message": "Hello, john! Your birthday is in N day(s)"}`

**Requirements:**
- Username: letters only
- Date: YYYY-MM-DD format, must be in the past

## Workflows

**Tests:** Runs on PRs and non-main pushes. Uses PostgreSQL 16 and Go 1.24.5.

**CI/CD:** Deploys to AWS ECS on main branch pushes. Builds Docker images to ECR and updates ECS service.

Both workflows can be triggered manually from the GitHub Actions tab.
