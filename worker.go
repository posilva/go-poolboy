package pool

import (
	"context"
	"errors"
	"fmt"
)

var (
	// ErrorTimeout represents the timeout error
	ErrorTimeout = errors.New("timeout")
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
	timeout   chan bool        // notify timeouts
	initFn    InitFun          // the function that enable to start the work
	initiated bool             // flag to mark that worker was initiated
}

// newWorker creates a worker
func newWorker(fn InitFun) *worker {
	return &worker{
		in:        make(chan WorkFun),
		out:       make(chan interface{}),
		err:       make(chan interface{}, 1),
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
		w.run()
	}
}

func (w *worker) run() {
	defer func() {
		if r := recover(); r != nil {
			w.err <- r
		}
	}()
	select {
	case workfn := <-w.in:
		r, err := workfn(w.state)
		if err != nil {
			w.err <- err
			return
		}
		select {
		case <-w.timeout:
			w.err <- ErrorTimeout
		default:
			w.out <- r
		}
	}
}

func (w *worker) do(ctx context.Context, work WorkFun) (interface{}, error) {
	if !w.initiated {
		panic("worker was not initiated")
	}
	w.in <- work
	select {
	case result := <-w.out:
		return result, nil
	case e := <-w.err:
		return nil, fmt.Errorf("%v", e)
	case <-ctx.Done():
		w.timeout <- true
		return nil, ErrorTimeout
	}
}
