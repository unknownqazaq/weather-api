package auth

import (
	"encoding/json"
	"net/http"
	"strings"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func AuthMiddleware(jwtManager *JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "missing authorization header"})
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "invalid authorization header format"})
				return
			}

			tokenStr := parts[1]
			claims, err := jwtManager.Validate(tokenStr)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "invalid or expired token"})
				return
			}

			ctx := ContextWithUserClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := UserClaimsFromContext(r.Context())
			if !ok {
				writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
				return
			}

			roleAllowed := false
			for _, role := range allowedRoles {
				if claims.Role == role {
					roleAllowed = true
					break
				}
			}

			if !roleAllowed {
				writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "access denied"})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
