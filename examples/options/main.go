package main

import (
	"errors"
	"github.com/mawngo/go-try/v2"
	"github.com/mawngo/go-try/v2/backoff"
	"io"
	"time"
)

func main() {
	// Create a shared option
	opt := try.NewOptions(
		try.WithAttempts(3),
		try.WithBackoff(backoff.NewBackoffWithJitter(backoff.NewFixedBackoff(1*time.Second), 100*time.Millisecond)),
		try.WithNoRetryFor(io.EOF),
	)

	i := 0
	err := try.DoWithOptions(func() error {
		if i >= 2 {
			return nil
		}
		i++
		return errors.New("failed")
	}, opt)

	println(err == nil)
	println(i == 2)

	i, err = try.GetWithOptions(func() (int, error) {
		return 2, nil
	}, opt)
	println(err == nil)
	println(i == 2)
}
