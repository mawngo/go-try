package try

import (
	"context"
	"errors"
	"github.com/mawngo/go-try/v2/backoff"
	"log/slog"
	"time"
)

const DefaultBackoff = 300 * time.Millisecond
const DefaultMaxAttempts = 5
const defaultMultiplier = 2

type Options struct {
	maxAttempts     int
	matcher         ErrorMatcher
	excludedMatcher ErrorMatcher
	backoffStrategy backoff.Strategy
	onRetry         OnRetryHandler
}

// ErrorMatcher match the error, return true if matched.
type ErrorMatcher func(err error) bool

// OnRetryHandler handler that will be called for each retry.
type OnRetryHandler func(ctx context.Context, err error, i int)

// RetryOption configure the Options.
type RetryOption func(options *Options)

// NewOptions create an Options.
// Defaults:
// - maxAttempts 5 times.
// - 300 ms backoff + 100 ms jitter.
func NewOptions(options ...RetryOption) Options {
	otp := Options{
		backoffStrategy: backoff.NewRandomBackoff(DefaultBackoff, 100*time.Millisecond),
		maxAttempts:     DefaultMaxAttempts,
	}
	for _, o := range options {
		o(&otp)
	}
	return otp
}

// WithAttempts specifies the maximum number of runs and retries.
// Total retry will be attempts - 1.
// Attempts = 1 means no retry, attempts = 0 mean retry infinity.
func WithAttempts(attempts int) RetryOption {
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
// If not specified, then all errors will be retried, except for context.* errors.
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
// The default jitter is half of the duration, if you need to customize this value, use WithBackoff with backoff.NewRandomBackoff.
func WithRandomBackoff(duration time.Duration) RetryOption {
	return func(options *Options) {
		options.backoffStrategy = backoff.NewRandomBackoff(duration, duration/2)
	}
}

// WithExponentialBackoff exponential wait time between retries.
// Default multiplier is 2, if you need to customize this value, use WithBackoff with backoff.NewExponentialBackoff.
func WithExponentialBackoff(initialBackoff time.Duration, maximumBackoff time.Duration) RetryOption {
	return func(options *Options) {
		options.backoffStrategy = backoff.NewExponentialRandomBackoff(initialBackoff, defaultMultiplier, maximumBackoff, initialBackoff/2)
	}
}

// WithExponentialRandomBackoff exponential wait time between retries with added jitter.
// Default multiplier is 2, if you need to customize this value, use WithBackoff with backoff.NewExponentialRandomBackoff.
// The default jitter is half of the initialBackoff, if you need to customize this value, use WithBackoff with backoff.NewExponentialRandomBackoff.
func WithExponentialRandomBackoff(initialBackoff time.Duration, maximumBackoff time.Duration) RetryOption {
	return func(options *Options) {
		options.backoffStrategy = backoff.NewExponentialBackoff(initialBackoff, defaultMultiplier, maximumBackoff)
	}
}

// WithOnRetry configure listener on each retry.
func WithOnRetry(handler OnRetryHandler, handlers ...OnRetryHandler) RetryOption {
	if len(handlers) == 0 {
		return func(options *Options) {
			options.onRetry = handler
		}
	}
	return func(options *Options) {
		handlers := append([]OnRetryHandler{handler}, handlers...)
		options.onRetry = func(ctx context.Context, err error, retry int) {
			for i := range handlers {
				handlers[i](ctx, err, retry)
			}
		}
	}
}

// WithOnRetryLogging return a RetryOption that log a message on each retry.
// The log level will automatically be changed to error when reach DefaultMaxAttempts.
func WithOnRetryLogging(level slog.Level, msg string) RetryOption {
	return WithOnRetry(NewOnRetryLoggingHandler(level, msg))
}

// NewOnRetryLoggingHandler return a OnRetryHandler that log a message on each retry.
func NewOnRetryLoggingHandler(level slog.Level, msg string) OnRetryHandler {
	return func(ctx context.Context, err error, i int) {
		if i >= DefaultMaxAttempts {
			level = slog.LevelError
		}
		slog.Log(ctx, level, msg, slog.Int("retry", i), slog.Any("err", err))
	}
}

// WithOptions copy all the specified Options value into this Options instance.
// Useful if you have a global Options somewhere and want to customize it for a local use case,
// otherwise use the DoWithOptions instead.
func WithOptions(opt Options) RetryOption {
	return func(options *Options) {
		*options = opt
	}
}

// ErrAs is an ErrorMatcher that match error using errors.As.
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

func (o Options) matchError(err error) bool {
	if o.excludedMatcher != nil && o.excludedMatcher(err) {
		return false
	}
	if o.matcher == nil {
		return true
	}
	return o.matcher(err)
}
