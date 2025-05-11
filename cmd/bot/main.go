package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/app"
	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/infrastructure/logging"
)

func main() {
	// Setup logger
	logger := logging.SetupLogger()

	// Load configuration
	config, err := app.LoadConfig()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		fmt.Printf("Error: %v\n", err)
		fmt.Println("Make sure you have a .env file in the project root with the required environment variables.")
		fmt.Println("You can copy .env.example to .env and fill in the values.")
		os.Exit(1)
	}

	// Create application
	application, err := app.NewApp(config, logger)
	if err != nil {
		logger.Error("Failed to create application", "error", err)
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Create context that listens for termination signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		logger.Info("Received termination signal, shutting down...")

		// Create a timeout context for shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := application.Shutdown(shutdownCtx); err != nil {
			logger.Error("Error during shutdown", "error", err)
		}

		cancel()
	}()

	// Start the application
	logger.Info("Starting Flash Cards Language Telegram Bot")
	if err := application.Start(ctx); err != nil {
		logger.Error("Application error", "error", err)
		os.Exit(1)
	}
}
