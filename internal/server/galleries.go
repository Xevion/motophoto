package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Xevion/motophoto/internal/service"
)

func (s *Server) handleListGalleries(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "eventId")

	galleries, err := s.galleries.List(r.Context(), eventID)
	if err != nil {
		writeServiceError(w, r, err, "list galleries")
		return
	}

	data := make([]GalleryResponse, 0, len(galleries))
	for _, g := range galleries {
		data = append(data, galleryResponseFromService(g))
	}

	writeJSON(w, http.StatusOK, ListResponse[GalleryResponse]{Data: data})
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
