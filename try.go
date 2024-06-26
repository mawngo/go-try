package go_try

import (
	"context"
	"errors"
	"time"
)

var ErrLimitExceed = errors.New("retry limit exceed")

// Do performs the given operation.
// Based on the retryOptions, it can retry the operation, if it failed.
// See RetryOption.
func Do(op func() error, retryOptions ...RetryOption) error {
	option := NewOptions(retryOptions...)
	return DoWithOptions(op, option)
}

// DoWithOptions performs the given operation.
// Based on the options, it can retry the operation, if it failed.
func DoWithOptions(op func() error, options Options) error {
	cnt := 0
	var lastErr error
	for {
		err := op()
		cnt++

		if err != nil {
			if !options.matchError(err) {
				return combineErr(err, lastErr)
			}
			if options.maxAttempts > 0 && cnt >= options.maxAttempts {
				return errors.Join(ErrLimitExceed, combineErr(err, lastErr))
			}
			if options.backoffStrategy != nil {
				time.Sleep(options.backoffStrategy(err, cnt))
			}
			if options.onRetry != nil {
				options.onRetry(err, cnt)
			}
			if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
				lastErr = err
			}
			continue
		}
		return nil
	}
}

// Get performs the given operation, and return the result.
// See DoReturnWithOptions.
func Get[T any](op func() (T, error), retryOptions ...RetryOption) (T, error) {
	option := NewOptions(retryOptions...)
	return GetWithOptions(op, option)
}

// GetWithOptions performs the given operation, and return the result.
// See DoWithOptions.
func GetWithOptions[T any](op func() (T, error), options Options) (T, error) {
	cnt := 0
	var lastErr error
	for {
		v, err := op()
		cnt++

		if err != nil {
			if !options.matchError(err) {
				return v, combineErr(err, lastErr)
			}
			if options.maxAttempts > 0 && cnt >= options.maxAttempts {
				return v, errors.Join(ErrLimitExceed, combineErr(err, lastErr))
			}
			if options.backoffStrategy != nil {
				time.Sleep(options.backoffStrategy(err, cnt))
			}
			if options.onRetry != nil {
				options.onRetry(err, cnt)
			}
			continue
		}
		return v, nil
	}
}

func combineErr(err error, last error) error {
	if last == nil {
		return err
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return errors.Join(err, last)
	}
	return err
}
