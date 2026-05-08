package internal

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/samber/do/v2"
	"github.com/willie68/schematic2/backend/internal/api"
	"github.com/willie68/schematic2/backend/internal/config"
	"github.com/willie68/schematic2/backend/internal/logging"
	"github.com/willie68/schematic2/backend/internal/repository/store"
	"github.com/willie68/schematic2/backend/internal/services/health"
	"github.com/willie68/schematic2/backend/internal/services/index"
	"github.com/willie68/schematic2/backend/internal/services/shttp"
	"github.com/willie68/schematic2/backend/internal/webclient"
)

var (
	logger = logging.New("di")
)

// Service is the standard service interface
type Service interface {
	Init() error
	Shutdown() error
}

// InitServices initialise the service system
func InitServices(inj do.Injector, cfg config.Config) error {
	logger.Debug("initialise services")
	err := InitHelperServices(inj, cfg)
	if err != nil {
		return err
	}

	do.ProvideValue(inj, store.NewMongoDocumentStore(inj))
	do.ProvideValue(inj, index.NewInMemoryIndex(inj))

	return InitRESTService(inj, cfg)
}

// InitHelperServices initialise the helper services like Healthsystem
func InitHelperServices(inj do.Injector, cfg config.Config) error {
	do.ProvideValue(inj, cfg)

	healthService := health.NewService(cfg.Healthcheck)
	do.ProvideValue(inj, healthService)

	return nil
}

// InitRESTService initialise REST Services
func InitRESTService(inj do.Injector, cfg config.Config) error {
	httpService := shttp.New(cfg.HTTP)
	do.ProvideValue(inj, httpService)
	return nil
}

func NewRouter(inj do.Injector) (http.Handler, error) {

	do.Provide(inj, func(i do.Injector) (*api.Handler, error) {
		return api.NewHandler(i), nil
	})

	h, err := do.Invoke[*api.Handler](inj)
	if err != nil {
		return nil, err
	}
	clientHandler, err := webclient.Handler()
	if err != nil {
		return nil, err
	}

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
		MaxAge:         int((10 * time.Minute).Seconds()),
	}))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/client", http.StatusTemporaryRedirect)
	})
	r.Get("/client", clientHandler.ServeHTTP)
	r.Handle("/client/*", clientHandler)

	h.RegisterRoutes(r)
	return r, nil
}
func ShutdownServices(inj do.Injector) {
	inj.Shutdown()
}

func newDocumentStore(inj do.Injector) error {
	mongoStore := store.NewMongoDocumentStore(inj)
	if err := mongoStore.Prepare(); err != nil {
		return err
	}

	do.ProvideValue(inj, mongoStore)
	return nil
}
