package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/samber/do/v2"
	"github.com/willie68/schematic2/backend/internal"
	"github.com/willie68/schematic2/backend/internal/config"
	"github.com/willie68/schematic2/backend/internal/logging"
	"github.com/willie68/schematic2/backend/internal/services/health"
	"github.com/willie68/schematic2/backend/internal/version"
)

var (
	inj    = do.New()
	logger = logging.New("main")
)

type shttpsrv interface {
	StartServers(router http.Handler, healthRouter http.Handler)
	ShutdownServers()
}

func main() {
	cfg := config.LoadFromEnv()

	// Log version and build information
	logFields := []any{
		"version", version.Version,
		"http_port", cfg.HTTP.Port,
		"https_port", cfg.HTTP.SSLPort,
	}
	if version.BuildTime != "" {
		logFields = append(logFields, "build_time", version.BuildTime)
	}
	if version.Commit != "" {
		logFields = append(logFields, "commit", version.Commit)
	}
	if cfg.ClientBasePath != "" {
		logFields = append(logFields, "client_base_path", cfg.ClientBasePath)
	}

	logger.Info("starting schematic2 backend", logFields...)

	// Log sanitized configuration (no passwords)
	logger.Info("configuration loaded",
		"mongodb_hosts", cfg.MongoDB.Hosts,
		"mongodb_database", cfg.MongoDB.Database,
		"repository_path", cfg.Repository.RepositoryPath,
		"repository_container_max_size_mb", cfg.Repository.ContainerMaxSizeMB,
		"healthcheck_period", cfg.Healthcheck.Period,
		"healthcheck_start_delay", cfg.Healthcheck.StartDelay,
	)

	err := internal.InitServices(inj, cfg)
	if err != nil {
		log.Fatalf("init services: %v", err)
	}

	router, err := internal.NewRouter(inj)
	if err != nil {
		log.Fatalf("create router: %v", err)
	}

	healthHandler := health.NewHandler(inj, cfg.HTTP.Servicename)

	httpService := do.MustInvokeAs[shttpsrv](inj)

	httpService.StartServers(router, healthHandler.Router())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	logger.Info("shutdown signal received, stopping servers")
	httpService.ShutdownServers()
}
