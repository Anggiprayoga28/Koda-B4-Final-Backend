package main

import (
	"context"
	"fmt"
	"koda-shortlink-backend/internal/config"
	"koda-shortlink-backend/internal/database"
	"koda-shortlink-backend/internal/handler"
	"koda-shortlink-backend/internal/middleware"
	"koda-shortlink-backend/internal/repository"
	"koda-shortlink-backend/internal/service"
	"koda-shortlink-backend/internal/utils"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	db, err := database.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	var redisClient *redis.Client

	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		opt, err := redis.ParseURL(redisURL)
		if err != nil {
			log.Fatalf("Invalid REDIS_URL: %v", err)
		}
		redisClient = redis.NewClient(opt)
		log.Println("Using Redis from REDIS_URL")
	} else {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})
		log.Println("Using Redis from REDIS_HOST/REDIS_PORT")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	log.Println("Redis connection established")

	jwtUtil := utils.NewJWTUtil(cfg.JWT.Secret, cfg.JWT.AccessExpiry, cfg.JWT.RefreshExpiry)

	userRepo := repository.NewUserRepository(db.DB)
	linkRepo := repository.NewShortLinkRepository(db.DB)
	sessionRepo := repository.NewSessionRepository(db.DB)
	clickRepo := repository.NewClickRepository(db.DB)

	authService := service.NewAuthService(userRepo, sessionRepo, jwtUtil)
	linkService := service.NewLinkService(linkRepo, clickRepo, redisClient, cfg.Server.BaseURL)

	authHandler := handler.NewAuthHandler(authService)
	linkHandler := handler.NewLinkHandler(linkService)
	redirectHandler := handler.NewRedirectHandler(linkService, redisClient)

	authMiddleware := middleware.NewAuthMiddleware(jwtUtil)
	rateLimiter := middleware.NewRateLimiter(redisClient, 100, time.Minute)

	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())

	router.GET("/:shortCode", redirectHandler.Redirect)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	api := router.Group("/api/v1")
	api.Use(rateLimiter.Limit())

	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/logout", authMiddleware.RequireAuth(), authHandler.Logout)

	}

	links := api.Group("/links")
	{
		links.POST("", authMiddleware.OptionalAuth(), linkHandler.CreateLink)
		links.GET("", authMiddleware.RequireAuth(), linkHandler.GetUserLinks)
		links.GET("/:shortCode", authMiddleware.RequireAuth(), linkHandler.GetLinkByShortCode)
		links.PUT("/:shortCode", authMiddleware.RequireAuth(), linkHandler.UpdateLink)
		links.DELETE("/:shortCode", authMiddleware.RequireAuth(), linkHandler.DeleteLink)
	}

	dashboard := api.Group("/dashboard")
	dashboard.Use(authMiddleware.RequireAuth())
	{
		dashboard.GET("/stats", linkHandler.GetDashboardStats)
	}

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
