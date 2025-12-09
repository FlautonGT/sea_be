package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"seaply/internal/config"
	"seaply/internal/database"
	"seaply/internal/middleware"
	"seaply/internal/payment"
	"seaply/internal/provider"
	"seaply/internal/router"
	"seaply/internal/services"
	"seaply/internal/storage"
	"seaply/internal/utils"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Setup logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if os.Getenv("ENVIRONMENT") == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	log.Info().
		Str("environment", cfg.Server.Environment).
		Str("port", cfg.Server.Port).
		Msg("Starting Seaply Backend API")

	// Initialize database
	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()
	log.Info().Msg("Connected to PostgreSQL")

	// Initialize Redis
	redis, err := database.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redis.Close()
	log.Info().Msg("Connected to Redis")

	// Initialize S3 Storage
	s3Storage, err := storage.NewS3Storage(cfg.S3)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to initialize S3 storage, some features may not work")
		// Create a nil storage - routes should handle this gracefully
		s3Storage = nil
	} else {
		log.Info().Msg("Initialized S3 storage")
	}

	// Initialize providers
	providerManager := initializeProviders(cfg)
	log.Info().Msg("Initialized product providers")

	// Initialize payment gateways
	paymentManager := initializePaymentGateways(cfg)
	log.Info().Msg("Initialized payment gateways")

	// Start provider and payment health checks
	ctx := context.Background()
	providerManager.StartHealthCheck(ctx, 5*time.Minute)
	paymentManager.StartHealthCheck(ctx, 5*time.Minute)

	// Initialize services
	jwtService := utils.NewJWTService(cfg.JWT)
	emailService := services.NewEmailService()
	log.Info().Msg("Initialized email service")

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService)
	rateLimiter := middleware.NewRateLimiter(redis)

	// Setup router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.SecurityHeaders)
	r.Use(chimiddleware.Compress(5))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.Server.AllowOrigins},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		utils.WriteSuccessJSON(w, map[string]string{
			"status":  "healthy",
			"version": "2.0.0",
		})
	})

	// API routes
	router.SetupRoutes(r, &router.Dependencies{
		Config:          cfg,
		DB:              db,
		Redis:           redis,
		S3:              s3Storage,
		JWTService:      jwtService,
		EmailService:    emailService,
		AuthMiddleware:  authMiddleware,
		RateLimiter:     rateLimiter,
		ProviderManager: providerManager,
		PaymentManager:  paymentManager,
	})

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info().Str("port", cfg.Server.Port).Msg("Server started")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited properly")
}

// printBanner prints the application banner
func printBanner() {
	banner := `
   ____       _       
  / ___| __ _| |_ ___ 
 | |  _ / _  | __/ _ \
 | |_| | (_| | ||  __/
  \____|\__,_|\__\___|
                      
 API Backend v2.0.0
 ====================
`
	fmt.Println(banner)
}

// initializeProviders initializes all product providers
func initializeProviders(cfg *config.Config) *provider.Manager {
	manager := provider.NewManager()

	// Initialize Digiflazz provider
	if cfg.Provider.Digiflazz.Username != "" {
		digiflazz := provider.NewDigiflazzProvider(
			cfg.Provider.Digiflazz.Username,
			cfg.Provider.Digiflazz.APIKey,
			cfg.Provider.Digiflazz.WebhookSecret,
			cfg.Server.Environment == "production",
		)
		manager.Register(digiflazz)
		log.Info().Msg("Registered Digiflazz provider")
	}

	// Initialize VIP Reseller provider
	if cfg.Provider.VIPReseller.APIID != "" {
		vipreseller := provider.NewVIPResellerProvider(
			cfg.Provider.VIPReseller.APIID,
			cfg.Provider.VIPReseller.APIKey,
		)
		manager.Register(vipreseller)
		log.Info().Msg("Registered VIP Reseller provider")
	}

	// Initialize BangJeff provider
	if cfg.Provider.BangJeff.MemberID != "" {
		bangjeff := provider.NewBangJeffProvider(
			cfg.Provider.BangJeff.MemberID,
			cfg.Provider.BangJeff.SecretKey,
			cfg.Provider.BangJeff.WebhookToken,
		)
		manager.Register(bangjeff)
		log.Info().Msg("Registered BangJeff provider")
	}

	return manager
}

// initializePaymentGateways initializes all payment gateways
func initializePaymentGateways(cfg *config.Config) *payment.Manager {
	manager := payment.NewManager()
	isProduction := cfg.Server.Environment == "production"

	// Initialize LinkQu gateway (QRIS)
	if cfg.Payment.LinkQu.ClientID != "" {
		linkqu := payment.NewLinkQuGateway(
			cfg.Payment.LinkQu.ClientID,
			cfg.Payment.LinkQu.ClientSecret,
			cfg.Payment.LinkQu.Username,
			cfg.Payment.LinkQu.PIN,
			isProduction,
		)
		manager.Register(linkqu)
		log.Info().Msg("Registered LinkQu gateway (QRIS)")
	}

	// Initialize BCA gateway (VA BCA)
	if cfg.Payment.BCA.ClientID != "" {
		bca := payment.NewBCAGateway(
			cfg.Payment.BCA.ClientID,
			cfg.Payment.BCA.ClientSecret,
			cfg.Payment.BCA.APIKey,
			cfg.Payment.BCA.APISecret,
			cfg.Payment.BCA.CorporateID,
			isProduction,
		)
		manager.Register(bca)
		log.Info().Msg("Registered BCA gateway (VA BCA)")
	}

	// Initialize BRI gateway (VA BRI via SNAP API)
	log.Info().
		Str("client_id", cfg.Payment.BRI.ClientID).
		Str("partner_id", cfg.Payment.BRI.PartnerID).
		Str("private_key_path", cfg.Payment.BRI.PrivateKeyPath).
		Str("base_url", cfg.Payment.BRI.BaseURL).
		Bool("has_client_id", cfg.Payment.BRI.ClientID != "").
		Bool("has_private_key_path", cfg.Payment.BRI.PrivateKeyPath != "").
		Msg("BRI gateway config check")

	if cfg.Payment.BRI.ClientID != "" && cfg.Payment.BRI.PrivateKeyPath != "" {
		bri, err := payment.NewBRIGateway(
			cfg.Payment.BRI.ClientID,
			cfg.Payment.BRI.ClientSecret,
			cfg.Payment.BRI.PartnerID,
			cfg.Payment.BRI.PrivateKeyPath,
			cfg.Payment.BRI.BaseURL,
			cfg.Payment.BRI.CallbackURL,
		)
		if err != nil {
			log.Error().
				Err(err).
				Str("private_key_path", cfg.Payment.BRI.PrivateKeyPath).
				Msg("Failed to initialize BRI gateway")
		} else {
			manager.Register(bri)
			log.Info().Msg("Registered BRI gateway (VA BRI via SNAP API)")
		}
	} else {
		log.Warn().
			Bool("has_client_id", cfg.Payment.BRI.ClientID != "").
			Bool("has_private_key_path", cfg.Payment.BRI.PrivateKeyPath != "").
			Msg("BRI gateway not initialized - missing required config")
	}

	// Initialize Xendit gateway (VA Permata, Mandiri, etc.)
	if cfg.Payment.Xendit.SecretKey != "" {
		xendit := payment.NewXenditGateway(
			cfg.Payment.Xendit.SecretKey,
			cfg.Payment.Xendit.CallbackToken,
			isProduction,
		)
		manager.Register(xendit)
		log.Info().Msg("Registered Xendit gateway (VA Permata, Mandiri, BNI)")
	}

	// Initialize Midtrans gateway (GoPay, ShopeePay)
	if cfg.Payment.Midtrans.ServerKey != "" {
		midtrans, err := payment.NewMidtransGateway(payment.MidtransGatewayConfig{
			ServerKey:    cfg.Payment.Midtrans.ServerKey,
			ClientKey:    cfg.Payment.Midtrans.ClientKey,
			IsProduction: isProduction,
			BaseURL:      cfg.Payment.Midtrans.BaseURL,
			CallbackURL:  cfg.Payment.Midtrans.CallbackURL,
		})
		if err != nil {
			log.Warn().Err(err).Msg("Failed to initialize Midtrans gateway")
		} else {
			manager.Register(midtrans)
			log.Info().Msg("Registered Midtrans gateway (GoPay, ShopeePay)")
		}
	}

	// Initialize DANA gateway (QRIS + DANA SNAP)
	if cfg.Payment.DANA.PartnerID != "" && cfg.Payment.DANA.MerchantID != "" {
		dana, err := payment.NewDANAGateway(payment.DANAGatewayConfig{
			PartnerID:      cfg.Payment.DANA.PartnerID,
			ClientSecret:   cfg.Payment.DANA.ClientSecret,
			MerchantID:     cfg.Payment.DANA.MerchantID,
			ShopId:         cfg.Payment.DANA.ShopId,
			ChannelID:      cfg.Payment.DANA.ChannelID,
			Origin:         cfg.Payment.DANA.Origin,
			Environment:    cfg.Payment.DANA.Environment,
			PrivateKey:     cfg.Payment.DANA.PrivateKey,
			PrivateKeyPath: cfg.Payment.DANA.PrivateKeyPath,
			CallbackURL:    cfg.Payment.DANA.CallbackURL,
			ReturnURL:      cfg.Payment.DANA.ReturnURL,
			DefaultMCC:     cfg.Payment.DANA.DefaultMCC,
			Debug:          cfg.Payment.DANA.Debug,
		})
		if err != nil {
			log.Warn().Err(err).Msg("Failed to initialize DANA gateway")
		} else {
			manager.Register(dana)
			log.Info().Msg("Registered DANA gateway")
		}
	}

	return manager
}
