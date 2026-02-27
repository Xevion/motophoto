package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"

	"github.com/Xevion/motophoto/internal/database/db"
	"github.com/Xevion/motophoto/internal/middleware"
	"github.com/Xevion/motophoto/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	router    *chi.Mux
	db        *pgxpool.Pool
	queries   *db.Queries
	events    *service.EventService
	galleries *service.GalleryService
	sessions  *scs.SessionManager
	port      string
}

func New(pool *pgxpool.Pool, sessions *scs.SessionManager) (*Server, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	q := db.New(pool)
	s := &Server{
		router:    chi.NewRouter(),
		port:      port,
		db:        pool,
		queries:   q,
		events:    service.NewEventService(q),
		galleries: service.NewGalleryService(q),
		sessions:  sessions,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s, nil
}

func (s *Server) setupMiddleware() {
	s.router.Use(middleware.RequestID)
	s.router.Use(chimw.RealIP)
	s.router.Use(middleware.RequestLogger)
	s.router.Use(chimw.Recoverer)
	s.router.Use(httprate.LimitByRealIP(100, time.Minute))
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	s.router.Use(chimw.Compress(5))
	s.router.Use(s.sessions.LoadAndSave)
}

func (s *Server) setupRoutes() {
	s.router.Route("/api", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, req *http.Request) {
			writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		})

		r.Route("/v1", func(r chi.Router) {
			r.Post("/auth/register", s.handleRegister)
			r.Post("/auth/login", s.handleLogin)
			r.Post("/auth/logout", s.handleLogout)

			r.Get("/events", s.handleListEvents)
			r.Post("/events", s.handleCreateEvent)
			r.Get("/events/{id}", s.handleGetEvent)
			r.Patch("/events/{id}", s.handleUpdateEvent)
			r.Delete("/events/{id}", s.handleDeleteEvent)
			r.Get("/events/{eventId}/galleries", s.handleListGalleries)
			r.Post("/events/{eventId}/galleries", s.handleCreateGallery)
			r.Patch("/events/{eventId}/galleries/{id}", s.handleUpdateGallery)
			r.Delete("/events/{eventId}/galleries/{id}", s.handleDeleteGallery)
		})
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		slog.Error("failed to encode JSON response", "error", err)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func writeServiceError(w http.ResponseWriter, r *http.Request, err error, action string) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case errors.Is(err, service.ErrConflict):
		writeError(w, http.StatusConflict, "already exists")
	default:
		middleware.LoggerFromContext(r.Context()).Error("service action failed", "action", action, "error", err)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to %s", action))
	}
}

func (s *Server) Router() http.Handler {
	return s.router
}

func (s *Server) Addr() string {
	return fmt.Sprintf(":%s", s.port)
}

func (s *Server) Close() {
	// Clean up resources here
}
