package medication

import (
	"errors"
)

var (
	ErrAlreadyExists = errors.New("conflict")
	ErrNotFound      = errors.New("not found")
	ErrBadInput      = errors.New("bad request")
)
