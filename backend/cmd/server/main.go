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
	log.Printf("schematic2 backend started: http=%d https=%d", cfg.HTTP.Port, cfg.HTTP.SSLPort)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	log.Printf("shutdown signal received, stopping servers")
	httpService.ShutdownServers()
}
