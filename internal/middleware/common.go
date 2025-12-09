package middleware

import (
	"context"
	"net/http"
	"runtime/debug"
	"time"

	"gate-v2/internal/utils"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const (
	RegionContextKey contextKey = "region"
	RequestIDContextKey contextKey = "request_id"
)

// RequestID adds a unique request ID to each request
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		w.Header().Set("X-Request-ID", requestID)
		ctx := context.WithValue(r.Context(), RequestIDContextKey, requestID)
		r = r.WithContext(ctx)
		
		next.ServeHTTP(w, r)
	})
}

// Recoverer recovers from panics
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Error().
					Interface("panic", rec).
					Str("stack", string(debug.Stack())).
					Str("path", r.URL.Path).
					Str("method", r.Method).
					Msg("Panic recovered")

				utils.WriteErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR",
					"Terjadi kesalahan internal server", "")
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}

// Timeout sets a timeout for request processing
func Timeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			r = r.WithContext(ctx)
			
			done := make(chan struct{})
			go func() {
				next.ServeHTTP(w, r)
				close(done)
			}()

			select {
			case <-done:
				return
			case <-ctx.Done():
				utils.WriteErrorJSON(w, http.StatusGatewayTimeout, "REQUEST_TIMEOUT",
					"Request timeout", "")
				return
			}
		})
	}
}

// RegionValidator validates and extracts region from query params
func RegionValidator(defaultRegion string) func(http.Handler) http.Handler {
	validRegions := map[string]bool{
		"ID": true,
		"MY": true,
		"PH": true,
		"SG": true,
		"TH": true,
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			region := r.URL.Query().Get("region")
			if region == "" {
				region = defaultRegion
			}

			if !validRegions[region] {
				utils.WriteErrorJSON(w, http.StatusBadRequest, "INVALID_REGION",
					"Region tidak valid",
					"Region code must be one of: ID, MY, PH, SG, TH")
				return
			}

			ctx := context.WithValue(r.Context(), RegionContextKey, region)
			r = r.WithContext(ctx)
			
			next.ServeHTTP(w, r)
		})
	}
}

// ContentType ensures Content-Type is application/json for POST/PUT/PATCH
func ContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			ct := r.Header.Get("Content-Type")
			if ct != "" && ct != "application/json" && 
			   !isMultipartFormData(ct) {
				utils.WriteErrorJSON(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE",
					"Content-Type must be application/json or multipart/form-data", "")
				return
			}
		}
		
		next.ServeHTTP(w, r)
	})
}

func isMultipartFormData(ct string) bool {
	return len(ct) >= 19 && ct[:19] == "multipart/form-data"
}

// GetRegionFromContext returns region from context
func GetRegionFromContext(ctx context.Context) string {
	region, ok := ctx.Value(RegionContextKey).(string)
	if !ok {
		return "ID"
	}
	return region
}

// GetRequestIDFromContext returns request ID from context
func GetRequestIDFromContext(ctx context.Context) string {
	requestID, ok := ctx.Value(RequestIDContextKey).(string)
	if !ok {
		return ""
	}
	return requestID
}

// Maintenance mode middleware
func MaintenanceMode(enabled bool, message string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if enabled {
				utils.WriteErrorJSON(w, http.StatusServiceUnavailable, "MAINTENANCE_MODE",
					"Sistem sedang dalam pemeliharaan", message)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Security headers middleware
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		
		next.ServeHTTP(w, r)
	})
}

