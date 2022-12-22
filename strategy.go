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

func Compose(strategies ...Strategy) *CompositeStrategy {
	strategy := compositeStrategyPool.Get().(*CompositeStrategy)
	strategy.strategies = append(strategy.strategies[:0], strategies...)
	return strategy
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

func Sequence(strategies ...Strategy) *SequentialStrategy {
	strategy := sequentialStrategyPool.Get().(*SequentialStrategy)
	strategy.strategies = append(strategy.strategies[:0], strategies...)
	return strategy
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
		return &DelayedStrategy{
			delays:  make([]time.Duration, 20),
			counter: 0,
		}
	},
}

func Delays(delays ...time.Duration) *DelayedStrategy {
	strategy := delayedStrategyPool.Get().(*DelayedStrategy)
	if len(delays) > cap(strategy.delays) {
		strategy.delays = make([]time.Duration, len(delays))
	} else {
		strategy.delays = strategy.delays[:len(delays)]
	}
	copy(strategy.delays, delays)
	strategy.counter = 0
	return strategy
}

type DelayedStrategy struct {
	delays  []time.Duration
	counter int
}

func (s *DelayedStrategy) Attempt(ctx context.Context, _ int) (attempt bool) {
	attempt = s.counter < len(s.delays)
	if !attempt {
		return
	}
	delay := s.delays[s.counter]
	Sleep(ctx, delay)
	s.counter += 1
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

func Function(f StrategyFunc) *FuncStrategy {
	strategy := functionStrategyPool.Get().(*FuncStrategy)
	strategy.f = f
	return strategy
}

type FuncStrategy struct {
	f StrategyFunc
}

func (s *FuncStrategy) Attempt(ctx context.Context, retryNumber int) (attempt bool) {
	return s.f(ctx, retryNumber)
}

func (s *FuncStrategy) Reset() {
	functionStrategyPool.Put(s)
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

func (s *InfiniteAttemptsStrategy) Reset() {
}

var maxRetriesStrategyPool = &sync.Pool{
	New: func() any {
		return &MaxRetriesStrategy{0, 0}
	},
}

func MaxAttempts(attempts int) *MaxRetriesStrategy {
	strategy := maxRetriesStrategyPool.Get().(*MaxRetriesStrategy)
	strategy.maxAttempts = attempts
	strategy.counter = 0
	return strategy
}

type MaxRetriesStrategy struct {
	maxAttempts int
	counter     int
}

func (s *MaxRetriesStrategy) Attempt(_ context.Context, _ int) (attempt bool) {
	// -1 because attempts parameter is max function call count, and maxAttempts parameter is max function rerun count
	attempt = s.counter < s.maxAttempts-1
	s.counter += 1
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

func FixedDelay(delay time.Duration) *FixedDelayStrategy {
	strategy := fixedDelayStrategyPool.Get().(*FixedDelayStrategy)
	strategy.delay = delay
	return strategy
}

type FixedDelayStrategy struct {
	delay time.Duration
}

func (s *FixedDelayStrategy) Attempt(ctx context.Context, _ int) (attempt bool) {
	return Sleep(ctx, s.delay)
}

func (s *FixedDelayStrategy) Reset() {
	fixedDelayStrategyPool.Put(s)
}

var randomDelayStrategyPool = &sync.Pool{
	New: func() any {
		return &RandomDelayStrategy{0, 0}
	},
}

func RandomDelay(min time.Duration, max time.Duration) *RandomDelayStrategy {
	strategy := randomDelayStrategyPool.Get().(*RandomDelayStrategy)
	strategy.minDelay = min
	strategy.maxDelay = max
	return strategy
}

type RandomDelayStrategy struct {
	minDelay time.Duration
	maxDelay time.Duration
}

func (s *RandomDelayStrategy) Attempt(ctx context.Context, _ int) (attempt bool) {
	delay := s.minDelay + time.Duration(rand.Int63()%int64(s.maxDelay-s.minDelay))
	return Sleep(ctx, delay)
}

func (s *RandomDelayStrategy) Reset() {
	randomDelayStrategyPool.Put(s)
}

var linearDelayStrategyPool = &sync.Pool{
	New: func() any {
		return &LinearDelayStrategy{0, 0}
	},
}

func LinearDelay(seed time.Duration, delta time.Duration) *LinearDelayStrategy {
	strategy := linearDelayStrategyPool.Get().(*LinearDelayStrategy)
	strategy.seed = seed
	strategy.delta = delta
	return strategy
}

type LinearDelayStrategy struct {
	seed  time.Duration
	delta time.Duration
}

func (s *LinearDelayStrategy) Attempt(ctx context.Context, retryNumber int) (attempt bool) {
	delay := s.seed + s.delta*time.Duration(retryNumber)
	return Sleep(ctx, delay)
}

func (s *LinearDelayStrategy) Reset() {
	linearDelayStrategyPool.Put(s)
}

func ExpDelay(seed time.Duration) *PowDelayStrategy {
	return PowDelay(seed, math.E)
}

var powDelayStrategyPool = &sync.Pool{
	New: func() any {
		return &PowDelayStrategy{0, 0}
	},
}

func PowDelay(seed time.Duration, base float64) *PowDelayStrategy {
	strategy := powDelayStrategyPool.Get().(*PowDelayStrategy)
	strategy.seed = seed
	strategy.base = base
	return strategy
}

type PowDelayStrategy struct {
	seed time.Duration
	base float64
}

func (s *PowDelayStrategy) Attempt(ctx context.Context, retryNumber int) (attempt bool) {
	delay := time.Duration(float64(s.seed) * math.Pow(s.base, float64(retryNumber)))
	return Sleep(ctx, delay)
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
