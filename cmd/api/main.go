package main

import (
	"os"

	"github.com/acidsoft/gorestteach/internal/config"
	"github.com/acidsoft/gorestteach/internal/database"
	"github.com/acidsoft/gorestteach/internal/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// ─── Logger ───────────────────────────────────────────────────────────────
	// Pretty console logging in development, JSON in production.
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// ─── Config ───────────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load configuration")
	}
	log.Info().Str("mode", cfg.Server.Mode).Int("port", cfg.Server.Port).Msg("Configuration loaded")

	// ─── Database ─────────────────────────────────────────────────────────────
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	log.Info().Msg("Database connected and migrated")

	// ─── Server ───────────────────────────────────────────────────────────────
	srv := server.New(cfg, db)
	if err := srv.Start(); err != nil {
		log.Fatal().Err(err).Msg("server stopped with error")
	}
}
