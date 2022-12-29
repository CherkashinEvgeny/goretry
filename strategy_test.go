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
		Function(func(ctx context.Context) (attempt bool) {
			return true
		}),
		Function(func(ctx context.Context) (attempt bool) {
			return false
		}),
	)
	assert.False(t, strategy.Attempt(context.Background()))
}

func TestDelayedStrategy(t *testing.T) {
	delays := []time.Duration{time.Second, time.Second / 2, time.Second / 4}
	strategy := Delays(time.Second, time.Second/2, time.Second/4)
	for _, delay := range delays {
		start := time.Now()
		assert.True(t, strategy.Attempt(context.Background()))
		stop := time.Now()
		assert.True(t, stop.Sub(start) >= delay)
	}
	assert.False(t, strategy.Attempt(context.Background()))
}

func TestFunctionStrategy(t *testing.T) {
	value := false
	strategy := Function(func(ctx context.Context) (attempt bool) {
		attempt = rand.Int()>>1 == 0
		value = attempt
		return
	})
	for i := 0; i < 1000; i++ {
		assert.Equal(t, value, strategy.Attempt(context.Background()))
	}
}

func TestInfiniteStrategy(t *testing.T) {
	strategy := Infinite()
	for i := 0; i < 1000; i++ {
		assert.True(t, strategy.Attempt(context.Background()))
	}
}

func TestMaxAttemptsStrategy(t *testing.T) {
	attempts := 10
	strategy := MaxAttempts(attempts)
	counter := 0
	for strategy.Attempt(context.Background()) {
		counter++
	}
	assert.Equal(t, attempts-1, counter)
}

func TestFixedDelayStrategy(t *testing.T) {
	delay := time.Second
	strategy := FixedDelay(delay)
	for i := 0; i < 3; i++ {
		start := time.Now()
		strategy.Attempt(context.Background())
		stop := time.Now()
		assert.True(t, stop.Sub(start) >= delay)
	}
}

func TestSleep(t *testing.T) {
	assert.True(t, Sleep(context.Background(), 2*time.Second))
}

func TestCancelSleep(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(time.Second)
		cancel()
	}()
	assert.False(t, Sleep(ctx, 2*time.Second))
}
