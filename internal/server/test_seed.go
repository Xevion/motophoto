package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Xevion/motophoto/internal/database/db"
)

// SeedTestData creates mock data in the database for testing photo filtering
// Only available in development mode (ENVIRONMENT=development)
func (s *Server) handleSeedTestData(w http.ResponseWriter, r *http.Request) {
	// Guard: only allow in development mode
	if os.Getenv("ENVIRONMENT") != "development" {
		writeError(w, http.StatusForbidden, "test endpoint only available in development mode")
		return
	}

	ctx := context.Background()

	result, err := seedTestPhotos(ctx, s.queries)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to seed data: %v", err))
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

type SeedResult struct {
	Message    string `json:"message"`
	UserID     string `json:"user_id"`
	EventID    string `json:"event_id"`
	GalleryID  string `json:"gallery_id"`
	PhotoCount int    `json:"photo_count"`
	TestURL    string `json:"test_url"`
}

func seedTestPhotos(ctx context.Context, q *db.Queries) (*SeedResult, error) {
	// Create test user
	userID, err := gonanoid.New()
	if err != nil {
		return nil, fmt.Errorf("generate user id: %w", err)
	}

	_, err = q.CreateUser(ctx, db.CreateUserParams{
		ID:          userID,
		Email:       fmt.Sprintf("test-photographer-%d@motophoto.local", time.Now().Unix()),
		PasswordHash: "$2a$10$dummy", // dummy hash
		DisplayName: "Test Photographer",
		Role:        db.UserRolePhotographer,
	})
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	// Create test event
	eventID, err := gonanoid.New()
	if err != nil {
		return nil, fmt.Errorf("generate event id: %w", err)
	}

	_, err = q.CreateEvent(ctx, db.CreateEventParams{
		ID:              eventID,
		PhotographerID:  userID,
		Slug:            fmt.Sprintf("test-event-%d", time.Now().Unix()),
		Name:            "Test Motocross Event",
		Sport:           "motocross",
		Location:        pgtype.Text{String: "Test Track, USA", Valid: true},
		Description:     pgtype.Text{String: "A test event for photo filtering", Valid: true},
		Tags:            []string{"test", "filters", "demo"},
		Status:          db.EventStatusPublished,
		Date:            pgtype.Date{Time: time.Now(), Valid: true},
		SortOrder:       1,
	})
	if err != nil {
		return nil, fmt.Errorf("create event: %w", err)
	}

	// Create test gallery
	galleryID, err := gonanoid.New()
	if err != nil {
		return nil, fmt.Errorf("generate gallery id: %w", err)
	}

	_, err = q.CreateGallery(ctx, db.CreateGalleryParams{
		ID:      galleryID,
		EventID: eventID,
		Slug:    "morning-heats",
		Name:    "Morning Heats",
		Description: pgtype.Text{
			String: "Photos from the morning heat races",
			Valid:  true,
		},
		SortOrder: 1,
	})
	if err != nil {
		return nil, fmt.Errorf("create gallery: %w", err)
	}

	// Create test photos with different capture times
	eventTime := time.Now().Add(-7 * 24 * time.Hour) // one week ago
	photoCount := 0

	photoTimes := []struct {
		name   string
		offset time.Duration
	}{
		{"morning-1", 0},
		{"morning-2", 30 * time.Minute},
		{"morning-3", 1 * time.Hour},
		{"lunch-break", 3 * time.Hour},
		{"afternoon-1", 5 * time.Hour},
		{"afternoon-2", 5*time.Hour + 30*time.Minute},
		{"afternoon-3", 6 * time.Hour},
		{"finals-1", 7 * time.Hour},
		{"finals-2", 7*time.Hour + 30*time.Minute},
		{"finals-3", 8 * time.Hour},
	}

	for _, photo := range photoTimes {
		photoID, err := gonanoid.New()
		if err != nil {
			return nil, fmt.Errorf("generate photo id: %w", err)
		}

		takenAtTime := eventTime.Add(photo.offset)
		takenAt := pgtype.Timestamptz{Time: takenAtTime, Valid: true}

		_, err = q.CreatePhoto(ctx, db.CreatePhotoParams{
			ID:          photoID,
			GalleryID:   galleryID,
			StorageKey:  fmt.Sprintf("test-photos/%s/%s.jpg", galleryID, photo.name),
			PreviewKey:  fmt.Sprintf("test-photos/%s/%s-preview.jpg", galleryID, photo.name),
			Filename:    fmt.Sprintf("%s.jpg", photo.name),
			ContentType: "image/jpeg",
			SizeBytes:   1024 * 100, // 100KB
		})
		if err != nil {
			return nil, fmt.Errorf("create photo: %w", err)
		}

		// Confirm the photo with dimensions and EXIF time
		_, err = q.ConfirmPhoto(ctx, db.ConfirmPhotoParams{
			ID:        photoID,
			GalleryID: galleryID,
			Width:     pgtype.Int4{Int32: 1920, Valid: true},
			Height:    pgtype.Int4{Int32: 1080, Valid: true},
			SizeBytes: 1024 * 100,
			TakenAt:   takenAt,
		})
		if err != nil {
			return nil, fmt.Errorf("confirm photo: %w", err)
		}

		photoCount++
	}

	return &SeedResult{
		Message: fmt.Sprintf("Successfully seeded %d test photos", photoCount),
		UserID:  userID,
		EventID: eventID,
		GalleryID: galleryID,
		PhotoCount: photoCount,
		TestURL: fmt.Sprintf("/api/v1/events/%s/galleries/%s/photos", eventID, galleryID),
	}, nil
}
