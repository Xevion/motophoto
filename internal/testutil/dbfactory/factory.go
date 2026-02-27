// Package dbfactory provides test data factories that insert real rows via
// the service layer or sqlc queries. Every factory function calls t.Fatal on
// error, keeping test call sites to a single line.
package dbfactory

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	gonanoid "github.com/matoous/go-nanoid/v2"

	"github.com/Xevion/motophoto/internal/database/db"
	"github.com/Xevion/motophoto/internal/service"
)

var seq int64

func nextSeq() int64 { return atomic.AddInt64(&seq, 1) }

// UserOpts overrides defaults for test user creation.
type UserOpts struct {
	ID          *string
	Email       *string
	DisplayName *string
	Role        *string
}

// User inserts a user row via the sqlc CreateUser query. Returns the user ID.
func User(ctx context.Context, t *testing.T, pool *pgxpool.Pool, opts *UserOpts) string {
	t.Helper()
	n := nextSeq()

	id, err := gonanoid.New()
	if err != nil {
		t.Fatalf("dbfactory.User: generate id: %v", err)
	}

	email := fmt.Sprintf("user-%06d@test.example", n)
	displayName := fmt.Sprintf("Test User %06d", n)
	role := db.UserRolePhotographer

	if opts != nil {
		if opts.ID != nil {
			id = *opts.ID
		}
		if opts.Email != nil {
			email = *opts.Email
		}
		if opts.DisplayName != nil {
			displayName = *opts.DisplayName
		}
		if opts.Role != nil {
			role = db.UserRole(*opts.Role)
		}
	}

	q := db.New(pool)
	_, err = q.CreateUser(ctx, db.CreateUserParams{
		ID:           id,
		Email:        email,
		PasswordHash: "hash-not-used-in-tests",
		DisplayName:  displayName,
		Role:         role,
	})
	if err != nil {
		t.Fatalf("dbfactory.User: insert: %v", err)
	}
	return id
}

// EventOpts overrides defaults for test event creation.
type EventOpts struct {
	PhotographerID *string
	Name           *string
	Slug           *string
	Sport          *string
	Status         *string
	Date           *string
	Location       *string
	Description    *string
	Tags           []string
}

// Event creates an event via the service layer. If no PhotographerID is
// provided, a new test user is created automatically.
func Event(ctx context.Context, t *testing.T, pool *pgxpool.Pool, svc *service.EventService, opts *EventOpts) *service.Event {
	t.Helper()
	n := nextSeq()

	photographerID := ""
	if opts != nil && opts.PhotographerID != nil {
		photographerID = *opts.PhotographerID
	}
	if photographerID == "" {
		photographerID = User(ctx, t, pool, nil)
	}

	name := fmt.Sprintf("Test Event %06d", n)
	slug := fmt.Sprintf("test-event-%06d", n)
	sport := "motocross"
	status := "published"

	if opts != nil {
		if opts.Name != nil {
			name = *opts.Name
		}
		if opts.Slug != nil {
			slug = *opts.Slug
		}
		if opts.Sport != nil {
			sport = *opts.Sport
		}
		if opts.Status != nil {
			status = *opts.Status
		}
	}

	params := service.CreateEventParams{
		PhotographerID: photographerID,
		Name:           name,
		Slug:           slug,
		Sport:          sport,
		Status:         &status,
		Location:       optString(opts, func(o *EventOpts) *string { return o.Location }),
		Description:    optString(opts, func(o *EventOpts) *string { return o.Description }),
		Date:           optString(opts, func(o *EventOpts) *string { return o.Date }),
		Tags:           optTags(opts),
	}

	event, err := svc.Create(ctx, params)
	if err != nil {
		t.Fatalf("dbfactory.Event: create: %v", err)
	}
	return event
}

// GalleryOpts overrides defaults for test gallery creation.
type GalleryOpts struct {
	Name        *string
	Slug        *string
	Description *string
	SortOrder   *int32
}

// Gallery creates a gallery under the given event using the service layer.
func Gallery(ctx context.Context, t *testing.T, svc *service.GalleryService, eventID string, opts *GalleryOpts) *service.Gallery {
	t.Helper()
	n := nextSeq()

	name := fmt.Sprintf("Test Gallery %06d", n)
	slug := fmt.Sprintf("test-gallery-%06d", n)

	if opts != nil {
		if opts.Name != nil {
			name = *opts.Name
		}
		if opts.Slug != nil {
			slug = *opts.Slug
		}
	}

	params := service.CreateGalleryParams{
		Name: name,
		Slug: slug,
	}
	if opts != nil {
		params.Description = opts.Description
		params.SortOrder = opts.SortOrder
	}

	g, err := svc.Create(ctx, eventID, params)
	if err != nil {
		t.Fatalf("dbfactory.Gallery: create: %v", err)
	}
	return g
}

func optString(opts *EventOpts, f func(*EventOpts) *string) *string {
	if opts == nil {
		return nil
	}
	return f(opts)
}

func optTags(opts *EventOpts) []string {
	if opts != nil && opts.Tags != nil {
		return opts.Tags
	}
	return []string{}
}
