package retry

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"reflect"
	"time"
)

type Strategy interface {
	Attempt(ctx context.Context, retryNumber int, err error) (attempt bool)
}

func Default() Strategy {
	return And(MaxAttempts(5), PowDelay(100*time.Millisecond, math.Sqrt2))
}

func And(strategies ...Strategy) (strategy AndStrategy) {
	return AndStrategy{strategies}
}

var _ Strategy = (*AndStrategy)(nil)

type AndStrategy struct {
	strategies []Strategy
}

func (s AndStrategy) Attempt(ctx context.Context, retryNumber int, err error) (attempt bool) {
	attempt = true
	for index := 0; attempt && index < len(s.strategies); index++ {
		attempt = s.strategies[index].Attempt(ctx, retryNumber, err)
	}
	return
}

func Or(strategies ...Strategy) (strategy OrStrategy) {
	return OrStrategy{strategies}
}

var _ Strategy = (*OrStrategy)(nil)

type OrStrategy struct {
	strategies []Strategy
}

func (s OrStrategy) Attempt(ctx context.Context, retryNumber int, err error) (attempt bool) {
	attempt = false
	for index := 0; !attempt && index < len(s.strategies); index++ {
		attempt = s.strategies[index].Attempt(ctx, retryNumber, err)
	}
	return
}

func Delays(delays ...time.Duration) (strategy *DelayedStrategy) {
	return &DelayedStrategy{0, delays}
}

var _ Strategy = (*DelayedStrategy)(nil)

type DelayedStrategy struct {
	index  int
	delays []time.Duration
}

func (s *DelayedStrategy) Attempt(ctx context.Context, _ int, _ error) (attempt bool) {
	attempt = s.index < len(s.delays)
	if !attempt {
		return
	}
	delay := s.delays[s.index]
	s.index++
	attempt = Sleep(ctx, delay)
	return
}

type StrategyFunc func(ctx context.Context, retryNumber int, _ error) (attempt bool)

func Function(retryFunc StrategyFunc) (strategy FuncStrategy) {
	return FuncStrategy{retryFunc}
}

var _ Strategy = (*FuncStrategy)(nil)

type FuncStrategy struct {
	retryFunc StrategyFunc
}

func (s FuncStrategy) Attempt(ctx context.Context, retryNumber int, err error) (attempt bool) {
	attempt = s.retryFunc(ctx, retryNumber, err)
	return
}

var infiniteAttemptStrategyPtr = &InfiniteAttemptsStrategy{}

func Infinite() (strategy *InfiniteAttemptsStrategy) {
	return infiniteAttemptStrategyPtr
}

var _ Strategy = (*InfiniteAttemptsStrategy)(nil)

type InfiniteAttemptsStrategy struct {
}

func (s *InfiniteAttemptsStrategy) Attempt(_ context.Context, _ int, _ error) (attempt bool) {
	attempt = true
	return
}

func MaxAttempts(attempts int) (strategy *MaxRetriesStrategy) {
	// -1 because attempts parameter is max function call count, and remainingAttempts parameter is max function rerun count
	return &MaxRetriesStrategy{attempts - 1}
}

var _ Strategy = (*MaxRetriesStrategy)(nil)

type MaxRetriesStrategy struct {
	remainingAttempts int
}

func (s *MaxRetriesStrategy) Attempt(_ context.Context, _ int, _ error) (attempt bool) {
	attempt = s.remainingAttempts > 0
	if attempt {
		s.remainingAttempts--
	}
	return
}

func FixedDelay(delay time.Duration) (strategy FixedDelayStrategy) {
	return FixedDelayStrategy{delay}
}

var _ Strategy = (*FixedDelayStrategy)(nil)

type FixedDelayStrategy struct {
	delay time.Duration
}

func (s FixedDelayStrategy) Attempt(ctx context.Context, _ int, _ error) (attempt bool) {
	attempt = Sleep(ctx, s.delay)
	return
}

func RandomDelay(minDelay time.Duration, maxDelay time.Duration) (strategy RandomDelayStrategy) {
	return RandomDelayStrategy{minDelay, maxDelay}
}

var _ Strategy = (*RandomDelayStrategy)(nil)

type RandomDelayStrategy struct {
	minDelay time.Duration
	maxDelay time.Duration
}

func (s RandomDelayStrategy) Attempt(ctx context.Context, _ int, _ error) (attempt bool) {
	var delay time.Duration
	if s.minDelay == s.maxDelay {
		delay = s.minDelay
	} else {
		delay = s.minDelay + time.Duration(rand.Int63()%int64(s.maxDelay-s.minDelay))
	}
	attempt = Sleep(ctx, delay)
	return
}

func LinearDelay(seed time.Duration, delta time.Duration) (strategy *LinearDelayStrategy) {
	return &LinearDelayStrategy{seed, delta}
}

var _ Strategy = (*LinearDelayStrategy)(nil)

type LinearDelayStrategy struct {
	delay time.Duration
	delta time.Duration
}

func (s *LinearDelayStrategy) Attempt(ctx context.Context, _ int, _ error) (attempt bool) {
	attempt = Sleep(ctx, s.delay)
	s.delay = s.delay + s.delta
	return
}

func ExpDelay(seed time.Duration) (strategy Strategy) {
	return PowDelay(seed, math.E)
}

func PowDelay(seed time.Duration, base float64) (strategy *PowDelayStrategy) {
	return &PowDelayStrategy{seed, base}
}

var _ Strategy = (*PowDelayStrategy)(nil)

type PowDelayStrategy struct {
	delay time.Duration
	base  float64
}

func (s *PowDelayStrategy) Attempt(ctx context.Context, _ int, _ error) (attempt bool) {
	attempt = Sleep(ctx, s.delay)
	s.delay = time.Duration(float64(s.delay) * s.base)
	return
}

func Is(err error) (strategy IsStrategy) {
	return IsStrategy{err}
}

var _ Strategy = (*IsStrategy)(nil)

type IsStrategy struct {
	err error
}

func (s *IsStrategy) Attempt(_ context.Context, _ int, err error) (attempt bool) {
	attempt = errors.Is(err, s.err)
	return
}

func Type(err error) (strategy TypeStrategy) {
	val := reflect.ValueOf(err)
	targetType := val.Type()
	return TypeStrategy{targetType}
}

var _ Strategy = (*TypeStrategy)(nil)

type TypeStrategy struct {
	targetType reflect.Type
}

func (s *TypeStrategy) Attempt(_ context.Context, _ int, err error) (attempt bool) {
	for err != nil {
		if reflect.TypeOf(err).AssignableTo(s.targetType) {
			attempt = true
			break
		}
		err = errors.Unwrap(err)
	}
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
