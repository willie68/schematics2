package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/samber/do/v2"
	"github.com/willie68/schematic2/backend/internal/auth"
	"github.com/willie68/schematic2/backend/internal/config"
	"github.com/willie68/schematic2/backend/internal/domain"
)

type documentStore interface {
	Upsert(doc domain.Document) error
}

type documentIndex interface {
	Upsert(doc domain.Document)
	Search(query string, tags []string) []domain.SearchResult
}

type Handler struct {
	cfg      config.Config
	docStore documentStore
	index    documentIndex
	adminPW  string
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func NewHandler(i do.Injector) *Handler {
	cfg := do.MustInvoke[config.Config](i)
	hash, _ := auth.HashPassword(cfg.AdminPass)
	return &Handler{
		cfg:      cfg,
		docStore: do.MustInvokeAs[documentStore](i),
		index:    do.MustInvokeAs[documentIndex](i),
		adminPW:  hash,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/health", h.health)

	r.Route("/api/v1", func(api chi.Router) {
		api.Post("/auth/login", h.login)

		api.Group(func(protected chi.Router) {
			protected.Use(h.authMiddleware)
			protected.Get("/auth/me", h.me)
			protected.Post("/documents/index", h.indexDocument)
			protected.Get("/documents/search", h.searchDocuments)
		})
	})
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if req.Username != h.cfg.AdminUser {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err := auth.CheckPassword(h.adminPW, req.Password); err != nil {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := auth.IssueToken(h.cfg.JWTSecret, req.Username, []string{"admin"}, 24*time.Hour)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "issue token")
		return
	}

	respondJSON(w, http.StatusOK, loginResponse{Token: token})
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	sub, _ := r.Context().Value(ctxSubjectKey{}).(string)
	roles, _ := r.Context().Value(ctxRolesKey{}).([]string)
	respondJSON(w, http.StatusOK, map[string]any{
		"subject": sub,
		"roles":   roles,
	})
}

func (h *Handler) indexDocument(w http.ResponseWriter, r *http.Request) {
	var doc domain.Document
	if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if strings.TrimSpace(doc.ID) == "" || strings.TrimSpace(doc.Title) == "" {
		respondError(w, http.StatusBadRequest, "id and title are required")
		return
	}
	if err := h.docStore.Upsert(doc); err != nil {
		respondError(w, http.StatusInternalServerError, "save document")
		return
	}
	h.index.Upsert(doc)
	respondJSON(w, http.StatusCreated, map[string]any{"status": "indexed", "id": doc.ID})
}

func (h *Handler) searchDocuments(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	tags := r.URL.Query()["tag"]
	results := h.index.Search(query, tags)
	respondJSON(w, http.StatusOK, map[string]any{"results": results, "count": len(results)})
}
