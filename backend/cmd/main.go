package main

import (
	"context"
	"dnsc_microservice/internal/app"
	"dnsc_microservice/internal/config"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	// TZ is set in config.Load() to match AppTimezone (same as PostgreSQL session in db.NewPostgresPool).

	application, err := app.NewApp(cfg)
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)

	go func() {
		errCh <- application.Run(ctx)
	}()

	sigCh := make(chan os.Signal, 1)

	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case sig := <-sigCh:
		log.Printf("received signal: %s", sig)

		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := application.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown error: %v", err)
		}

		select {
		case err := <-errCh:
			if err != nil {
				log.Printf("application stopped with error: %v", err)
				os.Exit(1)
			}
			log.Printf("application stopped gracefully")
		case <-shutdownCtx.Done():
			log.Printf("timed out waiting for application to stop")
			os.Exit(1)
		}

	case err := <-errCh:
		if err != nil {
			log.Printf("application stopped with error: %v", err)
			os.Exit(1)
		}
		log.Printf("application stopped gracefully")
	}
}
