package main

import (
	"os"

	"github.com/Pelfox/go-loyalty-system/internal"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stdout)

	config, err := internal.LoadConfig()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load config")
	}

	logger.Info().Str("run_addr", config.RunAddr).Msg("address to run")
}
