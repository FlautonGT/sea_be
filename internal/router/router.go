package router

import (
	"unsafe"

	"gate-v2/internal/config"
	"gate-v2/internal/database"
	"gate-v2/internal/middleware"
	"gate-v2/internal/payment"
	"gate-v2/internal/provider"
	"gate-v2/internal/router/admin"
	"gate-v2/internal/router/public"
	"gate-v2/internal/router/user"
	"gate-v2/internal/services"
	"gate-v2/internal/storage"
	"gate-v2/internal/utils"

	"github.com/go-chi/chi/v5"
)

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

// Helper functions to convert Dependencies to package-specific types
func toPublicDeps(deps *Dependencies) *public.Dependencies {
	return (*public.Dependencies)(unsafe.Pointer(deps))
}

func toAdminDeps(deps *Dependencies) *admin.Dependencies {
	return (*admin.Dependencies)(unsafe.Pointer(deps))
}

func toUserDeps(deps *Dependencies) *user.Dependencies {
	return (*user.Dependencies)(unsafe.Pointer(deps))
}

func SetupRoutes(r chi.Router, deps *Dependencies) {
	// Public API v2
	r.Route("/v2", func(r chi.Router) {
		// Region validator middleware
		r.Use(middleware.RegionValidator(deps.Config.App.DefaultRegion))

		// Public endpoints (no auth required)
		setupPublicRoutes(r, deps)

		// Auth endpoints
		r.Route("/auth", func(r chi.Router) {
			setupAuthRoutes(r, deps)
		})

		// User endpoints (auth required)
		r.Group(func(r chi.Router) {
			r.Use(deps.AuthMiddleware.RequireAuth)
			setupUserRoutes(r, deps)
		})

		// Transaction endpoints (optional auth)
		r.Group(func(r chi.Router) {
			r.Use(deps.AuthMiddleware.OptionalAuth)
			setupTransactionRoutes(r, deps)
		})
	})

	// Admin API v2
	r.Route("/admin/v2", func(r chi.Router) {
		// Admin auth endpoints (no auth required)
		r.Route("/auth", func(r chi.Router) {
			setupAdminAuthRoutes(r, deps)
		})

		// Admin protected routes
		r.Group(func(r chi.Router) {
			r.Use(deps.AuthMiddleware.RequireAdminAuth)
			setupAdminRoutes(r, deps)
		})
	})

	// Webhooks (no auth, signature validation)
	r.Route("/webhooks", func(r chi.Router) {
		setupWebhookRoutes(r, deps)
	})

	// QR Code generator
	r.Get("/v2/qr/generate", public.HandleQRGenerate(toPublicDeps(deps)))
}

func setupPublicRoutes(r chi.Router, deps *Dependencies) {
	// GET /v2/regions
	r.Get("/regions", public.HandleGetRegions(toPublicDeps(deps)))

	// GET /v2/languages
	r.Get("/languages", public.HandleGetLanguages(toPublicDeps(deps)))

	// GET /v2/contacts
	r.Get("/contacts", public.HandleGetContacts(toPublicDeps(deps)))

	// GET /v2/popups
	r.Get("/popups", public.HandleGetPopups(toPublicDeps(deps)))

	// GET /v2/banners
	r.Get("/banners", public.HandleGetBanners(toPublicDeps(deps)))

	// GET /v2/categories
	r.Get("/categories", public.HandleGetCategories(toPublicDeps(deps)))

	// GET /v2/products
	r.Get("/products", public.HandleGetProducts(toPublicDeps(deps)))

	// GET /v2/populars
	r.Get("/populars", public.HandleGetPopularProducts(toPublicDeps(deps)))

	// GET /v2/fields
	r.Get("/fields", public.HandleGetFields(toPublicDeps(deps)))

	// GET /v2/sections
	r.Get("/sections", public.HandleGetSections(toPublicDeps(deps)))

	// GET /v2/skus
	r.Get("/skus", public.HandleGetSKUs(toPublicDeps(deps)))

	// GET /v2/sku/promos
	r.Get("/sku/promos", public.HandleGetPromoSKUs(toPublicDeps(deps)))

	// GET /v2/payment-channel/categories
	r.Get("/payment-channel/categories", public.HandleGetPaymentCategories(toPublicDeps(deps)))

	// GET /v2/payment-channels
	r.Get("/payment-channels", public.HandleGetPaymentChannels(toPublicDeps(deps)))

	// GET /v2/promos
	r.Get("/promos", public.HandleGetPromos(toPublicDeps(deps)))

	// POST /v2/promos/validate
	r.Post("/promos/validate", public.HandleValidatePromo(toPublicDeps(deps)))

	// GET /v2/invoices
	r.Get("/invoices", public.HandleGetInvoice(toPublicDeps(deps)))

}

func setupAuthRoutes(r chi.Router, deps *Dependencies) {
	// Rate limit auth endpoints
	r.Use(deps.RateLimiter.IPRateLimit(middleware.AuthRateLimit))
	mainDeps := toPublicDeps(deps)

	// POST /v2/auth/register
	r.Post("/register", public.HandleRegister(mainDeps))

	// POST /v2/auth/register/google
	r.Post("/register/google", public.HandleRegisterGoogle(mainDeps))

	// POST /v2/auth/verify-email
	r.Post("/verify-email", public.HandleVerifyEmail(mainDeps))

	// POST /v2/auth/resend-verification
	r.Post("/resend-verification", public.HandleResendVerification(mainDeps))

	// POST /v2/auth/login
	r.Post("/login", public.HandleLogin(mainDeps))

	// POST /v2/auth/login/google
	r.Post("/login/google", public.HandleLoginGoogle(mainDeps))

	// POST /v2/auth/verify-mfa
	r.Post("/verify-mfa", public.HandleVerifyMFA(mainDeps))

	// POST /v2/auth/forgot-password
	r.Post("/forgot-password", public.HandleForgotPassword(mainDeps))

	// POST /v2/auth/reset-password
	r.Post("/reset-password", public.HandleResetPassword(mainDeps))

	// POST /v2/auth/refresh-token
	r.Post("/refresh-token", public.HandleRefreshToken(mainDeps))

	// Protected auth routes
	r.Group(func(r chi.Router) {
		r.Use(deps.AuthMiddleware.RequireAuth)

		// POST /v2/auth/mfa/enable
		r.Post("/mfa/enable", public.HandleEnableMFA(mainDeps))

		// POST /v2/auth/mfa/verify-setup
		r.Post("/mfa/verify-setup", public.HandleVerifyMFASetup(mainDeps))

		// POST /v2/auth/mfa/disable
		r.Post("/mfa/disable", public.HandleDisableMFA(mainDeps))

		// POST /v2/auth/logout
		r.Post("/logout", public.HandleLogout(mainDeps))
	})
}

func setupUserRoutes(r chi.Router, deps *Dependencies) {
	mainDeps := toPublicDeps(deps)

	// GET /v2/user/profile
	r.Get("/user/profile", public.HandleGetProfile(mainDeps))

	// PUT /v2/user/profile
	r.Put("/user/profile", public.HandleUpdateProfile(mainDeps))

	// POST /v2/user/change-password
	r.Post("/user/change-password", public.HandleChangePassword(mainDeps))

	// GET /v2/transactions
	r.Get("/transactions", public.HandleGetUserTransactions(mainDeps))

	// GET /v2/mutations
	r.Get("/mutations", public.HandleGetMutations(mainDeps))

	// GET /v2/reports
	r.Get("/reports", public.HandleGetReports(mainDeps))

	// GET /v2/deposits
	r.Get("/deposits", public.HandleGetDeposits(mainDeps))

	// POST /v2/deposits/inquiries
	r.Post("/deposits/inquiries", public.HandleDepositInquiry(mainDeps))

	// POST /v2/deposits
	r.Post("/deposits", public.HandleCreateDeposit(mainDeps))
}

func setupTransactionRoutes(r chi.Router, deps *Dependencies) {
	// Rate limit order endpoints
	r.Use(deps.RateLimiter.RateLimit(middleware.OrderRateLimit))
	mainDeps := toPublicDeps(deps)

	// POST /v2/account/inquiries
	r.Post("/account/inquiries", public.HandleAccountInquiry(mainDeps))

	// POST /v2/orders/inquiries
	r.Post("/orders/inquiries", public.HandleOrderInquiry(mainDeps))

	// POST /v2/orders
	r.Post("/orders", public.HandleCreateOrder(mainDeps))
}

func setupAdminAuthRoutes(r chi.Router, deps *Dependencies) {
	r.Use(deps.RateLimiter.IPRateLimit(middleware.AuthRateLimit))

	// POST /admin/v2/auth/login
	r.Post("/login", admin.HandleAdminLogin(toAdminDeps(deps)))

	// POST /admin/v2/auth/verify-mfa
	r.Post("/verify-mfa", admin.HandleAdminVerifyMFA(toAdminDeps(deps)))

	// POST /admin/v2/auth/refresh-token
	r.Post("/refresh-token", admin.HandleAdminRefreshToken(toAdminDeps(deps)))

	// Protected logout
	r.Group(func(r chi.Router) {
		r.Use(deps.AuthMiddleware.RequireAdminAuth)
		r.Post("/logout", admin.HandleAdminLogout(toAdminDeps(deps)))
	})
}

func setupAdminRoutes(r chi.Router, deps *Dependencies) {
	r.Use(deps.RateLimiter.RateLimit(middleware.AdminRateLimit))

	// Admin Management
	r.Route("/admins", func(r chi.Router) {
		r.With(deps.AuthMiddleware.RequirePermission("admin:read")).Get("/", admin.HandleGetAdmins(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("admin:read")).Get("/{adminId}", admin.HandleGetAdmin(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("admin:create")).Post("/", admin.HandleCreateAdmin(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("admin:update")).Put("/{adminId}", admin.HandleUpdateAdmin(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("admin:delete")).Delete("/{adminId}", admin.HandleDeleteAdmin(toAdminDeps(deps)))
	})

	// Roles
	r.Route("/roles", func(r chi.Router) {
		r.With(deps.AuthMiddleware.RequirePermission("role:manage")).Get("/", admin.HandleGetRoles(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("role:manage")).Put("/{roleCode}/permissions", admin.HandleUpdateRolePermissions(toAdminDeps(deps)))
	})

	// Providers
	r.Route("/providers", func(r chi.Router) {
		r.With(deps.AuthMiddleware.RequirePermission("provider:read")).Get("/", admin.HandleGetProviders(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("provider:read")).Get("/{providerId}", admin.HandleGetProvider(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("provider:create")).Post("/", admin.HandleCreateProvider(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("provider:update")).Put("/{providerId}", admin.HandleUpdateProvider(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("provider:delete")).Delete("/{providerId}", admin.HandleDeleteProvider(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("provider:update")).Post("/{providerId}/test", admin.HandleTestProvider(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("sku:sync")).Post("/{providerId}/sync", admin.HandleSyncProvider(toAdminDeps(deps)))
	})

	// Payment Channels
	r.Route("/payment-channels", func(r chi.Router) {
		r.With(deps.AuthMiddleware.RequirePermission("gateway:read")).Get("/", admin.HandleAdminGetPaymentChannels(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("gateway:read")).Get("/assignments", admin.HandleGetChannelAssignments(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("gateway:read")).Get("/{channelId}", admin.HandleGetPaymentChannel(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("gateway:create")).Post("/", admin.HandleCreatePaymentChannel(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("gateway:update")).Put("/{channelId}", admin.HandleUpdatePaymentChannel(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("gateway:update")).Put("/{paymentCode}/assignment", admin.HandleUpdateChannelAssignment(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("gateway:delete")).Delete("/{channelId}", admin.HandleDeletePaymentChannel(toAdminDeps(deps)))
	})

	// Payment Channel Categories
	r.Route("/payment-channel-categories", func(r chi.Router) {
		r.With(deps.AuthMiddleware.RequirePermission("gateway:read")).Get("/", admin.HandleGetPaymentChannelCategories(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("gateway:create")).Post("/", admin.HandleCreatePaymentChannelCategory(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("gateway:update")).Put("/{categoryId}", admin.HandleUpdatePaymentChannelCategory(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("gateway:delete")).Delete("/{categoryId}", admin.HandleDeletePaymentChannelCategory(toAdminDeps(deps)))
	})

	// Products
	r.Route("/products", func(r chi.Router) {
		r.With(deps.AuthMiddleware.RequirePermission("product:read")).Get("/", admin.HandleAdminGetProducts(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("product:read")).Get("/{productId}", admin.HandleAdminGetProduct(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("product:create")).Post("/", admin.HandleCreateProduct(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("product:update")).Put("/{productId}", admin.HandleUpdateProduct(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("product:delete")).Delete("/{productId}", admin.HandleDeleteProduct(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("product:read")).Get("/{productId}/fields", admin.HandleAdminGetFields(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("product:update")).Put("/{productId}/fields", admin.HandleUpdateFields(toAdminDeps(deps)))
	})

	// Categories
	r.Route("/categories", func(r chi.Router) {
		r.With(deps.AuthMiddleware.RequirePermission("product:read")).Get("/", admin.HandleAdminGetCategories(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("product:create")).Post("/", admin.HandleCreateCategory(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("product:update")).Put("/{categoryId}", admin.HandleUpdateCategory(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("product:delete")).Delete("/{categoryId}", admin.HandleDeleteCategory(toAdminDeps(deps)))
	})

	// Sections
	r.Route("/sections", func(r chi.Router) {
		r.With(deps.AuthMiddleware.RequirePermission("product:read")).Get("/", admin.HandleAdminGetSections(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("product:create")).Post("/", admin.HandleCreateSection(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("product:update")).Put("/{sectionId}", admin.HandleUpdateSection(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("product:delete")).Delete("/{sectionId}", admin.HandleDeleteSection(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("product:update")).Put("/{sectionId}/products", admin.HandleAssignSectionProducts(toAdminDeps(deps)))
	})

	// SKUs
	r.Route("/skus", func(r chi.Router) {
		r.With(deps.AuthMiddleware.RequirePermission("sku:read")).Get("/", admin.HandleAdminGetSKUs(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("sku:read")).Get("/images", admin.HandleAdminGetSKUImages(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("sku:read")).Get("/{skuId}", admin.HandleAdminGetSKU(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("sku:create")).Post("/", admin.HandleCreateSKU(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("sku:update")).Put("/{skuId}", admin.HandleUpdateSKU(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("sku:delete")).Delete("/{skuId}", admin.HandleDeleteSKU(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("sku:update")).Put("/bulk-price", admin.HandleBulkUpdatePrice(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("sku:sync")).Post("/sync", admin.HandleSyncSKUs(toAdminDeps(deps)))
	})

	// Transactions
	r.Route("/transactions", func(r chi.Router) {
		r.With(deps.AuthMiddleware.RequirePermission("transaction:read")).Get("/", admin.HandleAdminGetTransactions(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("transaction:read")).Get("/{transactionId}", admin.HandleAdminGetTransaction(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("transaction:update")).Put("/{transactionId}/status", admin.HandleUpdateTransactionStatus(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("transaction:refund")).Post("/{transactionId}/refund", admin.HandleRefundTransaction(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("transaction:manual")).Post("/{transactionId}/retry", admin.HandleRetryTransaction(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("transaction:manual")).Post("/{transactionId}/manual", admin.HandleManualProcess(toAdminDeps(deps)))
	})

	// Users
	r.Route("/users", func(r chi.Router) {
		r.With(deps.AuthMiddleware.RequirePermission("user:read")).Get("/", admin.HandleAdminGetUsers(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("user:read")).Get("/{userId}", admin.HandleAdminGetUser(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("user:suspend")).Put("/{userId}/status", admin.HandleUpdateUserStatus(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("user:balance")).Post("/{userId}/balance", admin.HandleAdjustBalance(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("user:read")).Get("/{userId}/transactions", admin.HandleUserTransactions(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("user:read")).Get("/{userId}/mutations", admin.HandleUserMutations(toAdminDeps(deps)))
	})

	// Promos
	r.Route("/promos", func(r chi.Router) {
		r.With(deps.AuthMiddleware.RequirePermission("promo:read")).Get("/", admin.HandleAdminGetPromos(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("promo:read")).Get("/{promoId}", admin.HandleAdminGetPromo(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("promo:create")).Post("/", admin.HandleCreatePromo(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("promo:update")).Put("/{promoId}", admin.HandleUpdatePromo(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("promo:delete")).Delete("/{promoId}", admin.HandleDeletePromo(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("promo:read")).Get("/{promoId}/stats", admin.HandleGetPromoStats(toAdminDeps(deps)))
	})

	// Banners
	r.Route("/banners", func(r chi.Router) {
		r.With(deps.AuthMiddleware.RequirePermission("content:banner")).Get("/", admin.HandleAdminGetBanners(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("content:banner")).Post("/", admin.HandleCreateBanner(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("content:banner")).Put("/{bannerId}", admin.HandleUpdateBanner(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("content:banner")).Delete("/{bannerId}", admin.HandleDeleteBanner(toAdminDeps(deps)))
	})

	// Popups
	r.Route("/popups", func(r chi.Router) {
		r.With(deps.AuthMiddleware.RequirePermission("content:popup")).Get("/", admin.HandleAdminGetPopups(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("content:popup")).Post("/", admin.HandleCreatePopup(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("content:popup")).Put("/{region}", admin.HandleUpdatePopup(toAdminDeps(deps)))
	})

	// Deposits
	r.Route("/deposits", func(r chi.Router) {
		r.With(deps.AuthMiddleware.RequirePermission("transaction:read")).Get("/", admin.HandleAdminGetDeposits(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("transaction:read")).Get("/{depositId}", admin.HandleAdminGetDeposit(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("transaction:manual")).Post("/{depositId}/confirm", admin.HandleConfirmDeposit(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("transaction:update")).Post("/{depositId}/cancel", admin.HandleCancelDeposit(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("transaction:refund")).Post("/{depositId}/refund", admin.HandleRefundDeposit(toAdminDeps(deps)))
	})

	// Invoices
	r.Route("/invoices", func(r chi.Router) {
		r.With(deps.AuthMiddleware.RequirePermission("transaction:read")).Get("/", admin.HandleAdminGetInvoices(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("transaction:read")).Get("/search", admin.HandleSearchInvoice(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("transaction:update")).Post("/{invoiceNumber}/send-email", admin.HandleSendInvoiceEmail(toAdminDeps(deps)))
	})

	// Reports
	r.Route("/reports", func(r chi.Router) {
		r.With(deps.AuthMiddleware.RequirePermission("report:read")).Get("/dashboard", admin.HandleGetDashboard(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("report:read")).Get("/revenue", admin.HandleGetRevenueReport(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("report:read")).Get("/transactions", admin.HandleGetTransactionReport(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("report:read")).Get("/products", admin.HandleGetProductReport(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("report:read")).Get("/providers", admin.HandleGetProviderReport(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("report:export")).Post("/export", admin.HandleExportReport(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("report:export")).Get("/export/{exportId}", admin.HandleGetExportStatus(toAdminDeps(deps)))
	})

	// Audit Logs
	r.Route("/audit-logs", func(r chi.Router) {
		r.With(deps.AuthMiddleware.RequirePermission("audit:read")).Get("/", admin.HandleGetAuditLogs(toAdminDeps(deps)))
	})

	// Settings
	r.Route("/settings", func(r chi.Router) {
		r.With(deps.AuthMiddleware.RequirePermission("setting:read")).Get("/", admin.HandleGetSettings(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("setting:update")).Put("/{category}", admin.HandleUpdateSettings(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("setting:read")).Get("/contacts", admin.HandleGetContactSettings(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("setting:update")).Put("/contacts", admin.HandleUpdateContactSettings(toAdminDeps(deps)))
	})

	// Regions
	r.Route("/regions", func(r chi.Router) {
		r.With(deps.AuthMiddleware.RequirePermission("setting:read")).Get("/", admin.HandleAdminGetRegions(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("setting:update")).Post("/", admin.HandleCreateRegion(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("setting:update")).Put("/{regionId}", admin.HandleUpdateRegion(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("setting:update")).Delete("/{regionId}", admin.HandleDeleteRegion(toAdminDeps(deps)))
	})

	// Languages
	r.Route("/languages", func(r chi.Router) {
		r.With(deps.AuthMiddleware.RequirePermission("setting:read")).Get("/", admin.HandleAdminGetLanguages(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("setting:update")).Post("/", admin.HandleCreateLanguage(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("setting:update")).Put("/{languageId}", admin.HandleUpdateLanguage(toAdminDeps(deps)))
		r.With(deps.AuthMiddleware.RequirePermission("setting:update")).Delete("/{languageId}", admin.HandleDeleteLanguage(toAdminDeps(deps)))
	})
}

func setupWebhookRoutes(r chi.Router, deps *Dependencies) {
	mainDeps := toPublicDeps(deps)
	
	// Provider webhooks
	r.Post("/digiflazz", public.HandleDigiflazzWebhook(mainDeps))
	r.Post("/vipreseller", public.HandleVIPResellerWebhook(mainDeps))
	r.Post("/bangjeff", public.HandleBangJeffWebhook(mainDeps))

	// Payment webhooks
	r.Post("/linkqu", public.HandleLinkQuWebhook(mainDeps))
	r.Post("/bca", public.HandleBCAWebhook(mainDeps))
	r.Post("/bri", public.HandleBRIWebhook(mainDeps))
	r.Post("/xendit", public.HandleXenditWebhook(mainDeps))
	r.Post("/midtrans", public.HandleMidtransWebhook(mainDeps))
	r.Post("/dana", public.HandleDANAWebhook(mainDeps))
}









