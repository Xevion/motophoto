package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Xevion/motophoto/internal/middleware"
	"github.com/Xevion/motophoto/internal/service"
)

func photoResponseFromService(p service.Photo) PhotoResponse {
	return PhotoResponse{
		ID:          p.ID,
		Filename:    p.Filename,
		ContentType: p.ContentType,
		PreviewURL:  p.PreviewURL,
		SizeBytes:   p.SizeBytes,
		Width:       p.Width,
		Height:      p.Height,
	}
}

func (s *Server) handleInitUpload(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "eventId")
	galleryID := chi.URLParam(r, "galleryId")
	user, _ := middleware.UserFromContext(r.Context())

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	var req InitUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err))
		return
	}

	result, err := s.photos.InitUpload(r.Context(), eventID, galleryID, user.ID, service.InitUploadParams{
		Filename:    req.Filename,
		ContentType: req.ContentType,
		SizeBytes:   req.SizeBytes,
	})
	if err != nil {
		writeServiceError(w, r, err, "init photo upload")
		return
	}

	writeJSON(w, http.StatusCreated, ItemResponse[InitUploadResponse]{Data: InitUploadResponse{
		PhotoID:   result.PhotoID,
		UploadURL: result.UploadURL,
	}})
}

func (s *Server) handleConfirmUpload(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "eventId")
	galleryID := chi.URLParam(r, "galleryId")
	photoID := chi.URLParam(r, "photoId")
	user, _ := middleware.UserFromContext(r.Context())

	photo, err := s.photos.ConfirmUpload(r.Context(), eventID, galleryID, photoID, user.ID)
	if err != nil {
		writeServiceError(w, r, err, "confirm photo upload")
		return
	}

	writeJSON(w, http.StatusOK, ItemResponse[PhotoResponse]{Data: photoResponseFromService(*photo)})
}
