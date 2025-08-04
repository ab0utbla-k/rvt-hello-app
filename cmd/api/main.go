package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ab0utbla-k/rvt-hello-app/internal/data"

	_ "github.com/lib/pq"
)

const (
	version = "1.0.0"
)

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
}

type application struct {
	config config
	logger *slog.Logger
	models data.Models
	wg     sync.WaitGroup
}

func main() {
	var cfg config

	cfg.port = getEnv("APP_PORT", 4000, parseInt)
	cfg.env = getEnv("ENVIRONMENT", "development", parseString)
	cfg.db.dsn = getEnv("DB_DSN", "", parseString)
	cfg.db.maxOpenConns = getEnv("DB_MAX_OPEN_CONNS", 25, parseInt)
	cfg.db.maxIdleConns = getEnv("DB_MAX_IDLE_CONNS", 25, parseInt)
	cfg.db.maxIdleTime = getEnv("DB_MAX_IDLE_TIME", 15*time.Minute, parseDuration)

	flag.IntVar(&cfg.port, "port", cfg.port, "API server port")
	flag.StringVar(&cfg.env, "env", cfg.env, "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", cfg.db.dsn, "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", cfg.db.maxOpenConns, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", cfg.db.maxIdleConns, "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", cfg.db.maxIdleTime, "PostgreSQL max connection idle time")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if strings.TrimSpace(cfg.db.dsn) == "" {
		logger.Error("missing required DB_DSN (PostgreSQL DSN)")
		os.Exit(1)
	}

	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	logger.Info("database connection pool established")

	expvar.NewString("version").Set(version)

	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))

	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	db.SetConnMaxIdleTime(cfg.db.maxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func getEnv[T any](key string, defaultValue T, parser func(string) (T, error)) T {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	parsed, err := parser(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

// parseString returns the input string as-is without validation.
// Empty strings are considered valid values.
func parseString(s string) (string, error) {
	return s, nil
}

func parseInt(s string) (int, error) {
	val, err := strconv.ParseInt(s, 10, 64)
	return int(val), err
}

func parseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}
