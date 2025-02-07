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

func (e ErrNotFound) Is(err error) bool {
	return true
}
