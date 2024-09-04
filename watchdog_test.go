package watchdog

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	watchdog = Default()
)

func TestWatchdog_Watch(t *testing.T) {
	var (
		key = "test"
		ctx = context.Background()
	)
	err := client.Del(ctx, key).Err()
	if err != nil {
		t.Fatal(err)
	}

	dog := watchdog.Clone()
	dog.IsUnlock = false
	err = dog.Watch(client, key, 6*time.Second, func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Error(err)
	}

	// test timeout
	err = dog.Watch(client, key, 2*time.Second, func(ctx context.Context) error {
		return nil
	})
	if err == nil || err.Error() != "watchdog: timeout" {
		t.Errorf("expect watchdog: timeout, got %v", err)
	}

	dog.IsUnlock = true
	err = dog.Watch(client, key, 6*time.Second, func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Error(err)
	}

	err = dog.Watch(client, key, 2*time.Second, func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}
