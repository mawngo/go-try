# Go Try

Simple retry helpers for go. Require go 1.22+

```shell
go get -u github.com/mawngo/go-try/v2
```

## Usage

The retry package provides a Do() function which can be used to execute a provided function until it succeeds.

```go
package main

import (
	"context"
	"errors"
	"fmt"
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
	}, try.WithOnRetry(func(_ context.Context, err error, i int) {
		fmt.Printf("Retries #%d %s\n", i, err)
	}))

	println(err == nil)
	println(i == 2)
}

```

See [options.go](options.go) for available options.

See [backoff.go](backoff/backoff.go) for built-in backoff support.