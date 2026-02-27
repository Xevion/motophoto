package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/Xevion/motophoto/internal/service"
)

var validate = validator.New()

const maxRequestBodySize = 1 << 20 // 1 MB

func eventResponseFromService(e service.Event) EventResponse {
	return EventResponse{
		ID: e.ID, Slug: e.Slug, Name: e.Name, Sport: e.Sport,
		Status: e.Status, Location: e.Location, Description: e.Description,
		Date: e.Date, Tags: e.Tags, PhotoCount: e.PhotoCount,
	}
}

func galleryResponseFromService(g service.Gallery) GalleryResponse {
	return GalleryResponse{
		ID: g.ID, Slug: g.Slug, Name: g.Name,
		Description: g.Description,
		SortOrder:   g.SortOrder, PhotoCount: g.PhotoCount,
	}
}

func (s *Server) handleListEvents(w http.ResponseWriter, r *http.Request) {
	cursorSortOrder, cursorID, limit := parsePaginationParams(r)

	result, err := s.events.List(r.Context(), cursorSortOrder, cursorID, limit)
	if err != nil {
		writeServiceError(w, r, err, "list events")
		return
	}

	resp := ListResponse[EventResponse]{Data: make([]EventResponse, 0, len(result.Events))}
	for _, e := range result.Events {
		resp.Data = append(resp.Data, eventResponseFromService(e))
	}

	if result.NextCursorID != nil {
		c := encodeCursor(*result.NextCursorSortOrder, *result.NextCursorID)
		resp.NextCursor = &c
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleGetEvent(w http.ResponseWriter, r *http.Request) {
	idOrSlug := chi.URLParam(r, "id")

	result, err := s.events.Get(r.Context(), idOrSlug)
	if err != nil {
		writeServiceError(w, r, err, "get event")
		return
	}

	resp := eventResponseFromService(result.Event)
	resp.Galleries = make([]GalleryResponse, 0, len(result.Galleries))
	for _, g := range result.Galleries {
		resp.Galleries = append(resp.Galleries, galleryResponseFromService(g))
	}

	writeJSON(w, http.StatusOK, ItemResponse[EventResponse]{Data: resp})
}

func (s *Server) handleCreateEvent(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	var req CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err))
		return
	}

	event, err := s.events.Create(r.Context(), service.CreateEventParams{
		Name: req.Name, Slug: req.Slug, Sport: req.Sport,
		Location: req.Location, Description: req.Description,
		Date: req.Date, Status: req.Status, Tags: req.Tags,
	})
	if err != nil {
		writeServiceError(w, r, err, "create event")
		return
	}

	writeJSON(w, http.StatusCreated, ItemResponse[EventResponse]{Data: eventResponseFromService(*event)})
}

func (s *Server) handleUpdateEvent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	var req UpdateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	event, err := s.events.Update(r.Context(), id, service.UpdateEventParams{
		Name: req.Name, Slug: req.Slug, Sport: req.Sport,
		Location: req.Location, Description: req.Description,
		Date: req.Date, Status: req.Status, Tags: req.Tags,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		writeServiceError(w, r, err, "update event")
		return
	}

	writeJSON(w, http.StatusOK, ItemResponse[EventResponse]{Data: eventResponseFromService(*event)})
}

func (s *Server) handleDeleteEvent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := s.events.Delete(r.Context(), id); err != nil {
		writeServiceError(w, r, err, "delete event")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
