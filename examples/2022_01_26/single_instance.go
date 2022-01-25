// main.go
package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

func main() {
	action := func(context.Context) (string, error) {
		return "done", nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	// Limiting action for 2 calls every second.
	limited := limit(ctx, 2, time.Second, action)

	// Launching 5 go routines every second.
	// We should see 2 successes and 3 errors every second.
	start := time.Now()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			for i := 0; i < 5; i++ {
				go func() {
					result, err := limited(ctx)
					since := time.Since(start).Milliseconds()
					fmt.Printf("%vms\t-> result: %v\t\t| error: %v\n", since, result, err)
				}()
			}
			time.Sleep(time.Second)
		}
	}
}

type actionFunc func(context.Context) (string, error)

// limit wraps a function and limits its calls over time respecting the
// max and refill period rate.
func limit(ctx context.Context, maxCalls int, refillPeriod time.Duration, action actionFunc) func(context.Context) (string, error) {
	// Start with a filled bucket.
	tokens := maxCalls
	// Creates a ticker to receive periodic events.
	ticker := time.NewTicker(refillPeriod)

	// We could use other synchronization mechanisms, but to keep it simple
	// lets use a mutex to avoid race conditions.
	var mx sync.Mutex
	go func() {
		// Defering the Stop call to make sure we don't leak the ticker.
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// On every tick fill the token bucket.
				mx.Lock()
				fmt.Println("\t-> TokenBucket: refilling!!")
				tokens = maxCalls
				mx.Unlock()
			case <-ctx.Done():
				// If the context is cancelled, we should return from the
				// function and avoid leaking the go routine.
				return
			}
		}
	}()

	// requestToken request a token from the bucket and
	// if returns an error if there are no tokens available.
	requestToken := func() error {
		mx.Lock()
		defer mx.Unlock()

		if tokens <= 0 {
			return errors.New("no tokens available")
		}

		// If there are tokens available,
		// removes one from the bucket and returns nil.
		tokens--
		return nil
	}

	return func(c context.Context) (string, error) {
		if err := requestToken(); err != nil {
			return "", err
		}

		return action(ctx)
	}
}
