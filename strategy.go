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

func Compose(strategies ...Strategy) (strategy Strategy) {
	return &compositeStrategy{strategies}
}

type compositeStrategy struct {
	strategies []Strategy
}

func (s *compositeStrategy) Attempt(ctx context.Context) (attempt bool) {
	attempt = true
	for index := 0; attempt && index < len(s.strategies); index++ {
		attempt = s.strategies[index].Attempt(ctx)
	}
	return
}

func Delays(delays ...time.Duration) (strategy Strategy) {
	return &delayedStrategy{0, delays}
}

type delayedStrategy struct {
	index  int
	delays []time.Duration
}

func (s *delayedStrategy) Attempt(ctx context.Context) (attempt bool) {
	attempt = s.index < len(s.delays)
	if !attempt {
		return
	}
	delay := s.delays[s.index]
	s.index++
	attempt = Sleep(ctx, delay)
	return
}

type StrategyFunc func(ctx context.Context) (attempt bool)

func Function(retryFunc StrategyFunc) (strategy Strategy) {
	return &funcStrategy{retryFunc}
}

type funcStrategy struct {
	retryFunc StrategyFunc
}

func (s *funcStrategy) Attempt(ctx context.Context) (attempt bool) {
	attempt = s.retryFunc(ctx)
	return
}

func Infinite() (strategy Strategy) {
	return infiniteAttemptStrategyPtr
}

var infiniteAttemptStrategyPtr = &infiniteAttemptsStrategy{}

type infiniteAttemptsStrategy struct {
}

func (s *infiniteAttemptsStrategy) Attempt(_ context.Context) (attempt bool) {
	attempt = true
	return
}

func MaxAttempts(attempts int) (strategy Strategy) {
	// -1 because attempts parameter is max function call count, and remainingAttempts parameter is max function rerun count
	return &maxRetriesStrategy{attempts - 1}
}

type maxRetriesStrategy struct {
	remainingAttempts int
}

func (s *maxRetriesStrategy) Attempt(_ context.Context) (attempt bool) {
	attempt = s.remainingAttempts > 0
	if attempt {
		s.remainingAttempts--
	}
	return
}

func FixedDelay(delay time.Duration) (strategy Strategy) {
	return &fixedDelayStrategy{delay}
}

type fixedDelayStrategy struct {
	delay time.Duration
}

func (s *fixedDelayStrategy) Attempt(ctx context.Context) (attempt bool) {
	attempt = Sleep(ctx, s.delay)
	return
}

func RandomDelay(minDelay time.Duration, maxDelay time.Duration) (strategy Strategy) {
	return &randomDelayStrategy{minDelay, maxDelay}
}

type randomDelayStrategy struct {
	minDelay time.Duration
	maxDelay time.Duration
}

func (s *randomDelayStrategy) Attempt(ctx context.Context) (attempt bool) {
	var delay time.Duration
	if s.minDelay == s.maxDelay {
		delay = s.minDelay
	} else {
		delay = s.minDelay + time.Duration(rand.Int63()%int64(s.maxDelay-s.minDelay))
	}
	attempt = Sleep(ctx, delay)
	return
}

func ExpDelay(seed time.Duration) (strategy Strategy) {
	return PowDelay(seed, math.E)
}

func PowDelay(seed time.Duration, base float64) (strategy Strategy) {
	return &powDelayStrategy{seed, base}
}

type powDelayStrategy struct {
	delay time.Duration
	base  float64
}

func (s *powDelayStrategy) Attempt(ctx context.Context) (attempt bool) {
	attempt = Sleep(ctx, s.delay)
	s.delay = time.Duration(float64(s.delay) * s.base)
	return
}

func Sleep(ctx context.Context, delay time.Duration) (ok bool) {
	select {
	case <-time.After(delay):
		ok = true
	case <-ctx.Done():
		ok = false
	}
	return
}
