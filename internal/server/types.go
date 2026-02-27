package server

// API response types -- tygo generates TypeScript bindings from this file.

type ListResponse[T any] struct {
	NextCursor *string `json:"next_cursor"`
	Data       []T     `json:"data"`
}

type ItemResponse[T any] struct {
	Data T `json:"data"`
}

type EventResponse struct {
	Location    *string           `json:"location"`
	Description *string           `json:"description"`
	Date        *string           `json:"date"`
	ID          string            `json:"id"`
	Slug        string            `json:"slug"`
	Name        string            `json:"name"`
	Sport       string            `json:"sport"`
	Status      string            `json:"status"`
	Galleries   []GalleryResponse `json:"galleries,omitempty"`
	Tags        []string          `json:"tags"`
	PhotoCount  int64             `json:"photo_count"`
}

type GalleryResponse struct {
	Description *string `json:"description"`
	ID          string  `json:"id"`
	Slug        string  `json:"slug"`
	Name        string  `json:"name"`
	PhotoCount  int64   `json:"photo_count"`
	SortOrder   int32   `json:"sort_order"`
}

type CreateEventRequest struct {
	Location    *string  `json:"location"`
	Description *string  `json:"description"`
	Date        *string  `json:"date"`
	Status      *string  `json:"status"`
	Name        string   `json:"name"        validate:"required"`
	Slug        string   `json:"slug"        validate:"required"`
	Sport       string   `json:"sport"       validate:"required"`
	Tags        []string `json:"tags"`
}

type UpdateEventRequest struct {
	Tags        *[]string `json:"tags"`
	Name        *string   `json:"name"`
	Slug        *string   `json:"slug"`
	Sport       *string   `json:"sport"`
	Location    *string   `json:"location"`
	Description *string   `json:"description"`
	Date        *string   `json:"date"`
	Status      *string   `json:"status"`
	SortOrder   *int32    `json:"sort_order"`
}

type CreateGalleryRequest struct {
	Description *string `json:"description"`
	SortOrder   *int32  `json:"sort_order"`
	Name        string  `json:"name"        validate:"required"`
	Slug        string  `json:"slug"        validate:"required"`
}

type UpdateGalleryRequest struct {
	Name        *string `json:"name"`
	Slug        *string `json:"slug"`
	Description *string `json:"description"`
	SortOrder   *int32  `json:"sort_order"`
}

type LoginRequest struct {
	Email string `json:"email"    validate:"required,email"`
	//nolint:gosec // G117: intentional request body field
	Password string `json:"password" validate:"required"`
}

type RegisterRequest struct {
	Email string `json:"email"        validate:"required,email"`
	//nolint:gosec // G117: intentional request body field
	Password    string `json:"password"     validate:"required,min=8"`
	DisplayName string `json:"display_name" validate:"required"`
	Role        string `json:"role"         validate:"required,oneof=photographer customer"`
}

type UserResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}
