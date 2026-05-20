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

	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/ai"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/config"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/enrichment"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/httpserver"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/items"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/photos"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/storage/localphotos"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/storage/sqlite"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg := config.Load(os.Getenv)
	startupCtx, cancelStartup := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelStartup()
	db, err := sqlite.Open(startupCtx, cfg.DatabasePath)
	if err != nil {
		logger.Error("database initialization failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	itemRepository := sqlite.NewItemRepository(db)
	itemService := items.NewService(itemRepository)
	photoRepository := sqlite.NewPhotoRepository(db)
	photoStorage := localphotos.NewStorage(cfg.PhotoStoragePath)
	photoService := photos.NewService(itemRepository, photoRepository, photoStorage)
	routerDependencies := httpserver.RouterDependencies{ItemService: &itemService, PhotoService: &photoService}
	if cfg.AIEnrichmentEnabled {
		enrichmentRepository := sqlite.NewEnrichmentRepository(db)
		enrichmentProvider := ai.FakeProvider{}
		enrichmentService := enrichment.NewService(
			itemRepository,
			photoRepository,
			enrichmentRepository,
			enrichmentProvider,
			enrichment.ProviderConfig{Provider: cfg.AIProvider, Model: cfg.AIModel},
		)
		routerDependencies.EnrichmentService = &enrichmentService
	}

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           httpserver.NewRouter(routerDependencies),
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
