package retry

import (
	"context"
	"github.com/stretchr/testify/assert"
	"math"
	"math/rand"
	"testing"
	"time"
)

func TestCompositeStrategy(t *testing.T) {
	retryNumber := 0
	strategy := Compose(
		Function(func(ctx context.Context) (attempt bool) {
			return true
		}),
		Function(func(ctx context.Context) (attempt bool) {
			return retryNumber < 5
		}),
	)
	for retryNumber < 5 {
		attempt := strategy.Attempt(context.Background())
		assert.True(t, attempt)
		retryNumber++
	}
	attempt := strategy.Attempt(context.Background())
	assert.False(t, attempt)
}

func TestDelayedStrategy(t *testing.T) {
	delays := []time.Duration{time.Second, time.Second / 2, time.Second / 4}
	strategy := Delays(time.Second, time.Second/2, time.Second/4)
	for _, delay := range delays {
		start := time.Now()
		attempt := strategy.Attempt(context.Background())
		stop := time.Now()
		assert.True(t, attempt)
		assert.InDelta(t, delay, stop.Sub(start), float64(50*time.Millisecond))
	}
	attempt := strategy.Attempt(context.Background())
	assert.False(t, attempt)
}

func TestFunctionStrategy(t *testing.T) {
	value := false
	strategy := Function(func(ctx context.Context) (attempt bool) {
		attempt = rand.Int()>>1 == 0
		value = attempt
		return
	})
	for retryNumber := 0; retryNumber < 1000; retryNumber++ {
		attempt := strategy.Attempt(context.Background())
		assert.Equal(t, value, attempt)
	}
}

func TestInfiniteStrategy(t *testing.T) {
	strategy := Infinite()
	for retryNumber := 0; retryNumber < 1000; retryNumber++ {
		attempt := strategy.Attempt(context.Background())
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
		attempt := strategy.Attempt(context.Background())
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
		attempt := strategy.Attempt(context.Background())
		stop := time.Now()
		assert.Equal(t, true, attempt)
		assert.InDelta(t, delay, stop.Sub(start), float64(50*time.Millisecond))
	}
}

func TestRandomDelayStrategy(t *testing.T) {
	minDelay := time.Second
	maxDelay := minDelay + 500*time.Millisecond
	strategy := RandomDelay(minDelay, maxDelay)
	for retryNumber := 0; retryNumber < 5; retryNumber++ {
		start := time.Now()
		attempt := strategy.Attempt(context.Background())
		stop := time.Now()
		assert.Equal(t, true, attempt)
		assert.GreaterOrEqual(t, stop.Sub(start), minDelay)
		assert.LessOrEqual(t, stop.Sub(start), maxDelay+50*time.Millisecond)
	}
}

func TestLinearDelayStrategy(t *testing.T) {
	delay := time.Second
	strategy := LinearDelay(delay, time.Second)
	for retryNumber := 0; retryNumber < 5; retryNumber++ {
		start := time.Now()
		attempt := strategy.Attempt(context.Background())
		stop := time.Now()
		assert.True(t, attempt)
		assert.InDelta(t, delay, stop.Sub(start), float64(50*time.Millisecond))
		delay += time.Second
	}
}

func TestPowDelayStrategy(t *testing.T) {
	delay := 100 * time.Millisecond
	strategy := PowDelay(delay, math.Sqrt2)
	for retryNumber := 0; retryNumber < 10; retryNumber++ {
		start := time.Now()
		attempt := strategy.Attempt(context.Background())
		stop := time.Now()
		assert.True(t, attempt)
		assert.InDelta(t, delay, stop.Sub(start), float64(50*time.Millisecond))
		delay = time.Duration(float64(delay) * math.Sqrt2)
	}
}

func TestSleep(t *testing.T) {
	ok := Sleep(context.Background(), 2*time.Second)
	assert.True(t, ok)
}

func TestCancelSleep(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(time.Second)
		cancel()
	}()
	ok := Sleep(ctx, 2*time.Second)
	assert.False(t, ok)
}
