package health

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/samber/do/v2"
)

// Service is the healthcheck service.
type Service struct {
	cfg         Config
	healthy     bool
	readyz      bool
	messages    []string
	checks      []Check
	lastChecked time.Time
	reg         sync.Mutex
}

// Message is the health response payload.
type Message struct {
	Messages  []string `json:"messages"`
	LastCheck string   `json:"lastCheck,omitempty"`
}

// NewService initializes the health system service.
func NewService(cfg Config) *Service {
	s := &Service{
		cfg:     cfg,
		healthy: false,
		checks:  make([]Check, 0),
		reg:     sync.Mutex{},
	}
	s.Init()
	return s
}

// Init starts periodic health checks.
func (h *Service) Init() {
	log.Printf("healthcheck starting with period: %d seconds", h.cfg.Period)
	h.messages = []string{"service starting"}
	h.readyz = false
	h.doCheck()
	h.lastChecked = time.Now()

	go func() {
		if h.cfg.StartDelay > 0 {
			time.Sleep(time.Duration(h.cfg.StartDelay) * time.Second)
		}
		if h.cfg.Period > 0 {
			background := time.NewTicker(time.Second * time.Duration(h.cfg.Period))
			defer background.Stop()
			for range background.C {
				h.doCheck()
			}
		}
	}()
}

// CheckHealthCheckTimer verifies the periodic health check is still running.
func (h *Service) CheckHealthCheckTimer() {
	t := time.Now()
	if h.cfg.Period <= 0 {
		return
	}
	if t.Sub(h.lastChecked) > (time.Second * time.Duration(2*h.cfg.Period)) {
		h.readyz = false
		h.messages = []string{"health check not running"}
		if t.Sub(h.lastChecked) > (time.Second * time.Duration(4*h.cfg.Period)) {
			panic("panic: health check is not running anymore")
		}
	}
}

// Register registers or replaces a health check.
func (h *Service) Register(check Check) {
	h.reg.Lock()
	defer h.reg.Unlock()
	for x, c := range h.checks {
		if c.CheckName() == check.CheckName() {
			h.checks[x] = check
			return
		}
	}
	h.checks = append(h.checks, check)
}

// Unregister removes a health check by name.
func (h *Service) Unregister(checkname string) bool {
	h.reg.Lock()
	defer h.reg.Unlock()
	for x := len(h.checks) - 1; x >= 0; x-- {
		c := h.checks[x]
		if c.CheckName() == checkname {
			h.checks = append(h.checks[:x], h.checks[x+1:]...)
			return true
		}
	}
	return false
}

// Message returns messages from the last health check.
func (h *Service) Message() Message {
	return Message{
		LastCheck: h.lastChecked.String(),
		Messages:  h.messages,
	}
}

func (h *Service) doCheck() {
	h.lastChecked = time.Now()
	h.messages = make([]string, 0)
	healthy := true

	h.reg.Lock()
	defer h.reg.Unlock()
	for _, c := range h.checks {
		ok, err := c.Check()
		if !ok {
			healthy = false
			if err != nil {
				h.messages = append(h.messages, fmt.Sprintf("%s: %s", c.CheckName(), err.Error()))
			} else {
				h.messages = append(h.messages, fmt.Sprintf("%s: unhealthy", c.CheckName()))
			}
		}
	}

	h.healthy = healthy
	if healthy {
		h.readyz = true
	}
}

// Healthyz returns the current liveness state.
func (h *Service) Healthyz() bool {
	return h.healthy
}

// Readyz returns the current readiness state.
func (h *Service) Readyz() bool {
	return h.readyz
}

// LastChecked returns the timestamp of the last check run.
func (h *Service) LastChecked() time.Time {
	return h.lastChecked
}

// Handler is the HTTP handler for health endpoints.
type Handler struct {
	health      *Service
	serviceName string
}

// NewHandler creates a new REST health handler.
func NewHandler(inj do.Injector, serviceName string) *Handler {
	health := do.MustInvoke[*Service](inj)
	return &Handler{health: health, serviceName: serviceName}
}

// Router returns all routes for the health endpoints.
func (h *Handler) Router() *chi.Mux {
	router := chi.NewRouter()
	router.Get("/", h.GetDefaultEndpoint)
	router.Get("/livez", h.GetLivenessEndpoint)
	router.Get("/readyz", h.GetReadinessEndpoint)
	router.Head("/livez", h.HeadLivenessEndpoint)
	router.Head("/readyz", h.HeadReadinessEndpoint)
	return router
}

// GetLivenessEndpoint exposes liveness probe state.
func (h *Handler) GetLivenessEndpoint(w http.ResponseWriter, r *http.Request) {
	if h.health.Healthyz() {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	_ = json.NewEncoder(w).Encode(h.health.Message())
}

// HeadLivenessEndpoint exposes liveness probe state as HEAD.
func (h *Handler) HeadLivenessEndpoint(w http.ResponseWriter, _ *http.Request) {
	if h.health.Healthyz() {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}

// GetReadinessEndpoint exposes readiness probe state.
func (h *Handler) GetReadinessEndpoint(w http.ResponseWriter, _ *http.Request) {
	h.health.CheckHealthCheckTimer()
	if h.health.Readyz() {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(Message{
			Messages:  []string{"main: service up and running"},
			LastCheck: h.health.LastChecked().String(),
		})
		return
	}
	w.WriteHeader(http.StatusServiceUnavailable)
	_ = json.NewEncoder(w).Encode(h.health.Message())
}

// HeadReadinessEndpoint exposes readiness probe state as HEAD.
func (h *Handler) HeadReadinessEndpoint(w http.ResponseWriter, _ *http.Request) {
	h.health.CheckHealthCheckTimer()
	if h.health.Readyz() {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}

// GetDefaultEndpoint exposes a simple text default response.
func (h *Handler) GetDefaultEndpoint(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(fmt.Sprintf("<b>%s</b>: http-server up and running!", h.serviceName)))
}
