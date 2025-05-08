package errs

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

type ErrNotFound struct {
	resource string
}

func (e ErrNotFound) Error() string {
	return e.resource + " not found"
}

func (e ErrNotFound) Is(target error) bool {
	_, ok := target.(ErrNotFound)
	return ok
}

func NewNotFound(resource string) ErrNotFound {
	return ErrNotFound{resource: resource}
}

type ErrValidation struct {
	err string
}

func (e ErrValidation) Error() string {
	return e.err
}

func (e ErrValidation) Is(target error) bool {
	_, ok := target.(ErrValidation)
	return ok
}

func NewValidationError(err string) ErrValidation {
	return ErrValidation{err: err}
}

func NewValidationErrorF(err string, a ...any) ErrValidation {
	return ErrValidation{err: fmt.Sprintf(err, a...)}
}
