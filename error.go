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

func getUnrecoverableErrorCause(unrecoverableErr error) (cause error, ok bool) {
	var unrecErr unrecoverableError
	unrecErr, ok = unrecoverableErr.(unrecoverableError)
	cause = unrecErr.error
	return
}
