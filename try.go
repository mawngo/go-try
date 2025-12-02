package try

import (
	"context"
	"errors"
	"time"
)

var ErrRetryAttemptsExceed = errors.New("retry attempts exceed")

// Do perform the given operation.
// Based on the retryOptions, it can retry the operation if it failed.
// See RetryOption.
func Do(op func() error, retryOptions ...RetryOption) error {
	option := NewOptions(retryOptions...)
	//nolint:staticcheck
	return DoCtxWithOptions(nil, op, option)
}

// DoCtx perform the given operation.
// Based on the retryOptions, it can retry the operation if it failed.
// See RetryOption.
// Does not retry on ctx error.
func DoCtx(ctx context.Context, op func() error, retryOptions ...RetryOption) error {
	option := NewOptions(retryOptions...)
	return DoCtxWithOptions(ctx, op, option)
}

// DoWithOptions performs the given operation.
// Based on the options, it can retry the operation if it failed.
func DoWithOptions(op func() error, options Options) error {
	//nolint:staticcheck
	return DoCtxWithOptions(nil, op, options)
}

// DoCtxWithOptions performs the given operation.
// Based on the options, it can retry the operation if it failed.
// Does not retry on ctx error.
func DoCtxWithOptions(ctx context.Context, op func() error, options Options) error {
	cnt := 0
	var lastErr error

	for {
		if ctx != nil {
			if err := ctx.Err(); err != nil {
				return combineErr(options.joinCtxErr, err, lastErr)
			}
		}

		err := op()
		cnt++

		if err != nil {
			if !options.matchError(err) {
				return combineErr(options.joinCtxErr, err, lastErr)
			}
			if options.maxAttempts > 0 && cnt >= options.maxAttempts {
				return errors.Join(ErrRetryAttemptsExceed, combineErr(options.joinCtxErr, err, lastErr))
			}
			if options.backoffStrategy != nil {
				backoff := options.backoffStrategy(err, cnt)
				time.Sleep(min(backoff, maximumBackoff))
			}
			if options.onRetry != nil {
				options.onRetry(ctx, err, cnt)
			}
			if options.joinCtxErr && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
				lastErr = err
			}
			continue
		}
		return nil
	}
}

// Get performs the given operation and return the result.
// Based on the retryOptions, it can retry the operation if it failed.
// See Do.
func Get[T any](op func() (T, error), retryOptions ...RetryOption) (T, error) {
	option := NewOptions(retryOptions...)
	//nolint:staticcheck
	return GetCtxWithOptions(nil, op, option)
}

// GetCtx performs the given operation, and return the result.
// Based on the retryOptions, it can retry the operation if it failed.
// Does not retry on ctx error.
// See Do.
func GetCtx[T any](ctx context.Context, op func() (T, error), retryOptions ...RetryOption) (T, error) {
	option := NewOptions(retryOptions...)
	return GetCtxWithOptions(ctx, op, option)
}

// GetWithOptions performs the given operation and returns the result.
// Based on the options, it can retry the operation if it failed.
// See DoWithOptions.
func GetWithOptions[T any](op func() (T, error), options Options) (T, error) {
	//nolint:staticcheck
	return GetCtxWithOptions(nil, op, options)
}

// GetCtxWithOptions performs the given operation and returns the result.
// Based on the options, it can retry the operation if it failed.
// Does not retry on ctx error.
// See DoCtxWithOptions.
func GetCtxWithOptions[T any](ctx context.Context, op func() (T, error), options Options) (T, error) {
	cnt := 0
	var lastErr error

	for {
		if ctx != nil {
			if err := ctx.Err(); err != nil {
				var empty T
				return empty, combineErr(options.joinCtxErr, err, lastErr)
			}
		}

		v, err := op()
		cnt++

		if err != nil {
			if !options.matchError(err) {
				return v, combineErr(options.joinCtxErr, err, lastErr)
			}
			if options.maxAttempts > 0 && cnt >= options.maxAttempts {
				return v, errors.Join(ErrRetryAttemptsExceed, combineErr(options.joinCtxErr, err, lastErr))
			}
			if options.backoffStrategy != nil {
				backoff := options.backoffStrategy(err, cnt)
				time.Sleep(min(backoff, maximumBackoff))
			}
			if options.onRetry != nil {
				options.onRetry(ctx, err, cnt)
			}
			if options.joinCtxErr && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
				lastErr = err
			}
			continue
		}
		return v, nil
	}
}

func combineErr(join bool, err error, last error) error {
	if last == nil {
		return err
	}
	if join && (errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled)) {
		return errors.Join(err, last)
	}
	return err
}
