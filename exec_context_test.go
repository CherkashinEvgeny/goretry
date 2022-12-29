package retry

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSuccessExecContext(t *testing.T) {
	counter := 0
	err := ExecContext(context.Background(), func(ctx context.Context, retryNumber int) (err error) {
		counter++
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, counter)
}

func TestSecondRetrySuccessExecContext(t *testing.T) {
	counter := 0
	err := ExecContext(context.Background(), func(ctx context.Context, retryNumber int) (err error) {
		counter++
		if retryNumber == 0 {
			return testError
		}
		return nil

	})
	assert.NoError(t, err)
	assert.Equal(t, 2, counter)
}

func TestUnrecoverableFailureExecContext(t *testing.T) {
	counter := 0
	err := ExecContext(context.Background(), func(ctx context.Context, retryNumber int) (err error) {
		counter++
		return Unrecoverable(testError)
	})
	assert.Equal(t, err, testError)
	assert.Equal(t, 1, counter)
}

func TestCancelExecContext(t *testing.T) {
	counter := 0
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(time.Second)
		cancel()
	}()
	err := ExecContext(ctx, func(ctx context.Context, retryNumber int) (err error) {
		time.Sleep(2 * time.Second)
		counter++
		return testError
	})
	assert.Equal(t, testError, err)
	assert.Equal(t, 1, counter)
}
