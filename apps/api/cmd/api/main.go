package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"

	"github.com/zenvikar/api/internal/platform"
	"github.com/zenvikar/api/internal/platform/config"
	"github.com/zenvikar/api/internal/platform/handlers"
	"github.com/zenvikar/api/internal/platform/logger"
	appmiddleware "github.com/zenvikar/api/internal/platform/middleware"
	appotel "github.com/zenvikar/api/internal/platform/otel"
	"github.com/zenvikar/api/migrations"
)

func main() {
	log := logger.New()

	cfg := config.Load()

	// Initialize OpenTelemetry trace + log providers
	providers, err := appotel.Init(cfg.OTelEndpoint)
	if err != nil {
		log.Error("failed to initialize otel", "error", err)
		os.Exit(1)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		providers.Shutdown(ctx)
	}()

	// Upgrade logger to also send logs to OTel Collector
	if providers.Logger != nil {
		log = logger.NewWithOTel(providers.Logger)
	}

	// Connect to PostgreSQL
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisURL})
	defer rdb.Close()

	deps := platform.Dependencies{
		DB:     db,
		Redis:  rdb,
		Logger: log,
		Tracer: providers.Tracer.Tracer("zenvikar-api"),
		Config: cfg,
	}

	// Register modules — modules are added here as they are implemented.
	modules := []platform.Module{}

	// Set up router
	router := chi.NewRouter()
	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.RealIP)
	router.Use(appmiddleware.Recovery(log))
	router.Use(appmiddleware.Logging(log))
	router.Use(appmiddleware.Tracing("zenvikar-api"))
	router.Use(appmiddleware.CORS(cfg.AllowedOrigins))

	// Health endpoints
	router.Get("/healthz", handlers.Liveness)
	router.Get("/readyz", handlers.Readiness(db, rdb))

	// Metrics endpoint
	router.Handle("/metrics", promhttp.Handler())

	// Register module routes
	for _, m := range modules {
		m.RegisterRoutes(router, deps)
		log.Info("registered module", "name", m.Name())
	}

	// Run migrations
	if err := migrations.RunAll(db, log); err != nil {
		log.Error("migration failed", "error", err)
		os.Exit(1)
	}

	// Start server with graceful shutdown
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Info("starting server", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-done
	log.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("server shutdown error", "error", err)
	}

	log.Info("server stopped")
}

