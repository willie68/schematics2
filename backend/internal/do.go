package internal

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/samber/do/v2"
	"github.com/willie68/schematic2/backend/internal/config"
	"github.com/willie68/schematic2/backend/internal/httpapi"
	"github.com/willie68/schematic2/backend/internal/index"
	"github.com/willie68/schematic2/backend/internal/store"
)

func NewRouter(cfg config.Config) (http.Handler, error) {
	injector := do.New()

	do.ProvideValue(injector, cfg)
	do.Provide(injector, func(i do.Injector) (*store.InMemoryDocumentStore, error) {
		return store.NewInMemoryDocumentStore(), nil
	})
	do.Provide(injector, func(i do.Injector) (*index.InMemoryIndex, error) {
		return index.NewInMemoryIndex(), nil
	})
	do.Provide(injector, func(i do.Injector) (*httpapi.Handler, error) {
		return httpapi.NewHandler(i), nil
	})

	h := do.MustInvoke[*httpapi.Handler](injector)
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
		MaxAge:         int((10 * time.Minute).Seconds()),
	}))

	h.RegisterRoutes(r)
	return r, nil
}
