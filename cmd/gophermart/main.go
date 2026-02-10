package main

import (
	"os"

	"github.com/Pelfox/go-loyalty-system/internal/app"
	"github.com/Pelfox/go-loyalty-system/internal/config"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	config, err := config.LoadConfig()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load application config")
	}

	if err := app.Run(config, logger); err != nil {
		logger.Fatal().Err(err).Msg("failed to start the application")
	}
}
