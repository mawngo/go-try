package try

import (
	"context"
	"errors"
	"github.com/mawngo/go-try/backoff"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var errFailed = errors.New("failed")

func TestDoRetry(t *testing.T) {
	i := 0
	err := Do(func() error {
		if i >= 2 {
			return nil
		}
		i++
		return errors.New("failed")
	})
	assert.Nil(t, err)
	assert.Equal(t, 2, i)
}

func TestDoRetryContext(t *testing.T) {
	i := 0
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := Do(func() error {
		return nil
	}, WithContext(ctx))
	assert.Equal(t, err, context.Canceled)
	assert.Equal(t, 0, i)
}

func TestDoRetryWithOnRetry(t *testing.T) {
	i := 0
	err := Do(func() error {
		return errFailed
	}, WithMaxAttempts(10), WithOnRetry(func(_ error, _ int) {
		i++
	}))
	assert.True(t, errors.Is(err, errFailed))
	assert.Equal(t, 9, i)
}

func TestDoRetryWithSpecificError(t *testing.T) {
	i := 0
	errAnother := errors.New("another")
	err := Do(func() error {
		if i >= 2 {
			return errAnother
		}
		i++
		return errFailed
	}, WithRetryFor(errFailed))
	assert.True(t, errors.Is(err, errAnother))
	assert.Equal(t, 2, i)
}

func TestDoRetryWithSpecificErrorExclude(t *testing.T) {
	i := 0
	errAnother := errors.New("another")
	err := Do(func() error {
		if i >= 2 {
			return errAnother
		}
		i++
		return errFailed
	}, WithNoRetryFor(errAnother))
	assert.True(t, errors.Is(err, errAnother))
	assert.Equal(t, 2, i)
}

func TestDoRetryLimited(t *testing.T) {
	i := 0
	err := Do(func() error {
		i++
		return errFailed
	}, WithMaxAttempts(10))
	assert.True(t, errors.Is(err, errFailed))
	assert.Equal(t, 10, i)
}

func TestDoRetryBackoff(t *testing.T) {
	start := time.Now()
	i := 0
	err := Do(func() error {
		i++
		return errFailed
	}, WithMaxAttempts(11), WithFixedBackoff(200*time.Millisecond))
	took := time.Since(start)

	assert.True(t, errors.Is(err, errFailed))
	assert.Equal(t, 11, i)
	// Expected total retry sleep took 2s
	assert.Greater(t, took, 2000*time.Millisecond)
	assert.Greater(t, 2100*time.Millisecond, took)
}

func TestDoRetryExponentialBackoff(t *testing.T) {
	start := time.Now()
	i := 0
	err := Do(func() error {
		i++
		return errFailed
	}, WithMaxAttempts(4), WithExponentialBackoff(200*time.Millisecond, 0))
	took := time.Since(start)

	assert.True(t, errors.Is(err, errFailed))
	assert.Equal(t, 4, i)
	// Expected total retry sleep took 1400ms
	assert.Greater(t, took, 1400*time.Millisecond)
	assert.Greater(t, 1500*time.Millisecond, took)
}

func TestDoRetryIncrementalBackoff(t *testing.T) {
	start := time.Now()
	i := 0
	err := Do(func() error {
		i++
		return errFailed
	}, WithMaxAttempts(5), WithBackoff(backoff.NewIncrementalBackoff(200*time.Millisecond, 200*time.Millisecond, 0)))
	took := time.Since(start)

	assert.True(t, errors.Is(err, errFailed))
	assert.Equal(t, 5, i)
	// Expected total retry sleep took 2000ms
	assert.Greater(t, time.Since(start), 2000*time.Millisecond)
	assert.Greater(t, 2100*time.Millisecond, took)
}

func TestGetRetry(t *testing.T) {
	i := 0
	num, err := Get(func() (int, error) {
		if i >= 2 {
			return i, nil
		}
		i++
		return 0, errors.New("failed")
	})
	assert.Nil(t, err)
	assert.Equal(t, 2, i)
	assert.Equal(t, 2, num)
}

func TestDoNotRetryOnContextErr(t *testing.T) {
	i := 0
	err := Do(func() error {
		if i >= 2 {
			return context.DeadlineExceeded
		}
		i++
		return errFailed
	})
	assert.True(t, errors.Is(err, context.DeadlineExceeded))
	assert.True(t, errors.Is(err, errFailed))
	assert.Equal(t, 2, i)
}

func TestNoRetry(t *testing.T) {
	i := 0
	err := Do(func() error {
		i++
		return errFailed
	}, WithAttempts(1))
	assert.True(t, errors.Is(err, errFailed))
	assert.Equal(t, 1, i)
}

func TestWithOptions(t *testing.T) {
	global := NewOptions(WithAttempts(1))
	i := 0
	err := Do(func() error {
		i++
		return errFailed
	}, WithOptions(global), WithAttempts(2))
	assert.True(t, errors.Is(err, errFailed))
	assert.Equal(t, 2, i)
	assert.Equal(t, 1, global.maxAttempts)
}
