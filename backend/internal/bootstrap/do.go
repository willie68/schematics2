package bootstrap

import (
	"net/http"
	"path"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/samber/do/v2"
	"github.com/willie68/gowillie68/pkg/measurement"
	"github.com/willie68/schematics2/backend/internal/api"
	"github.com/willie68/schematics2/backend/internal/config"
	"github.com/willie68/schematics2/backend/internal/domain/index"
	"github.com/willie68/schematics2/backend/internal/logging"
	"github.com/willie68/schematics2/backend/internal/repository/blob"
	"github.com/willie68/schematics2/backend/internal/repository/store"
	"github.com/willie68/schematics2/backend/internal/services/backup"
	"github.com/willie68/schematics2/backend/internal/services/hash"
	"github.com/willie68/schematics2/backend/internal/services/health"
	"github.com/willie68/schematics2/backend/internal/services/shttp"
	"github.com/willie68/schematics2/backend/internal/services/users"
	"github.com/willie68/schematics2/backend/internal/version"
	"github.com/willie68/schematics2/backend/internal/webclient"
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

	if err = newBlobStore(inj); err != nil {
		return err
	}

	if err = newDocumentStore(inj); err != nil {
		return err
	}

	if err = newUserService(inj); err != nil {
		return err
	}
	do.ProvideValue(inj, index.New(inj))

	if err = newBackup(inj, cfg); err != nil {
		return err
	}

	if err = newHasher(inj, cfg); err != nil {
		return err
	}

	migration(inj)

	return InitRESTService(inj, cfg)
}

// InitHelperServices initialise the helper services like Healthsystem
func InitHelperServices(inj do.Injector, cfg config.Config) error {
	logger.Debug("initialise helper services")

	do.ProvideValue(inj, cfg)

	measurement := measurement.New(true)
	do.ProvideValue(inj, measurement)

	healthService := health.NewService(cfg.Healthcheck)
	do.ProvideValue(inj, healthService)

	return nil
}

// InitRESTService initialise REST Services
func InitRESTService(inj do.Injector, cfg config.Config) error {
	logger.Debug("init rest services")

	httpService := shttp.New(cfg.HTTP)
	do.ProvideValue(inj, httpService)
	return nil
}

func NewRouter(inj do.Injector) (http.Handler, error) {
	logger.Debug("create router")

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
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
		MaxAge:         int((10 * time.Minute).Seconds()),
	}))

	// MaxBodySize middleware: limit request body to 50MB for file uploads
	const maxBodySize = 50 * 1024 * 1024 // 50MB
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			req.Body = http.MaxBytesReader(w, req.Body, maxBodySize)
			next.ServeHTTP(w, req)
		})
	})

	// Debug middleware: log every incoming request path
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Normalize double slashes caused by Apache ProxyPass without trailing slash
			if cleaned := path.Clean(req.URL.Path); cleaned != req.URL.Path {
				req.URL.Path = cleaned
			}
			logger.Info("incoming request", "method", req.Method, "path", req.URL.Path, "remote", req.RemoteAddr)
			next.ServeHTTP(w, req)
		})
	})

	clientRedirect := version.ClientBasePath + "/client"

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, clientRedirect, http.StatusTemporaryRedirect)
	})
	r.Get("/client", clientHandler.ServeHTTP)
	r.Handle("/client/*", clientHandler)

	h.RegisterRoutes(r)
	r.Mount("/metrics/measurement", measurement.Routes(inj))
	return r, nil
}

func ShutdownServices(inj do.Injector) {
	inj.Shutdown()
}

func newDocumentStore(inj do.Injector) error {
	mongoStore := store.NewMongoStore(inj)
	if err := mongoStore.Prepare(); err != nil {
		return err
	}

	do.ProvideValue(inj, mongoStore)
	return nil
}

func newBlobStore(inj do.Injector) error {
	blobStore := blob.New(inj)
	if err := blobStore.Prepare(); err != nil {
		return err
	}
	do.ProvideValue(inj, blobStore)
	return nil
}

func newUserService(inj do.Injector) error {
	// Use 10 seconds minimum duration for each registration request
	userSvc := users.NewService(inj, 10*time.Second)
	do.ProvideValue(inj, userSvc)

	return nil
}

func newBackup(inj do.Injector, cfg config.Config) error {
	backupSvc, err := backup.New(
		inj,
		backup.WithBackupEnable(cfg.Backup.Enable),
		backup.WithDuration(time.Duration(cfg.Backup.Duration)*time.Hour),
		backup.WithPath(cfg.Backup.BackupPath),
	)
	if err != nil {
		return err
	}
	do.ProvideValue(inj, backupSvc)

	return nil
}

func newHasher(inj do.Injector, cfg config.Config) error {
	hasherSvc := hash.New(inj)
	do.ProvideValue(inj, hasherSvc)

	return nil
}

type hasher interface {
	RebuildAllHashes() error
}

func migration(inj do.Injector) error {
	logger.Debug("start migration")
	srv := do.MustInvokeAs[hasher](inj)
	srv.RebuildAllHashes()
	return nil
}
