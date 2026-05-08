package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/willie68/schematic2/backend/internal/auth"
)

type ctxSubjectKey struct{}
type ctxRolesKey struct{}

func (h *Handler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			respondError(w, http.StatusUnauthorized, "invalid authorization header")
			return
		}
		claims, err := auth.VerifyToken(h.cfg.JWTSecret, parts[1])
		if err != nil {
			respondError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), ctxSubjectKey{}, claims.Subject)
		ctx = context.WithValue(ctx, ctxRolesKey{}, claims.Roles)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
