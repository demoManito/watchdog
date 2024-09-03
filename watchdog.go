package watchdog

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	defaultWatchdog = &Watchdog{
		RetryMaxWaitDuration:  2 * time.Second,
		RetryIntervalDuration: 100 * time.Millisecond,
		RetainLock:            false,
	}
)

// Watchdog is a tool to monitor the status of a service.
type Watchdog struct {
	RetryMaxWaitDuration  time.Duration
	RetryIntervalDuration time.Duration
	RetainLock            bool // retain the lock after the lock is acquired
}

func New() *Watchdog {
	return defaultWatchdog
}

func (w *Watchdog) Watch(client *redis.Client, key string, duration time.Duration, fn func(ctx context.Context) error) error {
	val, err := uuid.NewV7()
	if err != nil {
		return err
	}
	newCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	returyCtx, cancel := context.WithTimeout(newCtx, w.RetryMaxWaitDuration)
	defer cancel()
	closeChan := make(chan struct{})
	defer close(closeChan)
	ctx := &Context{
		client:         client,
		ctx:            newCtx,
		key:            key,
		val:            val.String(),
		expireDuration: duration,
	}
	var ok bool
	for {
		ok, err = w.Lock(ctx)
		if err != nil {
			select {
			case <-returyCtx.Done():
				return errors.New("watchdog: timeout")
			case <-time.NewTimer(w.RetryIntervalDuration).C:
				continue // NOTICE: lock reentrant
			}
		}
		break
	}
	if !ok {
		return errors.New("watchdog: lock failed")
	}

	wg := &sync.WaitGroup{}
	defer func() {
		wg.Wait()
		if !w.RetainLock {
			_, _ = w.Unlock(ctx)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(duration / 3)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_ = w.Expire(ctx)
			case <-closeChan:
				return
			case <-newCtx.Done():
				return
			}
		}
	}()
	return fn(ctx.ctx)
}
