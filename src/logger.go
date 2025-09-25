package main

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

// initLogger initializes the global logger with the specified configuration
func initLogger() {
	// Check for debug environment variable
	debugMode := os.Getenv("DEBUG") == "true" || os.Getenv("DEBUG") == "1"

	var level slog.Level
	if debugMode {
		level = slog.LevelDebug
	} else {
		level = slog.LevelInfo
	}

	// Create a handler that writes to stdout with the specified level
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	// Create the logger
	Logger = slog.New(handler)

	// Set as default logger
	slog.SetDefault(Logger)
}
