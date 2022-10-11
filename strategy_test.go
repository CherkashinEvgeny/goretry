package retry

import (
	"context"
	"gotest.tools/assert"
	"math/rand"
	"testing"
	"time"
)

func TestCompositeStrategy(t *testing.T) {
	strategy := Compose(
		Function(func(ctx context.Context, _ int) (attempt bool) {
			return true
		}),
		Function(func(ctx context.Context, _ int) (attempt bool) {
			return false
		}),
	)
	assert.Assert(t, !strategy.Attempt(context.Background(), 0))
}

func TestDelayedStrategy(t *testing.T) {
	delays := []time.Duration{time.Second, time.Second / 2, time.Second / 4}
	strategy := Delays(time.Second, time.Second/2, time.Second/4)
	for retryNumber, delay := range delays {
		start := time.Now()
		assert.Assert(t, strategy.Attempt(context.Background(), retryNumber))
		stop := time.Now()
		assert.Assert(t, stop.Sub(start) >= delay)
	}
	assert.Assert(t, !strategy.Attempt(context.Background(), len(delays)))
}

func TestFunctionStrategy(t *testing.T) {
	value := false
	strategy := Function(func(ctx context.Context, _ int) (attempt bool) {
		attempt = rand.Int()>>1 == 0
		value = attempt
		return
	})
	for retryNumber := 0; retryNumber < 1000; retryNumber++ {
		assert.Equal(t, value, strategy.Attempt(context.Background(), retryNumber))
	}
}

func TestInfiniteStrategy(t *testing.T) {
	strategy := Infinite()
	for retryNumber := 0; retryNumber < 1000; retryNumber++ {
		assert.Assert(t, strategy.Attempt(context.Background(), retryNumber))
	}
}

func TestMaxAttemptsStrategy(t *testing.T) {
	attempts := 10
	strategy := MaxAttempts(attempts)
	retryCount := 0
	retryNumber := 0
	for {
		retryCount++
		if !strategy.Attempt(context.Background(), retryNumber) {
			break
		}
		retryNumber++
	}
	assert.Equal(t, attempts, retryCount)
}

func TestFixedDelayStrategy(t *testing.T) {
	delay := time.Second
	strategy := FixedDelay(delay)
	for retryNumber := 0; retryNumber < 3; retryNumber++ {
		start := time.Now()
		strategy.Attempt(context.Background(), retryNumber)
		stop := time.Now()
		assert.Assert(t, stop.Sub(start) >= delay)
	}
}

func TestSleep(t *testing.T) {
	assert.Assert(t, Sleep(context.Background(), 2*time.Second))
}

func TestCancelSleep(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(time.Second)
		cancel()
	}()
	assert.Assert(t, !Sleep(ctx, 2*time.Second))
}
