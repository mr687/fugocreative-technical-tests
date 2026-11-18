package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/mr687/fugocreative-technical-tests/internal/config"
	"github.com/mr687/fugocreative-technical-tests/internal/database"
	"github.com/mr687/fugocreative-technical-tests/internal/handler"
	"github.com/mr687/fugocreative-technical-tests/internal/middleware"
	"github.com/mr687/fugocreative-technical-tests/internal/repository"
	"github.com/mr687/fugocreative-technical-tests/internal/service"
	"github.com/mr687/fugocreative-technical-tests/internal/worker"
)

func main() {
	config.InitLogging()
	if err := run(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startupCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	db, err := database.Open(startupCtx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()
	slog.Info("PostgreSQL connected")
	if err := database.Migrate(startupCtx, db, "migrations"); err != nil {
		return err
	}
	slog.Info("database migrations applied")

	briefRepository := repository.NewBriefRepository(db)
	jobRepository := repository.NewAIJobRepository(db)
	llmProvider := service.NewOpenAIProvider(cfg.LLMEndpoint, cfg.LLMModel, cfg.LLMAPIKey)
	aiWorker := worker.New(jobRepository, llmProvider, cfg.LLMTimeout)
	briefService := service.NewBriefService(briefRepository, aiWorker.Wake)
	briefHandler := handler.NewBriefHandler(briefService)
	adminBriefRepository := repository.NewAdminBriefRepository(db)
	adminBriefService := service.NewAdminBriefService(adminBriefRepository, aiWorker.Wake, llmProvider, cfg.LLMTimeout)
	adminBriefHandler := handler.NewAdminBriefHandler(adminBriefService, cfg.JWTSecret)
	if err := aiWorker.Start(ctx); err != nil {
		return err
	}

	adminRepository := repository.NewAdminRepository(cfg.AdminUsername, cfg.AdminPassword)
	authService := service.NewAuthService(adminRepository, cfg.JWTSecret, cfg.JWTTTL)
	authHandler := handler.NewAuthHandler(authService, cfg.JWTSecret)

	router := chi.NewRouter()
	router.Use(middleware.Logger("/ping"))
	router.Use(chimiddleware.Recoverer)
	router.Use(chimiddleware.AllowContentType(
		"application/json",
	))
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.CORSAllowedOrigin},
		AllowedMethods:   []string{"HEAD", "GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"*"},
		AllowCredentials: true,
	}))
	router.Use(chimiddleware.ClientIPFromHeader("CF-Connecting-IP"))
	router.Use(chimiddleware.CleanPath)
	router.Use(chimiddleware.Timeout(200 * time.Second))
	router.Use(chimiddleware.Heartbeat("/ping"))

	router.Route("/api", func(r chi.Router) {
		r.Mount(authHandler.Path(), authHandler.Start())
		r.Mount(briefHandler.Path(), briefHandler.Start())
		r.Mount(adminBriefHandler.Path(), adminBriefHandler.Start())
	})

	server := &http.Server{Addr: cfg.ServerAddr, Handler: router, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("HTTP shutdown failed", "error", err)
		}
	}()

	slog.Info("HTTP server listening", "address", cfg.ServerAddr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
