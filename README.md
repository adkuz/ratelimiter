# ratelimiter


Install package:

- `go get github.com/Alex-Kuz/ratelimiter`

Test package:

- `git clone github.com/Alex-Kuz/ratelimiter`
- `cd ratelimiter`
- `go mod vendor`
- `go test`

Usage: 

```golang
package main

import (
	"fmt"
	"time"

	"github.com/Alex-Kuz/ratelimiter"
)

func main() {
	option := &ratelimiter.RateLimiterOptions{
		// maximum number of simultaneous tasks
		PeakLoad: 5,

		// at most 10 tasks can be performed in any interval of 30 seconds
		IntervalOpt: ratelimiter.IntervalOption{
			Interval: 30 * time.Second,
			Limit:    10,
		},
	}

	limiter, err := ratelimiter.MakeRateLimiter(option)
	if err != nil {
		panic(err) // or handle
	}

	for i := 0; i < 20; i++ {

		functionWrapper := func(id int) func() {
			return func() {
				fmt.Printf("[%s] hello from task %d\n", time.Now().String(), id)
			}
		}

		function := functionWrapper(i)

		// add function to queue (fifo is not guaranteed) for performing
		limiter.Perform(function)
	}

	for !limiter.Empty() {
		time.Sleep(time.Second)
		fmt.Printf("[%s] main: limiter is not empty\n", time.Now().String())
	}
}
```
