package main

import (
	"errors"
	"github.com/mawngo/go-try/v2"
	"time"
)

func main() {
	i := 0
	err := try.Do(func() error {
		if i >= 2 {
			return nil
		}
		i++
		return errors.New("failed")

		// 100ms back off.
	}, try.WithFixedBackoff(100*time.Millisecond))

	println(err == nil)
	println(i == 2)
}
