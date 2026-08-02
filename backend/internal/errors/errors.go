package errors

import "errors"

var (
	ErrNotFound   = errors.New("not_found")
	ErrValidation = errors.New("validation_failed")
	ErrConflict   = errors.New("conflict")
)
