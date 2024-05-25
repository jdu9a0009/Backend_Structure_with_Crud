package postgres

import (
	"errors"
)

var (
	// ErrNotFound is used when a specific Product is requested but does not exist.
	ErrNotFound = errors.New("not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")

	// ErrForbidden occurs when a area tries to do something that is forbidden to them according to our access control policies.
	ErrForbidden = errors.New("attempted action is not allowed")
)
