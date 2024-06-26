package try

import (
	"context"
	"errors"
	"github.com/mawngo/go-try/backoff"
	"log/slog"
	"strconv"
	"time"
)

const DefaultBackoff = 200 * time.Millisecond
const DefaultMaxAttempts = 5
const defaultMultiplier = 2

type Options struct {
	context          context.Context
	maxAttempts      int
	matcher          ErrorMatcher
	excludedMatcher  ErrorMatcher
	backoffStrategy  backoff.Strategy
	onRetry          OnRetryHandler
	skipContextError bool
}

// ErrorMatcher match the error, return true if matched.
type ErrorMatcher func(err error) bool

// ErrAs is a ErrorMatcher that match error using errors.As.
func ErrAs[T error](err error) bool {
	var e T
	return errors.As(err, &e)
}

// ErrIs return a ErrorMatcher that match error using errors.Is.
func ErrIs(err error) ErrorMatcher {
	return func(e error) bool {
		return errors.Is(e, err)
	}
}

// OnRetryHandler handler that will be called for each retry.
type OnRetryHandler func(err error, i int)

// WithOnRetryLogging return a OnRetryHandler that log a message.
// The log level will automatically be changed to error when reach DefaultMaxAttempts.
func WithOnRetryLogging(level slog.Level, msg string) OnRetryHandler {
	return func(err error, i int) {
		if i >= DefaultMaxAttempts {
			level = slog.LevelError
		}
		slog.Log(context.Background(), level, msg+" - retries #"+strconv.Itoa(i)+" "+err.Error())
	}
}

// RetryOption configure the Options.
type RetryOption func(options *Options)

// WithContext set context of retry.
func WithContext(ctx context.Context) RetryOption {
	return func(options *Options) {
		options.context = ctx
	}
}

// WithMaxAttempts specifies the maximum number retries.
func WithMaxAttempts(attempts int) RetryOption {
	return func(options *Options) {
		options.maxAttempts = attempts
	}
}

// WithUnlimitedAttempts configure unlimited retries.
func WithUnlimitedAttempts() RetryOption {
	return func(options *Options) {
		options.maxAttempts = 0
	}
}

// WithRetryIf match the error for retry.
// If not specified, then all error will be retried, except for context.* errors.
func WithRetryIf(matcher ErrorMatcher, matchers ...ErrorMatcher) RetryOption {
	if len(matchers) == 0 {
		return func(options *Options) {
			options.matcher = matcher
		}
	}
	return func(options *Options) {
		matchers := append([]ErrorMatcher{matcher}, matchers...)
		options.matcher = func(err error) bool {
			for i := range matchers {
				if matchers[i](err) {
					return true
				}
			}
			return false
		}
	}
}

// WithRetryFor match the error for retry using errors.Is.
func WithRetryFor(err error, errs ...error) RetryOption {
	if len(errs) == 0 {
		return func(options *Options) {
			options.matcher = func(e error) bool {
				return errors.Is(e, err)
			}
		}
	}
	return func(options *Options) {
		errs := append([]error{err}, errs...)
		options.matcher = func(e error) bool {
			for i := range errs {
				return errors.Is(e, errs[i])
			}
			return false
		}
	}
}

// WithNoRetryIf exclude the error that matched by matcher.
func WithNoRetryIf(matcher ErrorMatcher, matchers ...ErrorMatcher) RetryOption {
	if len(matchers) == 0 {
		return func(options *Options) {
			options.excludedMatcher = matcher
		}
	}
	return func(options *Options) {
		matchers := append([]ErrorMatcher{matcher}, matchers...)
		options.excludedMatcher = func(err error) bool {
			for i := range matchers {
				if matchers[i](err) {
					return true
				}
			}
			return false
		}
	}
}

// WithNoRetryFor exclude the error that matched by error.Is.
func WithNoRetryFor(err error, errs ...error) RetryOption {
	if len(errs) == 0 {
		return func(options *Options) {
			options.excludedMatcher = func(e error) bool {
				return errors.Is(e, err)
			}
		}
	}
	return func(options *Options) {
		errs := append([]error{err}, errs...)
		options.excludedMatcher = func(e error) bool {
			for i := range errs {
				return errors.Is(e, errs[i])
			}
			return false
		}
	}
}

// WithBackoff configure a BackoffStrategy.
// See backoff.Strategy.
func WithBackoff(strategy backoff.Strategy) RetryOption {
	return func(options *Options) {
		options.backoffStrategy = strategy
	}
}

// WithNoBackoff disabling backoff.
func WithNoBackoff() RetryOption {
	return func(options *Options) {
		options.backoffStrategy = nil
	}
}

// WithFixedBackoff fixed wait time between retries.
func WithFixedBackoff(duration time.Duration) RetryOption {
	return func(options *Options) {
		options.backoffStrategy = backoff.NewFixedBackoff(duration)
	}
}

// WithRandomBackoff fixed wait time between retries with added jitter.
// Default jitter is half of the duration, if you need customize this value, use WithBackoff with backoff.NewRandomBackoff.
func WithRandomBackoff(duration time.Duration) RetryOption {
	return func(options *Options) {
		options.backoffStrategy = backoff.NewRandomBackoff(duration, duration/2)
	}
}

// WithExponentialBackoff exponential wait time between retries.
// Default multiplier is 2, if you need customize this value, use WithBackoff with backoff.NewExponentialBackoff.
func WithExponentialBackoff(initialBackoff time.Duration, maximumBackoff time.Duration) RetryOption {
	return func(options *Options) {
		options.backoffStrategy = backoff.NewExponentialRandomBackoff(initialBackoff, defaultMultiplier, maximumBackoff, initialBackoff/2)
	}
}

// WithExponentialRandomBackoff exponential wait time between retries with added jitter.
// Default multiplier is 2, if you need customize this value, use WithBackoff with backoff.NewExponentialRandomBackoff.
// Default jitter is half of the initialBackoff, if you need customize this value, use WithBackoff with backoff.NewExponentialRandomBackoff.
func WithExponentialRandomBackoff(initialBackoff time.Duration, maximumBackoff time.Duration) RetryOption {
	return func(options *Options) {
		options.backoffStrategy = backoff.NewExponentialBackoff(initialBackoff, defaultMultiplier, maximumBackoff)
	}
}

// WithOnRetry configure listener on each retry.
func WithOnRetry(handler OnRetryHandler) RetryOption {
	return func(options *Options) {
		options.onRetry = handler
	}
}

// WithRetryOnContextError enable retry for context.* errors.
func WithRetryOnContextError() RetryOption {
	return func(options *Options) {
		options.skipContextError = false
	}
}

func (o Options) matchError(err error) bool {
	if o.excludedMatcher != nil && o.excludedMatcher(err) {
		return false
	}
	if o.skipContextError {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return false
		}
	}
	if o.matcher == nil {
		return true
	}
	return o.matcher(err)
}

// WithOptions copy all the specified Options value into this options.
// Useful if you have a global Options somewhere and want to customize it for local use case,
// otherwise just use the DoWithOptions instead.
func WithOptions(opt Options) RetryOption {
	return func(options *Options) {
		*options = opt
	}
}

// NewOptions create an Options.
// Defaults:
// - maxAttempts 5 times.
// - 200ms backoff
// - does not retry on context error, retry on every other error.
func NewOptions(options ...RetryOption) Options {
	otp := Options{
		backoffStrategy:  backoff.NewFixedBackoff(DefaultBackoff),
		maxAttempts:      DefaultMaxAttempts,
		skipContextError: true,
	}
	for _, o := range options {
		o(&otp)
	}
	return otp
}
