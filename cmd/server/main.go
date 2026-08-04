package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"nex_play_auth/github.com/config"
	"nex_play_auth/github.com/internal/handler/auth"
	"nex_play_auth/github.com/internal/handler/middleware"
	"nex_play_auth/github.com/internal/repository/db"
	authservice "nex_play_auth/github.com/internal/service/auth_service"
	"nex_play_auth/github.com/pkg/jwt"
	"nex_play_auth/github.com/pkg/mailer"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func setUpLogger() {
	var h slog.Handler
	if os.Getenv("LOG_FORMAT") == "json" {
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		h = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	slog.SetDefault(slog.New(h))
}

func main() {
	//logger
	setUpLogger()

	//Config
	cfg := config.Load()

	//Database
	database, err := db.NewDB(cfg.DatabasePath)

	if err != nil {

		slog.Error("failed to open database", "error", err)

		os.Exit(1)
	}
	defer database.Close()

	//Repositories
	userRepo := db.NewUserRepo(database)

	otpRepo := db.NewOTPRepo(database)

	// Background job to clean up expired or used OTPs
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if err := otpRepo.DeleteExpiredOrUsed(context.Background()); err != nil {
				slog.Error("failed to delete expired OTPs", "error", err)
			}
		}
	}()

	//pkg

	//JWT
	jwtManager := jwt.NewManager(cfg.JWTSecret, cfg.JWTExpiry, cfg.RefreshTokenExpire)

	//Mailer
	mailClient := mailer.New(
		mailer.MailerConfig{
			APIKey:   cfg.MailerSendAPIKey,
			From:     cfg.MailFrom,
			FromName: cfg.MailFromName,
		},
	)

	//Services

	authSvc := authservice.NewService(userRepo, otpRepo, jwtManager, mailClient, cfg)

	//Handlers
	mux := http.NewServeMux()

	// Health check  (useful for load balancers and Docker health checks)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	//middlewares
	authMiddleware := middleware.RequiredAuth(jwtManager)
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst)

	//Handlers
	auth.NewHander(authSvc).Register(mux, authMiddleware)

	handler := middleware.Chain(mux,
		middleware.RecoverPanic,
		middleware.SecurityHeaders,
		middleware.CORS,
		middleware.Logger,
		rateLimiter.Middleware,
	)

	//HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start the server in a goroutine so we can listen for OS signals below
	go func() {
		slog.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	//Block until SIGINT or SIGTERM is received
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	slog.Info("shutting down connections (30s timeout)")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {

		slog.Error("server forced to shutdown", "error", err)
	}

	slog.Info("server stopped")
}
