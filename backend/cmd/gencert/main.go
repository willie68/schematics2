package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/willie68/schematic2/backend/internal/config"
	"github.com/willie68/schematic2/backend/internal/services/shttp"
	"gopkg.in/yaml.v3"
)

func main() {
	cfgPath := flag.String("config", "configs/service.yaml", "path to service config")
	certPath := flag.String("cert", "configs/cert/server.crt", "path to certificate output file")
	keyPath := flag.String("key", "configs/cert/server.key", "path to private key output file")
	flag.Parse()

	cfg, err := loadConfig(*cfgPath)
	if err != nil {
		fatalf("load config %s: %v", *cfgPath, err)
	}

	certPEM, keyPEM, err := shttp.GenerateCertificateFromConfig(cfg.HTTP)
	if err != nil {
		fatalf("generate certificate: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(*certPath), 0o755); err != nil {
		fatalf("create cert directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(*keyPath), 0o755); err != nil {
		fatalf("create key directory: %v", err)
	}

	if err := os.WriteFile(*certPath, certPEM, 0o644); err != nil {
		fatalf("write cert file %s: %v", *certPath, err)
	}
	if err := os.WriteFile(*keyPath, keyPEM, 0o600); err != nil {
		fatalf("write key file %s: %v", *keyPath, err)
	}

	fmt.Printf("certificate written to %s\n", *certPath)
	fmt.Printf("private key written to %s\n", *keyPath)
}

func loadConfig(path string) (config.Config, error) {
	cfg := config.Config{}
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
