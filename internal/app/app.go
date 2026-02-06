package app

import (
	"context"
	"net/http"

	"github.com/Pelfox/go-loyalty-system/internal"
	"github.com/Pelfox/go-loyalty-system/internal/controllers"
	"github.com/Pelfox/go-loyalty-system/internal/middlewares"
	"github.com/Pelfox/go-loyalty-system/internal/services"
	"github.com/Pelfox/go-loyalty-system/internal/storage"
	"github.com/Pelfox/go-loyalty-system/internal/storage/postgres"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

// Run открывает подключение к базе данных и запускает воркер и HTTP-сервер.
func Run(config internal.Config, logger zerolog.Logger) error {
	pool, err := storage.NewPostgresPool(context.Background(), config.DatabaseURI, storage.DefaultPostgresConnSettings)
	if err != nil {
		return err
	}

	usersRepository := postgres.NewUsersRepository(pool)

	router := chi.NewRouter()
	router.Use(middlewares.LoggerMiddleware(logger))

	router.Route("/api", func(router chi.Router) {
		usersService := services.NewUsersService(logger, []byte(config.JWTSecret), usersRepository)
		usersController := controllers.NewUsersController(logger, usersService)
		router.Route("/user", usersController.ApplyRoutes)

		// TODO: Use router.Use(middlewares.AuthMiddleware([]byte(config.JWTSecret)))
	})

	// запускаем основной HTTP-сервер
	logger.Info().Str("addr", config.RunAddr).Msg("starting HTTP server")
	if err := http.ListenAndServe(config.RunAddr, router); err != nil {
		return err
	}

	return nil
}
