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
	strategy := Compose(
		Function(func(ctx context.Context, _ int) (attempt bool) {
			return true
		}),
		Function(func(ctx context.Context, retryNumber int) (attempt bool) {
			return retryNumber < 5
		}),
	)
	var retryNumber int
	for retryNumber < 5 {
		attempt := strategy.Attempt(context.Background(), 0)
		assert.True(t, attempt)
		retryNumber++
	}
	attempt := strategy.Attempt(context.Background(), retryNumber)
	assert.False(t, attempt)
}

func BenchmarkCompositeStrategyAllocations(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = Exec(func(retryNumber int) (err error) {
				if retryNumber > 5 {
					return nil
				}
				return testError
			}, Compose(
				Function(func(ctx context.Context, _ int) (attempt bool) {
					return true
				}),
				Function(func(ctx context.Context, _ int) (attempt bool) {
					return false
				}),
			))
		}
	})
}

func TestSequentialStrategy(t *testing.T) {
	strategy := Sequence(
		Function(func(ctx context.Context, retryNumber int) (attempt bool) {
			return retryNumber < 5
		}),
		Function(func(ctx context.Context, retryNumber int) (attempt bool) {
			return retryNumber < 10
		}),
	)
	var retryNumber int
	for retryNumber < 10 {
		attempt := strategy.Attempt(context.Background(), 0)
		assert.True(t, attempt)
		retryNumber++
	}
	attempt := strategy.Attempt(context.Background(), retryNumber)
	assert.False(t, attempt)
}

func BenchmarkSequentialStrategyAllocations(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = Exec(func(retryNumber int) (err error) {
				if retryNumber > 5 {
					return nil
				}
				return testError
			}, Sequence(
				Function(func(ctx context.Context, _ int) (attempt bool) {
					return true
				}),
				Function(func(ctx context.Context, _ int) (attempt bool) {
					return false
				}),
			))
		}
	})
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

func BenchmarkDelayedStrategyAllocations(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = Exec(func(retryNumber int) (err error) {
				if retryNumber > 5 {
					return nil
				}
				return testError
			}, Delays(0, 0, 0, 0, 0, 0, 0, 0, 0, 0))
		}
	})
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

func BenchmarkFunctionStrategyAllocations(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = Exec(func(retryNumber int) (err error) {
				if retryNumber > 5 {
					return nil
				}
				return testError
			}, Function(func(ctx context.Context, retryNumber int) (attempt bool) {
				return true
			}))
		}
	})
}

func TestInfiniteStrategy(t *testing.T) {
	strategy := Infinite()
	for retryNumber := 0; retryNumber < 1000; retryNumber++ {
		attempt := strategy.Attempt(context.Background(), retryNumber)
		assert.True(t, attempt)
	}
}

func BenchmarkInfiniteStrategyAllocations(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = Exec(func(retryNumber int) (err error) {
				if retryNumber > 5 {
					return nil
				}
				return testError
			}, Infinite())
		}
	})
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

func BenchmarkMaxAttemptsStrategyAllocations(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = Exec(func(retryNumber int) (err error) {
				if retryNumber > 5 {
					return nil
				}
				return testError
			}, MaxAttempts(10))
		}
	})
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

func BenchmarkFixedDelayStrategyAllocations(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = Exec(func(retryNumber int) (err error) {
				if retryNumber > 5 {
					return nil
				}
				return testError
			}, FixedDelay(0))
		}
	})
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

func BenchmarkLinearDelayStrategyAllocations(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = Exec(func(retryNumber int) (err error) {
				if retryNumber > 5 {
					return nil
				}
				return testError
			}, LinearDelay(0, 0))
		}
	})
}

func TestPowDelayStrategy(t *testing.T) {
	delay := 100 * time.Millisecond
	strategy := PowDelay(delay, math.Sqrt2)
	for retryNumber := 0; retryNumber < 10; retryNumber++ {
		start := time.Now()
		attempt := strategy.Attempt(context.Background(), retryNumber)
		stop := time.Now()
		assert.True(t, attempt)
		assert.True(t, stop.Sub(start) >= delay)
		delay = time.Duration(float64(delay) * math.Sqrt2)
	}
}

func BenchmarkPowDelayStrategyAllocations(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = Exec(func(retryNumber int) (err error) {
				if retryNumber > 5 {
					return nil
				}
				return testError
			}, PowDelay(0, 0))
		}
	})
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
