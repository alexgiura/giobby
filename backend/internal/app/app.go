package app

import (
	"context"
	"fmt"
	"log"

	"dnsc_microservice/internal/config"
	"dnsc_microservice/internal/db"
	"dnsc_microservice/internal/repository"
	"dnsc_microservice/internal/routes"
	"dnsc_microservice/internal/server"
	"dnsc_microservice/internal/services"

	"github.com/jackc/pgx/v4/pgxpool"
)

type App struct {
	server *server.Server
	dbPool *pgxpool.Pool
}

func NewApp(cfg *config.Config) (*App, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	pool, err := db.NewPostgresPool(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	repo := repository.NewRepository(pool)
	appServices := services.NewAppServices(repo, cfg)
	handler := routes.RegisterRoutes(appServices, cfg)

	srv, err := server.NewServer(cfg.AppSettings.ServerPort, handler)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("create server: %w", err)
	}

	return &App{
		server: srv,
		dbPool: pool,
	}, nil
}

func (app *App) Run(ctx context.Context) error {
	if app.server == nil {
		return fmt.Errorf("server is nil")
	}

	log.Println("starting HTTP server")

	if err := app.server.Start(); err != nil {
		return fmt.Errorf("start server: %w", err)
	}

	return nil
}

func (app *App) Shutdown(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("shutdown context is nil")
	}

	log.Println("shutting down application")

	var shutdownErr error

	if app.server != nil {
		if err := app.server.Shutdown(ctx); err != nil {
			log.Printf("error shutting down server: %v", err)
			if shutdownErr == nil {
				shutdownErr = fmt.Errorf("shutdown server: %w", err)
			}
		} else {
			log.Println("server stopped successfully")
		}
	}

	if app.dbPool != nil {
		app.dbPool.Close()
		log.Println("postgres connection pool closed")
	}

	return shutdownErr
}
