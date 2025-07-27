package errs

import "errors"

var (
	ErrDBOperation = errors.New("database operation failed")
	ErrNotFound    = errors.New("record not found")
	ErrUniqueName  = errors.New("duplicate name")
)
