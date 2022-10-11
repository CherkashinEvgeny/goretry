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

func Compose(strategies ...Strategy) Strategy {
	return compositeStrategy{strategies: strategies}
}

type compositeStrategy struct {
	strategies []Strategy
}

func (s compositeStrategy) Attempt(ctx context.Context, retryNumber int) (attempt bool) {
	attempt = true
	for index := 0; attempt && index < len(s.strategies); index++ {
		attempt = s.strategies[index].Attempt(ctx, retryNumber)
	}
	return
}

func Delays(delays ...time.Duration) Strategy {
	return delayedStrategy{delays}
}

type delayedStrategy struct {
	delays []time.Duration
}

func (s delayedStrategy) Attempt(ctx context.Context, retryNumber int) (attempt bool) {
	attempt = retryNumber < len(s.delays)
	if !attempt {
		return
	}
	delay := s.delays[retryNumber]
	Sleep(ctx, delay)
	return
}

type StrategyFunc func(ctx context.Context, retryNumber int) (attempt bool)

func Function(retryFunc StrategyFunc) Strategy {
	return funcStrategy{retryFunc}
}

type funcStrategy struct {
	retryFunc StrategyFunc
}

func (s funcStrategy) Attempt(ctx context.Context, retryNumber int) (attempt bool) {
	return s.retryFunc(ctx, retryNumber)
}

func Infinite() Strategy {
	return infiniteAttemptStrategyPtr
}

var infiniteAttemptStrategyPtr = &infiniteAttemptsStrategy{}

type infiniteAttemptsStrategy struct {
	remainingAttempts int
}

func (s *infiniteAttemptsStrategy) Attempt(_ context.Context, _ int) (attempt bool) {
	return true
}

func MaxAttempts(attempts int) Strategy {
	return maxAttemptsStrategy{
		// -1 because attempts parameter is max function call count, and maxRetryCount parameter is max function rerun count
		maxRetryCount: attempts - 1,
	}
}

type maxAttemptsStrategy struct {
	maxRetryCount int
}

func (s maxAttemptsStrategy) Attempt(_ context.Context, retryNumber int) (attempt bool) {
	attempt = retryNumber < s.maxRetryCount
	return
}

func FixedDelay(delay time.Duration) Strategy {
	return fixedDelayStrategy{delay: delay}
}

type fixedDelayStrategy struct {
	delay time.Duration
}

func (s fixedDelayStrategy) Attempt(ctx context.Context, _ int) (attempt bool) {
	attempt = true
	return Sleep(ctx, s.delay)
}

func RandomDelay(min time.Duration, max time.Duration) Strategy {
	return randomDelayStrategy{
		min: min,
		max: max,
	}
}

type randomDelayStrategy struct {
	min time.Duration
	max time.Duration
}

func (s randomDelayStrategy) Attempt(ctx context.Context, _ int) (attempt bool) {
	delay := s.min + time.Duration(rand.Int63()%int64(s.max-s.min))
	return Sleep(ctx, delay)
}

func ExpDelay(seed time.Duration) Strategy {
	return PowDelay(seed, math.E)
}

func PowDelay(seed time.Duration, base float64) Strategy {
	return powStrategy{base: base, seed: seed}
}

type powStrategy struct {
	base float64
	seed time.Duration
}

func (e powStrategy) Attempt(ctx context.Context, retryNumber int) (attempt bool) {
	return Sleep(ctx, e.seed*time.Duration(math.Pow(e.base, float64(retryNumber))))
}

func Sleep(ctx context.Context, delay time.Duration) bool {
	select {
	case <-time.After(delay):
		return true
	case <-ctx.Done():
		return false
	}
}
