package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/samber/do/v2"
	"github.com/willie68/schematics2/backend/internal/auth"
	"github.com/willie68/schematics2/backend/internal/config"
	"github.com/willie68/schematics2/backend/internal/domain/model"
	"github.com/willie68/schematics2/backend/internal/logging"
	"github.com/willie68/schematics2/backend/internal/repository/connector"
	"github.com/willie68/schematics2/backend/internal/services/users"
	"github.com/willie68/schematics2/backend/internal/version"
)

type documentStore interface {
	Upsert(doc model.Document) error
	ListTags(ctx context.Context) ([]model.Tag, error)
	SuggestTags(ctx context.Context, prefix string, limit int) ([]model.Tag, error)
	SuggestManufacturers(ctx context.Context, prefix string, limit int) ([]string, error)
	GetByID(ctx context.Context, id string) (model.Document, error)
	DeleteByID(ctx context.Context, id string) error
}

type effectStore interface {
	SearchEffects(ctx context.Context, query string, skip, limit int64, sortField, sortOrder string) (model.PagedEffects, error)
	GetEffectByID(ctx context.Context, id string) (*model.Effect, error)
	CreateEffect(ctx context.Context, effect *model.Effect) error
	UpdateEffect(ctx context.Context, effect *model.Effect) error
	UpdateManufacturer(ctx context.Context, manufacturer string) error
	DeleteEffect(ctx context.Context, id string) error
}

type effectTypeStore interface {
	GetAllEffectTypes(ctx context.Context) ([]model.EffectType, error)
}

type userStore interface {
	CreateUser(ctx context.Context, user model.User) error
	GetUserByEmail(ctx context.Context, email string) (model.User, bool)
	UpdateUser(ctx context.Context, user model.User) error
}

type documentIndex interface {
	Search(query string, tags []string, skip, limit int64, sortField string, sortOrder int, privateOnly, isAuthenticated bool, username string) model.PagedSearchResult
}

type blobStore interface {
	Save(data []byte, mimeType, filename string) (*model.ContainerInfo, error)
	Load(ci *model.ContainerInfo) ([]byte, error)
	DeleteByInfo(ci *model.ContainerInfo) error
}

type usersService interface {
	Register(ctx context.Context, req users.RegisterRequest) (model.User, error)
	Authenticate(ctx context.Context, email, password string) (model.User, error)
}

type Handler struct {
	cfg             config.Config
	log             *slog.Logger
	docStore        documentStore
	effectStore     effectStore
	effectTypeStore effectTypeStore
	index           documentIndex
	blob            blobStore
	userSvc         usersService
	userStore       userStore
	adminPW         string
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

type changePasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

func NewHandler(i do.Injector) *Handler {
	cfg := do.MustInvoke[config.Config](i)
	hash, _ := auth.HashPassword(cfg.AdminPass)
	return &Handler{
		cfg:             cfg,
		log:             logging.New("api-handler"),
		docStore:        do.MustInvokeAs[documentStore](i),
		effectStore:     do.MustInvokeAs[effectStore](i),
		effectTypeStore: do.MustInvokeAs[effectTypeStore](i),
		index:           do.MustInvokeAs[documentIndex](i),
		blob:            do.MustInvokeAs[blobStore](i),
		userSvc:         do.MustInvokeAs[usersService](i),
		userStore:       do.MustInvokeAs[userStore](i),
		adminPW:         hash,
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
		api.Get("/effecttypes", h.listEffectTypes)
		api.Get("/documents/search", h.searchDocuments)
		api.Get("/documents/{id}/files/{filename}", h.downloadFile)
		api.Get("/effects/search", h.searchEffects)
		api.Get("/effects/{id}/image", h.getEffectImage)
		api.Get("/effects/{id}", h.getEffect)
		api.Get("/connectors/{name}", h.getConnectorImage)

		api.Group(func(protected chi.Router) {
			protected.Use(h.authMiddleware)
			protected.Get("/users/me", h.me)
			protected.Post("/users/change-password", h.changePassword)
			protected.Post("/documents/index", h.indexDocument)
			protected.Patch("/documents/{id}", h.updateDocument)
			protected.Delete("/documents/{id}", h.deleteDocument)
			protected.Post("/effects", h.createEffect)
			protected.Patch("/effects/{id}", h.updateEffect)
			protected.Delete("/effects/{id}", h.deleteEffect)
		})
	})
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (h *Handler) info(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, infoResponse{
		Version: version.Version,
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

	// Check if admin
	isAdmin := false
	for _, role := range roles {
		if role == "admin" {
			isAdmin = true
			break
		}
	}

	// Try to fetch user from database
	var userData map[string]any
	user, exists := h.userStore.GetUserByEmail(r.Context(), sub)

	if exists {
		// User found in database
		userData = map[string]any{
			"id":        user.ID,
			"email":     user.Email,
			"firstName": user.FirstName,
			"lastName":  user.LastName,
			"address":   user.Address,
			"created":   user.Created,
			"updated":   user.Updated,
		}
	} else if isAdmin {
		// Admin not in database, return minimal but complete data
		userData = map[string]any{
			"id":        "admin",
			"email":     "admin",
			"firstName": "Administrator",
			"lastName":  "",
			"address":   nil,
			"created":   0,
			"updated":   0,
		}
	} else {
		// Regular user not found
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	respondJSON(w, http.StatusOK, userData)
}

func (h *Handler) changePassword(w http.ResponseWriter, r *http.Request) {
	sub, _ := r.Context().Value(ctxSubjectKey{}).(string)
	roles, _ := r.Context().Value(ctxRolesKey{}).([]string)

	// Admin cannot change password via this endpoint
	for _, role := range roles {
		if role == "admin" {
			respondError(w, http.StatusForbidden, "admin password change not supported")
			return
		}
	}

	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	// Validate input
	if req.OldPassword == "" {
		respondError(w, http.StatusBadRequest, "old password required")
		return
	}
	if req.NewPassword == "" {
		respondError(w, http.StatusBadRequest, "new password required")
		return
	}
	if len(req.NewPassword) < 8 {
		respondError(w, http.StatusBadRequest, "new password must be at least 8 characters long")
		return
	}

	// Get current user from database
	user, exists := h.userStore.GetUserByEmail(r.Context(), sub)
	if !exists {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	// Verify old password
	if err := auth.CheckPassword(user.Password, req.OldPassword); err != nil {
		respondError(w, http.StatusUnauthorized, "invalid current password")
		return
	}

	// Hash new password
	hashedPassword, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	// Update password in database
	user.Password = hashedPassword
	user.Updated = time.Now().UTC().Unix()
	if err := h.userStore.UpdateUser(r.Context(), user); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update password")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"message": "password changed successfully",
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

	var doc model.Document
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

	// Set owner to currently authenticated user
	user := h.getAuthenticatedUser(r)
	if strings.TrimSpace(user) == "" {
		respondError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}
	doc.Owner = user

	if err := h.docStore.Upsert(doc); err != nil {
		respondError(w, http.StatusInternalServerError, "save document")
		return
	}
	respondJSON(w, http.StatusCreated, map[string]any{"status": "indexed", "id": doc.ID})
}

func (h *Handler) updateDocument(w http.ResponseWriter, r *http.Request) {
	docID := chi.URLParam(r, "id")
	if strings.TrimSpace(docID) == "" {
		respondError(w, http.StatusBadRequest, "document id is required")
		return
	}

	// Get current document
	doc, err := h.docStore.GetByID(r.Context(), docID)
	if err != nil {
		respondError(w, http.StatusNotFound, "document not found")
		return
	}

	// Check permissions (admin or owner)
	user := r.Context().Value(ctxSubjectKey{}).(string)
	roles := r.Context().Value(ctxRolesKey{}).([]string)
	isAdmin := slices.Contains(roles, "admin")

	if !isAdmin && doc.Owner != user {
		respondError(w, http.StatusForbidden, "not authorized to update this document")
		return
	}

	// Parse update payload
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	// Update editable fields
	if manufacturer, ok := payload["manufacturer"].(string); ok && strings.TrimSpace(manufacturer) != "" {
		doc.Manufacturer = manufacturer
	}

	if model, ok := payload["model"].(string); ok && strings.TrimSpace(model) != "" {
		doc.Model = model
	}

	if subtitle, ok := payload["subtitle"].(string); ok {
		doc.Subtitle = subtitle
	}

	if description, ok := payload["description"].(string); ok {
		doc.Description = description
	}

	// Update tags
	doc.Tags = parseTags(payload["tags"])

	// Handle new files (with base64 data)
	newFiles, ok := payload["newFiles"].([]any)
	if ok {
		for _, nf := range newFiles {
			fileMap, ok := nf.(map[string]any)
			if !ok {
				continue
			}

			// Create file entry
			file := model.DocumentFile{
				Name:     toString(fileMap["name"]),
				Type:     toString(fileMap["type"]),
				MIMEType: toString(fileMap["mimetype"]),
				Page:     int(toInt64(fileMap["page"])),
			}

			if dataStr, ok := fileMap["data"].(string); ok && dataStr != "" {
				// Store in blob and get ContainerInfo
				dataBytes, err := base64.StdEncoding.DecodeString(dataStr)
				if err != nil {
					respondError(w, http.StatusBadRequest, "invalid base64 data")
					return
				}

				ci, err := h.blob.Save(dataBytes, file.MIMEType, file.Name)
				if err != nil {
					respondError(w, http.StatusInternalServerError, "save file")
					return
				}
				file.Container = ci
			}

			doc.Files = append(doc.Files, file)
		}
	}

	// Handle deleted files
	deletedFiles, ok := payload["deletedFiles"].([]any)
	if ok {
		for _, df := range deletedFiles {
			fileMap, ok := df.(map[string]any)
			if !ok {
				continue
			}

			deletedName := toString(fileMap["name"])

			// Find and delete file from blob store
			for i := range doc.Files {
				if doc.Files[i].Name == deletedName && doc.Files[i].Container != nil {
					if err := h.blob.DeleteByInfo(doc.Files[i].Container); err != nil {
						h.log.Warn("failed to mark file as deleted in blob store", "error", err, "container", doc.Files[i].Container.ContainerNumber)
					}
					doc.Files[i].Container.Deleted = true
				}
			}

			// Remove file from document
			doc.Files = slices.DeleteFunc(doc.Files, func(f model.DocumentFile) bool {
				return f.Name == deletedName
			})
		}
	}

	// Update timestamps
	doc.LastModifiedAt = time.Now()

	// Save updated document
	if err := h.docStore.Upsert(doc); err != nil {
		respondError(w, http.StatusInternalServerError, "save document")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"status": "updated", "id": doc.ID})
}

func toString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func toInt64(v any) int64 {
	if f, ok := v.(float64); ok {
		return int64(f)
	}
	return 0
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

func (h *Handler) deleteDocument(w http.ResponseWriter, r *http.Request) {
	docID := chi.URLParam(r, "id")
	if strings.TrimSpace(docID) == "" {
		respondError(w, http.StatusBadRequest, "document id is required")
		return
	}

	ctx := r.Context()
	user := h.getAuthenticatedUser(r)

	// Get the document to check permissions
	doc, err := h.docStore.GetByID(ctx, docID)
	if err != nil {
		respondError(w, http.StatusNotFound, "document not found")
		return
	}

	// Check permissions: only admin or document owner can delete
	roles, ok := r.Context().Value(ctxRolesKey{}).([]string)
	isAdmin := false
	if ok {
		for _, role := range roles {
			if role == "admin" {
				isAdmin = true
				break
			}
		}
	}

	// Admin can delete everything, users can only delete their own
	if !isAdmin && doc.Owner != user {
		respondError(w, http.StatusForbidden, "not authorized to delete this document")
		return
	}

	// Mark all files as deleted in the blob container metadata (update .inf files)
	for i := range doc.Files {
		if doc.Files[i].Container != nil {
			if err := h.blob.DeleteByInfo(doc.Files[i].Container); err != nil {
				// Log the error but continue deleting other files
				h.log.Warn("failed to mark file as deleted in blob store", "error", err, "container", doc.Files[i].Container.ContainerNumber)
			}
		}
	}

	// Delete the document
	if err := h.docStore.DeleteByID(ctx, docID); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete document")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"status": "deleted", "id": docID})
}

func (h *Handler) storeBlobFiles(doc *model.Document) error {
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

		info, err := h.blob.Save(data, doc.Files[i].MIMEType, doc.Files[i].Name)
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

	// Extract authentication from Authorization header
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

	// Check if user has access (guest can only see public, authenticated can see public + own private)
	if doc.PrivateFile {
		if !isAuthenticated || username != doc.Owner {
			respondError(w, http.StatusForbidden, "access denied")
			return
		}
	}

	// Find the file
	var file *model.DocumentFile
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

func (h *Handler) searchEffects(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	skip := parseInt64Param(r, "skip", 0)
	limit := parseInt64Param(r, "limit", 20)
	sort := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order") // "asc" or "desc"

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	result, err := h.effectStore.SearchEffects(ctx, query, skip, limit, sort, order)
	if err != nil {
		h.log.Error("search effects failed", "error", err, "query", query, "skip", skip, "limit", limit)
		respondError(w, http.StatusInternalServerError, "search effects failed")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"results": result.Items,
		"count":   len(result.Items),
		"total":   result.Total,
		"skip":    result.Skip,
		"limit":   result.Limit,
	})
}

func (h *Handler) getEffect(w http.ResponseWriter, r *http.Request) {
	effectID := chi.URLParam(r, "id")
	if strings.TrimSpace(effectID) == "" {
		respondError(w, http.StatusBadRequest, "effect id is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	effect, err := h.effectStore.GetEffectByID(ctx, effectID)
	if err != nil {
		respondError(w, http.StatusNotFound, "effect not found")
		return
	}

	respondJSON(w, http.StatusOK, effect)
}

func (h *Handler) getEffectImage(w http.ResponseWriter, r *http.Request) {
	effectID := chi.URLParam(r, "id")
	if strings.TrimSpace(effectID) == "" {
		respondError(w, http.StatusBadRequest, "effect id is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Get effect from store
	effect, err := h.effectStore.GetEffectByID(ctx, effectID)
	if err != nil {
		respondError(w, http.StatusNotFound, "effect not found")
		return
	}

	// Get first image if available
	if effect.Image == nil {
		respondError(w, http.StatusNotFound, "no images available")
		return
	}

	data, err := h.blob.Load(effect.Image)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load image")
		return
	}

	w.Header().Set("Content-Type", effect.Image.MIMEType)
	w.Header().Set("Cache-Control", "public, max-age=3600")
	if effect.Image.Name != "" {
		filename := url.PathEscape(effect.Image.Name)
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filename))
	} else {
		w.Header().Set("Content-Disposition", "inline")
	}
	w.Write(data)
}

func (h *Handler) getConnectorImage(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	// Decode URL-encoded characters (e.g., %2B → +)
	// Use PathUnescape (not QueryUnescape) so '+' stays '+' and is not converted to space
	decodedName, err := url.PathUnescape(name)
	if err != nil {
		decodedName = name
	}

	if strings.TrimSpace(decodedName) == "" {
		respondError(w, http.StatusBadRequest, "connector name is required")
		return
	}

	// Try to find the image with different extensions
	var data []byte
	var fileErr error
	var mimeType string

	// Try .png first
	data, fileErr = connector.GetImage(decodedName + ".png")
	if fileErr == nil {
		mimeType = "image/png"
	} else {
		// Try .jpg
		data, fileErr = connector.GetImage(decodedName + ".jpg")
		if fileErr == nil {
			mimeType = "image/jpeg"
		} else {
			// Try .jpeg
			data, fileErr = connector.GetImage(decodedName + ".jpeg")
			if fileErr == nil {
				mimeType = "image/jpeg"
			}
		}
	}

	if fileErr != nil {
		respondError(w, http.StatusNotFound, "connector image not found")
		return
	}

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(data)
}

func (h *Handler) listEffectTypes(w http.ResponseWriter, r *http.Request) {
	effectTypes, err := h.effectTypeStore.GetAllEffectTypes(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "list effect types failed")
		return
	}
	respondJSON(w, http.StatusOK, effectTypes)
}

func (h *Handler) createEffect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse multipart form (max 128MB)
	if err := r.ParseMultipartForm(128 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "failed to parse form")
		return
	}

	// Extract form values
	effectType := r.FormValue("effectType")
	manufacturer := r.FormValue("manufacturer")
	emodel := r.FormValue("model")
	voltage := r.FormValue("voltage")
	current := r.FormValue("current")
	connector := r.FormValue("connector")

	// Validate required fields
	if effectType == "" || manufacturer == "" || emodel == "" || connector == "" {
		respondError(w, http.StatusBadRequest, "missing required fields")
		return
	}

	// Create new effect
	now := time.Now()
	effect := &model.Effect{
		CreatedAt:      now,
		LastModifiedAt: now,
		EffectType:     effectType,
		Manufacturer:   manufacturer,
		Model:          emodel,
		Voltage:        voltage,
		Current:        current,
		Connector:      connector,
		Image:          nil,
	}

	// Handle image upload if provided
	file, handler, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		// Read file data
		buf, err := io.ReadAll(file)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to read image")
			return
		}

		// Save to blob store
		mimeType := handler.Header.Get("Content-Type")
		if mimeType == "" {
			mimeType = "image/jpeg"
		}
		containerInfo, err := h.blob.Save(buf, mimeType, handler.Filename)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to save image")
			return
		}
		effect.Image = containerInfo
	}

	// Save effect to database
	if err := h.effectStore.CreateEffect(ctx, effect); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create effect")
		return
	}

	// Add manufacturer to manufacturers collection if not already exists
	if err := h.effectStore.UpdateManufacturer(ctx, effect.Manufacturer); err != nil {
		// Log but don't fail - manufacturer update is secondary
		h.log.Warn("failed to update manufacturer", "error", err, "manufacturer", effect.Manufacturer)
	}

	respondJSON(w, http.StatusCreated, effect)
}

func (h *Handler) updateEffect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	effectID := chi.URLParam(r, "id")

	if effectID == "" {
		respondError(w, http.StatusBadRequest, "effect ID is required")
		return
	}

	// Get existing effect
	effect, err := h.effectStore.GetEffectByID(ctx, effectID)
	if err != nil {
		respondError(w, http.StatusNotFound, "effect not found")
		return
	}

	// Parse multipart form (max 32MB)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "failed to parse form")
		return
	}

	// Update form values if provided
	if effectType := r.FormValue("effectType"); effectType != "" {
		effect.EffectType = effectType
	}
	if manufacturer := r.FormValue("manufacturer"); manufacturer != "" {
		effect.Manufacturer = manufacturer
	}
	if model := r.FormValue("model"); model != "" {
		effect.Model = model
	}
	if voltage := r.FormValue("voltage"); voltage != "" {
		effect.Voltage = voltage
	}
	if current := r.FormValue("current"); current != "" {
		effect.Current = current
	}
	if connector := r.FormValue("connector"); connector != "" {
		effect.Connector = connector
	}

	// Update comment if provided
	if comment := r.FormValue("comment"); comment != "" {
		effect.Comment = comment
	}

	// Handle image upload if provided
	file, handler, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		// Read file data
		buf, err := io.ReadAll(file)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to read image")
			return
		}

		// Delete old image if it exists
		if effect.Image != nil {
			if err := h.blob.DeleteByInfo(effect.Image); err != nil {
				h.log.Warn("failed to delete old image", "error", err)
				// Continue anyway - don't fail the update
			}
		}

		// Save to blob store
		mimeType := handler.Header.Get("Content-Type")
		if mimeType == "" {
			mimeType = "image/jpeg"
		}
		containerInfo, err := h.blob.Save(buf, mimeType, handler.Filename)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to save image")
			return
		}

		// Always replace the single image
		effect.Image = containerInfo
	}

	// Update effect in database
	if err := h.effectStore.UpdateEffect(ctx, effect); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update effect")
		return
	}

	// Add manufacturer to manufacturers collection if not already exists
	if err := h.effectStore.UpdateManufacturer(ctx, effect.Manufacturer); err != nil {
		// Log but don't fail - manufacturer update is secondary
		h.log.Warn("failed to update manufacturer", "error", err, "manufacturer", effect.Manufacturer)
	}

	respondJSON(w, http.StatusOK, effect)
}

func (h *Handler) deleteEffect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	effectID := chi.URLParam(r, "id")

	if effectID == "" {
		respondError(w, http.StatusBadRequest, "effect ID is required")
		return
	}

	// Get existing effect to find associated images
	effect, err := h.effectStore.GetEffectByID(ctx, effectID)
	if err != nil {
		respondError(w, http.StatusNotFound, "effect not found")
		return
	}

	// Mark associated images as deleted in blob store
	if effect.Image != nil {
		if err := h.blob.DeleteByInfo(effect.Image); err != nil {
			h.log.Warn("failed to mark effect image as deleted in blob store", "error", err, "container", effect.Image.ContainerNumber)
		}
	}

	// Delete effect from database
	if err := h.effectStore.DeleteEffect(ctx, effectID); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete effect")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"status": "deleted", "id": effectID})
}
