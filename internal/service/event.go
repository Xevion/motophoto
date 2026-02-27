package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	gonanoid "github.com/matoous/go-nanoid/v2"

	"github.com/Xevion/motophoto/internal/database/db"
)

// Event is the service-layer representation with plain Go types.
type Event struct {
	ID          string
	Slug        string
	Name        string
	Sport       string
	Status      string
	Location    *string
	Description *string
	Date        *string
	Tags        []string
	PhotoCount  int64
}

// EventWithGalleries wraps an Event with its child galleries.
type EventWithGalleries struct {
	Galleries []Gallery
	Event
}

// CreateEventParams holds inputs for creating an event.
type CreateEventParams struct {
	PhotographerID string
	Name           string
	Slug           string
	Sport          string
	Location       *string
	Description    *string
	Date           *string
	Status         *string
	Tags           []string
}

// UpdateEventParams holds optional fields for patching an event.
type UpdateEventParams struct {
	Name        *string
	Slug        *string
	Sport       *string
	Location    *string
	Description *string
	Date        *string
	Status      *string
	Tags        *[]string
	SortOrder   *int32
}

// EventListResult holds a page of events plus an optional cursor for the next page.
type EventListResult struct {
	NextCursorID        *string
	NextCursorSortOrder *int32
	Events              []Event
}

type EventService struct {
	queries *db.Queries
}

func NewEventService(q *db.Queries) *EventService {
	return &EventService{queries: q}
}

func (s *EventService) List(ctx context.Context, cursorSortOrder *int32, cursorID *string, limit int32) (*EventListResult, error) {
	params := db.ListEventsParams{LimitVal: limit}
	if cursorSortOrder != nil && cursorID != nil {
		params.CursorSortOrder = pgtype.Int4{Int32: *cursorSortOrder, Valid: true}
		params.CursorID = pgtype.Text{String: *cursorID, Valid: true}
	}

	rows, err := s.queries.ListEvents(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}

	result := &EventListResult{
		Events: make([]Event, 0, len(rows)),
	}
	for _, r := range rows {
		result.Events = append(result.Events, eventFromListRow(r))
	}

	if len(rows) == int(limit) {
		last := rows[len(rows)-1]
		result.NextCursorSortOrder = &last.SortOrder
		result.NextCursorID = &last.ID
	}

	return result, nil
}

func (s *EventService) Get(ctx context.Context, idOrSlug string) (*EventWithGalleries, error) {
	row, err := s.queries.GetEvent(ctx, idOrSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get event %q: %w", idOrSlug, err)
	}

	galRows, err := s.queries.ListGalleriesByEvent(ctx, row.ID)
	if err != nil {
		return nil, fmt.Errorf("list galleries for event %q: %w", row.ID, err)
	}

	galleries := make([]Gallery, 0, len(galRows))
	for _, g := range galRows {
		galleries = append(galleries, galleryFromListRow(g))
	}

	e := EventFromGetRow(row)
	return &EventWithGalleries{Event: e, Galleries: galleries}, nil
}

func (s *EventService) Create(ctx context.Context, params CreateEventParams) (*Event, error) {
	id, err := gonanoid.New()
	if err != nil {
		return nil, fmt.Errorf("generate event id: %w", err)
	}

	date, err := toPgDate(params.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	status := db.EventStatusDraft
	if params.Status != nil {
		if !validEventStatus(*params.Status) {
			return nil, fmt.Errorf("invalid status %q: must be draft, published, or archived", *params.Status)
		}
		status = db.EventStatus(*params.Status)
	}

	tags := params.Tags
	if tags == nil {
		tags = []string{}
	}

	row, err := s.queries.CreateEvent(ctx, db.CreateEventParams{
		ID:             id,
		PhotographerID: params.PhotographerID,
		Slug:           params.Slug,
		Name:           params.Name,
		Sport:          params.Sport,
		Location:       toPgText(params.Location),
		Description:    toPgText(params.Description),
		Tags:           tags,
		Status:         status,
		Date:           date,
		SortOrder:      0,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrConflict
		}
		return nil, fmt.Errorf("create event: %w", err)
	}

	e := eventFromModel(row)
	return &e, nil
}

func (s *EventService) Update(ctx context.Context, id string, params UpdateEventParams) (*Event, error) {
	date, err := toPgDate(params.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	var statusParam db.NullEventStatus
	if params.Status != nil {
		if !validEventStatus(*params.Status) {
			return nil, fmt.Errorf("invalid status %q: must be draft, published, or archived", *params.Status)
		}
		statusParam = db.NullEventStatus{EventStatus: db.EventStatus(*params.Status), Valid: true}
	}

	var sortOrder pgtype.Int4
	if params.SortOrder != nil {
		sortOrder = pgtype.Int4{Int32: *params.SortOrder, Valid: true}
	}

	var tags []string
	if params.Tags != nil {
		tags = *params.Tags
	}

	_, err = s.queries.UpdateEvent(ctx, db.UpdateEventParams{
		Name:        toPgText(params.Name),
		Slug:        toPgText(params.Slug),
		Sport:       toPgText(params.Sport),
		Location:    toPgText(params.Location),
		Description: toPgText(params.Description),
		Tags:        tags,
		Status:      statusParam,
		Date:        date,
		SortOrder:   sortOrder,
		ID:          id,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update event %q: %w", id, err)
	}

	updated, err := s.queries.GetEvent(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetch updated event %q: %w", id, err)
	}

	e := EventFromGetRow(updated)
	return &e, nil
}

func (s *EventService) Delete(ctx context.Context, id string) error {
	if err := s.queries.DeleteEvent(ctx, id); err != nil {
		return fmt.Errorf("delete event %q: %w", id, err)
	}
	return nil
}

// TextPtr converts a pgtype.Text to a *string, returning nil if not valid.
func TextPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

// DatePtr converts a pgtype.Date to a *string in "2006-01-02" format, returning nil if not valid.
func DatePtr(d pgtype.Date) *string {
	if !d.Valid {
		return nil
	}
	s := d.Time.Format("2006-01-02")
	return &s
}

func toPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func toPgDate(s *string) (pgtype.Date, error) {
	if s == nil {
		return pgtype.Date{}, nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return pgtype.Date{}, err
	}
	return pgtype.Date{Time: t, Valid: true}, nil
}

func validEventStatus(s string) bool {
	switch db.EventStatus(s) {
	case db.EventStatusDraft, db.EventStatusPublished, db.EventStatusArchived:
		return true
	}
	return false
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func eventFromListRow(r db.ListEventsRow) Event {
	tags := r.Tags
	if tags == nil {
		tags = []string{}
	}
	return Event{
		ID: r.ID, Slug: r.Slug, Name: r.Name, Sport: r.Sport,
		Status: string(r.Status), Location: TextPtr(r.Location),
		Description: TextPtr(r.Description), Date: DatePtr(r.Date),
		Tags: tags, PhotoCount: r.PhotoCount,
	}
}

// EventFromGetRow converts a database GetEventRow into the service-layer Event type.
func EventFromGetRow(r db.GetEventRow) Event {
	tags := r.Tags
	if tags == nil {
		tags = []string{}
	}
	return Event{
		ID: r.ID, Slug: r.Slug, Name: r.Name, Sport: r.Sport,
		Status: string(r.Status), Location: TextPtr(r.Location),
		Description: TextPtr(r.Description), Date: DatePtr(r.Date),
		Tags: tags, PhotoCount: r.PhotoCount,
	}
}

func eventFromModel(r db.Event) Event {
	tags := r.Tags
	if tags == nil {
		tags = []string{}
	}
	return Event{
		ID: r.ID, Slug: r.Slug, Name: r.Name, Sport: r.Sport,
		Status: string(r.Status), Location: TextPtr(r.Location),
		Description: TextPtr(r.Description), Date: DatePtr(r.Date),
		Tags: tags, PhotoCount: 0,
	}
}
