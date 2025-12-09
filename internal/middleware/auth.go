package middleware

import (
	"context"
	"net/http"
	"strings"

	"seaply/internal/domain"
	"seaply/internal/utils"
)

type contextKey string

const (
	UserContextKey   contextKey = "user"
	AdminContextKey  contextKey = "admin"
	ClaimsContextKey contextKey = "claims"
)

type AuthMiddleware struct {
	jwtService utils.JWTService
}

func NewAuthMiddleware(jwtService utils.JWTService) *AuthMiddleware {
	return &AuthMiddleware{jwtService: jwtService}
}

// OptionalAuth allows requests with or without authentication
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token != "" {
			claims, err := m.jwtService.ValidateAccessToken(token)
			if err == nil {
				ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAuth requires a valid user authentication
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			utils.WriteErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", "")
			return
		}

		claims, err := m.jwtService.ValidateAccessToken(token)
		if err != nil {
			utils.WriteErrorJSON(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired token", "")
			return
		}

		// Check if it's a user token (not admin)
		if claims.Type != "user" {
			utils.WriteErrorJSON(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid token type", "")
			return
		}

		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// RequireAdminAuth requires a valid admin authentication
func (m *AuthMiddleware) RequireAdminAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			utils.WriteErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "Admin authentication required", "")
			return
		}

		claims, err := m.jwtService.ValidateAccessToken(token)
		if err != nil {
			utils.WriteErrorJSON(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired token", "")
			return
		}

		// Check if it's an admin token
		if claims.Type != "admin" {
			utils.WriteErrorJSON(w, http.StatusUnauthorized, "INVALID_TOKEN", "Admin token required", "")
			return
		}

		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// RequirePermission checks if admin has required permission
func (m *AuthMiddleware) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaimsFromContext(r.Context())
			if claims == nil {
				utils.WriteErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", "")
				return
			}

			// Check permission
			hasPermission := false
			for _, p := range claims.Permissions {
				if p == permission {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				utils.WriteErrorJSON(w, http.StatusForbidden, "PERMISSION_DENIED",
					"Anda tidak memiliki akses untuk melakukan aksi ini",
					"Required permission: "+permission)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission checks if admin has any of the required permissions
func (m *AuthMiddleware) RequireAnyPermission(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaimsFromContext(r.Context())
			if claims == nil {
				utils.WriteErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", "")
				return
			}

			// Check if has any permission
			hasPermission := false
			for _, required := range permissions {
				for _, p := range claims.Permissions {
					if p == required {
						hasPermission = true
						break
					}
				}
				if hasPermission {
					break
				}
			}

			if !hasPermission {
				utils.WriteErrorJSON(w, http.StatusForbidden, "PERMISSION_DENIED",
					"Anda tidak memiliki akses untuk melakukan aksi ini",
					"Required one of: "+strings.Join(permissions, ", "))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Helper functions
func extractToken(r *http.Request) string {
	// Check Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	// Check query parameter (for WebSocket)
	token := r.URL.Query().Get("token")
	if token != "" {
		return token
	}

	return ""
}

func GetClaimsFromContext(ctx context.Context) *utils.TokenClaims {
	claims, ok := ctx.Value(ClaimsContextKey).(*utils.TokenClaims)
	if !ok {
		return nil
	}
	return claims
}

func GetUserIDFromContext(ctx context.Context) string {
	claims := GetClaimsFromContext(ctx)
	if claims == nil {
		return ""
	}
	return claims.Subject()
}

func GetAdminIDFromContext(ctx context.Context) string {
	claims := GetClaimsFromContext(ctx)
	if claims == nil || claims.Type != "admin" {
		return ""
	}
	return claims.Subject()
}

func GetAdminRoleFromContext(ctx context.Context) domain.RoleCode {
	claims := GetClaimsFromContext(ctx)
	if claims == nil || claims.Type != "admin" {
		return ""
	}
	return domain.RoleCode(claims.Role)
}

func GetPermissionsFromContext(ctx context.Context) []string {
	claims := GetClaimsFromContext(ctx)
	if claims == nil {
		return nil
	}
	return claims.Permissions
}

func HasPermission(ctx context.Context, permission string) bool {
	permissions := GetPermissionsFromContext(ctx)
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}
