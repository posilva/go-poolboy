package pool

import (
	"context"
	"math"
	"time"
)

// Pool defines the data to manage a pool of workers
type Pool struct {
	cancel    chan struct{}
	size      int
	available chan *worker
	initiated bool
}

// NewPool creates a new pool of workers
func NewPool(sz int) *Pool {
	pool := &Pool{
		size:      sz,
		available: make(chan *worker, sz),
		cancel:    make(chan struct{}, 1),
		initiated: false,
	}
	return pool
}

// NewPoolWithInit starts a pool with a defined size and initiates
func NewPoolWithInit(fn InitFun, sz int) (*Pool, error) {
	pool := NewPool(sz)
	err := pool.Init(fn)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

// Init the pool of workers
func (p *Pool) Init(fn InitFun) error {
	for index := 0; index < p.size; index++ {
		w := newWorker(fn, p.cancel)
		err := w.init()
		if err != nil {
			return err
		}
		p.checkin(w)
	}
	return nil
}

// Checkout recruits a worker to do a task
func (p *Pool) checkout(ctx context.Context) (*worker, error) {
	select {
	case w := <-p.available:
		return w, nil
	case <-ctx.Done():
		return nil, ErrorTimeout
	}
}

// Checkin releases the worker back to pool
func (p *Pool) checkin(w *worker) {
	if !w.canceled {
		p.available <- w
	}
}

// ExecuteWithTimeout executes a work function with a timeout in milliseconds
func (p *Pool) ExecuteWithTimeout(fn WorkFun, timeout uint64) (interface{}, error) {
	t := time.Duration(math.MaxInt64)
	if timeout > 0 {
		t = time.Duration(timeout) * time.Millisecond
	}
	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()
	w, err := p.checkout(ctx)
	if err != nil {
		return nil, err
	}
	defer p.checkin(w)
	return w.do(ctx, fn)
}

// Execute executes a work function with no timeout
func (p *Pool) Execute(fn WorkFun) (interface{}, error) {
	return p.ExecuteWithTimeout(fn, 0)
}

// Cancel the workers in the pool
func (p *Pool) Cancel() {
	close(p.cancel)
}
