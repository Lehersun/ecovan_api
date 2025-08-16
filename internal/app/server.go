package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	appconfig "eco-van-api/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	middlewareTimeout = 30 * time.Second
)

// Server represents the HTTP server
type Server struct {
	router *chi.Mux
	server *http.Server
	config *appconfig.Config
}

// NewServer creates a new Server instance
func NewServer(cfg *appconfig.Config) *Server {
	router := chi.NewRouter()

	// Add middleware placeholders
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Timeout(middlewareTimeout))

	// Setup routes
	setupRoutes(router)

	server := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	return &Server{
		router: router,
		server: server,
		config: cfg,
	}
}

// setupRoutes configures the application routes
func setupRoutes(router chi.Router) {
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

	// API v1 routes placeholder
	router.Route("/api/v1", func(r chi.Router) {
		// TODO: Add API routes here
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"message":"API v1","endpoints":["/healthz"]}`)
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
