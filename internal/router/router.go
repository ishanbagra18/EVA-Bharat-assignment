package router

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"ticket-system/internal/auth"
	"ticket-system/internal/handlers"
)

func SetupRouter(
	healthHandler *handlers.HealthHandler,
	authHandler *handlers.AuthHandler,
	ticketHandler *handlers.TicketHandler,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", healthHandler.HealthCheck)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
	})

	r.Route("/tickets", func(r chi.Router) {
		r.Use(auth.AuthMiddleware)
		r.Post("/", ticketHandler.CreateTicket)
		r.Get("/", ticketHandler.GetTickets)
		r.Get("/{id}", ticketHandler.GetTicketByID)
		r.Patch("/{id}/status", ticketHandler.UpdateTicketStatus)
	})

	// Serve Static Frontend UI
	workDir, _ := os.Getwd()
	webDir := http.Dir(filepath.Join(workDir, "web"))
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		fs := http.FileServer(webDir)
		fs.ServeHTTP(w, r)
	})

	return r
}
