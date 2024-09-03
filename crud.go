package watchdog

import "github.com/redis/go-redis/v9"

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

	deleteScript = redis.NewScript(`
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

func (w *Watchdog) Lock(ctx *Context) (bool, error) {
	return lockScript.Run(ctx.ctx, ctx.client, []string{ctx.key}, ctx.val, ctx.expireDuration.Milliseconds()).Bool()
}

func (w *Watchdog) Unlock(ctx *Context) (bool, error) {
	return deleteScript.Run(ctx.ctx, ctx.client, []string{ctx.key}, ctx.val).Bool()
}

func (w *Watchdog) Expire(ctx *Context) error {
	return expireScript.Run(ctx.ctx, ctx.client, []string{ctx.key}, ctx.val, ctx.expireDuration.Milliseconds()).Err()
}
