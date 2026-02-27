package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	gonanoid "github.com/matoous/go-nanoid/v2"

	"github.com/Xevion/motophoto/internal/database/db"
)

// Gallery is the service-layer representation with plain Go types.
type Gallery struct {
	Description *string
	ID          string
	Slug        string
	Name        string
	PhotoCount  int64
	SortOrder   int32
}

// CreateGalleryParams holds inputs for creating a gallery.
type CreateGalleryParams struct {
	Description *string
	SortOrder   *int32
	Name        string
	Slug        string
}

// UpdateGalleryParams holds optional fields for patching a gallery.
type UpdateGalleryParams struct {
	Name        *string
	Slug        *string
	Description *string
	SortOrder   *int32
}

type GalleryService struct {
	queries *db.Queries
}

func NewGalleryService(q *db.Queries) *GalleryService {
	return &GalleryService{queries: q}
}

// List returns all galleries for an event. Returns ErrNotFound if the event doesn't exist.
func (s *GalleryService) List(ctx context.Context, eventID string) ([]Gallery, error) {
	if err := s.verifyEventExists(ctx, eventID); err != nil {
		return nil, err
	}

	rows, err := s.queries.ListGalleriesByEvent(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("list galleries for event %q: %w", eventID, err)
	}

	galleries := make([]Gallery, 0, len(rows))
	for _, r := range rows {
		galleries = append(galleries, galleryFromListRow(r))
	}
	return galleries, nil
}

// Create persists a new gallery under the given event. Returns ErrNotFound if event missing, ErrConflict on duplicate slug.
func (s *GalleryService) Create(ctx context.Context, eventID string, params CreateGalleryParams) (*Gallery, error) {
	if err := s.verifyEventExists(ctx, eventID); err != nil {
		return nil, err
	}

	id, err := gonanoid.New()
	if err != nil {
		return nil, fmt.Errorf("generate gallery id: %w", err)
	}

	var sortOrder int32
	if params.SortOrder != nil {
		sortOrder = *params.SortOrder
	}

	row, err := s.queries.CreateGallery(ctx, db.CreateGalleryParams{
		ID:          id,
		EventID:     eventID,
		Slug:        params.Slug,
		Name:        params.Name,
		Description: toPgText(params.Description),
		SortOrder:   sortOrder,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrConflict
		}
		return nil, fmt.Errorf("create gallery: %w", err)
	}

	g := galleryFromModel(row)
	return &g, nil
}

// Update applies partial updates to a gallery. Returns ErrNotFound if missing.
func (s *GalleryService) Update(ctx context.Context, eventID, id string, params UpdateGalleryParams) (*Gallery, error) {
	var sortOrder pgtype.Int4
	if params.SortOrder != nil {
		sortOrder = pgtype.Int4{Int32: *params.SortOrder, Valid: true}
	}

	_, err := s.queries.UpdateGallery(ctx, db.UpdateGalleryParams{
		Name:        toPgText(params.Name),
		Slug:        toPgText(params.Slug),
		Description: toPgText(params.Description),
		SortOrder:   sortOrder,
		ID:          id,
		EventID:     eventID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update gallery %q: %w", id, err)
	}

	updated, err := s.queries.GetGallery(ctx, db.GetGalleryParams{ID: id, EventID: eventID})
	if err != nil {
		return nil, fmt.Errorf("fetch updated gallery %q: %w", id, err)
	}

	g := galleryFromGetRow(updated)
	return &g, nil
}

// Delete removes a gallery by ID and event ID.
func (s *GalleryService) Delete(ctx context.Context, eventID, id string) error {
	if err := s.queries.DeleteGallery(ctx, db.DeleteGalleryParams{ID: id, EventID: eventID}); err != nil {
		return fmt.Errorf("delete gallery %q: %w", id, err)
	}
	return nil
}

func (s *GalleryService) verifyEventExists(ctx context.Context, eventID string) error {
	_, err := s.queries.GetEvent(ctx, eventID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("verify event %q: %w", eventID, err)
	}
	return nil
}

func galleryFromListRow(r db.ListGalleriesByEventRow) Gallery {
	return Gallery{
		ID: r.ID, Slug: r.Slug, Name: r.Name,
		Description: TextPtr(r.Description),
		SortOrder:   r.SortOrder, PhotoCount: r.PhotoCount,
	}
}

func galleryFromGetRow(r db.GetGalleryRow) Gallery {
	return Gallery{
		ID: r.ID, Slug: r.Slug, Name: r.Name,
		Description: TextPtr(r.Description),
		SortOrder:   r.SortOrder, PhotoCount: r.PhotoCount,
	}
}

func galleryFromModel(r db.Gallery) Gallery {
	return Gallery{
		ID: r.ID, Slug: r.Slug, Name: r.Name,
		Description: TextPtr(r.Description),
		SortOrder:   r.SortOrder, PhotoCount: 0,
	}
}
