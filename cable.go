package cable

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	processManager *ProcessManager
	startOnce      = sync.Once{}
)

type (
	RunnableProcess func(context.Context) error
	CleanupProcess  func() error
)

type ProcessManager struct {
	lock *sync.RWMutex

	logger Logger

	cleanCtx       context.Context
	cleanCtxCancel context.CancelFunc
	doneCtx        context.Context
	doneCtxCancel  context.CancelFunc

	runner *process

	errors           []error
	cleanupProcesses []CleanupProcess
}

func newProcessManager(option ...Option) *ProcessManager {
	startOnce.Do(func() {
		opts := newOptions(option...)

		processManager = &ProcessManager{
			lock:   &sync.RWMutex{},
			logger: opts.logger,
			errors: make([]error, 0),
			runner: newProcess(),
		}

		processManager.start(opts.ctx)
	})

	return processManager
}

func NewProcessManager(option ...Option) *ProcessManager {
	return newProcessManager(option...)
}

func NewProcessManagerWithContext(ctx context.Context, option ...Option) *ProcessManager {
	return newProcessManager(append(option, WithContext(ctx))...)
}

func GetProcessManager() *ProcessManager {
	if processManager == nil {
		panic("ProcessManager is not initialized")
	}

	return processManager
}

func (p *ProcessManager) AddRunnableProcess(fn RunnableProcess) {
	p.runner.Run(func() {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Errorf("Panic in running process: %s", err)
				p.logger.Error(message)
				p.lock.Lock()
				p.errors = append(p.errors, message)
				p.lock.Unlock()
			}
		}()

		err := fn(p.cleanCtx)
		if err != nil {
			p.lock.Lock()
			p.errors = append(p.errors, err)
			p.lock.Unlock()
		}
	})
}

func (p *ProcessManager) AddCleanupProcess(fn CleanupProcess) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.cleanupProcesses = append(p.cleanupProcesses, fn)
}

func (p *ProcessManager) doCleanupProcess(fn CleanupProcess) {
	defer func() {
		if err := recover(); err != nil {
			message := fmt.Errorf("Panic in cleanup process: %s", err)
			p.logger.Error(message)
			p.lock.Lock()
			p.errors = append(p.errors, message)
			p.lock.Unlock()
		}
	}()

	err := fn()
	if err != nil {
		p.lock.Lock()
		p.errors = append(p.errors, err)
		p.lock.Unlock()
	}
}

func (p *ProcessManager) doGracefulShutdown() {
	p.cleanCtxCancel()

	for _, cleanupProcess := range p.cleanupProcesses {
		func(c CleanupProcess) {
			p.runner.Run(func() {
				p.doCleanupProcess(c)
			})
		}(cleanupProcess)
	}

	go func() {
		p.runner.Wait()
		p.lock.Lock()
		p.doneCtxCancel()
		p.lock.Unlock()
	}()
}

func (p *ProcessManager) handleInterruptSignals(ctx context.Context) {
	interrupt := make(chan os.Signal, 1)

	signal.Notify(interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGTSTP)
	defer signal.Stop(interrupt)

	pid := syscall.Getpid()

	for {
		select {
		case sig := <-interrupt:
			switch sig {
			case syscall.SIGINT:
				p.logger.Infof("Received SIGINT for process: %d. Terminating process...", pid)
				p.doGracefulShutdown()
				return
			case syscall.SIGTERM:
				p.logger.Infof("Received SIGTERM for process: %d. Terminating process...", pid)
				p.doGracefulShutdown()
				return
			case syscall.SIGTSTP:
				p.logger.Infof("Received SIGTSTP for process: %d. Terminating process...", pid)
			default:
				p.logger.Infof("Received signal: %s for process: %d. Terminating process...", sig.String(), pid)
			}
		case <-ctx.Done():
			p.logger.Infof("Received context done signal(%s) for process: %d. Terminating process...", ctx.Err(), pid)
		}
	}
}

func (p *ProcessManager) start(ctx context.Context) {
	p.cleanCtx, p.cleanCtxCancel = context.WithCancel(ctx)
	p.doneCtx, p.doneCtxCancel = context.WithCancel(context.Background())

	go p.handleInterruptSignals(ctx)
}

func (p *ProcessManager) wait() {
	p.runner.Wait()
}

func (p *ProcessManager) Done() <-chan struct{} {
	return p.doneCtx.Done()
}

func (p *ProcessManager) CleanupCtx() context.Context {
	return p.cleanCtx
}
