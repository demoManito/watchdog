## watchdog

> The watchdog strategy is a mechanism to automatically detect and handle expired keys.

## Quick Start

> Based on `github.com/redis/go-redis/v9`

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/demoManito/watchdog"
	"github.com/go-redis/redis/v9"
)

var (
	client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	
	dog = watchdog.Default()
)

func main() {
	// Add a key with a timeout of 2 second，Renew every (2/3)s intervals
	dog.Watch(client, "key", 2*time.Second, func(ctx context.Context) error {
		// Do something
		return nil
	})
}
```