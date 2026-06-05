// Package workerpool provides a fixed-size goroutine pool with blocking submit.
package workerpool

import (
	"context"
	"sync"

	"github.com/ygpkg/yg-go/logs"
)

// Task is a unit of work executed by the pool.
type Task func(ctx context.Context) error

// Pool is a fixed-size goroutine pool. Submit blocks when all workers are busy,
// providing natural backpressure.
type Pool struct {
	tasks chan Task
	wg    sync.WaitGroup
}

// New creates a fixed-size goroutine pool with size workers.
// The internal queue capacity equals size — Submit blocks when all workers busy.
func New(size int) *Pool {
	if size <= 0 {
		size = 1
	}
	p := &Pool{
		tasks: make(chan Task, size),
	}
	p.wg.Add(size)
	for i := 0; i < size; i++ {
		go p.worker(i)
	}
	return p
}

// Submit enqueues a task. Blocks until a worker is available, providing backpressure.
func (p *Pool) Submit(fn Task) {
	p.tasks <- fn
}

// Close stops accepting new tasks and waits for all in-flight tasks to complete.
func (p *Pool) Close() {
	close(p.tasks)
	p.wg.Wait()
}

func (p *Pool) worker(id int) {
	defer p.wg.Done()
	for fn := range p.tasks {
		p.safeRun(id, fn)
	}
}

func (p *Pool) safeRun(id int, fn Task) {
	defer func() {
		if r := recover(); r != nil {
			logs.Errorf("workerpool: worker-%d panicked: %v", id, r)
		}
	}()
	if err := fn(context.Background()); err != nil {
		logs.Errorf("workerpool: worker-%d task failed: %v", id, err)
	}
}
