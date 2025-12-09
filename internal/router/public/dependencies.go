package public

import (
	"gate-v2/internal/config"
	"gate-v2/internal/database"
	"gate-v2/internal/middleware"
	"gate-v2/internal/payment"
	"gate-v2/internal/provider"
	"gate-v2/internal/services"
	"gate-v2/internal/storage"
	"gate-v2/internal/utils"
)

// Dependencies matches router.Dependencies structure
type Dependencies struct {
	Config          *config.Config
	DB              *database.PostgresDB
	Redis           *database.RedisClient
	S3              *storage.S3Storage
	JWTService      utils.JWTService
	EmailService    *services.EmailService
	AuthMiddleware  *middleware.AuthMiddleware
	RateLimiter     *middleware.RateLimiter
	ProviderManager *provider.Manager
	PaymentManager  *payment.Manager
}

