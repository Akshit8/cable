package cable

import "sync"

type process struct {
	wg sync.WaitGroup
}

func newProcess() *process {
	return &process{
		wg: sync.WaitGroup{},
	}
}

func (p *process) Run(fn func()) {
	p.wg.Add(1)

	go func() {
		defer p.wg.Done()
		fn()
	}()
}

func (p *process) Wait() {
	p.wg.Wait()
}
