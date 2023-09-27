package retry

import "context"

type RetryableContextFunc func(ctx context.Context, retryNumber int) (err error)

func ExecContext(
	ctx context.Context,
	retryableFunc RetryableContextFunc,
	strategies ...Strategy,
) (err error) {
	strategy := And(strategies...)
	retryNumber := 0
	for {
		err = retryableFunc(ctx, retryNumber)
		if err == nil {
			return nil
		}
		if isUnrecoverable(err) {
			return getUnrecoverableErrorCause(err)
		}
		if isContextCancelled(ctx) {
			return err
		}
		if !strategy.Attempt(ctx, retryNumber, err) {
			return err
		}
		retryNumber++
	}
}

func isContextCancelled(ctx context.Context) (cancelled bool) {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
