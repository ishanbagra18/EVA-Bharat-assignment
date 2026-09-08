package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"ticket-system/internal/db"
	"ticket-system/internal/handlers"
	"ticket-system/internal/repository"
	"ticket-system/internal/router"
	"ticket-system/internal/service"
)

func main() {
	// Attempt to load .env file if available
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/tickets.db"
	}

	log.Printf("Initializing database at: %s", dbPath)
	database, err := db.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer database.Close()

	// Dependency Injection Wiring
	userRepo := repository.NewUserRepository(database)
	ticketRepo := repository.NewTicketRepository(database)

	authService := service.NewAuthService(userRepo)
	ticketService := service.NewTicketService(ticketRepo)

	healthHandler := handlers.NewHealthHandler()
	authHandler := handlers.NewAuthHandler(authService)
	ticketHandler := handlers.NewTicketHandler(ticketService)

	appRouter := router.SetupRouter(healthHandler, authHandler, ticketHandler)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      appRouter,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server listening on port %s...", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced shutdown error: %v", err)
	}

	log.Println("Server exited cleanly.")
}
