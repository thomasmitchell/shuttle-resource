package driver

type ErrNotFound struct {
	message string
}

func (e *ErrNotFound) Error() string {
	return e.message
}

func IsNotFound(err error) bool {
	_, is := err.(*ErrNotFound)
	return is
}
