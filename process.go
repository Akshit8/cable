package cable

import "sync"

type processRunner struct {
	wg sync.WaitGroup
}

func newProcessRunner() *processRunner {
	return &processRunner{
		wg: sync.WaitGroup{},
	}
}

func (p *processRunner) Run(fn func()) {
	p.wg.Add(1)

	go func() {
		defer p.wg.Done()
		fn()
	}()
}

func (p *processRunner) Wait() {
	p.wg.Wait()
}
