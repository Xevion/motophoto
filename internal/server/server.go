package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Server struct {
	router *chi.Mux
	port   string
}

func New() (*Server, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	s := &Server{
		router: chi.NewRouter(),
		port:   port,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s, nil
}

func (s *Server) setupMiddleware() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Compress(5))
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
}

func (s *Server) setupRoutes() {
	s.router.Route("/api", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, req *http.Request) {
			writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		})

		r.Route("/v1", func(r chi.Router) {
			r.Get("/events", handleListEvents)
			r.Get("/events/{id}", handleGetEvent)
		})
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// Demo data — replace with database queries once sqlc is wired up

var demoEvents = []Event{
	{
		ID: 1, Name: "Spring MX Championship", Sport: "Motocross",
		Location: "Thunder Valley MX Park, CO", Date: "2026-03-15",
		PhotoCount: 847, Galleries: 3,
		Description: "Round 1 of the Rocky Mountain Motocross Series featuring 250 and 450 classes.",
		Tags:        []string{"motocross", "mx", "250", "450", "championship"},
	},
	{
		ID: 2, Name: "BMX Freestyle Invitational", Sport: "BMX",
		Location: "Austin, TX", Date: "2026-02-28",
		PhotoCount: 312, Galleries: 2,
		Description: "Top riders compete in park and street disciplines at the annual invitational.",
		Tags:        []string{"bmx", "freestyle", "park", "street"},
	},
	{
		ID: 3, Name: "Lone Star Rodeo Finals", Sport: "Rodeo",
		Location: "Fort Worth Stockyards, TX", Date: "2026-02-14",
		PhotoCount: 523, Galleries: 4,
		Description: "Season-ending championship rodeo with bull riding, barrel racing, and roping events.",
		Tags:        []string{"rodeo", "bull riding", "barrel racing", "roping"},
	},
	{
		ID: 4, Name: "Regional Swim Meet", Sport: "Swimming",
		Location: "Barton Springs Aquatic Center, TX", Date: "2026-01-20",
		PhotoCount: 1204, Galleries: 6,
		Description: "High school regional qualifiers — all strokes and relay events.",
		Tags:        []string{"swimming", "high school", "regional", "relay"},
	},
}

func handleListEvents(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"events": demoEvents,
		"total":  len(demoEvents),
	})
}

func handleGetEvent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	for _, e := range demoEvents {
		if fmt.Sprintf("%d", e.ID) == idStr {
			writeJSON(w, http.StatusOK, e)
			return
		}
	}
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "event not found"})
}

func (s *Server) Router() http.Handler {
	return s.router
}

func (s *Server) Addr() string {
	return fmt.Sprintf(":%s", s.port)
}

func (s *Server) Close() {
	// Clean up resources (db connections, etc.) here
}
