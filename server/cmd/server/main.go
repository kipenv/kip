package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/kipenv/kip/server/internal/handler"
	"github.com/kipenv/kip/server/internal/middleware"
	"github.com/kipenv/kip/server/internal/store"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run wires up the stores, routes and server. It exists so that deferred
// cleanup (Redis and SQLite connections) actually runs on failure — os.Exit
// in main would skip it.
func run() error {
	addr := flag.String("addr", ":8080", "listen address")
	redisURL := flag.String("redis", "", "redis URL (e.g. redis://localhost:6379). If empty, uses in-memory store")
	dbPath := flag.String("db", "", "SQLite database path. If empty, uses ~/.local/share/kip/kip.db")
	rateLimit := flag.Int("rate-limit", 30, "requests per minute per IP")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Secret store (Redis or in-memory)
	var secretStore store.SecretStore
	if *redisURL != "" {
		rs, err := store.NewRedisStore(*redisURL)
		if err != nil {
			return fmt.Errorf("connect to redis: %w", err)
		}
		defer rs.Close()
		secretStore = rs
		logger.Info("using redis store", "url", *redisURL)
	} else {
		secretStore = store.NewMemoryStore()
		logger.Info("using in-memory store (development mode)")
	}

	// Team store (SQLite)
	sqlitePath := *dbPath
	if sqlitePath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("resolve home directory: %w", err)
		}
		sqlitePath = filepath.Join(home, ".local", "share", "kip", "kip.db")
	}
	if err := os.MkdirAll(filepath.Dir(sqlitePath), 0o700); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}

	teamStore, err := store.NewSQLiteStore(sqlitePath)
	if err != nil {
		return fmt.Errorf("open sqlite: %w", err)
	}
	defer teamStore.Close()
	logger.Info("using sqlite store", "path", sqlitePath)

	secretHandler := handler.NewSecretHandler(secretStore, logger)
	teamHandler := handler.NewTeamHandler(teamStore, logger)
	shareHandler := handler.NewShareHandler(teamStore, secretStore, logger)

	mux := http.NewServeMux()

	// Secret endpoints
	mux.HandleFunc("POST /api/v1/secret", secretHandler.Create)
	mux.HandleFunc("GET /api/v1/secret/{id}", secretHandler.Get)
	mux.HandleFunc("DELETE /api/v1/secret/{id}", secretHandler.Delete)

	// Team endpoints
	mux.HandleFunc("POST /api/v1/teams", teamHandler.Create)
	mux.HandleFunc("POST /api/v1/teams/join", teamHandler.Join)
	mux.HandleFunc("GET /api/v1/teams", teamHandler.ListTeams)
	mux.HandleFunc("DELETE /api/v1/teams/{id}/leave", teamHandler.Leave)
	mux.HandleFunc("GET /api/v1/teams/{id}/members", teamHandler.Members)

	// Team sharing endpoints
	mux.HandleFunc("POST /api/v1/teams/{id}/share", shareHandler.Share)
	mux.HandleFunc("GET /api/v1/teams/{id}/inbox", shareHandler.Inbox)
	mux.HandleFunc("GET /api/v1/teams/{id}/log", shareHandler.Log)
	mux.HandleFunc("PUT /api/v1/teams/{id}/pin", shareHandler.PinUpsert)
	mux.HandleFunc("GET /api/v1/teams/{id}/pin", shareHandler.PinGet)

	// Health
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	var h http.Handler = mux
	h = middleware.CORS(h)
	h = middleware.RateLimit(*rateLimit)(h)
	h = middleware.Logger(logger)(h)

	// Explicit timeouts: the zero-value http.Server has none, which lets a
	// slow client hold a connection open indefinitely (Slowloris). Payloads
	// here are small encrypted blobs, so these limits are generous.
	srv := &http.Server{
		Addr:              *addr,
		Handler:           h,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	logger.Info("server starting", "addr", *addr)
	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}
