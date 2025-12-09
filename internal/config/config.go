package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	S3       S3Config
	Provider ProviderConfig
	Payment  PaymentConfig
	App      AppConfig
}

type ServerConfig struct {
	Port         string
	Environment  string
	AllowOrigins string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	DBName       string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	SecretKey             string
	AccessTokenExpiry     time.Duration
	RefreshTokenExpiry    time.Duration
	MFATokenExpiry        time.Duration
	ValidationTokenExpiry time.Duration
}

type S3Config struct {
	Endpoint        string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	BaseURL         string
	// Folder paths
	ProductFolder string
	SKUFolder     string
	BannerFolder  string
	PopupFolder   string
	ProfileFolder string
	FlagFolder    string
	IconFolder    string
	PaymentFolder string
	ExportFolder  string
}

type ProviderConfig struct {
	Digiflazz   DigiflazzConfig
	VIPReseller VIPResellerConfig
	BangJeff    BangJeffConfig
}

type DigiflazzConfig struct {
	Username      string
	APIKey        string
	WebhookSecret string
	BaseURL       string
}

type VIPResellerConfig struct {
	APIID   string
	APIKey  string
	BaseURL string
}

type BangJeffConfig struct {
	MemberID     string
	SecretKey    string
	WebhookToken string
	BaseURL      string
}

type PaymentConfig struct {
	LinkQu   LinkQuConfig
	BCA      BCAConfig
	BRI      BRIConfig
	Xendit   XenditConfig
	Midtrans MidtransConfig
	DANA     DANAConfig
}

type LinkQuConfig struct {
	ClientID     string
	ClientSecret string
	Username     string
	PIN          string
	BaseURL      string
	CallbackURL  string
}

type BCAConfig struct {
	ClientID     string
	ClientSecret string
	APIKey       string
	APISecret    string
	CorporateID  string
	BaseURL      string
	CallbackURL  string
}

type BRIConfig struct {
	ClientID       string
	ClientSecret   string
	PartnerID      string // X-PARTNER-ID (similar to institution code)
	PrivateKeyPath string
	BaseURL        string
	CallbackURL    string
}

type XenditConfig struct {
	SecretKey     string
	CallbackToken string
	BaseURL       string
	CallbackURL   string
}

type MidtransConfig struct {
	ServerKey    string
	ClientKey    string
	IsProduction bool
	BaseURL      string
	CallbackURL  string
}

type DANAConfig struct {
	PartnerID      string
	ClientSecret   string
	MerchantID     string
	ShopId         string
	ChannelID      string
	Origin         string
	Environment    string
	PrivateKey     string
	PrivateKeyPath string
	CallbackURL    string
	ReturnURL      string
	DefaultMCC     string
	Debug          bool
}

type AppConfig struct {
	Name               string
	BaseURL            string // API Gateway URL (e.g., https://gateway.seaply.co)
	FrontendBaseURL    string // Frontend URL (e.g., https://gate.co.id)
	AdminBaseURL       string
	DefaultRegion      string
	DefaultLanguage    string
	OrderExpiryMinutes int
	MaxLoginAttempts   int
	LockoutDuration    time.Duration
	SessionTimeout     time.Duration
	MFARequired        bool
	MaintenanceMode    bool
	MaintenanceMessage string
	InquiryBaseURL     string
	InquiryKey         string
}

func Load() (*Config, error) {
	// Load .env file if exists
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			Environment:  getEnv("ENVIRONMENT", "development"),
			AllowOrigins: getEnv("CORS_ALLOW_ORIGINS", "*"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 30*time.Second),
		},
		Database: DatabaseConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnv("DB_PORT", "5432"),
			User:         getEnv("DB_USER", "seaply"),
			Password:     getEnv("DB_PASSWORD", "seaply_secret_password"),
			DBName:       getEnv("DB_NAME", "seaply_db"),
			SSLMode:      getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns: getIntEnv("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getIntEnv("DB_MAX_IDLE_CONNS", 10),
			MaxLifetime:  getDurationEnv("DB_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getIntEnv("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			SecretKey:             getEnv("JWT_SECRET_KEY", "your-super-secret-key-change-in-production"),
			AccessTokenExpiry:     getDurationEnv("JWT_ACCESS_TOKEN_EXPIRY", 1*time.Hour),
			RefreshTokenExpiry:    getDurationEnv("JWT_REFRESH_TOKEN_EXPIRY", 7*24*time.Hour),
			MFATokenExpiry:        getDurationEnv("JWT_MFA_TOKEN_EXPIRY", 5*time.Minute),
			ValidationTokenExpiry: getDurationEnv("JWT_VALIDATION_TOKEN_EXPIRY", 30*time.Minute),
		},
		S3: S3Config{
			Endpoint:        getEnv("S3_ENDPOINT", ""),
			Region:          getEnv("S3_REGION", "ap-southeast-1"),
			AccessKeyID:     getEnv("S3_ACCESS_KEY", getEnv("S3_ACCESS_KEY_ID", "")),
			SecretAccessKey: getEnv("S3_SECRET_KEY", getEnv("S3_SECRET_ACCESS_KEY", "")),
			Bucket:          getEnv("S3_BUCKET", "gate"),
			BaseURL:         getEnv("S3_BASE_URL", ""),
			ProductFolder:   getEnv("S3_PRODUCT_FOLDER", "products"),
			SKUFolder:       getEnv("S3_SKU_FOLDER", "skus"),
			BannerFolder:    getEnv("S3_BANNER_FOLDER", "banners"),
			PopupFolder:     getEnv("S3_POPUP_FOLDER", "popups"),
			ProfileFolder:   getEnv("S3_PROFILE_FOLDER", "profiles"),
			FlagFolder:      getEnv("S3_FLAG_FOLDER", "flags"),
			IconFolder:      getEnv("S3_ICON_FOLDER", "icons"),
			PaymentFolder:   getEnv("S3_PAYMENT_FOLDER", "payment"),
			ExportFolder:    getEnv("S3_EXPORT_FOLDER", "exports"),
		},
		Provider: ProviderConfig{
			Digiflazz: DigiflazzConfig{
				Username:      getEnv("DIGIFLAZZ_USERNAME", ""),
				APIKey:        getEnv("DIGIFLAZZ_API_KEY", ""),
				WebhookSecret: getEnv("DIGIFLAZZ_WEBHOOK_SECRET", ""),
				BaseURL:       getEnv("DIGIFLAZZ_BASE_URL", "https://api.digiflazz.com/v1"),
			},
			VIPReseller: VIPResellerConfig{
				APIID:   getEnv("VIPRESELLER_API_ID", ""),
				APIKey:  getEnv("VIPRESELLER_API_KEY", ""),
				BaseURL: getEnv("VIPRESELLER_BASE_URL", "https://vip-reseller.co.id/api"),
			},
			BangJeff: BangJeffConfig{
				MemberID:     getEnv("BANGJEFF_MEMBER_ID", ""),
				SecretKey:    getEnv("BANGJEFF_SECRET_KEY", ""),
				WebhookToken: getEnv("BANGJEFF_WEBHOOK_TOKEN", ""),
				BaseURL:      getEnv("BANGJEFF_BASE_URL", "https://api.bangjeff.com"),
			},
		},
		Payment: PaymentConfig{
			LinkQu: LinkQuConfig{
				ClientID:     getEnv("LINKQU_CLIENT_ID", ""),
				ClientSecret: getEnv("LINKQU_CLIENT_SECRET", ""),
				Username:     getEnv("LINKQU_USERNAME", ""),
				PIN:          getEnv("LINKQU_PIN", ""),
				BaseURL:      getEnv("LINKQU_BASE_URL", "https://api.linkqu.id"),
				CallbackURL:  getEnv("LINKQU_CALLBACK_URL", ""),
			},
			BCA: BCAConfig{
				ClientID:     getEnv("BCA_CLIENT_ID", ""),
				ClientSecret: getEnv("BCA_CLIENT_SECRET", ""),
				APIKey:       getEnv("BCA_API_KEY", ""),
				APISecret:    getEnv("BCA_API_SECRET", ""),
				CorporateID:  getEnv("BCA_CORPORATE_ID", ""),
				BaseURL:      getEnv("BCA_BASE_URL", "https://sandbox.bca.co.id"),
				CallbackURL:  getEnv("BCA_CALLBACK_URL", ""),
			},
			BRI: BRIConfig{
				ClientID:       getEnv("BRI_CLIENT_ID", ""),
				ClientSecret:   getEnv("BRI_CLIENT_SECRET", ""),
				PartnerID:      getEnv("BRI_PARTNER_ID", ""),
				PrivateKeyPath: getEnv("BRI_PRIVATE_KEY_PATH", "./keys/bri/rsa_private_key.pem"),
				BaseURL:        getEnv("BRI_ENDPOINT", "https://sandbox.partner.api.bri.co.id"),
				CallbackURL:    getEnv("BRI_CALLBACK_URL", ""),
			},
			Xendit: XenditConfig{
				SecretKey:     getEnv("XENDIT_SECRET_KEY", ""),
				CallbackToken: getEnv("XENDIT_CALLBACK_TOKEN", ""),
				BaseURL:       getEnv("XENDIT_BASE_URL", "https://api.xendit.co"),
				CallbackURL:   getEnv("XENDIT_CALLBACK_URL", ""),
			},
			Midtrans: MidtransConfig{
				ServerKey:    getEnv("MIDTRANS_SERVER_KEY", ""),
				ClientKey:    getEnv("MIDTRANS_CLIENT_KEY", ""),
				IsProduction: getBoolEnv("MIDTRANS_IS_PRODUCTION", false),
				BaseURL:      getEnv("MIDTRANS_BASE_URL", "https://api.sandbox.midtrans.com"),
				CallbackURL:  getEnv("MIDTRANS_CALLBACK_URL", ""),
			},
			DANA: DANAConfig{
				PartnerID:      getEnv("DANA_X_PARTNER_ID", getEnv("DANA_CLIENT_ID", "")),
				ClientSecret:   getEnv("DANA_CLIENT_SECRET", ""),
				MerchantID:     getEnv("DANA_MERCHANT_ID", ""),
				ShopId:         getEnv("DANA_SHOP_ID", ""),
				ChannelID:      getEnv("DANA_CHANNEL_ID", ""),
				Origin:         getEnv("DANA_ORIGIN", getEnv("APP_URL", "")),
				Environment:    getEnv("DANA_ENV", "SANDBOX"),
				PrivateKey:     getEnv("DANA_PRIVATE_KEY", ""),
				PrivateKeyPath: getEnv("DANA_PRIVATE_KEY_PATH", ""),
				CallbackURL:    getEnv("DANA_CALLBACK_URL", ""),
				ReturnURL:      getEnv("DANA_RETURN_URL", ""),
				DefaultMCC:     getEnv("DANA_DEFAULT_MCC", "6012"),
				Debug:          getBoolEnv("DANA_DEBUG", false),
			},
		},
		App: AppConfig{
			Name:               getEnv("APP_NAME", "Seaply"),
			BaseURL:            getEnv("APP_BASE_URL", "https://gateway.seaply.co"),
			FrontendBaseURL:    getEnv("APP_FRONTEND_URL", "https://seaply.co"),
			AdminBaseURL:       getEnv("APP_ADMIN_BASE_URL", "https://gateway.seaply.co/admin"),
			DefaultRegion:      getEnv("APP_DEFAULT_REGION", "ID"),
			DefaultLanguage:    getEnv("APP_DEFAULT_LANGUAGE", "id"),
			OrderExpiryMinutes: getIntEnv("APP_ORDER_EXPIRY_MINUTES", 60),
			MaxLoginAttempts:   getIntEnv("APP_MAX_LOGIN_ATTEMPTS", 5),
			LockoutDuration:    getDurationEnv("APP_LOCKOUT_DURATION", 15*time.Minute),
			SessionTimeout:     getDurationEnv("APP_SESSION_TIMEOUT", 1*time.Hour),
			MFARequired:        getBoolEnv("APP_MFA_REQUIRED", false),
			MaintenanceMode:    getBoolEnv("APP_MAINTENANCE_MODE", false),
			MaintenanceMessage: getEnv("APP_MAINTENANCE_MESSAGE", ""),
			InquiryBaseURL:     getEnv("INQUIRY_BASE_URL", "https://inquiry.seaply.co/game"),
			InquiryKey:         getEnv("INQUIRY_KEY", ""),
		},
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
