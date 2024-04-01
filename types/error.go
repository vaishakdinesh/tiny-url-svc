package types

import "errors"

var (
	InternalServerError = 99
	InputError          = 100
	ValidationError     = 101
	NoRouteError        = 102
	NotFoundError       = 103

	ErrNoPath           = errors.New("no route to the path. Check the URI")
	ErrDocumentNotFound = errors.New("no entry found for the key")
	ErrCacheNotFound    = errors.New("no cache found for the key")
	ErrInvalidScheme    = errors.New("unsupported scheme")
	ErrInvalidInput     = errors.New("invalid input")
)
