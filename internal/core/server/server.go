package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"superaib/internal/core/config"
	"superaib/internal/core/logger"
	// We no longer explicitly need mux here, as NewServer now takes http.Handler
	// "github.com/gorilla/mux"
)

// Server holds the HTTP server and its configuration
type Server struct {
	handler http.Handler // Changed from *mux.Router
	cfg     *config.Config
	http    *http.Server
}

// NewServer creates a new HTTP server instance
// It now accepts a generic http.Handler, which will be our router wrapped with CORS
func NewServer(handler http.Handler, cfg *config.Config) *Server { // <--- SIGNATURE CHANGE
	return &Server{
		handler: handler, // Assign the passed handler
		cfg:     cfg,
		http: &http.Server{
			Addr:         fmt.Sprintf("%s:%s", cfg.ServerHost, cfg.ServerPort),
			Handler:      handler, // Use the provided handler
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

// Run starts the HTTP server
func (s *Server) Run() {
	logger.Log.Infof("Server starting on %s:%s", s.cfg.ServerHost, s.cfg.ServerPort)

	// Start server in a goroutine so it doesn't block
	go func() {
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatalf("Could not listen on %s: %v\n", s.http.Addr, err)
		}
	}()

	// Graceful shutdown
	s.gracefulShutdown()
}

func (s *Server) gracefulShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM) // Capture Ctrl+C and SIGTERM

	<-quit // Block until a signal is received
	logger.Log.Info("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.http.Shutdown(ctx); err != nil {
		logger.Log.Fatalf("Server forced to shutdown: %v", err)
	}
	logger.Log.Info("Server gracefully stopped.")
}
