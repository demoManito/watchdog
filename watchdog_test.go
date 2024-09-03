package watchdog

import (
	"github.com/redis/go-redis/v9"
)

var (
	client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
)
