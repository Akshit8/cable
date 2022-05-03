package cable

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func setup() {
	startOnce = sync.Once{}
}

func TestMissingGetProcessManager(t *testing.T) {
	t.Parallel()

	setup()

	defer func() {
		if err := recover(); err == nil {
			t.Error("Expected panic, got nil")
		}
	}()

	_ = GetProcessManager()
}

func TestInitialisedGetProcessManager(t *testing.T) {
	t.Parallel()

	setup()

	_ = NewProcessManager()

	m := GetProcessManager()
	if m == nil {
		t.Error("Expected ProcessManager, got nil")
	}
}

func TestRunnableJob(t *testing.T) {
	t.Parallel()

	var count int32 = 0

	setup()

	m := NewProcessManager()

	m.AddRunnableProcess(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
				atomic.AddInt32(&count, 1)
				time.Sleep(time.Millisecond * 200)
			}
		}
	})

	go func() {
		time.Sleep(time.Millisecond * 50)
		m.doGracefulShutdown()
	}()

	<-m.Done()

	if atomic.LoadInt32(&count) != 1 {
		t.Errorf("Expected 1, got %d", atomic.LoadInt32(&count))
	}
}

func TestCleanupJob(t *testing.T) {
	t.Parallel()

	var count int32 = 0

	setup()

	m := NewProcessManager()

	m.AddRunnableProcess(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
				atomic.AddInt32(&count, 1)
				time.Sleep(time.Millisecond * 200)
			}
		}
	})

	m.AddCleanupProcess(func() error {
		time.Sleep(time.Millisecond * 10)
		atomic.AddInt32(&count, 1)
		return nil
	})

	go func() {
		time.Sleep(time.Millisecond * 50)
		m.doGracefulShutdown()
	}()

	<-m.Done()

	if atomic.LoadInt32(&count) != 2 {
		t.Errorf("Expected 2, got %d", atomic.LoadInt32(&count))
	}
}

func TestWithCustomContext(t *testing.T) {
	t.Parallel()

	var count int32 = 0

	setup()

	ctx, cancel := context.WithCancel(context.Background())
	m := NewProcessManagerWithContext(ctx)

	m.AddRunnableProcess(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
				atomic.AddInt32(&count, 1)
				time.Sleep(time.Millisecond * 200)
			}
		}
	})

	m.AddCleanupProcess(func() error {
		time.Sleep(time.Millisecond * 10)
		atomic.AddInt32(&count, 1)
		return nil
	})

	go func() {
		time.Sleep(time.Millisecond * 50)
		cancel()
	}()

	<-m.Done()

	if atomic.LoadInt32(&count) != 2 {
		t.Errorf("Expected 2, got %d", atomic.LoadInt32(&count))
	}
}
