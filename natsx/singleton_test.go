package natsx_test

import (
	"context"
	"testing"
	"time"

	"github.com/altairsix/pkg/natsx"
	"github.com/nats-io/go-nats"
	"github.com/savaki/randx"
)

func TestAsSingleton(t *testing.T) {
	nc, _ := nats.Connect(nats.DefaultURL)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	subject := randx.AlphaN(20)
	interval := time.Millisecond * 25

	// Setup Action 1
	//
	done1, fn1 := natsx.Singleton(nc, subject, interval, nopFunc())
	go fn1(ctx)

	// Setup Action 2
	//
	done2, fn2 := natsx.Singleton(nc, subject, interval, nopFunc())
	go fn2(ctx)

	// Give nats time to fight it out
	//
	time.Sleep(interval * 4)

	// Then
	//
	select {
	case <-done1:
		t.Errorf("expected ctx1 to still be valid as action1 started first")
		return
	case <-time.After(time.Millisecond):
	}

	select {
	case <-done2:
	case <-time.After(time.Millisecond):
		t.Errorf("expected ctx2 to be canceled as action1 started first")
		return
	}
}

func nopFunc() func(ctx context.Context) error {
	return func(ctx context.Context) error {
		select {
		case <-time.After(time.Minute):
		case <-ctx.Done():
		}

		return nil
	}
}
