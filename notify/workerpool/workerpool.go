package workerpool

import (
	"sync"
	"time"
)

type Task struct {
	Err error
	f   func() error
}

func NewTask(f func() error) *Task {
	return &Task{f: f}
}

func (t *Task) Run(wg *sync.WaitGroup, backoffCfg BackoffConfig) {
	defer wg.Done()
	var err error
	backoff := backoffCfg.InitialDelay
	for i := 0; i < backoffCfg.MaxRetries+1; i++ {
		err = t.f()
		if err == nil {
			t.Err = nil
			return
		}
		if i < backoffCfg.MaxRetries {
			time.Sleep(backoff)
			backoff = time.Duration(float64(backoff) * backoffCfg.Factor)
		}
	}
	t.Err = err
}

type BackoffConfig struct {
	// InitialDelay is the initial wait time before the first retry attempt. Recommended: 100 * time.Millisecond
	InitialDelay time.Duration
	// MaxRetries is the maximum number of retry attempts after the initial try. Recommended: 3
	MaxRetries int
	// Factor is the multiplier applied to the delay after each failed attempt. Recommended: 2.0
	Factor float64
}

// Pool manages a set of worker goroutines to process tasks.
type Pool struct {
	concurrency int
	tasksChan   chan *Task
	wg          sync.WaitGroup
	closeOnce   sync.Once
	closed      chan struct{}
	Backoff     BackoffConfig
}

// NewPool creates a new Pool with the given concurrency and backoff config.
func NewPool(concurrency int, backoff BackoffConfig) *Pool {
	return &Pool{
		concurrency: concurrency,
		tasksChan:   make(chan *Task),
		closed:      make(chan struct{}),
		Backoff:     backoff,
	}
}

// Start launches the worker goroutines.
func (p *Pool) Start() {
	for i := 0; i < p.concurrency; i++ {
		go p.work()
	}
}

// AddTask adds a new task to the pool and increments the wait group.
func (p *Pool) AddTask(task *Task) {
	p.wg.Add(1)
	p.tasksChan <- task
}

// Wait blocks until all tasks have been completed.
func (p *Pool) Wait() {
	p.wg.Wait()
}

// Close gracefully shuts down the pool, closing the task channel.
func (p *Pool) Close() {
	p.closeOnce.Do(func() {
		close(p.tasksChan)
		close(p.closed)
	})
}

func (p *Pool) work() {
	for task := range p.tasksChan {
		task.Run(&p.wg, p.Backoff)
	}
}
