package main

import (
	"log"
	"net/http"

	"github.com/willie68/schematic2/backend/internal"
	"github.com/willie68/schematic2/backend/internal/config"
)

func main() {
	cfg := config.LoadFromEnv()

	router, err := internal.NewRouter(cfg)
	if err != nil {
		log.Fatalf("create router: %v", err)
	}

	log.Printf("schematic2 backend listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
