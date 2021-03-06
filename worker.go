package pool

import (
	"context"
	"errors"
	"fmt"
)

var (
	// ErrorTimeout represents the timeout error
	ErrorTimeout = errors.New("timeout")
	// ErrorCanceled represents the cancelation error
	ErrorCanceled = errors.New("canceled")
)

// InitFun defines a function to initiate the worker
type InitFun func() (interface{}, error)

//WorkFun represents a unit of work to execute by the worker
type WorkFun func(interface{}) (interface{}, error)

// worker represents a worker enable to execute
type worker struct {
	state     interface{}      // save the state to be used later
	in        chan WorkFun     // to receive work
	out       chan interface{} // to return work results
	err       chan interface{} // to return errors
	cancel    <-chan struct{}  // allow to receive cancel
	canceled  bool             // this worker was canceled
	timeout   chan bool        // notify timeouts
	initFn    InitFun          // the function that enable to start the work
	initiated bool             // flag to mark that worker was initiated
}

// newWorker creates a worker
func newWorker(fn InitFun, cancel <-chan struct{}) *worker {
	return &worker{
		in:        make(chan WorkFun),
		out:       make(chan interface{}),
		err:       make(chan interface{}, 1),
		cancel:    cancel,
		timeout:   make(chan bool, 1),
		initFn:    fn,
		initiated: false,
	}
}

func (w *worker) init() error {
	s, err := w.initFn()
	if err != nil {
		return err
	}
	w.state = s
	go w.start()
	w.initiated = true
	return nil
}

func (w *worker) start() {
	for {
		select {
		case <-w.cancel:
			return
		default:
			if !w.run() || w.canceled {
				return
			}
		}
	}
}

func (w *worker) run() bool {
	defer func() {
		if r := recover(); r != nil {
			w.err <- r
		}
	}()
	select {
	case <-w.cancel:
		w.canceled = true
		return false
	case workfn := <-w.in:
		r, err := workfn(w.state)
		if err != nil {
			w.err <- err
			return true
		}
		select {
		case <-w.timeout:
			w.err <- ErrorTimeout
		default:
			w.out <- r
		}
	}
	return true
}

func (w *worker) do(ctx context.Context, work WorkFun) (interface{}, error) {
	if !w.initiated {
		panic("worker was not initiated")
	}
	select {
	case w.in <- work:
	case <-w.cancel:
		w.canceled = true
		return nil, ErrorCanceled
	}
	select {
	case <-w.cancel:
		w.canceled = true
		return nil, ErrorCanceled
	case result := <-w.out:
		return result, nil
	case e := <-w.err:
		return nil, fmt.Errorf("%v", e)
	case <-ctx.Done():
		w.timeout <- true
		return nil, ErrorTimeout
	}
}
