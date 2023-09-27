package retry

type unrecoverableError struct {
	error
}

func Unrecoverable(err error) error {
	return unrecoverableError{err}
}

func isUnrecoverable(err error) bool {
	_, ok := err.(unrecoverableError)
	return ok
}

func getUnrecoverableErrorCause(unrecoverableErr error) (cause error) {
	if unrecErr, ok := unrecoverableErr.(unrecoverableError); ok {
		return unrecErr.error
	}
	return nil
}
