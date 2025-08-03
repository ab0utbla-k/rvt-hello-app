package testutils

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const userTableMigrationPath = "../../migrations/000001_create_users_table.up.sql"

func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	if testing.Short() {
		t.Skip("Skipping DB setup in short test mode")
	}
	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:16-alpine3.22",
		postgres.WithDatabase("hello_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("password"),
		postgres.WithInitScripts(userTableMigrationPath),
		testcontainers.WithWaitStrategy(
			// PostgreSQL logs this twice: during startup and when fully ready
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second)),
	)
	require.NoError(t, err, "Failed to start PostgreSQL test container (is Docker running?)")

	t.Cleanup(func() {
		require.NoError(t, container.Terminate(ctx))
	})

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()

	require.NoError(t, db.PingContext(pingCtx))

	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	return db
}

func CleanupDB(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`TRUNCATE TABLE users RESTART IDENTITY CASCADE`)
	require.NoError(t, err)
}
