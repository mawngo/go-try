package try

import (
	"context"
	"errors"
	"github.com/mawngo/go-try/v2/backoff"
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
	if err != nil {
		t.Fatal()
	}
	if i != 2 {
		t.Fatal("retry times not match")
	}
}

func TestDoCtxRetry(t *testing.T) {
	t.Run("CancelledContext", func(t *testing.T) {
		i := 0
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := DoCtx(ctx, func() error {
			i++
			return nil
		})
		if !errors.Is(err, context.Canceled) {
			t.Fatal("context error swallowed")
		}
		if i != 0 {
			t.Fatal("must not execute on cancelled context")
		}
	})

	t.Run("ContextCancelled", func(t *testing.T) {
		i := 0
		ctx, cancel := context.WithCancel(context.Background())
		err := DoCtx(ctx, func() error {
			i++
			cancel()
			return errFailed
		})
		if !errors.Is(err, context.Canceled) {
			t.Fatal("context error swallowed")
		}
		if i != 1 {
			t.Fatal("must not retry on cancelled context")
		}
	})
}

func TestDoRetryWithOnRetry(t *testing.T) {
	i := 0
	err := Do(func() error {
		return errFailed
	}, WithAttempts(10), WithOnRetry(func(_ context.Context, _ error, _ int) {
		i++
	}))
	if !errors.Is(err, errFailed) {
		t.Fatal()
	}
	if i != 9 {
		t.Fatal("onRetry not executed")
	}
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
	if !errors.Is(err, errAnother) {
		t.Fatal()
	}
	if i != 2 {
		t.Fatal("WithRetryFor not work")
	}
}

func TestDoRetryWithSpecificErrorExclude(t *testing.T) {
	i := 0
	errAnother := errors.New("another")
	err := Do(func() error {
		i++
		if i >= 2 {
			return errAnother
		}
		return errFailed
	}, WithNoRetryFor(errAnother))
	if !errors.Is(err, errAnother) {
		t.Fatal()
	}
	if i != 2 {
		t.Fatal("WithNoRetryFor not work")
	}
}

func TestDoRetryLimited(t *testing.T) {
	t.Run("MaxAttempts", func(t *testing.T) {
		i := 0
		err := Do(func() error {
			i++
			return errFailed
		}, WithAttempts(10))
		if !errors.Is(err, errFailed) {
			t.Fatal()
		}
		if i != 10 {
			t.Fatal("WithAttempts not work")
		}
	})

	t.Run("1 Attempts (NoRetry)", func(t *testing.T) {
		i := 0
		err := Do(func() error {
			i++
			return errFailed
		}, WithAttempts(1))
		if !errors.Is(err, errFailed) {
			t.Fatal()
		}
		if i != 1 {
			t.Fatal("WithAttempts not work")
		}
	})
	t.Run("Unlimited Attempts (0 Attempts)", func(t *testing.T) {
		i := 0
		err := Do(func() error {
			i++
			if i > 1000 {
				return nil
			}
			return errFailed
		}, WithAttempts(0), WithNoBackoff())
		if err != nil {
			t.Fatal()
		}
		if i != 1001 {
			t.Fatal("WithAttempts not work")
		}
	})
	t.Run("Unlimited Attempts", func(t *testing.T) {
		i := 0
		err := Do(func() error {
			i++
			if i > 1000 {
				return nil
			}
			return errFailed
		}, WithUnlimitedAttempts(), WithNoBackoff())
		if err != nil {
			t.Fatal()
		}
		if i != 1001 {
			t.Fatal("WithAttempts not work")
		}
	})
}

func TestDoRetryBackoff(t *testing.T) {
	start := time.Now()
	i := 0
	err := Do(func() error {
		i++
		return errFailed
	}, WithAttempts(11), WithFixedBackoff(200*time.Millisecond))
	took := time.Since(start)
	if !errors.Is(err, errFailed) {
		t.Fatal()
	}
	if i != 11 {
		t.Fatal()
	}
	// Expected total retry sleep took 2s. 100ms buffer for execution time.
	if took <= 2000*time.Millisecond || took > 2100*time.Millisecond {
		t.Fatal("backoff not work")
	}
}

func TestDoRetryExponentialBackoff(t *testing.T) {
	start := time.Now()
	i := 0
	err := Do(func() error {
		i++
		return errFailed
	}, WithAttempts(4), WithExponentialBackoff(200*time.Millisecond, 0))
	took := time.Since(start)
	if !errors.Is(err, errFailed) {
		t.Fatal()
	}
	if i != 4 {
		t.Fatal()
	}
	// Expected total retry sleep took 1400ms. 100ms buffer for execution time.
	if took <= 1400*time.Millisecond || took >= 1500*time.Millisecond {
		t.Fatal("backoff not work")
	}
}

func TestDoRetryIncrementalBackoff(t *testing.T) {
	start := time.Now()
	i := 0
	err := Do(func() error {
		i++
		return errFailed
	}, WithAttempts(5), WithBackoff(backoff.NewIncrementalBackoff(200*time.Millisecond, 200*time.Millisecond, 0)))
	took := time.Since(start)
	if !errors.Is(err, errFailed) {
		t.Fatal()
	}
	if i != 5 {
		t.Fatal()
	}
	// Expected total retry sleep took 2s. 100ms buffer for execution time.
	if took <= 2000*time.Millisecond || took > 2100*time.Millisecond {
		t.Fatal("backoff not work")
	}
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
	if err != nil {
		t.Fatal()
	}
	if num != 2 {
		t.Fatal("not retried")
	}
}

func TestJoinErr(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		err := DoCtx(ctx, func() error {
			cancel()
			return errFailed
		})
		if !errors.Is(err, context.Canceled) {
			t.Fatal("context error swallowed")
		}
		if errors.Is(err, errFailed) {
			t.Fatal("context error must not joined")
		}
	})
	t.Run("Join", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		err := DoCtx(ctx, func() error {
			cancel()
			return errFailed
		}, WithJoinCtxErr())
		if !errors.Is(err, context.Canceled) {
			t.Fatal("context error swallowed")
		}
		if !errors.Is(err, errFailed) {
			t.Fatal("context error not joined")
		}
	})
	t.Run("DefaultGet", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		_, err := GetCtx(ctx, func() (int, error) {
			cancel()
			return 0, errFailed
		})
		if !errors.Is(err, context.Canceled) {
			t.Fatal("context error swallowed")
		}
		if errors.Is(err, errFailed) {
			t.Fatal("context error must not joined")
		}
	})
	t.Run("JoinGet", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		_, err := GetCtx(ctx, func() (int, error) {
			cancel()
			return 0, errFailed
		}, WithJoinCtxErr())
		if !errors.Is(err, context.Canceled) {
			t.Fatal("context error swallowed")
		}
		if !errors.Is(err, errFailed) {
			t.Fatal("context error not joined")
		}
	})
}

func TestWithOptions(t *testing.T) {
	global := NewOptions(WithAttempts(1))
	i := 0
	err := Do(func() error {
		i++
		return errFailed
	}, WithOptions(global), WithAttempts(2))
	if !errors.Is(err, errFailed) {
		t.Fatal()
	}
	if global.maxAttempts != 1 {
		t.Fatal()
	}
	if i != 2 {
		t.Fatal("WithAttempts must override global options")
	}
}
