package app

import (
	"context"
	"fmt"
	"net/http"

	"eco-van-api/internal/adapter/auth"
	httpmiddleware "eco-van-api/internal/adapter/http"
	"eco-van-api/internal/adapter/repo/pg"
	"eco-van-api/internal/adapter/telemetry"
	appconfig "eco-van-api/internal/config"
	"eco-van-api/internal/service"

	"github.com/go-chi/chi/v5"
)

// Server represents the HTTP server
type Server struct {
	router    *chi.Mux
	server    *http.Server
	config    *appconfig.Config
	telemetry *telemetry.Manager
	db        *pg.DB
}

// NewServer creates a new Server instance
func NewServer(cfg *appconfig.Config, telemetry *telemetry.Manager, db *pg.DB) *Server {
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
	setupRoutes(router, telemetry, db, cfg)

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
		db:        db,
	}
}

// setupRoutes configures the application routes
func setupRoutes(router chi.Router, telemetry *telemetry.Manager, db *pg.DB, cfg *appconfig.Config) {
	// Health check endpoint
	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok"}`)
	})

	// Readiness check endpoint
	router.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		// Check database connectivity
		if err := db.Ping(r.Context()); err != nil {
			// Use Problem JSON format for database down scenario
			httpmiddleware.WriteCustomProblem(w,
				"/errors/service-unavailable",
				"Service unavailable",
				http.StatusServiceUnavailable,
				"database ping failed",
				"/readyz")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ready"}`)
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

	// API v1 routes
	router.Route("/api/v1", func(r chi.Router) {
		// Public endpoints
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			endpoints := `["/healthz","/metrics","/auth/login","/auth/refresh","/users","/clients","/warehouses","/equipment","/drivers"]`
			fmt.Fprintf(w, `{"message":"API v1","endpoints":%s}`, endpoints)
		})

		// Metrics endpoint
		if telemetry.IsMetricsEnabled() {
			r.Get("/metrics", telemetry.Metrics.GetHandler().ServeHTTP)
		}

		// Authentication routes
		r.Route("/auth", func(r chi.Router) {
			// Create auth handler
			userRepo := pg.NewUserRepository(db)
			jwtManager := auth.NewDefaultJWTManager(cfg.Auth.JWTSecret)
			authService := service.NewAuthService(userRepo, jwtManager)
			authHandler := httpmiddleware.NewAuthHandler(authService)

			// Public auth endpoints
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.Refresh)
		})

		// Protected user management endpoints
		r.Route("/users", func(r chi.Router) {
			// Create auth handler and middleware
			userRepo := pg.NewUserRepository(db)
			jwtManager := auth.NewDefaultJWTManager(cfg.Auth.JWTSecret)
			authService := service.NewAuthService(userRepo, jwtManager)
			authHandler := httpmiddleware.NewAuthHandler(authService)
			authMiddleware := httpmiddleware.NewAuthMiddleware(jwtManager)
			rbacMiddleware := httpmiddleware.NewRBACMiddleware()

			// Require authentication for all user endpoints
			r.Use(authMiddleware.RequireAuth)

			// Read endpoints - accessible by all authenticated users (ADMIN, DISPATCHER, DRIVER, VIEWER)
			r.With(rbacMiddleware.RequireReadAccess).Group(func(r chi.Router) {
				r.Get("/", authHandler.ListUsers)
				r.Get("/{id}", authHandler.GetUser)
			})

			// Write endpoints - ADMIN only
			r.With(rbacMiddleware.RequireWriteAccess).Group(func(r chi.Router) {
				r.Post("/", authHandler.CreateUser)
				r.Delete("/{id}", authHandler.DeleteUser)
			})
		})

		// Protected client management endpoints
		r.Route("/clients", func(r chi.Router) {
			// Create client handler and middleware
			clientRepo := pg.NewClientRepository(db.GetPool())
			clientService := service.NewClientService(clientRepo)
			clientHandler := httpmiddleware.NewClientHandler(clientService)
			clientJWTManager := auth.NewDefaultJWTManager(cfg.Auth.JWTSecret)
			authMiddleware := httpmiddleware.NewAuthMiddleware(clientJWTManager)
			rbacMiddleware := httpmiddleware.NewRBACMiddleware()

			// Require authentication for all client endpoints
			r.Use(authMiddleware.RequireAuth)

			// Read endpoints - accessible by all authenticated users
			r.With(rbacMiddleware.RequireReadAccess).Group(func(r chi.Router) {
				r.Get("/", clientHandler.ListClients)
				r.Get("/{id}", clientHandler.GetClient)
			})

			// Write endpoints - ADMIN and DISPATCHER only
			r.With(rbacMiddleware.RequireWriteAccess).Group(func(r chi.Router) {
				r.Post("/", clientHandler.CreateClient)
				r.Put("/{id}", clientHandler.UpdateClient)
				r.Delete("/{id}", clientHandler.DeleteClient)
				r.Post("/{id}/restore", clientHandler.RestoreClient)
			})

			// Client Objects routes - nested under clients
			r.Route("/{clientId}/objects", func(r chi.Router) {
				// Create client object handler
				clientObjectRepo := pg.NewClientObjectRepository(db.GetPool())
				clientObjectService := service.NewClientObjectService(clientObjectRepo, clientRepo)
				clientObjectHandler := httpmiddleware.NewClientObjectHandler(clientObjectService)

				// Read endpoints - accessible by all authenticated users
				r.With(rbacMiddleware.RequireReadAccess).Group(func(r chi.Router) {
					r.Get("/", clientObjectHandler.ListClientObjects)
					r.Get("/{id}", clientObjectHandler.GetClientObject)
				})

				// Write endpoints - ADMIN and DISPATCHER only
				r.With(rbacMiddleware.RequireWriteAccess).Group(func(r chi.Router) {
					r.Post("/", clientObjectHandler.CreateClientObject)
					r.Put("/{id}", clientObjectHandler.UpdateClientObject)
					r.Delete("/{id}", clientObjectHandler.DeleteClientObject)
					r.Post("/{id}/restore", clientObjectHandler.RestoreClientObject)
				})
			})
		})

		// Protected warehouse management endpoints
		//nolint:dupl // Similar route pattern across resources but with different handlers and services
		r.Route("/warehouses", func(r chi.Router) {
			// Create warehouse handler and middleware
			warehouseRepo := pg.NewWarehouseRepository(db.GetPool())
			warehouseService := service.NewWarehouseService(warehouseRepo)
			warehouseHandler := httpmiddleware.NewWarehouseHandler(warehouseService)
			warehouseJWTManager := auth.NewDefaultJWTManager(cfg.Auth.JWTSecret)
			authMiddleware := httpmiddleware.NewAuthMiddleware(warehouseJWTManager)
			rbacMiddleware := httpmiddleware.NewRBACMiddleware()

			// Require authentication for all warehouse endpoints
			r.Use(authMiddleware.RequireAuth)

			// Read endpoints - accessible by all authenticated users
			r.With(rbacMiddleware.RequireReadAccess).Group(func(r chi.Router) {
				r.Get("/", warehouseHandler.ListWarehouses)
				r.Get("/{id}", warehouseHandler.GetWarehouse)
			})

			// Write endpoints - ADMIN and DISPATCHER only
			r.With(rbacMiddleware.RequireWriteAccess).Group(func(r chi.Router) {
				r.Post("/", warehouseHandler.CreateWarehouse)
				r.Put("/{id}", warehouseHandler.UpdateWarehouse)
				r.Delete("/{id}", warehouseHandler.DeleteWarehouse)
				r.Post("/{id}/restore", warehouseHandler.RestoreWarehouse)
			})
		})

		// Protected equipment management endpoints
		//nolint:dupl // Similar route pattern across resources but with different handlers and services
		r.Route("/equipment", func(r chi.Router) {
			// Create equipment handler and middleware
			equipmentRepo := pg.NewEquipmentRepository(db.GetPool())
			equipmentService := service.NewEquipmentService(equipmentRepo)
			equipmentHandler := httpmiddleware.NewEquipmentHandler(equipmentService)
			equipmentJWTManager := auth.NewDefaultJWTManager(cfg.Auth.JWTSecret)
			authMiddleware := httpmiddleware.NewAuthMiddleware(equipmentJWTManager)
			rbacMiddleware := httpmiddleware.NewRBACMiddleware()

			// Require authentication for all equipment endpoints
			r.Use(authMiddleware.RequireAuth)

			// Read endpoints - accessible by all authenticated users
			r.With(rbacMiddleware.RequireReadAccess).Group(func(r chi.Router) {
				r.Get("/", equipmentHandler.ListEquipment)
				r.Get("/{id}", equipmentHandler.GetEquipment)
			})

			// Write endpoints - ADMIN and DISPATCHER only
			r.With(rbacMiddleware.RequireWriteAccess).Group(func(r chi.Router) {
				r.Post("/", equipmentHandler.CreateEquipment)
				r.Put("/{id}", equipmentHandler.UpdateEquipment)
				r.Delete("/{id}", equipmentHandler.DeleteEquipment)
				r.Post("/{id}/restore", equipmentHandler.RestoreEquipment)
			})
		})

		// Protected driver management endpoints
		//nolint:dupl // Similar route pattern across resources but with different handlers and services
		r.Route("/drivers", func(r chi.Router) {
			// Create driver handler and middleware
			driverRepo := pg.NewDriverRepository(db.GetPool())
			driverService := service.NewDriverService(driverRepo)
			driverHandler := httpmiddleware.NewDriverHandler(driverService)
			driverJWTManager := auth.NewDefaultJWTManager(cfg.Auth.JWTSecret)
			authMiddleware := httpmiddleware.NewAuthMiddleware(driverJWTManager)
			rbacMiddleware := httpmiddleware.NewRBACMiddleware()

			// Require authentication for all driver endpoints
			r.Use(authMiddleware.RequireAuth)

			// Read endpoints - accessible by all authenticated users
			r.With(rbacMiddleware.RequireReadAccess).Group(func(r chi.Router) {
				r.Get("/", driverHandler.ListDrivers)
				r.Get("/{id}", driverHandler.GetDriver)
				r.Get("/available", driverHandler.ListAvailableDrivers)
			})

			// Write endpoints - ADMIN and DISPATCHER only
			r.With(rbacMiddleware.RequireWriteAccess).Group(func(r chi.Router) {
				r.Post("/", driverHandler.CreateDriver)
				r.Put("/{id}", driverHandler.UpdateDriver)
				r.Delete("/{id}", driverHandler.DeleteDriver)
				r.Post("/{id}/restore", driverHandler.RestoreDriver)
			})
		})
	})

	// 404 handler for unmatched routes
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		httpmiddleware.WriteNotFound(w, "The requested resource was not found")
	})

	// Method not allowed handler
	router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		httpmiddleware.WriteProblemWithType(w, http.StatusMethodNotAllowed,
			"/errors/method-not-allowed",
			"The HTTP method is not allowed for this resource")
	})
}

// Start starts the HTTP server
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
