##### February 7th, 2022

# Distributed rate limiting in Go
#### The token bucket pattern implementation with a single source of truth.

In the distributed system era some problems get a whole new perspective. One of these common problems is the denial of service by excessive calls. We always think about how this can affect our systems, but what if we were the bad actors?

This blog post covers a token bucket implementation that is a simple, but effective, pattern to avoid overwhelming the services you depend on. I will focus on a client implementation, but the concept can be used to limit incoming calls and protect your service resources also.

## The token bucket

![Token Bucket](/images/token_bucket.svg)

This pattern focuses on limiting a resource rate of consumption by giving a number of tokens periodically and rejecting any token requisition when the bucket is empty. This rate limiting method is able to impose a periodic limit, but cannot shape or throttle the traffic. If a high number of token requisitions arrives in a full bucket it will authorize every one of them until there are no more tokens in it, creating a burst behavior.

![Token Bucket Timeframe](/images/token_bucket_timeframe.svg)

Every time the bucket is refilled, it's capacity is available for use. This method does not carry state between refills, so if in a previous cycle tokens were not used they are "thrown away" and will not be available after refiling.


A single instance implementation is quite simple (see the example below), since all the needed state is stored in the shared memory.

```go
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
    // Deferring the Stop call to make sure we don't leak the ticker.
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
        // If the context is canceled, we should return from the
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

// Output: (something like this...)
// 0ms     -> result: done         | error: <nil>
// 0ms     -> result:              | error: no tokens available
// 0ms     -> result: done         | error: <nil>
// 0ms     -> result:              | error: no tokens available
// 0ms     -> result:              | error: no tokens available
//         -> TokenBucket: refilling!!
// 1000ms  -> result: done         | error: <nil>
// 1000ms  -> result: done         | error: <nil>
// 1000ms  -> result:              | error: no tokens available
// 1000ms  -> result:              | error: no tokens available
// 1000ms  -> result:              | error: no tokens available
//         -> TokenBucket: refilling!!
// 2000ms  -> result: done         | error: <nil>
// 2000ms  -> result: done         | error: <nil>
// 2000ms  -> result:              | error: no tokens available
// 2000ms  -> result:              | error: no tokens available
// 2000ms  -> result:              | error: no tokens available
//         -> TokenBucket: refilling!!
```

What if, like most of the systems being developed now-a-days, we could not rely on sharing memory (therefore having more than one instance) to build our solution?

If we used a simple token bucket like the one in the example, we would need to control every running instance and make sure they respect their share of the quota. In a situation where the load increases and we need more service replicas, we would have to code a way of broadcasting the updated call limit to every instance to avoid restarting services. Even if we achieved this goal, we would still have to build a central controller to broadcast this information to the replicas.

One way of doing it (and the one we will focus on in this blog post) is use a database to persist and centralize the token bucket's state.

## A centralized token bucket

The first thing we need to do is create the data model for the token bucket. This model should also consider how the refill bucket action will take place.

There are some alternatives to implement the bucket refill action. The more straightforward one would be creating a scheduled procedure, that runs periodically, for every bucket and simply writing its capacity to the available tokens field. Another, and more powerful, idea would be registering every requested token and using a moving time window to calculate if there are available tokens (this could even be used for throttling eventually). Both methods can lead to a functional token bucket implementation, but the first has a big cost (since every bucket would have its own scheduled task) and the second can be quite complex (and would force us to store every call to the limited resource in the database).

A third way, which is a combination of the first and second, is how we will implement the pattern. Instead of creating a scheduled procedure we could, in every token requisition, calculate if we should refill the bucket or not before actually removing a token from the bucket. This can be achieved by storing the last time we refilled the bucket and its refill period, then in every token request calculate if it is time to refill.

![Token Bucket Request and Refill](/images/token_bucket_request_and_refill.svg)

With the refill process defined, we can work on the data model. We already know we need to store the number of available tokens, the token capacity of the bucket and the last time it was refilled. To make it easier to use we also should store a bucket ID and its creation date and time. I will use SQL to describe the data model, since it is so ubiquitous.

```sql
CREATE TABLE token_buckets (
  id             CHAR(36) PRIMARY KEY,
  available      INT UNSIGNED,
  capacity       INT UNSIGNED,
  refill_seconds INT UNSIGNED,
  last_refill_at TIMESTAMP(3)
)
```

Now having buckets to use we could define a stored procedure to simplify the token requisitions.

```sql
CREATE PROCEDURE request_token (IN id CHAR(36))
BEGIN
  -- First try to refill the bucket. The WHERE clause should select only
  -- a bucket that can be refilled.
  UPDATE token_buckets tb
    SET 
      tb.available      = tb.capacity,
      tb.last_refill_at = NOW() 
    WHERE 
      tb.id = id
      AND TIMESTAMPADD(SECOND, tb.refill_seconds, tb.last_refill_at) <= NOW();
  -- Then if there are available tokens remove one.
  UPDATE token_buckets tb
    SET
      tb.available = tb.available - 1
    WHERE 
      tb.id = id
      AND tb.available > 0;
END
```

When the **request_token** Since a bucket is only refilled when there is a token requisition, you can't just query the available tokens to check if you could make a call to the limited resource and you avoid making useless queries to update unused buckets.

## The distributed implementation

![Token Bucket Central](/images/token_bucket_central.svg)

The centralized token bucket repository is a shared state that is accessed by the service replicas. To integrate that in the example we should isolate the parts where this state is accessed and mutated. To decouple the token requisition from the call authorization we can use a **TokenRequester** interface.

```go
type TokenRequester interface {
  // RequestToken requests a token from the bucket.
  // This interface assumes that the implementation knows
  // to which bucket it should make the token request.
  RequestToken(ctx context.Context) error
}
```

The new **limit** function is really simple, because all the refilling and token requisition logic was extracted and is abstracted by the interface.


```go
// limit wraps a function and limits its calls over time respecting the
// max and refill period rate.
func limit(ctx context.Context, tokenRequester TokenRequester, action actionFunc) func(context.Context) (string, error) {
  return func(c context.Context) (string, error) {
    // Using closure to capture the TokenRequester to have access to 
    // the RequestToken method.
    if err := tokenRequester.RequestToken(c); err != nil {
      return "", err
    }

    return action(ctx)
  }
}
```

To code a **TokenRequester** implementation we can leverage the standard library SQL package to access the database. Since the token buckets are shared by the service replicas, they should be inserted in the tables before running the rate limiting logic.

```go
// SQLTokenRequester implements the TokenRequester interface
// by using a sql compliant database as persistence layer.
type SQLTokenRequester struct {
	db       *sql.DB
	bucketID string
}

// NewSQLTokenRequester creates a new SQLTokenRequester and
// returns a pointer to it.
// Considering that the bucket is already created.
func NewSQLTokenRequester(ctx context.Context, db *sql.DB, bucketID string) *SQLTokenRequester {
	return &SQLTokenRequester{
		db:       db,
		bucketID: bucketID,
	}
}

// RequestToken requests a token from the shared bucket.
// This implementation assumes that the token bucket and the stored
// procedures were created beforehand in the database.
func (tr *SQLTokenRequester) RequestToken(ctx context.Context) error {
	q := `
		CALL request_token(?);
	`

	result, err := tr.db.ExecContext(ctx, q, tr.bucketID)
	if err != nil {
		return err
	}

	// If a row is affected, the stored procedure call was successful,
	// meaning that a token was given. If an error occurs or no
	// rows were affected the procedure last update was not successful
	// therefore no token was available.
	if rows, err := result.RowsAffected(); err != nil || rows <= 0 {
		return errors.New("no token available")
	}

	return nil
}
```

This implementation is a dynamic approach to the problem, using ID's to access the buckets and being able to control the usage of several limited resources, but it introduces some initialization complexity. The database also might not be the ideal way of sharing the state, since this is volatile data and nothing is achieved by storing this transient information about the tokens used if the limited resource is not active anymore. These are all implementation details that may be influenced on how you design this pattern to fit your system.

This blog post's main proposition is exploring and showing the ins and outs of a rate limit algorithm implementation and how it could be extended to cover more complex use cases, like the distributed case. Maybe your case has a more static context and you can use a single bucket, maybe you need to protect a resource inside your system from client abuse. How would that change the pattern?
