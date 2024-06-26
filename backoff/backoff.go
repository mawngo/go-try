package backoff

import (
	"math"
	"math/rand"
	"time"
)

// Strategy is a function that calculate the backoff.
type Strategy func(err error, i int) time.Duration

// NewFixedBackoff return a BackoffStrategy that backoff at a fixed rate.
func NewFixedBackoff(backoff time.Duration) Strategy {
	return func(_ error, _ int) time.Duration {
		return backoff
	}
}

// NewRandomBackoff return a NewFixedBackoff with added random jitter.
func NewRandomBackoff(minBackoff time.Duration, jitter time.Duration) Strategy {
	return NewBackoffWithJitter(NewFixedBackoff(minBackoff), jitter)
}

// NewBackoffWithJitter add random jitter to existing BackoffStrategy.
// The jitter is always added, which may not respect configuration of existing BackoffStrategy,
// for example, ExponentialBackoff max wait time may > maximumBackoff because of the jitter.
//
// This construct is intended to easily adding jitter to user defined backoff Strategy.
// For built-in Strategy you better use the RandomBackoff variant of it.
func NewBackoffWithJitter(backoff Strategy, jitter time.Duration) Strategy {
	return func(err error, i int) time.Duration {
		return backoff(err, i) + time.Duration(rand.Int63n(int64(jitter)))
	}
}

// NewExponentialBackoff return a BackoffStrategy that backoff at an exponential rate.
func NewExponentialBackoff(initialBackoff time.Duration, multiplier int, maximumBackoff time.Duration) Strategy {
	return func(_ error, i int) time.Duration {
		exponential := math.Pow(float64(multiplier), float64(i-1))
		backoff := initialBackoff * time.Duration(exponential)
		if maximumBackoff == 0 {
			return backoff
		}
		return min(backoff, maximumBackoff)
	}
}

// NewExponentialRandomBackoff return a ExponentialBackoff with added random jitter, and respect the maximum backoff.
func NewExponentialRandomBackoff(initialBackoff time.Duration, multiplier int, maximumBackoff time.Duration, jitter time.Duration) Strategy {
	return func(_ error, i int) time.Duration {
		exponential := math.Pow(float64(multiplier), float64(i-1))
		jitter := time.Duration(rand.Int63n(int64(jitter)))
		backoff := initialBackoff * time.Duration(exponential)
		if maximumBackoff == 0 {
			return backoff
		}
		if backoff >= maximumBackoff {
			return max(backoff - jitter + 0)
		}
		return min(backoff+jitter, maximumBackoff)
	}
}

// NewIncrementalBackoff return a BackoffStrategy that increment backoff every retry.
func NewIncrementalBackoff(initialBackoff time.Duration, incremental time.Duration, maximumBackoff time.Duration) Strategy {
	return func(_ error, i int) time.Duration {
		inc := incremental * time.Duration(i-1)
		backoff := initialBackoff + inc
		if maximumBackoff == 0 {
			return backoff
		}
		return min(backoff, maximumBackoff)
	}
}

// NewIncrementalRandomBackoff return an IncrementalBackoff with added random jitter, and respect the maximum backoff.
func NewIncrementalRandomBackoff(initialBackoff time.Duration, incremental time.Duration, maximumBackoff time.Duration, jitter time.Duration) Strategy {
	return func(_ error, i int) time.Duration {
		inc := incremental * time.Duration(i-1)
		jitter := time.Duration(rand.Int63n(int64(jitter)))
		backoff := initialBackoff + inc
		if maximumBackoff == 0 {
			return backoff
		}
		if backoff >= maximumBackoff {
			return max(backoff - jitter + 0)
		}
		return min(backoff+jitter, maximumBackoff)
	}
}
