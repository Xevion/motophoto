package service

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrConflict   = errors.New("resource already exists")
	ErrForbidden  = errors.New("forbidden")
	ErrValidation = errors.New("validation error")
)

// NewValidationError wraps ErrValidation with a descriptive message.
func NewValidationError(msg string) error {
	return fmt.Errorf("%s: %w", msg, ErrValidation)
}
