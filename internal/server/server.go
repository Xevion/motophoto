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
	"github.com/Xevion/motophoto/internal/shutdown"
	"github.com/Xevion/motophoto/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Options configures optional Server features that vary between dev and production.
type Options struct{}

type Server struct {
	router       *chi.Mux
	db           *pgxpool.Pool
	queries      *db.Queries
	events       *service.EventService
	galleries    *service.GalleryService
	photos       *service.PhotoService
	sessions     *scs.SessionManager
	auth         *middleware.Auth
	shutdown     *shutdown.Tracker
	privateStore storage.Store
	publicStore  storage.Store
	port         string
}

func New(pool *pgxpool.Pool, sessions *scs.SessionManager, tracker *shutdown.Tracker, privateStore, publicStore storage.Store, opts Options) (*Server, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	q := db.New(pool)
	s := &Server{
		router:       chi.NewRouter(),
		port:         port,
		db:           pool,
		queries:      q,
		events:       service.NewEventService(q),
		galleries:    service.NewGalleryService(q),
		photos:       service.NewPhotoService(q, privateStore, publicStore),
		sessions:     sessions,
		auth:         middleware.NewAuth(sessions, q),
		shutdown:     tracker,
		privateStore: privateStore,
		publicStore:  publicStore,
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
			// Development/testing endpoints
			r.Post("/test/seed", s.handleSeedTestData)

			r.Post("/auth/register", s.handleRegister)
			r.Post("/auth/login", s.handleLogin)
			r.Post("/auth/logout", s.handleLogout)

			// Public read endpoints
			r.Get("/events", s.handleListEvents)
			r.Get("/events/{id}", s.handleGetEvent)
			r.Get("/events/{eventId}/galleries", s.handleListGalleries)
			r.Get("/events/{eventId}/galleries/{galleryId}", s.handleGetGallery)
			r.Get("/events/{eventId}/galleries/{galleryId}/photos", s.handleListPhotos)

			// Authenticated routes (any role)
			r.Group(func(r chi.Router) {
				r.Use(s.auth.RequireAuth)
				r.Get("/me", s.handleGetMe)
			})

			// Photographer-only write endpoints
			r.Group(func(r chi.Router) {
				r.Use(s.auth.RequireRole(db.UserRolePhotographer))
				r.Post("/events", s.handleCreateEvent)
				r.Patch("/events/{id}", s.handleUpdateEvent)
				r.Delete("/events/{id}", s.handleDeleteEvent)
				r.Post("/events/{eventId}/galleries", s.handleCreateGallery)
				r.Patch("/events/{eventId}/galleries/{id}", s.handleUpdateGallery)
				r.Delete("/events/{eventId}/galleries/{id}", s.handleDeleteGallery)
				r.Post("/events/{eventId}/galleries/{galleryId}/photos/upload", s.handleInitUpload)
				r.Post("/events/{eventId}/galleries/{galleryId}/photos/{photoId}/confirm", s.handleConfirmUpload)
			})
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
	case errors.Is(err, service.ErrForbidden):
		writeError(w, http.StatusForbidden, "forbidden")
	case errors.Is(err, service.ErrValidation):
		writeError(w, http.StatusBadRequest, err.Error())
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
