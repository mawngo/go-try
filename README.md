# Go Try

Simple retry helpers for go. Require go 1.22+

```shell
go get -u github.com/mawngo/go-try/v2
```

## Usage

The retry package provides a `Do()` and `Get()` function which can be used to execute a provided function until it
succeeds.

```go
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

```

By default, the function will be retried 5 times max with a 300ms + 100ms jitter backoff.

### Context

This library provide `*Ctx()` and `GetCtx()` variants that accept a `Context` parameter.
Those functions do not retry on context errors.

Use `WithJoinCtxErr()` to join the last error with the context error.

## Documentation

See [examples](examples/) for more usage examples.

See [options.go](options.go) for available options.

See [backoff.go](backoff/backoff.go) for built-in backoff support.