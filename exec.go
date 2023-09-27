package retry

import (
	"context"
)

type RetryableFunc func(retryNumber int) (err error)

func Exec(
	retryableFunc RetryableFunc,
	strategies ...Strategy,
) (err error) {
	return ExecContext(context.Background(), func(ctx context.Context, retryNumber int) (err error) {
		return retryableFunc(retryNumber)
	}, strategies...)
}
