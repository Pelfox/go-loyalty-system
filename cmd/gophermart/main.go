package main

import (
	"os"

	"github.com/Pelfox/go-loyalty-system/internal"
	"github.com/Pelfox/go-loyalty-system/internal/app"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	config, err := internal.LoadConfig()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load application config")
	}

	if err := app.Run(config, logger); err != nil {
		logger.Fatal().Err(err).Msg("failed to start the application")
	}
}
