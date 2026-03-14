package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"net/http"
	"time"

	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"

	"github.com/disintegration/imaging"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/rwcarlsen/goexif/exif"

	"github.com/Xevion/motophoto/internal/database/db"
	"github.com/Xevion/motophoto/internal/storage"
)

// Photo is the service-layer representation with plain Go types.
type Photo struct {
	Width       *int32
	Height      *int32
	PreviewURL  string
	ID          string
	Filename    string
	ContentType string
	SizeBytes   int64
}

// InitUploadResult is returned by InitUpload with the presigned URL.
type InitUploadResult struct {
	PhotoID   string
	UploadURL string
}

// InitUploadParams holds validated inputs for starting an upload.
type InitUploadParams struct {
	Filename    string
	ContentType string
	SizeBytes   int64
}

var allowedContentTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

const (
	maxPhotoSize       = 50 << 20 // 50 MB
	presignExpiry      = 15 * time.Minute
	previewMaxWidth    = 1200
	previewJPEGQuality = 80
)

type PhotoService struct {
	queries      *db.Queries
	privateStore storage.Store
	publicStore  storage.Store
}

func NewPhotoService(q *db.Queries, privateStore, publicStore storage.Store) *PhotoService {
	return &PhotoService{queries: q, privateStore: privateStore, publicStore: publicStore}
}

// InitUpload validates the request, creates a pending photo row, and returns a
// presigned PUT URL for the client to upload directly to R2.
func (s *PhotoService) InitUpload(ctx context.Context, eventID, galleryID, photographerID string, params InitUploadParams) (*InitUploadResult, error) {
	if err := s.verifyOwnership(ctx, eventID, photographerID); err != nil {
		return nil, err
	}
	if err := s.verifyGallery(ctx, galleryID, eventID); err != nil {
		return nil, err
	}

	if !allowedContentTypes[params.ContentType] {
		return nil, NewValidationError(fmt.Sprintf("unsupported content type: %q", params.ContentType))
	}
	if params.SizeBytes <= 0 || params.SizeBytes > maxPhotoSize {
		return nil, NewValidationError(fmt.Sprintf("invalid size: %d bytes (max %d)", params.SizeBytes, maxPhotoSize))
	}

	id, err := gonanoid.New()
	if err != nil {
		return nil, fmt.Errorf("generate photo id: %w", err)
	}

	key := fmt.Sprintf("photos/%s/%s", galleryID, id)

	_, err = s.queries.CreatePhoto(ctx, db.CreatePhotoParams{
		ID:          id,
		GalleryID:   galleryID,
		StorageKey:  key,
		PreviewKey:  key,
		Filename:    params.Filename,
		ContentType: params.ContentType,
		SizeBytes:   params.SizeBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("create photo row: %w", err)
	}

	uploadURL, err := s.privateStore.PresignedPUT(ctx, key, params.ContentType, presignExpiry)
	if err != nil {
		return nil, fmt.Errorf("generate presigned PUT URL: %w", err)
	}

	return &InitUploadResult{
		PhotoID:   id,
		UploadURL: uploadURL,
	}, nil
}

// ConfirmUpload downloads the uploaded original from the private bucket,
// extracts metadata, generates a watermarked preview, and finalizes the
// photo row.
func (s *PhotoService) ConfirmUpload(ctx context.Context, eventID, galleryID, photoID, photographerID string) (*Photo, error) {
	if err := s.verifyOwnership(ctx, eventID, photographerID); err != nil {
		return nil, err
	}

	row, err := s.queries.GetPhoto(ctx, db.GetPhotoParams{ID: photoID, GalleryID: galleryID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get photo: %w", err)
	}
	if row.Status != "pending" {
		return nil, ErrConflict
	}

	body, err := s.privateStore.Download(ctx, row.StorageKey)
	if err != nil {
		return nil, fmt.Errorf("download original from private store: %w", err)
	}

	data, readErr := io.ReadAll(body)
	_ = body.Close()
	if readErr != nil {
		return nil, fmt.Errorf("read original: %w", readErr)
	}

	detectedType := http.DetectContentType(data[:min(512, len(data))])
	if !allowedContentTypes[detectedType] {
		return nil, fmt.Errorf("uploaded file has unsupported content type: %q", detectedType)
	}

	var width, height *int32
	var takenAt pgtype.Timestamptz

	img, _, decodeErr := image.Decode(bytes.NewReader(data))
	if decodeErr == nil {
		bounds := img.Bounds()
		w, h := safeIntToInt32(bounds.Dx()), safeIntToInt32(bounds.Dy())
		width = &w
		height = &h
	}

	exifData, exifErr := exif.Decode(bytes.NewReader(data))
	if exifErr == nil {
		if dt, dtErr := exifData.DateTime(); dtErr == nil {
			takenAt = pgtype.Timestamptz{Time: dt, Valid: true}
		}
	}

	previewData, previewErr := generateWatermarkedPreview(img, decodeErr)
	if previewErr != nil {
		return nil, fmt.Errorf("generate watermarked preview: %w", previewErr)
	}
	if uploadErr := s.publicStore.Upload(ctx, row.PreviewKey, bytes.NewReader(previewData), "image/jpeg"); uploadErr != nil {
		return nil, fmt.Errorf("upload preview to public store: %w", uploadErr)
	}

	var pgWidth, pgHeight pgtype.Int4
	if width != nil {
		pgWidth = pgtype.Int4{Int32: *width, Valid: true}
	}
	if height != nil {
		pgHeight = pgtype.Int4{Int32: *height, Valid: true}
	}

	confirmed, err := s.queries.ConfirmPhoto(ctx, db.ConfirmPhotoParams{
		ID:        photoID,
		GalleryID: galleryID,
		Width:     pgWidth,
		Height:    pgHeight,
		SizeBytes: int64(len(data)),
		TakenAt:   takenAt,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrConflict
		}
		return nil, fmt.Errorf("confirm photo row: %w", err)
	}

	return &Photo{
		ID:          confirmed.ID,
		Filename:    confirmed.Filename,
		ContentType: confirmed.ContentType,
		PreviewURL:  s.publicStore.PublicURL(confirmed.PreviewKey),
		SizeBytes:   confirmed.SizeBytes,
		Width:       width,
		Height:      height,
	}, nil
}

func (s *PhotoService) verifyOwnership(ctx context.Context, eventID, photographerID string) error {
	ownerID, err := s.queries.GetEventOwner(ctx, eventID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("get event owner: %w", err)
	}
	if ownerID != photographerID {
		return ErrForbidden
	}
	return nil
}

func (s *PhotoService) verifyGallery(ctx context.Context, galleryID, eventID string) error {
	_, err := s.queries.GetGallery(ctx, db.GetGalleryParams{ID: galleryID, EventID: eventID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("get gallery: %w", err)
	}
	return nil
}

// safeIntToInt32 clamps an int to the int32 range.
func safeIntToInt32(v int) int32 {
	if v > math.MaxInt32 {
		return math.MaxInt32
	}
	if v < math.MinInt32 {
		return math.MinInt32
	}
	return int32(v) //nolint:gosec // clamped above
}

func generateWatermarkedPreview(img image.Image, decodeErr error) ([]byte, error) {
	if decodeErr != nil || img == nil {
		placeholder := imaging.New(100, 100, color.Gray{Y: 128})
		return encodeJPEG(placeholder)
	}

	preview := imaging.Resize(img, previewMaxWidth, 0, imaging.Lanczos)
	watermark := imaging.New(preview.Bounds().Dx(), preview.Bounds().Dy(), color.NRGBA{R: 255, G: 255, B: 255, A: 40})
	preview = imaging.Overlay(preview, watermark, image.Pt(0, 0), 1.0)

	return encodeJPEG(preview)
}

func encodeJPEG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := imaging.Encode(&buf, img, imaging.JPEG, imaging.JPEGQuality(previewJPEGQuality)); err != nil {
		return nil, fmt.Errorf("encode JPEG: %w", err)
	}
	return buf.Bytes(), nil
}
