# Try

Simple retry helpers for go.

## Usage

The retry package provides a Do() function which can be used to execute a provided function until it succeds.

```go
package main

import (
	"errors"
	"fmt"
	"github.com/mawngo/go-try"
)

func main() {
	i := 0
	err := try.Do(func() error {
		if i >= 2 {
			return nil
		}
		i++
		return errors.New("failed")
	}, try.WithOnRetry(func(err error, i int) {
		fmt.Printf("Retries #%d %s\n", i, err)
	}))

	println(err == nil)
	println(i == 2)
}
```

See [options.go](options.go) for available options.