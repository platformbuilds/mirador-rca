package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/miradorstack/mirador-rca/internal/api"
	"github.com/miradorstack/mirador-rca/internal/config"
	"github.com/miradorstack/mirador-rca/internal/engine"
	"github.com/miradorstack/mirador-rca/internal/extractors"
	"github.com/miradorstack/mirador-rca/internal/repo"
	"github.com/miradorstack/mirador-rca/internal/services"
	"github.com/miradorstack/mirador-rca/internal/utils"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to configuration file")
	flag.Parse()

	cfg, err := config.Load(configPath)
	if err != nil {
		slog.Error("failed to load config", slog.String("path", configPath), slog.Any("error", err))
		os.Exit(1)
	}

	logger := utils.NewLogger(cfg.Logging.Level, cfg.Logging.JSON)
	logger.Info("starting mirador-rca", slog.String("address", cfg.Server.Address))

	coreClient := repo.NewMiradorCoreClient(
		cfg.Clients.Core.BaseURL,
		cfg.Clients.Core.MetricsPath,
		cfg.Clients.Core.LogsPath,
		cfg.Clients.Core.TracesPath,
		cfg.Clients.Core.ServiceGraphPath,
		cfg.Clients.Core.Timeout,
	)

	weaviateRepo := repo.NewWeaviateRepo(
		cfg.Weaviate.Endpoint,
		cfg.Weaviate.APIKey,
		cfg.Weaviate.Timeout,
	)

	ruleEngine, err := engine.NewRuleEngine(cfg.Rules.Path, logger)
	if err != nil {
		logger.Error("failed to load rule pack", slog.Any("error", err))
		os.Exit(1)
	}
	causalityEngine := engine.NewCausalityEngine(logger)

	pipeline := engine.NewPipeline(
		logger,
		coreClient,
		weaviateRepo,
		ruleEngine,
		causalityEngine,
		extractors.NewMetricExtractor(),
		extractors.NewLogsExtractor(),
		extractors.NewTracesExtractor(),
	)

	rcaService := services.NewRCAService(logger, coreClient, pipeline, weaviateRepo)

	server, err := api.NewServer(cfg.Server, rcaService)
	if err != nil {
		logger.Error("failed to create gRPC server", slog.Any("error", err))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if serveErr := server.Start(); serveErr != nil {
			logger.Error("gRPC server exited", slog.Any("error", serveErr))
			stop()
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.GracefulTimeout)
	defer cancel()
	server.Shutdown(shutdownCtx)

	// Give remaining goroutines time to finish logging
	time.Sleep(100 * time.Millisecond)
	logger.Info("mirador-rca stopped")
}
