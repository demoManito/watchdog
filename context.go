package watchdog

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Context struct {
	ctx            context.Context
	client         *redis.Client
	key            string
	val            string
	expireDuration time.Duration
}
