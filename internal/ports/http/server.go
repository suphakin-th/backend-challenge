package http

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/yourusername/userapi/pkg/auth"
)

// Server represents the HTTP server
type Server struct {
	handler *Handler
	jwtAuth *auth.JWTAuth
	router  *chi.Mux
	server  *http.Server
}

// NewServer creates a new HTTP server
func NewServer(handler *Handler, jwtAuth *auth.JWTAuth, addr string) *Server {
	s := &Server{
		handler: handler,
		jwtAuth: jwtAuth,
		router:  chi.NewRouter(),
	}

	s.setupRoutes()

	s.server = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	return s
}

// setupRoutes configures the HTTP routes
func (s *Server) setupRoutes() {
	// Global middleware
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(LoggingMiddleware)
	s.router.Use(middleware.Recoverer)

	// Public routes
	s.router.Post("/register", s.handler.RegisterHandler)
	s.router.Post("/login", s.handler.LoginHandler)

	// Protected routes
	s.router.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(s.jwtAuth))

		r.Get("/users", s.handler.GetAllUsersHandler)
		r.Post("/users", s.handler.RegisterHandler) // Create user is same as register
		r.Get("/users/{id}", s.handler.GetUserHandler)
		r.Put("/users/{id}", s.handler.UpdateUserHandler)
		r.Delete("/users/{id}", s.handler.DeleteUserHandler)
	})
}

// Start starts the HTTP server
func (s *Server) Start() {
	// Start the server in a goroutine
	go func() {
		log.Printf("Starting server on %s", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline
	if err := s.server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
