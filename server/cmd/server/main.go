package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/kipenv/kip/server/internal/handler"
	"github.com/kipenv/kip/server/internal/middleware"
	"github.com/kipenv/kip/server/internal/store"
)

func main() {
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
			logger.Error("failed to connect to redis", "error", err)
			os.Exit(1)
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
		dataDir, err := os.UserHomeDir()
		if err != nil {
			logger.Error("failed to get home directory", "error", err)
			os.Exit(1)
		}
		sqlitePath = filepath.Join(dataDir, ".local", "share", "kip", "kip.db")
	}
	if err := os.MkdirAll(filepath.Dir(sqlitePath), 0o700); err != nil {
		logger.Error("failed to create data directory", "error", err)
		os.Exit(1)
	}

	teamStore, err := store.NewSQLiteStore(sqlitePath)
	if err != nil {
		logger.Error("failed to open sqlite", "error", err)
		os.Exit(1)
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
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	var h http.Handler = mux
	h = middleware.CORS(h)
	h = middleware.RateLimit(*rateLimit)(h)
	h = middleware.Logger(logger)(h)

	logger.Info("server starting", "addr", *addr)
	if err := http.ListenAndServe(*addr, h); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}
