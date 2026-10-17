// @title           YouPass API
// @version         1.0
// @description     YouPass Technical Assignment backend API (URL Shortener + Seamless Authentication).
// @termsOfService  https://youpass.vn/terms
// @host            localhost:8080
// @BasePath        /
// @schemes         http https
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/mongo"

	_ "github.com/youpass/backend/docs"

	"github.com/youpass/backend/configs"
	"github.com/youpass/backend/internal/auth"
	"github.com/youpass/backend/internal/middleware"
	"github.com/youpass/backend/internal/share"
	"github.com/youpass/backend/internal/submissions"
	"github.com/youpass/backend/internal/tracking"
	"github.com/youpass/backend/internal/users"
	"github.com/youpass/backend/pkg/cache"
	"github.com/youpass/backend/pkg/database"
	"github.com/youpass/backend/pkg/logger"
	"github.com/youpass/backend/pkg/redis"
)

const (
	shortCodeLength = 10
	codeMaxRetries  = 5
	workerMaxRetries = 5
)

func main() {
	cfg := configs.Load()
	log := logger.New()

	startupCtx, startupCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer startupCancel()

	mongoClient, err := database.Connect(startupCtx, cfg.MongoURI)
	if err != nil {
		log.Error("failed to connect to MongoDB", "error", err)
		os.Exit(1)
	}
	defer mongoClient.Disconnect(context.Background())

	db := mongoClient.Database(cfg.MongoDatabase)
	if err := database.EnsureIndexes(startupCtx, db); err != nil {
		log.Error("failed to ensure indexes", "error", err)
		os.Exit(1)
	}

	rdb := redis.NewClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err := redis.Ping(startupCtx, rdb); err != nil {
		log.Error("failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	defer rdb.Close()

	router := buildRouter(cfg, db, rdb, log)

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: router}

	go func() {
		log.Info("api: listening", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("api: listen failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("api: shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("api: forced shutdown", "error", err)
	}
	log.Info("api: stopped")
}

// buildRouter wires all dependencies and routes.
func buildRouter(cfg *configs.Config, db *mongo.Database, rdb *redis.Client, log *slog.Logger) *gin.Engine {
	tokenManager := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTIssuer, cfg.JWTAudience)

	usersRepo := users.NewRepository(db)
	sessionsRepo := auth.NewRepository(db)
	authSvc := auth.NewService(usersRepo, sessionsRepo, tokenManager, cfg.RefreshTokenTTL, cfg.RefreshGracePeriod)
	authHandler := auth.NewHandler(authSvc, auth.CookieConfig{
		Secure:   cfg.CookieSecure,
		SameSite: parseSameSite(cfg.CookieSameSite),
		Domain:   cfg.CookieDomain,
	}, cfg.RefreshTokenTTL, log)

	submissionsRepo := submissions.NewRepository(db)
	submissionsHandler := submissions.NewHandler(submissionsRepo)

	shareRepo := share.NewRepository(db)
	shareCache := cache.New(rdb)
	shareSvc := share.NewService(shareRepo, submissionsRepo, shareCache, shortCodeLength, codeMaxRetries)
	shareHandler := share.NewHandler(shareSvc)

	trackingRepo := tracking.NewRepository(db)
	trackingSvc := tracking.NewService(shareSvc, trackingRepo, rdb, log)
	trackingHandler := tracking.NewHandler(trackingSvc)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.CORS(cfg.AllowedOrigins))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Swagger / OpenAPI.
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Authentication.
	authGroup := router.Group("/api/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", middleware.RateLimit(rdb, "rl:login", cfg.LoginRateLimit, cfg.RateLimitWindow, log), authHandler.Login)
		authGroup.POST("/refresh",
			middleware.OriginCheck(cfg.AllowedOrigins),
			middleware.RateLimit(rdb, "rl:refresh", cfg.RefreshRateLimit, cfg.RateLimitWindow, log),
			authHandler.Refresh,
		)
		authGroup.POST("/logout", middleware.OriginCheck(cfg.AllowedOrigins), authHandler.Logout)
		authGroup.GET("/me", middleware.Auth(tokenManager), authHandler.Me)
	}

	// Authenticated resources.
	protected := router.Group("/api")
	protected.Use(middleware.Auth(tokenManager))
	{
		protected.GET("/submissions", submissionsHandler.List)
		protected.POST("/submissions", submissionsHandler.Create)
		protected.POST("/submissions/:id/share", shareHandler.Create)

		protected.GET("/share-links", shareHandler.List)
		protected.GET("/share-links/:code", shareHandler.Get)
		protected.PATCH("/share-links/:code", shareHandler.Update)
		protected.DELETE("/share-links/:code", shareHandler.Delete)
		protected.GET("/share-links/:code/stats", trackingHandler.GetStats)
	}

	// Public endpoints.
	router.GET("/s/:code", shareHandler.Redirect)
	public := router.Group("/api/public")
	{
		public.GET("/shares/:code", shareHandler.PublicShare)
		public.POST("/shares/:code/views", trackingHandler.PostView)
	}

	return router
}

func parseSameSite(v string) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

