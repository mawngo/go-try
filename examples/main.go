package main

import (
	"errors"
	"github.com/mawngo/go-try/v2"
)

func main() {
	i := 0
	err := try.Do(func() error {
		if i >= 2 {
			return nil
		}
		i++
		return errors.New("failed")
	})

	println(err == nil)
	println(i == 2)
}
