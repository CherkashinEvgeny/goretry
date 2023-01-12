package retry

import "context"

type RetryableContextFunc func(ctx context.Context, retryNumber int) (err error)

func ExecContext(
	ctx context.Context,
	retryableFunc RetryableContextFunc,
	strategies ...Strategy,
) (err error) {
	strategy := Compose(strategies...)
	retryNumber := 0
	for {
		err = retryableFunc(ctx, retryNumber)
		if err == nil {
			return
		}
		if isUnrecoverable(err) {
			err = getUnrecoverableErrorCause(err)
			return
		}
		if isContextCancelled(ctx) {
			return
		}
		if !strategy.Attempt(ctx, retryNumber) {
			return
		}
		retryNumber++
	}
}

func isContextCancelled(ctx context.Context) (cancelled bool) {
	select {
	case <-ctx.Done():
		cancelled = true
		break
	default:
		break
	}
	return
}
