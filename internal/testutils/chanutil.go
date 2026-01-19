package testutils

import (
	"context"
	"fmt"
	"time"
)

// WaitRecv waits for one value from ch until timeout.
// Returns (value, true) if received; otherwise (zero, false).
func WaitRecv[T any](ch <-chan T, timeout time.Duration) (T, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return WaitRecvCtx(ctx, ch)
}

// WaitRecvCtx waits for one value from ch until ctx is done.
// Returns (value, true) if received; otherwise (zero, false).
func WaitRecvCtx[T any](ctx context.Context, ch <-chan T) (T, bool) {
	select {
	case v, ok := <-ch:
		if !ok {
			var zero T
			return zero, false
		}
		return v, true
	case <-ctx.Done():
		var zero T
		return zero, false
	}
}

// WaitNoRecv returns true if *no* value is received from ch for the entire duration.
// If ch closes, this returns true (i.e., "no value arrived").
func WaitNoRecv[T any](ch <-chan T, dur time.Duration) bool {
	timer := time.NewTimer(dur)
	defer timer.Stop()

	select {
	case <-ch:
		return false
	case <-timer.C:
		return true
	}
}

// Drain drains ch without blocking until it would block.
// Returns number of values drained. If ch is closed, it stops.
func Drain[T any](ch <-chan T) int {
	n := 0
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return n
			}
			n++
		default:
			return n
		}
	}
}

// DrainN drains up to max values without blocking.
// Returns number drained. If max <= 0, drains until would block.
func DrainN[T any](ch <-chan T, max int) int {
	if max <= 0 {
		return Drain(ch)
	}
	n := 0
	for n < max {
		select {
		case _, ok := <-ch:
			if !ok {
				return n
			}
			n++
		default:
			return n
		}
	}
	return n
}

// CollectN collects exactly n values from ch, waiting up to timeout total.
// Returns the collected slice or an error on timeout/close.
func CollectN[T any](ch <-chan T, n int, timeout time.Duration) ([]T, error) {
	if n <= 0 {
		return nil, nil
	}

	deadline := time.NewTimer(timeout)
	defer deadline.Stop()

	out := make([]T, 0, n)
	for len(out) < n {
		select {
		case v, ok := <-ch:
			if !ok {
				return out, fmt.Errorf("channel closed after %d/%d values", len(out), n)
			}
			out = append(out, v)
		case <-deadline.C:
			return out, fmt.Errorf("timeout waiting for %d values; got %d", n, len(out))
		}
	}
	return out, nil
}

// Eventually calls fn repeatedly until it returns nil or timeout expires.
// interval is the sleep between attempts (if <=0, defaults to 10ms).
func Eventually(timeout, interval time.Duration, fn func() error) error {
	if interval <= 0 {
		interval = 10 * time.Millisecond
	}
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// try immediately first
	if err := fn(); err == nil {
		return nil
	}

	for {
		select {
		case <-ticker.C:
			if err := fn(); err == nil {
				return nil
			}
		case <-deadline.C:
			return fmt.Errorf("eventually: timeout after %s", timeout)
		}
	}
}
