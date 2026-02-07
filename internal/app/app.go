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
	ordersRepository := postgres.NewOrdersRepository(pool)

	router := chi.NewRouter()
	router.Use(middlewares.LoggerMiddleware(logger))

	router.Route("/api/user", func(router chi.Router) {
		ordersService := services.NewOrdersService(logger, ordersRepository)
		ordersController := controllers.NewOrdersController(logger, ordersService)

		// защищённые пути
		router.Group(func(r chi.Router) {
			r.Use(middlewares.AuthMiddleware([]byte(config.JWTSecret)))
			ordersController.ApplyRoutes(r)
		})

		usersService := services.NewUsersService(logger, []byte(config.JWTSecret), usersRepository)
		usersController := controllers.NewUsersController(logger, usersService)
		router.Group(usersController.ApplyRoutes)
	})

	// запускаем основной HTTP-сервер
	logger.Info().Str("addr", config.RunAddr).Msg("starting HTTP server")
	if err := http.ListenAndServe(config.RunAddr, router); err != nil {
		return err
	}

	return nil
}
