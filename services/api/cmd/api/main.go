package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/httpserver"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           httpserver.NewRouter(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errs := make(chan error, 1)
	go func() {
		logger.Info("api server listening", "addr", server.Addr)
		errs <- server.ListenAndServe()
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("api server shutdown failed", "error", err)
			os.Exit(1)
		}
	case err := <-errs:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("api server failed", "error", err)
			os.Exit(1)
		}
	}
}
