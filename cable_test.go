package cable

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func setup() {
	startOnce = sync.Once{}
}

func TestMissingGetProcessManager(t *testing.T) {
	setup()

	defer func() {
		if err := recover(); err == nil {
			t.Error("Expected panic, got nil")
		}
	}()

	_ = GetProcessManager()
}

func TestInitialisedGetProcessManager(t *testing.T) {
	setup()

	_ = NewProcessManager()

	m := GetProcessManager()
	if m == nil {
		t.Error("Expected ProcessManager, got nil")
	}
}

func TestRunnableJob(t *testing.T) {
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

func TestWithErrors(t *testing.T) {
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
				return errors.New("error from runnable process")
			}
		}
	})

	m.AddRunnableProcess(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
				atomic.AddInt32(&count, 1)
				time.Sleep(time.Millisecond * 200)
				panic("panic from runnable process")
			}
		}
	})

	m.AddCleanupProcess(func() error {
		time.Sleep(time.Millisecond * 10)
		atomic.AddInt32(&count, 1)
		return errors.New("error from cleanup process")
	})

	m.AddCleanupProcess(func() error {
		time.Sleep(time.Millisecond * 10)
		atomic.AddInt32(&count, 1)
		panic("panic from cleanup process")
	})

	go func() {
		time.Sleep(time.Millisecond * 50)
		cancel()
	}()

	<-m.Done()

	if atomic.LoadInt32(&count) != 4 {
		t.Errorf("Expected 4, got %d", atomic.LoadInt32(&count))
	}

	if len(m.GetErrors()) != 4 {
		t.Errorf("Expected 4 errors, got %d", len(m.GetErrors()))
	}
}

func TestCleanupContext(t *testing.T) {
	var count int32 = 0

	setup()

	ctx, cancel := context.WithCancel(context.Background())
	m := NewProcessManagerWithContext(ctx)

	m.AddCleanupProcess(func() error {
		time.Sleep(time.Millisecond * 10)
		atomic.AddInt32(&count, 1)
		return nil
	})

	m.AddCleanupProcess(func() error {
		<-m.CleanupCtx().Done()
		atomic.AddInt32(&count, 1)
		return nil
	})

	go func() {
		time.Sleep(time.Millisecond * 1000)
		cancel()
	}()

	<-m.Done()

	if atomic.LoadInt32(&count) != 2 {
		t.Errorf("Expected 2, got %d", atomic.LoadInt32(&count))
	}
}

func TestWithSIGINT(t *testing.T) {
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
		if err := syscall.Kill(syscall.Getpid(), syscall.SIGINT); err != nil {
			t.Errorf("Error sending SIGINT: %s", err)
		}
	}()

	<-m.Done()

	if atomic.LoadInt32(&count) != 2 {
		t.Errorf("Expected 2, got %d", atomic.LoadInt32(&count))
	}
}

func TestWithSIGTERM(t *testing.T) {
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
		if err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM); err != nil {
			t.Errorf("Error sending SIGTERM: %s", err)
		}
	}()

	<-m.Done()

	if atomic.LoadInt32(&count) != 2 {
		t.Errorf("Expected 2, got %d", atomic.LoadInt32(&count))
	}
}
