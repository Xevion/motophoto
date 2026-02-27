package service

import "errors"

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("resource already exists")
)
