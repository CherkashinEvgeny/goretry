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
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		default:
			break
		}
		if err == nil {
			return
		}
		if isUnrecoverable(err) {
			err, _ = getUnrecoverableErrorCause(err)
			return
		}
		if !strategy.Attempt(ctx) {
			return
		}
		retryNumber += 1
	}
}
