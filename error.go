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
	unrecErr, ok := unrecoverableErr.(unrecoverableError)
	if ok {
		cause = unrecErr.error
	}
	return
}
