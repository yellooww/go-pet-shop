package storage

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrInvalidInput      = errors.New("invalid input")
	ErrInsufficientStock = errors.New("insufficient stock")
)
