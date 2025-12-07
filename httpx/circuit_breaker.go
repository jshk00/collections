package httpx

import (
	"errors"
	"net/http"
	"sync/atomic"
	"time"
)

type BreakerConfig struct {
	SuccessThreshold uint32
	FailureThreshold uint32
	Timeout          time.Duration
	TripFunc         func(*http.Response) bool
}

var ErrCircuitBreakerOpen = errors.New("httpx: circuit breaker open")

type State int

func (s State) String() string {
	return [...]string{"closed", "open", "half-open"}[s]
}

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker is implements circuit breaking pattern for improving system resiliency
// CircuitBreaker is only used as client
type CircuitBreaker struct {
	config       BreakerConfig
	failureCount atomic.Uint32
	successCount atomic.Uint32
	state        atomic.Value
}

const (
	defaultFailureThreshold uint32 = 3
	defaultSuccessThreshold uint32 = 1
	defaultTimeout                 = 2 * time.Second
)

func NewCircuitBreaker(config BreakerConfig) *CircuitBreaker {
	if config.FailureThreshold == 0 {
		config.FailureThreshold = defaultFailureThreshold
	}
	if config.SuccessThreshold == 0 {
		config.SuccessThreshold = defaultSuccessThreshold
	}
	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}
	if config.TripFunc == nil {
		config.TripFunc = defaultTripFunc
	}
	cb := &CircuitBreaker{config: config}
	cb.state.Store(StateClosed)
	return cb
}

func (cb *CircuitBreaker) Execute(r *http.Response) {
	if cb.config.TripFunc(r) {
	}
}

func (cb *CircuitBreaker) changeState(s State) {
	cb.failureCount.Store(0)
	cb.successCount.Store(0)
	cb.state.Store(s)
}

func defaultTripFunc(r *http.Response) bool {
	return r.StatusCode > 499
}
