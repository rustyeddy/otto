package rules

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rustyeddy/otto/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testRule struct {
	name string
	run  func(ctx context.Context) error
}

func (r testRule) Name() string { return r.name }
func (r testRule) Run(ctx context.Context) error {
	return r.run(ctx)
}

func TestRunnerReturnsFirstError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(cancel)

	runner := NewRunner()
	wantErr := errors.New("boom")
	runner.Add(testRule{name: "err", run: func(context.Context) error { return wantErr }})
	runner.Add(testRule{name: "ok", run: func(context.Context) error { return nil }})

	err := runner.Run(ctx)
	require.ErrorIs(t, err, wantErr)
}

func TestRunnerContextCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(cancel)
	waitCtx, waitCancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(waitCancel)

	started := make(chan struct{})
	runner := NewRunner()
	runner.Add(testRule{name: "block", run: func(ctx context.Context) error {
		close(started)
		<-ctx.Done()
		return nil
	}})

	done := make(chan error, 1)
	go func() { done <- runner.Run(ctx) }()

	select {
	case <-started:
	case <-waitCtx.Done():
		require.Fail(t, "rule did not start before timeout")
	}

	cancel()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-waitCtx.Done():
		require.Fail(t, "runner did not exit after cancel")
	}
}

func TestFollowForwards(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(cancel)

	src := testutils.NewSource[bool]("src", 8)
	dst := testutils.NewSink[bool]("sink", 8)
	//src := &testutils.Source[bool]{name: "button", out: make(chan bool, 1), events: make(chan devices.Event)}
	//dst := &testutils.Sink[bool]{name: "relay", in: make(chan bool, 1), events: make(chan devices.Event)}

	rule := NewFollow("follow", src, dst)

	done := make(chan error, 1)
	go func() { done <- rule.Run(ctx) }()

	src.Set() <- true

	select {
	case got := <-dst.Get():
		require.True(t, got)
	case <-ctx.Done():
		require.Fail(t, "did not receive forwarded value")
	}

	src.Close()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-ctx.Done():
		require.Fail(t, "follow did not exit after source closed")
	}
}
