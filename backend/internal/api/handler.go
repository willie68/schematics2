package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/samber/do/v2"
	"github.com/willie68/schematic2/backend/internal/auth"
	"github.com/willie68/schematic2/backend/internal/config"
	"github.com/willie68/schematic2/backend/internal/domain"
	"github.com/willie68/schematic2/backend/internal/services/users"
)

const (
	// BackendVersion should be synced with HISTORY.md
	BackendVersion = "0.2.0"
)

type documentStore interface {
	Upsert(doc domain.Document) error
	ListTags(ctx context.Context) ([]domain.Tag, error)
	SuggestTags(ctx context.Context, prefix string, limit int) ([]domain.Tag, error)
	SuggestManufacturers(ctx context.Context, prefix string, limit int) ([]string, error)
	GetByID(ctx context.Context, id string) (domain.Document, error)
}

type userStore interface {
	CreateUser(ctx context.Context, user domain.User) error
	GetUserByEmail(ctx context.Context, email string) (domain.User, bool)
}

type documentIndex interface {
	Search(query string, tags []string, skip, limit int64, sortField string, sortOrder int, privateOnly, isAuthenticated bool, username string) domain.PagedSearchResult
}

type blobStore interface {
	Save(data []byte, mimeType string) (*domain.ContainerInfo, error)
	Load(ci *domain.ContainerInfo) ([]byte, error)
}

type Handler struct {
	cfg      config.Config
	docStore documentStore
	index    documentIndex
	blob     blobStore
	userSvc  *users.Service
	adminPW  string
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

type infoResponse struct {
	Version string `json:"version"`
	Status  string `json:"status"`
}

func NewHandler(i do.Injector) *Handler {
	cfg := do.MustInvoke[config.Config](i)
	hash, _ := auth.HashPassword(cfg.AdminPass)
	return &Handler{
		cfg:      cfg,
		docStore: do.MustInvokeAs[documentStore](i),
		index:    do.MustInvokeAs[documentIndex](i),
		blob:     do.MustInvokeAs[blobStore](i),
		userSvc:  do.MustInvokeAs[*users.Service](i),
		adminPW:  hash,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/health", h.health)

	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/info", h.info)
		api.Post("/auth/login", h.login)
		api.Post("/auth/register", h.register)
		api.Get("/tags", h.listTags)
		api.Get("/tags/suggest", h.suggestTags)
		api.Get("/manufacturers/suggest", h.suggestManufacturers)
		api.Get("/documents/search", h.searchDocuments)
		api.Get("/documents/{id}/files/{filename}", h.downloadFile)

		api.Group(func(protected chi.Router) {
			protected.Use(h.authMiddleware)
			protected.Get("/auth/me", h.me)
			protected.Post("/documents/index", h.indexDocument)
		})
	})
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (h *Handler) info(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, infoResponse{
		Version: BackendVersion,
		Status:  "ok",
	})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	// Check if it's an admin login
	if req.Username == h.cfg.AdminUser {
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
		return
	}

	// Try regular user login (email as username)
	user, err := h.userSvc.Authenticate(r.Context(), req.Username, req.Password)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := auth.IssueToken(h.cfg.JWTSecret, user.Email, []string{"user"}, 24*time.Hour)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "issue token")
		return
	}

	respondJSON(w, http.StatusOK, loginResponse{Token: token})
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req users.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	user, err := h.userSvc.Register(r.Context(), req)
	if err != nil {
		// Check if it's a validation error
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "email already exists") || strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "at least 8 characters") {
			statusCode = http.StatusBadRequest
		}
		respondError(w, statusCode, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]any{
		"id":        user.ID,
		"email":     user.Email,
		"firstName": user.FirstName,
		"lastName":  user.LastName,
		"created":   user.Created,
	})
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
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	var doc domain.Document
	if err := json.Unmarshal(raw, &doc); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	doc.Tags = parseTags(payload["tags"])

	if err := h.storeBlobFiles(&doc); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(doc.ID) == "" || strings.TrimSpace(doc.Manufacturer) == "" || strings.TrimSpace(doc.Model) == "" {
		respondError(w, http.StatusBadRequest, "id, manufacturer and model are required")
		return
	}
	if doc.PrivateFile && strings.TrimSpace(doc.Owner) == "" {
		respondError(w, http.StatusBadRequest, "owner is required for private documents")
		return
	}
	if err := h.docStore.Upsert(doc); err != nil {
		respondError(w, http.StatusInternalServerError, "save document")
		return
	}
	respondJSON(w, http.StatusCreated, map[string]any{"status": "indexed", "id": doc.ID})
}

func parseTags(raw any) []string {
	items, ok := raw.([]any)
	if !ok {
		return nil
	}

	seen := make(map[string]struct{}, len(items))
	tags := make([]string, 0, len(items))

	for _, item := range items {
		var tag string
		switch v := item.(type) {
		case string:
			tag = strings.TrimSpace(v)
		case map[string]any:
			if name, ok := v["name"].(string); ok {
				tag = strings.TrimSpace(name)
			}
		}

		if tag == "" {
			continue
		}
		normalized := strings.ToLower(tag)
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		tags = append(tags, tag)
	}

	if len(tags) == 0 {
		return nil
	}

	return tags
}

func (h *Handler) storeBlobFiles(doc *domain.Document) error {
	if doc == nil {
		return errors.New("document is nil")
	}
	if h.blob == nil {
		return errors.New("blob store not initialized")
	}

	for i := range doc.Files {
		encoded := strings.TrimSpace(doc.Files[i].Data)
		if encoded == "" {
			continue
		}

		data, err := decodeBase64File(encoded)
		if err != nil {
			return fmt.Errorf("invalid file data for %q: %w", doc.Files[i].Name, err)
		}

		info, err := h.blob.Save(data, doc.Files[i].MIMEType)
		if err != nil {
			return fmt.Errorf("store file %q: %w", doc.Files[i].Name, err)
		}

		doc.Files[i].Container = info
		doc.Files[i].Data = ""
	}

	return nil
}

func decodeBase64File(input string) ([]byte, error) {
	encoded := input
	if idx := strings.Index(encoded, ","); idx >= 0 && strings.Contains(encoded[:idx], ";base64") {
		encoded = encoded[idx+1:]
	}
	return base64.StdEncoding.DecodeString(encoded)
}

func (h *Handler) searchDocuments(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	tags := r.URL.Query()["tag"]
	skip := parseInt64Param(r, "skip", 0)
	limit := parseInt64Param(r, "limit", 20)
	sortField := r.URL.Query().Get("sortField")
	sortOrder := 1
	if r.URL.Query().Get("sortOrder") == "-1" {
		sortOrder = -1
	}
	privateOnly := r.URL.Query().Get("privateOnly") == "true"

	// Extract username from Authorization header (if present)
	username := ""
	isAuthenticated := false
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			claims, err := auth.VerifyToken(h.cfg.JWTSecret, parts[1])
			if err == nil {
				isAuthenticated = true
				username = claims.Subject
			}
		}
	}

	paged := h.index.Search(query, tags, skip, limit, sortField, sortOrder, privateOnly, isAuthenticated, username)
	respondJSON(w, http.StatusOK, map[string]any{
		"results": paged.Results,
		"count":   len(paged.Results),
		"total":   paged.Total,
		"skip":    paged.Skip,
		"limit":   paged.Limit,
	})
}

func (h *Handler) downloadFile(w http.ResponseWriter, r *http.Request) {
	docID := chi.URLParam(r, "id")
	filename := chi.URLParam(r, "filename")

	if strings.TrimSpace(docID) == "" || strings.TrimSpace(filename) == "" {
		respondError(w, http.StatusBadRequest, "id and filename are required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Load the document
	doc, err := h.docStore.GetByID(ctx, docID)
	if err != nil {
		respondError(w, http.StatusNotFound, "document not found")
		return
	}

	// Check if user has access (guest can only see public, authenticated can see public + own private)
	isAuthenticated := h.isAuthenticated(r)
	username := h.getAuthenticatedUser(r)

	if doc.PrivateFile {
		if !isAuthenticated || username != doc.Owner {
			respondError(w, http.StatusForbidden, "access denied")
			return
		}
	}

	// Find the file
	var file *domain.DocumentFile
	for i := range doc.Files {
		if doc.Files[i].Name == filename {
			file = &doc.Files[i]
			break
		}
	}

	if file == nil {
		respondError(w, http.StatusNotFound, "file not found")
		return
	}

	// Load file data from blob store
	if file.Container == nil {
		respondError(w, http.StatusInternalServerError, "file has no container info")
		return
	}

	data, err := h.blob.Load(file.Container)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load file")
		return
	}

	// Return as JSON with base64-encoded data
	w.Header().Set("Content-Type", "application/json")
	response := map[string]any{
		"name":     file.Name,
		"type":     file.Type,
		"mimetype": file.MIMEType,
		"page":     file.Page,
		"data":     base64.StdEncoding.EncodeToString(data),
	}
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) listTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.docStore.ListTags(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "list tags failed")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"tags": tags, "count": len(tags)})
}

func (h *Handler) suggestTags(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("q")
	limitStr := r.URL.Query().Get("limit")
	limit := 10

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	tags, err := h.docStore.SuggestTags(r.Context(), prefix, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "suggest tags failed")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"tags": tags, "count": len(tags)})
}

func (h *Handler) suggestManufacturers(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("q")
	limitStr := r.URL.Query().Get("limit")
	limit := 10

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	manufacturers, err := h.docStore.SuggestManufacturers(r.Context(), prefix, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "suggest manufacturers failed")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"manufacturers": manufacturers, "count": len(manufacturers)})
}
