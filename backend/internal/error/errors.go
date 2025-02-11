package errs

type ErrNotFound struct {
	resource string
}

func NewNotFound(resource string) ErrNotFound {
	return ErrNotFound{resource: resource}
}

func (e ErrNotFound) Error() string {
	return e.resource + " not found"
}

func (e ErrNotFound) Is(target error) bool {
	if _, ok := target.(ErrNotFound); ok {
		return true
	}

	return false
}
