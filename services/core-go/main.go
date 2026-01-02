package main

import (
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"luma/core/internal/ai"
	"luma/core/internal/db"
	"luma/core/internal/httpapi"
)

func main() {
	zerolog.TimeFieldFormat = time.RFC3339Nano

	port := getenv("CORE_PORT", "8081")
	aiURL := getenv("AI_URL", "http://127.0.0.1:8788")
	dbPath := getenv("DB_PATH", "./data/luma.db")

	store, err := db.Open(dbPath)
	if err != nil {
		log.Fatal().Err(err).Msg("db init failed")
	}

	aiClient := ai.NewClient(aiURL)
	handler := httpapi.NewHandler(store, aiClient, log.Logger)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler.Router(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	log.Info().Msgf("core service listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("server crashed")
	}
}

func getenv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
