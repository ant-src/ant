package value

import (
	"context"
	"testing"
	"time"
)

// A tick the ticker had in flight as its context ended, imitated by a send after the cancel, is
// never received, whether the receive starts after it or was already waiting, and the channel
// reports none held.
func TestTickerChanDiscardsStale(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	c := NewTickerChan(ctx)
	c.Raw() <- "held"
	if c.Len() != 1 {
		t.Fatalf("Len before the cancel = %d, want 1", c.Len())
	}
	cancel()
	if c.Len() != 0 {
		t.Fatalf("Len after the cancel = %d, want 0", c.Len())
	}
	wait := func() (Value, bool, bool) {
		rctx, stop := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer stop()
		return c.Receive(rctx)
	}
	if v, ok, gaveUp := wait(); ok || !gaveUp {
		t.Fatalf("receive of the held tick = %v, %v, %v; want a give-up", v, ok, gaveUp)
	}

	got := make(chan bool)
	go func() {
		_, ok, _ := wait()
		got <- ok
	}()
	time.Sleep(5 * time.Millisecond) // the receive is waiting
	c.Raw() <- "late"
	if <-got {
		t.Fatal("a waiting receive took the late tick")
	}
	if c.Len() != 0 {
		t.Fatalf("Len after the late tick = %d, want 0", c.Len())
	}
}

// A live ticker's channel delivers, and only receives.
func TestTickerChanLive(t *testing.T) {
	c := NewTickerChan(context.Background())
	c.Raw() <- "tick"
	if v, ok, _ := c.Receive(context.Background()); !ok || v != "tick" {
		t.Fatalf("Receive = %v, %v; want tick", v, ok)
	}
	if _, err := c.Send(context.Background(), "x"); err != ErrTicker {
		t.Fatalf("Send err = %v, want ErrTicker", err)
	}
	if err := c.Close(); err != ErrTicker {
		t.Fatalf("Close err = %v, want ErrTicker", err)
	}
}
