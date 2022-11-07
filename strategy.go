package retry

import (
	"context"
	"math"
	"math/rand"
	"time"
)

type Strategy interface {
	Attempt(ctx context.Context, retryNumber int) (attempt bool)
}

func DefaultStrategy() Strategy {
	return Compose(MaxAttempts(10), PowDelay(100*time.Millisecond, math.Sqrt2))
}

func Compose(strategies ...Strategy) CompositeStrategy {
	return CompositeStrategy{strategies}
}

type CompositeStrategy struct {
	Strategies []Strategy
}

func (s CompositeStrategy) Attempt(ctx context.Context, retryNumber int) (attempt bool) {
	attempt = true
	for index := 0; attempt && index < len(s.Strategies); index++ {
		attempt = s.Strategies[index].Attempt(ctx, retryNumber)
	}
	return
}

func Delays(delays ...time.Duration) DelayedStrategy {
	return DelayedStrategy{delays}
}

type DelayedStrategy struct {
	Delays []time.Duration
}

func (s DelayedStrategy) Attempt(ctx context.Context, retryNumber int) (attempt bool) {
	attempt = retryNumber < len(s.Delays)
	if !attempt {
		return
	}
	delay := s.Delays[retryNumber]
	Sleep(ctx, delay)
	return
}

type StrategyFunc func(ctx context.Context, retryNumber int) (attempt bool)

func Function(retryFunc StrategyFunc) FuncStrategy {
	return FuncStrategy{retryFunc}
}

type FuncStrategy struct {
	Func StrategyFunc
}

func (s FuncStrategy) Attempt(ctx context.Context, retryNumber int) (attempt bool) {
	return s.Func(ctx, retryNumber)
}

func Infinite() *InfiniteAttemptsStrategy {
	return infiniteAttemptStrategyPtr
}

var infiniteAttemptStrategyPtr = &InfiniteAttemptsStrategy{}

type InfiniteAttemptsStrategy struct {
}

func (s *InfiniteAttemptsStrategy) Attempt(_ context.Context, _ int) (attempt bool) {
	return true
}

func MaxAttempts(attempts int) MaxRetriesStrategy {
	return MaxRetriesStrategy{attempts}
}

type MaxRetriesStrategy struct {
	MaxAttempts int
}

func (s MaxRetriesStrategy) Attempt(_ context.Context, retryNumber int) (attempt bool) {
	// -1 because attempts parameter is max function call count, and MaxAttempts parameter is max function rerun count
	return retryNumber < s.MaxAttempts-1
}

func FixedDelay(delay time.Duration) FixedDelayStrategy {
	return FixedDelayStrategy{delay}
}

type FixedDelayStrategy struct {
	Delay time.Duration
}

func (s FixedDelayStrategy) Attempt(ctx context.Context, _ int) (attempt bool) {
	return Sleep(ctx, s.Delay)
}

func RandomDelay(min time.Duration, max time.Duration) RandomDelayStrategy {
	return RandomDelayStrategy{
		MinDelay: min,
		MaxDelay: max,
	}
}

type RandomDelayStrategy struct {
	MinDelay time.Duration
	MaxDelay time.Duration
}

func (s RandomDelayStrategy) Attempt(ctx context.Context, _ int) (attempt bool) {
	delay := s.MinDelay + time.Duration(rand.Int63()%int64(s.MaxDelay-s.MinDelay))
	return Sleep(ctx, delay)
}

func LinearDelay(seed time.Duration, base time.Duration) LinearDelayStrategy {
	return LinearDelayStrategy{Seed: seed, Delta: base}
}

type LinearDelayStrategy struct {
	Seed  time.Duration
	Delta time.Duration
}

func (s LinearDelayStrategy) Attempt(ctx context.Context, retryNumber int) (attempt bool) {
	delay := s.Seed + s.Delta*time.Duration(retryNumber)
	return Sleep(ctx, delay)
}

func ExpDelay(seed time.Duration) PowDelayStrategy {
	return PowDelay(seed, math.E)
}

func PowDelay(seed time.Duration, base float64) PowDelayStrategy {
	return PowDelayStrategy{Seed: seed, Base: base}
}

type PowDelayStrategy struct {
	Seed time.Duration
	Base float64
}

func (e PowDelayStrategy) Attempt(ctx context.Context, retryNumber int) (attempt bool) {
	delay := e.Seed * time.Duration(math.Pow(e.Base, float64(retryNumber)))
	return Sleep(ctx, delay)
}

func Sleep(ctx context.Context, delay time.Duration) bool {
	select {
	case <-time.After(delay):
		return true
	case <-ctx.Done():
		return false
	}
}
