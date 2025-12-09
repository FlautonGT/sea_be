package admin

import (
	"seaply/internal/config"
	"seaply/internal/database"
	"seaply/internal/middleware"
	"seaply/internal/payment"
	"seaply/internal/provider"
	"seaply/internal/services"
	"seaply/internal/storage"
	"seaply/internal/utils"
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
