package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/samber/do/v2"
	flag "github.com/spf13/pflag"
	"github.com/willie68/schematics2/backend/internal/bootstrap"
	"github.com/willie68/schematics2/backend/internal/config"
	"github.com/willie68/schematics2/backend/internal/logging"
	"github.com/willie68/schematics2/backend/internal/services/health"
	"github.com/willie68/schematics2/backend/internal/version"
)

var (
	inj        = do.New()
	logger     *slog.Logger
	configFile string
)

type shttpsrv interface {
	StartServers(router http.Handler, healthRouter http.Handler)
	ShutdownServers()
}

type backupService interface {
	Start() error
}

func init() {
	// variables for parameter override
	logging.Root.Info("init service")
	flag.StringVarP(&configFile, "config", "c", config.File, "this is the path and filename to the config file")
}

func main() {
	flag.Parse()

	if configFile != "" {
		config.File = configFile
	}
	cfg := config.LoadFromEnv()

	// Initialize logging first
	logging.Init(cfg.Logging, cfg.HTTP.Servicename)
	logger = logging.New("main")

	// Log version and build information
	startupLog(cfg)

	// Log admin credentials status
	maskedPass := ""
	if cfg.AdminPass != "" {
		maskedPass = ""
		for i := 0; i < len(cfg.AdminPass); i++ {
			maskedPass += "*"
		}
	}
	logger.Info("admin credentials",
		"admin_user", cfg.AdminUser,
		"admin_pass_masked", maskedPass,
	)

	err := bootstrap.InitServices(inj, cfg)
	if err != nil {
		log.Fatalf("init services: %v", err)
	}

	router, err := bootstrap.NewRouter(inj)
	if err != nil {
		log.Fatalf("create router: %v", err)
	}

	healthHandler := health.NewHandler(inj, cfg.HTTP.Servicename)

	httpService := do.MustInvokeAs[shttpsrv](inj)

	httpService.StartServers(router, healthHandler.Router())

	backupSvc := do.MustInvokeAs[backupService](inj)
	if err := backupSvc.Start(); err != nil {
		log.Fatalf("start backup service: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	logger.Info("shutdown signal received, stopping servers")
	httpService.ShutdownServers()
}

func startupLog(cfg config.Config) {
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
	if version.ClientBasePath != "" {
		logFields = append(logFields, "client_base_path", version.ClientBasePath)
	}

	logger.Info("starting schematics2 backend", logFields...)
}
