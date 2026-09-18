package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"livepoll/config"
	"livepoll/internal/auth"
	"livepoll/internal/database"
	"livepoll/internal/middleware"
	"livepoll/internal/poll"
	"livepoll/internal/realtime"
	"livepoll/internal/vote"
)

func main() {
	cfg := config.Load()
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET must be set — refusing to start with an empty signing secret")
	}

	db, disconnect, err := database.ConnectMongo(cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		log.Fatalf("mongodb: %v", err)
	}

	rdb, err := database.ConnectRedis(cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}

	// --- wire dependencies -------------------------------------------------
	counters := realtime.NewCounters(rdb)
	hub := realtime.NewHub(counters)

	authRepo := auth.NewRepository(db)
	authSvc := auth.NewService(authRepo, cfg.JWTSecret)
	authHandler := auth.NewHandler(authSvc, auth.CookieOptions{Secure: cfg.Env == "production"})

	pollRepo := poll.NewRepository(db)
	pollSvc := poll.NewService(pollRepo, counters)
	pollHandler := poll.NewHandler(pollSvc)

	voteRepo := vote.NewRepository(db)
	voteSvc := vote.NewService(voteRepo, pollRepo, counters)
	voteHandler := vote.NewHandler(voteSvc)

	pollSvc.SetVoteCounter(voteSvc.CountByOptionAdapter)

	wsHandler := realtime.NewHandler(hub, counters, cfg.FrontendURL)

	// --- Redis recovery: rebuild active polls' counters from MongoDB ------
	// so a Redis restart never permanently loses live counts (spec §36).
	startupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	if err := pollSvc.RebuildActiveCounters(startupCtx, voteSvc.CountByOptionAdapter); err != nil {
		log.Printf("warning: could not rebuild Redis counters from MongoDB at startup: %v", err)
	}
	cancel()

	// --- HTTP wiring ---------------------------------------------------------
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery(), middleware.StructuredLogger(), middleware.CORS(cfg.FrontendURL))

	parseToken := func(token string) (interface{}, error) {
		id, err := authSvc.ParseToken(token)
		if err != nil {
			return nil, err
		}
		return id, nil
	}
	requireAuth := middleware.RequireAuth(parseToken, auth.CookieName)

	r.GET("/health", func(c *gin.Context) {
		status := gin.H{"status": "ok", "service": "livepoll"}
		c.JSON(http.StatusOK, status)
	})

	api := r.Group("/api")
	{
		authRoutes := api.Group("/auth")
		authRoutes.POST("/signup", middleware.RateLimit(rate.Limit(0.5), 5), authHandler.Signup)
		authRoutes.POST("/login", middleware.RateLimit(rate.Limit(0.5), 10), authHandler.Login)
		authRoutes.POST("/logout", authHandler.Logout)
		authRoutes.GET("/me", requireAuth, authHandler.Me)

		polls := api.Group("/polls")
		polls.POST("", requireAuth, pollHandler.Create)
		polls.GET("", requireAuth, pollHandler.List)
		polls.GET("/:id", pollHandler.Get) // public
		polls.GET("/:id/export", pollHandler.Export) // public
		polls.PATCH("/:id", requireAuth, pollHandler.Update)
		polls.DELETE("/:id", requireAuth, pollHandler.Delete)

		polls.POST("/:id/vote", middleware.RateLimit(rate.Limit(2), 10), voteHandler.Cast) // public
		polls.GET("/:id/vote-status", voteHandler.Status)                                  // public
		polls.GET("/:id/ws", wsHandler.Serve)                                              // public
	}

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("livepoll backend listening on :%s (env=%s)", cfg.Port, cfg.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// --- graceful shutdown ---------------------------------------------------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("http server shutdown error: %v", err)
	}
	if err := disconnect(ctx); err != nil {
		log.Printf("mongodb disconnect error: %v", err)
	}
	if err := rdb.Close(); err != nil {
		log.Printf("redis close error: %v", err)
	}
	log.Println("shutdown complete")
}
