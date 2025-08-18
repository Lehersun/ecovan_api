package app

import (
	"context"
	"fmt"
	"log"
	"net/http"

	httpmiddleware "eco-van-api/internal/adapter/http"
	"eco-van-api/internal/adapter/telemetry"
	appconfig "eco-van-api/internal/config"

	"github.com/go-chi/chi/v5"
)

// Server represents the HTTP server
type Server struct {
	router    *chi.Mux
	server    *http.Server
	config    *appconfig.Config
	telemetry *telemetry.Manager
}

// NewServer creates a new Server instance
func NewServer(cfg *appconfig.Config, telemetry *telemetry.Manager) *Server {
	router := chi.NewRouter()

	// Create middleware instance
	mw := httpmiddleware.NewMiddleware(telemetry.Logger, telemetry.Tracer, telemetry.Metrics)

	// Add custom middleware
	router.Use(mw.RequestID())
	router.Use(mw.Recover())
	router.Use(mw.AccessLog())
	router.Use(mw.Trace())
	router.Use(mw.CORS(cfg.HTTP.CORSOrigins))
	router.Use(mw.RateLimit())
	router.Use(mw.MetricsInFlight())

	// Setup routes
	setupRoutes(router, telemetry)

	server := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	return &Server{
		router:    router,
		server:    server,
		config:    cfg,
		telemetry: telemetry,
	}
}

// setupRoutes configures the application routes
func setupRoutes(router chi.Router, telemetry *telemetry.Manager) {
	// Health check endpoint
	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok"}`)
	})

	// Root endpoint
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message":"Eco Van API","version":"1.0.0"}`)
	})

	// Metrics endpoint
	if telemetry.IsMetricsEnabled() {
		router.Get("/metrics", telemetry.Metrics.GetHandler().ServeHTTP)
	}

	// API v1 routes placeholder
	router.Route("/api/v1", func(r chi.Router) {
		// TODO: Add API routes here
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"message":"API v1","endpoints":["/healthz","/metrics"]}`)
		})
	})
}

// Start starts the HTTP server
func (s *Server) Start() error {
	log.Printf("Starting server on %s", s.config.HTTP.Addr)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down server...")
	return s.server.Shutdown(ctx)
}
