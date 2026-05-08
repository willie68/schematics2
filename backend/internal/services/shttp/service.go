package shttp

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// SHttp encapsulates an http/https server pair.
type SHttp struct {
	cfg     Config
	useSSL  bool
	sslsrv  *http.Server
	srv     *http.Server
	Started bool
}

// New creates a new shttp service instance.
func New(cfg Config) *SHttp {
	sh := &SHttp{cfg: cfg}
	sh.init()
	return sh
}

func (s *SHttp) init() {
	s.useSSL = s.cfg.SSLPort > 0
	s.Started = false
}

// StartServers starts all configured servers.
// If SSL is enabled, appHandler is served via HTTPS and healthHandler via HTTP.
func (s *SHttp) StartServers(appHandler, healthHandler http.Handler) {
	if s.useSSL {
		s.startHTTPSServer(appHandler)
		s.startHTTPServer(healthHandler)
	} else {
		s.startHTTPServer(appHandler)
	}
	s.Started = true
}

// ShutdownServers gracefully stops all running servers.
func (s *SHttp) ShutdownServers() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if s.srv != nil {
		if err := s.srv.Shutdown(ctx); err != nil {
			log.Printf("shutdown http server error: %v", err)
		}
	}
	if s.useSSL && s.sslsrv != nil {
		if err := s.sslsrv.Shutdown(ctx); err != nil {
			log.Printf("shutdown https server error: %v", err)
		}
	}
	s.Started = false
}

func (s *SHttp) startHTTPSServer(handler http.Handler) {
	tlsConfig, err := s.resolveTLSConfig()
	if err != nil {
		panic(fmt.Errorf("could not create tls config: %w", err))
	}

	s.sslsrv = &http.Server{
		Addr:         "0.0.0.0:" + strconv.Itoa(s.cfg.SSLPort),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      handler,
		TLSConfig:    tlsConfig,
	}

	go func() {
		log.Printf("starting https server on address: %s", s.sslsrv.Addr)
		if err := s.sslsrv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			log.Printf("error starting https server: %v", err)
		}
	}()
}

func (s *SHttp) resolveTLSConfig() (*tls.Config, error) {
	if s.cfg.Certificate != "" && s.cfg.Key != "" {
		tlsCert, err := tls.LoadX509KeyPair(s.cfg.Certificate, s.cfg.Key)
		if err != nil {
			return nil, err
		}
		return &tls.Config{Certificates: []tls.Certificate{tlsCert}}, nil
	}

	host := s.cfg.ServiceURL
	if u, err := url.Parse(host); err == nil {
		host = u.Hostname()
	}

	gc := generateCertificate{
		ServiceName:  s.cfg.Servicename,
		Host:         host,
		ValidFor:     10 * 365 * 24 * time.Hour,
		IsCA:         false,
		EcdsaCurve:   "P384",
		Ed25519Key:   false,
		DNSnames:     s.cfg.DNSNames,
		IPs:          s.cfg.IPAddresses,
		Organization: s.cfg.Servicename,
	}
	return gc.GenerateTLSConfig()
}

func (s *SHttp) startHTTPServer(handler http.Handler) {
	s.srv = &http.Server{
		Addr:         "0.0.0.0:" + strconv.Itoa(s.cfg.Port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      handler,
	}

	go func() {
		log.Printf("starting http server on address: %s", s.srv.Addr)
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("error starting http server: %v", err)
		}
	}()
}
