package main

import (
	"errors"
	"github.com/mawngo/go-try/v2"
	"log/slog"
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
	},
		try.WithAttempts(10),
		try.WithFixedBackoff(300*time.Millisecond),
		try.WithOnRetryLogging(slog.LevelInfo, "retrying..."),
	)

	//2025/12/02 15:55:08 INFO retrying... retry=1 err=failed
	//2025/12/02 15:55:08 INFO retrying... retry=2 err=failed

	println(err == nil) // true
	println(i == 2)     //true
}
