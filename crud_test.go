package watchdog

import (
	"context"
	"testing"
	"time"
)

func TestWatchdog_Lock(t *testing.T) {
	ctx := &Context{
		ctx:            context.Background(),
		client:         client,
		key:            "test",
		val:            "test1",
		expireDuration: time.Second,
	}
	err := ctx.client.Del(ctx.ctx, ctx.key).Err()
	if err != nil {
		t.Fatal(err)
	}

	ok, err := New().Lock(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("lock failed")
	}

	ctx.val = "test2"
	ok, err = New().Lock(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("lock reentrant")
	}
}

func TestWatchdog_Unlock(t *testing.T) {
	ctx := &Context{
		ctx:            context.Background(),
		client:         client,
		key:            "test",
		val:            "test1",
		expireDuration: time.Second,
	}
	err := ctx.client.Del(ctx.ctx, ctx.key).Err()
	if err != nil {
		t.Fatal(err)
	}

	ok, err := New().Lock(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("lock failed")
	}

	ctx.val = "test2"
	ok, err = New().Unlock(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Log("unlock reentrant")
	}

	ctx.val = "test1"
	ok, err = New().Unlock(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("unlock failed")
	}
}
