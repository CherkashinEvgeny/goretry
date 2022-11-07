package retry

import (
	"context"
	"github.com/stretchr/testify/assert"
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
	attempt := strategy.Attempt(context.Background(), 0)
	assert.False(t, attempt)
}

func TestDelayedStrategy(t *testing.T) {
	delays := []time.Duration{time.Second, time.Second / 2, time.Second / 4}
	strategy := Delays(time.Second, time.Second/2, time.Second/4)
	for retryNumber, delay := range delays {
		start := time.Now()
		attempt := strategy.Attempt(context.Background(), retryNumber)
		stop := time.Now()
		assert.True(t, attempt)
		assert.True(t, stop.Sub(start) >= delay)
	}
	attempt := strategy.Attempt(context.Background(), len(delays))
	assert.False(t, attempt)
}

func TestFunctionStrategy(t *testing.T) {
	value := false
	strategy := Function(func(ctx context.Context, _ int) (attempt bool) {
		attempt = rand.Int()>>1 == 0
		value = attempt
		return
	})
	for retryNumber := 0; retryNumber < 1000; retryNumber++ {
		attempt := strategy.Attempt(context.Background(), retryNumber)
		assert.Equal(t, value, attempt)
	}
}

func TestInfiniteStrategy(t *testing.T) {
	strategy := Infinite()
	for retryNumber := 0; retryNumber < 1000; retryNumber++ {
		attempt := strategy.Attempt(context.Background(), retryNumber)
		assert.True(t, attempt)
	}
}

func TestMaxAttemptsStrategy(t *testing.T) {
	attempts := 10
	strategy := MaxAttempts(attempts)
	retryCount := 0
	retryNumber := 0
	for {
		retryCount++
		attempt := strategy.Attempt(context.Background(), retryNumber)
		if !attempt {
			break
		}
		retryNumber++
	}
	assert.Equal(t, attempts, retryCount)
}

func TestFixedDelayStrategy(t *testing.T) {
	delay := time.Second
	strategy := FixedDelay(delay)
	for retryNumber := 0; retryNumber < 5; retryNumber++ {
		start := time.Now()
		attempt := strategy.Attempt(context.Background(), retryNumber)
		stop := time.Now()
		assert.Equal(t, true, attempt)
		assert.True(t, stop.Sub(start) >= delay)
	}
}

func TestLinearDelayStrategy(t *testing.T) {
	delay := time.Second
	strategy := LinearDelay(delay, time.Second)
	for retryNumber := 0; retryNumber < 5; retryNumber++ {
		start := time.Now()
		attempt := strategy.Attempt(context.Background(), retryNumber)
		stop := time.Now()
		assert.True(t, attempt)
		assert.True(t, stop.Sub(start) >= delay)
		delay += time.Second
	}
}

func TestSleep(t *testing.T) {
	success := Sleep(context.Background(), 2*time.Second)
	assert.True(t, success)
}

func TestCancelSleep(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(time.Second)
		cancel()
	}()
	success := Sleep(ctx, 2*time.Second)
	assert.False(t, success)
}
