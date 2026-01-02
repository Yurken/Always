package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"luma/core/internal/ai"
	"luma/core/internal/db"
	"luma/core/internal/httpapi"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	port := getenv("CORE_PORT", "8081")
	aiURL := getenv("AI_URL", "http://127.0.0.1:8788")
	dbPath := getenv("DB_PATH", "./data/luma.db")

	store, err := db.Open(dbPath)
	if err != nil {
		logger.Error("db init failed", slog.Any("error", err))
		os.Exit(1)
	}

	aiClient := ai.NewClient(aiURL)
	handler := httpapi.NewHandler(store, aiClient, logger)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler.Router(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	logger.Info("core service listening", slog.String("addr", server.Addr))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server crashed", slog.Any("error", err))
		os.Exit(1)
	}
}

func getenv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
