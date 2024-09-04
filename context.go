package watchdog

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	lockScript = redis.NewScript(`
if redis.call("SET", KEYS[1], ARGV[1], "NX", "PX", ARGV[2]) then
	return 1
else
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		if redis.call("PEXPIRE", KEYS[1], ARGV[2]) then
			return 1
		else
			return 0
		end
	else
		return 0
	end
end`)

	unlockScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
else
	return 0
end`)

	expireScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("PEXPIRE", KEYS[1], ARGV[2])
else
	return 0
end`)
)

type ctx struct {
	ctx            context.Context
	client         *redis.Client
	key            string
	val            string
	expireDuration time.Duration
}

func (c *ctx) Lock() (bool, error) {
	return lockScript.Run(c.ctx, c.client, []string{c.key}, c.val, c.expireDuration.Milliseconds()).Bool()
}

func (c *ctx) Unlock() (bool, error) {
	return unlockScript.Run(c.ctx, c.client, []string{c.key}, c.val).Bool()
}

func (c *ctx) Expire() error {
	return expireScript.Run(c.ctx, c.client, []string{c.key}, c.val, c.expireDuration.Milliseconds()).Err()
}
