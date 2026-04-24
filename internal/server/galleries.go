package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Xevion/motophoto/internal/database/db"
	"github.com/Xevion/motophoto/internal/service"
)

// pgtypeTextToPtr converts a pgtype.Text to a pointer to string
func pgtypeTextToPtr(v pgtype.Text) *string {
	if !v.Valid {
		return nil
	}
	return &v.String
}

func galleryResponseFromService(g service.Gallery) GalleryResponse {
	var earliestTime, latestTime *string
	if g.EarliestPhotoTime != nil {
		t := g.EarliestPhotoTime.UTC().Format(time.RFC3339)
		earliestTime = &t
	}
	if g.LatestPhotoTime != nil {
		t := g.LatestPhotoTime.UTC().Format(time.RFC3339)
		latestTime = &t
	}

	return GalleryResponse{
		ID:                g.ID,
		Slug:              g.Slug,
		Name:              g.Name,
		Description:       g.Description,
		PhotoCount:        g.PhotoCount,
		SortOrder:         g.SortOrder,
		EarliestPhotoTime: earliestTime,
		LatestPhotoTime:   latestTime,
	}
}

func (s *Server) handleListGalleries(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "eventId")

	galleries, err := s.galleries.List(r.Context(), eventID)
	if err != nil {
		writeServiceError(w, r, err, "list galleries")
		return
	}

	data := make([]GalleryResponse, 0, len(galleries))
	for _, g := range galleries {
		// Add photo time range for each gallery
		earliest, latest, _ := s.galleries.GetPhotoTimeRange(r.Context(), g.ID)
		g.EarliestPhotoTime = earliest
		g.LatestPhotoTime = latest
		data = append(data, galleryResponseFromService(g))
	}

	writeJSON(w, http.StatusOK, ListResponse[GalleryResponse]{Data: data})
}

func (s *Server) handleGetGallery(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "eventId")
	galleryID := chi.URLParam(r, "galleryId")

	// Get gallery directly from database
	dbGallery, err := s.queries.GetGallery(r.Context(), db.GetGalleryParams{ID: galleryID, EventID: eventID})
	if err != nil {
		writeServiceError(w, r, err, "get gallery")
		return
	}

	gallery := service.Gallery{
		ID:          dbGallery.ID,
		Slug:        dbGallery.Slug,
		Name:        dbGallery.Name,
		Description: pgtypeTextToPtr(dbGallery.Description),
		PhotoCount:  dbGallery.PhotoCount,
		SortOrder:   dbGallery.SortOrder,
	}

	// Get photo time range
	earliest, latest, _ := s.galleries.GetPhotoTimeRange(r.Context(), galleryID)
	gallery.EarliestPhotoTime = earliest
	gallery.LatestPhotoTime = latest

	writeJSON(w, http.StatusOK, ItemResponse[GalleryResponse]{Data: galleryResponseFromService(gallery)})
}

func (s *Server) handleCreateGallery(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "eventId")

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	var req CreateGalleryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err))
		return
	}

	gallery, err := s.galleries.Create(r.Context(), eventID, service.CreateGalleryParams{
		Name: req.Name, Slug: req.Slug,
		Description: req.Description, SortOrder: req.SortOrder,
	})
	if err != nil {
		writeServiceError(w, r, err, "create gallery")
		return
	}

	writeJSON(w, http.StatusCreated, ItemResponse[GalleryResponse]{Data: galleryResponseFromService(*gallery)})
}

func (s *Server) handleUpdateGallery(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "eventId")
	id := chi.URLParam(r, "id")

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	var req UpdateGalleryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	gallery, err := s.galleries.Update(r.Context(), eventID, id, service.UpdateGalleryParams{
		Name: req.Name, Slug: req.Slug,
		Description: req.Description, SortOrder: req.SortOrder,
	})
	if err != nil {
		writeServiceError(w, r, err, "update gallery")
		return
	}

	writeJSON(w, http.StatusOK, ItemResponse[GalleryResponse]{Data: galleryResponseFromService(*gallery)})
}

func (s *Server) handleDeleteGallery(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "eventId")
	id := chi.URLParam(r, "id")

	if err := s.galleries.Delete(r.Context(), eventID, id); err != nil {
		writeServiceError(w, r, err, "delete gallery")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
