package repository

import (
	"sync"
	"time"
)

type breakerState string

const (
	stateClosed   breakerState = "CLOSED"
	stateOpen     breakerState = "OPEN"
	stateHalfOpen breakerState = "HALF_OPEN"
)

type CircuitBreaker struct {
	mu sync.Mutex

	state breakerState

	failCount int
	threshold int

	openUntil time.Time
	cooldown  time.Duration

	halfOpenInFlight bool
}

func NewCircuitBreaker(threshold int, cooldown time.Duration) *CircuitBreaker {
	if threshold <= 0 {
		threshold = 3
	}
	if cooldown <= 0 {
		cooldown = 5 * time.Second
	}
	return &CircuitBreaker{
		state:     stateClosed,
		threshold: threshold,
		cooldown:  cooldown,
	}
}

// Allow returns whether the call is allowed now.
// When OPEN and cooldown passed, it moves to HALF_OPEN and allows exactly one in-flight request.
func (b *CircuitBreaker) Allow(now time.Time) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case stateClosed:
		return true
	case stateOpen:
		if now.After(b.openUntil) {
			b.state = stateHalfOpen
			b.halfOpenInFlight = false
			// allow 1 probe request
		} else {
			return false
		}
		fallthrough
	case stateHalfOpen:
		if b.halfOpenInFlight {
			return false
		}
		b.halfOpenInFlight = true
		return true
	default:
		return true
	}
}

func (b *CircuitBreaker) OnSuccess(now time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.failCount = 0

	if b.state == stateHalfOpen {
		b.state = stateClosed
		b.halfOpenInFlight = false
	}
}

func (b *CircuitBreaker) OnFailure(now time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// if HALF_OPEN probe failed -> OPEN immediately
	if b.state == stateHalfOpen {
		b.state = stateOpen
		b.failCount = b.threshold
		b.openUntil = now.Add(b.cooldown)
		b.halfOpenInFlight = false
		return
	}

	b.failCount++
	if b.failCount >= b.threshold {
		b.state = stateOpen
		b.openUntil = now.Add(b.cooldown)
	}
}

// For tests/observability (optional)
func (b *CircuitBreaker) State() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return string(b.state)
}
