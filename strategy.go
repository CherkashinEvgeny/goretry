package retry

import (
	"context"
	"math"
	"math/rand"
	"time"
)

type Strategy interface {
	Attempt(ctx context.Context) (attempt bool)
}

func DefaultStrategy() Strategy {
	return Compose(MaxAttempts(10), PowDelay(100*time.Millisecond, math.Sqrt2))
}

func Compose(strategies ...Strategy) CompositeStrategy {
	return CompositeStrategy{strategies}
}

type CompositeStrategy struct {
	strategies []Strategy
}

func (s CompositeStrategy) Attempt(ctx context.Context) (attempt bool) {
	attempt = true
	for index := 0; attempt && index < len(s.strategies); index++ {
		attempt = s.strategies[index].Attempt(ctx)
	}
	return
}

func Delays(delays ...time.Duration) *DelayedStrategy {
	return &DelayedStrategy{
		index:  0,
		delays: delays,
	}
}

type DelayedStrategy struct {
	index  int
	delays []time.Duration
}

func (s *DelayedStrategy) Attempt(ctx context.Context) (attempt bool) {
	attempt = s.index < len(s.delays)
	if !attempt {
		return
	}
	delay := s.delays[s.index]
	s.index += 1
	Sleep(ctx, delay)
	return
}

type StrategyFunc func(ctx context.Context) (attempt bool)

func Function(retryFunc StrategyFunc) FuncStrategy {
	return FuncStrategy{retryFunc}
}

type FuncStrategy struct {
	retryFunc StrategyFunc
}

func (s FuncStrategy) Attempt(ctx context.Context) (attempt bool) {
	return s.retryFunc(ctx)
}

func Infinite() *InfiniteAttemptsStrategy {
	return infiniteAttemptStrategyPtr
}

var infiniteAttemptStrategyPtr = &InfiniteAttemptsStrategy{}

type InfiniteAttemptsStrategy struct {
}

func (s *InfiniteAttemptsStrategy) Attempt(_ context.Context) (attempt bool) {
	return true
}

func MaxAttempts(attempts int) *MaxRetriesStrategy {
	return &MaxRetriesStrategy{
		// -1 because attempts parameter is max function call count, and remainingAttempts parameter is max function rerun count
		remainingAttempts: attempts - 1,
	}
}

type MaxRetriesStrategy struct {
	remainingAttempts int
}

func (s *MaxRetriesStrategy) Attempt(_ context.Context) (attempt bool) {
	attempt = s.remainingAttempts > 0
	if attempt {
		s.remainingAttempts--
	}
	return
}

func FixedDelay(delay time.Duration) FixedDelayStrategy {
	return FixedDelayStrategy{delay}
}

type FixedDelayStrategy struct {
	delay time.Duration
}

func (s FixedDelayStrategy) Attempt(ctx context.Context) (attempt bool) {
	return Sleep(ctx, s.delay)
}

func RandomDelay(min time.Duration, max time.Duration) RandomDelayStrategy {
	return RandomDelayStrategy{
		min: min,
		max: max,
	}
}

type RandomDelayStrategy struct {
	min time.Duration
	max time.Duration
}

func (s RandomDelayStrategy) Attempt(ctx context.Context) (attempt bool) {
	delay := s.min + time.Duration(rand.Int63()%int64(s.max-s.min))
	return Sleep(ctx, delay)
}

func LinearDelay(seed time.Duration, delta time.Duration) *LinearDelayStrategy {
	return &LinearDelayStrategy{
		delay: seed,
		delta: delta,
	}
}

type LinearDelayStrategy struct {
	delay time.Duration
	delta time.Duration
}

func (s *LinearDelayStrategy) Attempt(ctx context.Context) (attempt bool) {
	s.delay = s.delay + s.delta
	return Sleep(ctx, s.delay)
}

func ExpDelay(seed time.Duration) *PowDelayStrategy {
	return PowDelay(seed, math.E)
}

func PowDelay(seed time.Duration, base float64) *PowDelayStrategy {
	return &PowDelayStrategy{delay: seed, base: base}
}

type PowDelayStrategy struct {
	delay time.Duration
	base  float64
}

func (e *PowDelayStrategy) Attempt(ctx context.Context) (attempt bool) {
	e.delay = time.Duration(float64(e.delay) * e.base)
	return Sleep(ctx, e.delay)
}

func Sleep(ctx context.Context, delay time.Duration) bool {
	select {
	case <-time.After(delay):
		return true
	case <-ctx.Done():
		return false
	}
}
