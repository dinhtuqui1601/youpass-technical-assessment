package configs

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds the runtime configuration for the backend services.
type Config struct {
	// AppEnv is the application environment (development, production, ...).
	AppEnv string
	// Port is the HTTP port the API server listens on.
	Port string

	// MongoURI is the MongoDB connection string.
	MongoURI string
	// MongoDatabase is the name of the application database.
	MongoDatabase string

	// RedisAddr is the Redis server address (host:port).
	RedisAddr string
	// RedisPassword is the optional Redis password.
	RedisPassword string
	// RedisDB is the Redis database index.
	RedisDB int

	// AllowedOrigins is a comma-separated list of CORS origins.
	AllowedOrigins string

	// JWTSecret signs access tokens.
	JWTSecret string
	// JWTAccessTTL is the access token lifetime.
	JWTAccessTTL time.Duration
	// JWTIssuer and JWTAudience are embedded in access token claims.
	JWTIssuer   string
	JWTAudience string

	// RefreshTokenTTL is the opaque refresh token lifetime.
	RefreshTokenTTL time.Duration
	// RefreshGracePeriod tolerates legitimate concurrent refresh requests
	// (multi-tab) before treating a rotated-out token as reuse.
	RefreshGracePeriod time.Duration

	// CookieSecure enables the Secure flag on the refresh-token cookie.
	CookieSecure bool
	// CookieSameSite is one of lax, strict, none.
	CookieSameSite string
	// CookieDomain optionally scopes the refresh-token cookie to a domain.
	CookieDomain string

	// LoginRateLimit / RefreshRateLimit bound requests per RateLimitWindow.
	LoginRateLimit   int
	RefreshRateLimit int
	RateLimitWindow  time.Duration

	// ViewWorkerConcurrency is the size of the bounded worker pool.
	ViewWorkerConcurrency int
}

// Load reads configuration from the environment. When a .env file is present
// it is loaded first (development convenience only; production should rely on
// real environment variables).
func Load() *Config {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			log.Printf("config: failed to load .env: %v", err)
		}
	}

	return &Config{
		AppEnv:         getEnv("APP_ENV", "development"),
		Port:           getEnv("PORT", "8080"),
		MongoURI:       getEnv("MONGODB_URI", getEnv("MONGO_URI", "mongodb://localhost:27017")),
		MongoDatabase:  getEnv("MONGODB_DATABASE", getEnv("MONGO_DATABASE", "youpass")),
		RedisAddr:      getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		RedisDB:        getEnvInt("REDIS_DB", 0),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:5173"),

		JWTSecret:    getEnv("JWT_ACCESS_SECRET", "dev-secret-change-me"),
		JWTAccessTTL: getDuration("JWT_ACCESS_TTL", 15*time.Minute),
		JWTIssuer:    getEnv("JWT_ISSUER", "youpass"),
		JWTAudience:  getEnv("JWT_AUDIENCE", "youpass-web"),

		RefreshTokenTTL:    getDuration("REFRESH_TOKEN_TTL", 7*24*time.Hour),
		RefreshGracePeriod: getDuration("REFRESH_GRACE_PERIOD", 30*time.Second),

		CookieSecure:   getBool("COOKIE_SECURE", false),
		CookieSameSite: getEnv("COOKIE_SAMESITE", "lax"),
		CookieDomain:   getEnv("COOKIE_DOMAIN", ""),

		LoginRateLimit:   getEnvInt("LOGIN_RATE_LIMIT", 10),
		RefreshRateLimit: getEnvInt("REFRESH_RATE_LIMIT", 30),
		RateLimitWindow:  getDuration("RATE_LIMIT_WINDOW", time.Minute),

		ViewWorkerConcurrency: getEnvInt("VIEW_WORKER_CONCURRENCY", 5),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return fallback
}

func getBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return fallback
}
