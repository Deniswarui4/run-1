package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Server
	Port        string
	GinMode     string
	Environment string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// JWT
	JWTSecret      string
	JWTExpiryHours int

	// Authboss
	CookieSecret  string
	SessionSecret string

	// Pesapal
	PesapalConsumerKey    string
	PesapalConsumerSecret string
	PesapalCallbackURL    string
	PesapalIPNID          string
	PesapalEnvironment    string // "sandbox" or "live"

	// Resend
	ResendAPIKey string
	FromEmail    string
	FromName     string

	// Storage
	StorageType        string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSRegion          string
	AWSBucketName      string
	AWSEndpoint        string
	LocalStoragePath   string

	// Rate Limiting
	RateLimitRequests int
	RateLimitWindow   time.Duration

	// Platform
	DefaultPlatformFeePercentage   float64
	DefaultWithdrawalFeePercentage float64
	Currency                       string

	// Frontend
	FrontendURL string
}

func LoadConfig() *Config {
	rateLimitWindow, _ := time.ParseDuration(getEnv("RATE_LIMIT_WINDOW", "1m"))
	jwtExpiry, _ := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	rateLimitReq, _ := strconv.Atoi(getEnv("RATE_LIMIT_REQUESTS", "100"))
	platformFee, _ := strconv.ParseFloat(getEnv("DEFAULT_PLATFORM_FEE_PERCENTAGE", "5.0"), 64)
	withdrawalFee, _ := strconv.ParseFloat(getEnv("DEFAULT_WITHDRAWAL_FEE_PERCENTAGE", "2.5"), 64)

	return &Config{
		Port:        getEnv("PORT", "8080"),
		GinMode:     getEnv("GIN_MODE", "debug"),
		Environment: getEnv("ENVIRONMENT", "development"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "event_ticketing"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		JWTSecret:      getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		JWTExpiryHours: jwtExpiry,

		CookieSecret:  getEnv("AUTHBOSS_COOKIE_SECRET", ""),
		SessionSecret: getEnv("AUTHBOSS_SESSION_SECRET", ""),

		PesapalConsumerKey:    getEnv("PESAPAL_CONSUMER_KEY", ""),
		PesapalConsumerSecret: getEnv("PESAPAL_CONSUMER_SECRET", ""),
		PesapalCallbackURL:    getEnv("PESAPAL_CALLBACK_URL", "http://localhost:3000/payment/callback"),
		PesapalIPNID:          getEnv("PESAPAL_IPN_ID", ""),
		PesapalEnvironment:    getEnv("PESAPAL_ENVIRONMENT", "sandbox"),

		ResendAPIKey: getEnv("RESEND_API_KEY", ""),
		FromEmail:    getEnv("FROM_EMAIL", "noreply@example.com"),
		FromName:     getEnv("FROM_NAME", "Event Ticketing"),

		StorageType:        getEnv("STORAGE_TYPE", "local"),
		AWSAccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
		AWSRegion:          getEnv("AWS_REGION", "us-east-1"),
		AWSBucketName:      getEnv("AWS_BUCKET_NAME", ""),
		AWSEndpoint:        getEnv("AWS_ENDPOINT", ""),
		LocalStoragePath:   getEnv("LOCAL_STORAGE_PATH", "./storage"),

		RateLimitRequests: rateLimitReq,
		RateLimitWindow:   rateLimitWindow,

		DefaultPlatformFeePercentage:   platformFee,
		DefaultWithdrawalFeePercentage: withdrawalFee,
		Currency:                       getEnv("CURRENCY", "NGN"),

		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
