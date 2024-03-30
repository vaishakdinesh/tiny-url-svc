package types

import "errors"

var (
	InternalServerError = 99
	ValidationError     = 101
	NoRouteError        = 102

	ErrInternalServer = errors.New("unable to perform the request action")
	ErrNoPath         = errors.New("no route to the path. Check the URI")
)
