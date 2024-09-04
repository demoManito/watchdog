package watchdog

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Watchdog is a tool to monitor the status of a service.
type Watchdog struct {
	RetryMaxWaitDuration  time.Duration
	RetryIntervalDuration time.Duration
	IsUnlock              bool // wether to unlock the event after the execution is completed, default is true
}

func Default() *Watchdog {
	return &Watchdog{
		RetryMaxWaitDuration:  5 * time.Second,
		RetryIntervalDuration: 100 * time.Millisecond,
		IsUnlock:              true,
	}
}

func (w *Watchdog) Clone() *Watchdog {
	return &Watchdog{
		RetryMaxWaitDuration:  w.RetryMaxWaitDuration,
		RetryIntervalDuration: w.RetryIntervalDuration,
		IsUnlock:              w.IsUnlock,
	}
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
	var (
		c = &ctx{
			client:         client,
			ctx:            newCtx,
			key:            key,
			val:            val.String(),
			expireDuration: duration,
		}
		ok bool
	)
	for {
		ok, err = c.Lock()
		if err != nil || !ok {
			select {
			case <-returyCtx.Done():
				return errors.New("watchdog: timeout")
			case <-time.NewTimer(w.RetryIntervalDuration).C:
				continue // lock reentrant
			}
		}
		break
	}
	if !ok {
		return errors.New("watchdog: lock failed")
	}

	go func() {
		ticker := time.NewTicker(duration / 3)
		defer func() {
			ticker.Stop()
			if w.IsUnlock {
				_, _ = c.Unlock()
			}
		}()
		for {
			select {
			case <-ticker.C:
				_ = c.Expire()
			case <-closeChan:
				return
			case <-newCtx.Done():
				return
			}
		}
	}()
	defer func() {
		if w.IsUnlock {
			_, _ = c.Unlock()
		}
	}()
	return fn(c.ctx)
}
