package retry

import (
	"context"
	"math"
	"math/rand"
	"sync"
	"time"
)

type Strategy interface {
	Attempt(ctx context.Context, retryNumber int) (attempt bool)
}

type Resetter interface {
	Reset()
}

func DefaultStrategy() Strategy {
	return Compose(MaxAttempts(10), PowDelay(100*time.Millisecond, math.Sqrt2))
}

var compositeStrategyPool = &sync.Pool{
	New: func() any {
		return &CompositeStrategy{
			strategies: make([]Strategy, 10),
		}
	},
}

func Compose(strategies ...Strategy) (strategy *CompositeStrategy) {
	strategy = compositeStrategyPool.Get().(*CompositeStrategy)
	strategy.strategies = append(strategy.strategies[:0], strategies...)
	return
}

type CompositeStrategy struct {
	strategies []Strategy
}

func (s *CompositeStrategy) Attempt(ctx context.Context, retryNumber int) (attempt bool) {
	attempt = true
	for index := 0; attempt && index < len(s.strategies); index++ {
		attempt = s.strategies[index].Attempt(ctx, retryNumber)
	}
	return
}

func (s *CompositeStrategy) Reset() {
	for _, strategy := range s.strategies {
		if resetter, ok := strategy.(Resetter); ok {
			resetter.Reset()
		}
	}
	compositeStrategyPool.Put(s)
}

var sequentialStrategyPool = &sync.Pool{
	New: func() any {
		return &SequentialStrategy{
			strategies: make([]Strategy, 10),
		}
	},
}

func Sequence(strategies ...Strategy) (strategy *SequentialStrategy) {
	strategy = sequentialStrategyPool.Get().(*SequentialStrategy)
	strategy.strategies = append(strategy.strategies[:0], strategies...)
	return
}

type SequentialStrategy struct {
	strategies []Strategy
}

func (s *SequentialStrategy) Attempt(ctx context.Context, retryNumber int) (attempt bool) {
	attempt = false
	for index := 0; !attempt && index < len(s.strategies); index++ {
		attempt = s.strategies[index].Attempt(ctx, retryNumber)
	}
	return
}

func (s *SequentialStrategy) Reset() {
	for _, strategy := range s.strategies {
		if resetter, ok := strategy.(Resetter); ok {
			resetter.Reset()
		}
	}
	sequentialStrategyPool.Put(s)
}

var delayedStrategyPool = &sync.Pool{
	New: func() any {
		return &DelayedStrategy{0, make([]time.Duration, 20)}
	},
}

func Delays(delays ...time.Duration) (strategy *DelayedStrategy) {
	strategy = delayedStrategyPool.Get().(*DelayedStrategy)
	strategy.index = 0
	strategy.delays = append(strategy.delays[:0], delays...)
	return
}

type DelayedStrategy struct {
	index  int
	delays []time.Duration
}

func (s *DelayedStrategy) Attempt(ctx context.Context, _ int) (attempt bool) {
	attempt = s.index < len(s.delays)
	if !attempt {
		return
	}
	delay := s.delays[s.index]
	attempt = Sleep(ctx, delay)
	s.index++
	return
}

func (s *DelayedStrategy) Reset() {
	delayedStrategyPool.Put(s)
}

var functionStrategyPool = &sync.Pool{
	New: func() any {
		return &FuncStrategy{nil}
	},
}

type StrategyFunc func(ctx context.Context, retryNumber int) (attempt bool)

func Function(f StrategyFunc) (strategy *FuncStrategy) {
	strategy = functionStrategyPool.Get().(*FuncStrategy)
	strategy.retryFunc = f
	return
}

type FuncStrategy struct {
	retryFunc StrategyFunc
}

func (s *FuncStrategy) Attempt(ctx context.Context, retryNumber int) (attempt bool) {
	attempt = s.retryFunc(ctx, retryNumber)
	return
}

func (s *FuncStrategy) Reset() {
	functionStrategyPool.Put(s)
}

func Infinite() (strategy *InfiniteAttemptsStrategy) {
	strategy = infiniteAttemptStrategyPtr
	return
}

var infiniteAttemptStrategyPtr = &InfiniteAttemptsStrategy{}

type InfiniteAttemptsStrategy struct {
}

func (s *InfiniteAttemptsStrategy) Attempt(_ context.Context, _ int) (attempt bool) {
	attempt = true
	return
}

func (s *InfiniteAttemptsStrategy) Reset() {
}

var maxRetriesStrategyPool = &sync.Pool{
	New: func() any {
		return &MaxRetriesStrategy{0}
	},
}

func MaxAttempts(attempts int) (strategy *MaxRetriesStrategy) {
	strategy = maxRetriesStrategyPool.Get().(*MaxRetriesStrategy)
	// -1 because attempts parameter is max function call count, and remainingAttempts parameter is max function rerun count
	strategy.remainingAttempts = attempts - 1
	return
}

type MaxRetriesStrategy struct {
	remainingAttempts int
}

func (s *MaxRetriesStrategy) Attempt(_ context.Context, _ int) (attempt bool) {
	attempt = s.remainingAttempts > 0
	if attempt {
		s.remainingAttempts--
	}
	return
}

func (s *MaxRetriesStrategy) Reset() {
	maxRetriesStrategyPool.Put(s)
}

var fixedDelayStrategyPool = &sync.Pool{
	New: func() any {
		return &FixedDelayStrategy{0}
	},
}

func FixedDelay(delay time.Duration) (strategy *FixedDelayStrategy) {
	strategy = fixedDelayStrategyPool.Get().(*FixedDelayStrategy)
	strategy.delay = delay
	return
}

type FixedDelayStrategy struct {
	delay time.Duration
}

func (s *FixedDelayStrategy) Attempt(ctx context.Context, _ int) (attempt bool) {
	attempt = Sleep(ctx, s.delay)
	return
}

func (s *FixedDelayStrategy) Reset() {
	fixedDelayStrategyPool.Put(s)
}

var randomDelayStrategyPool = &sync.Pool{
	New: func() any {
		return &RandomDelayStrategy{0, 0}
	},
}

func RandomDelay(min time.Duration, max time.Duration) (strategy *RandomDelayStrategy) {
	strategy = randomDelayStrategyPool.Get().(*RandomDelayStrategy)
	strategy.minDelay = min
	strategy.maxDelay = max
	return
}

type RandomDelayStrategy struct {
	minDelay time.Duration
	maxDelay time.Duration
}

func (s *RandomDelayStrategy) Attempt(ctx context.Context, _ int) (attempt bool) {
	var delay time.Duration
	if s.minDelay == s.maxDelay {
		delay = s.minDelay
	} else {
		delay = s.minDelay + time.Duration(rand.Int63()%int64(s.maxDelay-s.minDelay))
	}
	attempt = Sleep(ctx, delay)
	return
}

func (s *RandomDelayStrategy) Reset() {
	randomDelayStrategyPool.Put(s)
}

var linearDelayStrategyPool = &sync.Pool{
	New: func() any {
		return &LinearDelayStrategy{0, 0}
	},
}

func LinearDelay(seed time.Duration, delta time.Duration) (strategy *LinearDelayStrategy) {
	strategy = linearDelayStrategyPool.Get().(*LinearDelayStrategy)
	strategy.delay = seed
	strategy.delta = delta
	return
}

type LinearDelayStrategy struct {
	delay time.Duration
	delta time.Duration
}

func (s *LinearDelayStrategy) Attempt(ctx context.Context, _ int) (attempt bool) {
	attempt = Sleep(ctx, s.delay)
	s.delay = s.delay + s.delta
	return
}

func (s *LinearDelayStrategy) Reset() {
	linearDelayStrategyPool.Put(s)
}

func ExpDelay(seed time.Duration) (strategy *PowDelayStrategy) {
	strategy = PowDelay(seed, math.E)
	return
}

var powDelayStrategyPool = &sync.Pool{
	New: func() any {
		return &PowDelayStrategy{0, 0}
	},
}

func PowDelay(seed time.Duration, base float64) (strategy *PowDelayStrategy) {
	strategy = powDelayStrategyPool.Get().(*PowDelayStrategy)
	strategy.delay = seed
	strategy.base = base
	return
}

type PowDelayStrategy struct {
	delay time.Duration
	base  float64
}

func (s *PowDelayStrategy) Attempt(ctx context.Context, _ int) (attempt bool) {
	attempt = Sleep(ctx, s.delay)
	s.delay = time.Duration(float64(s.delay) * s.base)
	return
}

func (s *PowDelayStrategy) Reset() {
	powDelayStrategyPool.Put(s)
}

func Sleep(ctx context.Context, delay time.Duration) (ok bool) {
	timer := newTimerFromPool(delay)
	select {
	case <-timer.C:
		ok = true
	case <-ctx.Done():
		ok = false
		timer.Stop()
	}
	returnTimerToPool(timer)
	return
}

var timerPool = &sync.Pool{}

func newTimerFromPool(delay time.Duration) (timer *time.Timer) {
	timerFromPool := timerPool.Get()
	if timerFromPool == nil {
		timer = time.NewTimer(delay)
	} else {
		timer = timerFromPool.(*time.Timer)
		timer.Reset(delay)
	}
	return
}

func returnTimerToPool(timer *time.Timer) {
	// stop timer if it has not been stopped or fired
	timer.Stop()
	// drain timer channel
	select {
	case <-timer.C:
		break
	default:
		break
	}
	timerPool.Put(timer)
	return
}
