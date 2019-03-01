package pool

import (
	"context"
	"errors"

	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSpawnWorkerOK(t *testing.T) {
	c := make(chan struct{})
	w := newWorker(func() (interface{}, error) {
		return nil, nil
	}, c)

	w.init()

	work := func(interface{}) (interface{}, error) {
		return "ok", nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	r, e := w.do(ctx, work)

	assert.Nil(t, e, "no error should be returned")
	assert.Equal(t, r, "ok", "Result should be equal")
}

func TestSpawnWorkerWithTimeout(t *testing.T) {
	c := make(chan struct{})
	w := newWorker(func() (interface{}, error) {
		return nil, nil
	}, c)

	w.init()

	work := func(interface{}) (interface{}, error) {
		time.Sleep(2 * time.Second)
		return "ok", nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	r, e := w.do(ctx, work)

	assert.NotNil(t, e, "error should be returned")
	assert.Nil(t, r, "no result to return")
	assert.EqualError(t, e, ErrorTimeout.Error(), "timeout error should be returned")
}
func TestSpawnWorkerNotInitiated(t *testing.T) {
	c := make(chan struct{})
	w := newWorker(func() (interface{}, error) {
		return nil, nil
	}, c)

	work := func(interface{}) (interface{}, error) {
		e0 := errors.New("err")
		return nil, e0
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	assert.Panics(t, func() {
		_, e := w.do(ctx, work)
		if e == nil {
			panic("shouln't be nil")
		}
	}, "function should panic")

}
func TestSpawnWorkerWithError(t *testing.T) {
	c := make(chan struct{})
	w := newWorker(func() (interface{}, error) {
		return nil, nil
	}, c)

	w.init()
	work := func(interface{}) (interface{}, error) {
		e0 := errors.New("err")
		return nil, e0
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	r, e := w.do(ctx, work)
	if e == nil {
		panic("shouln't be nil")
	}
	assert.NotNil(t, e, "error should be returned")
	assert.Nil(t, r, "no result to return")
	//assert.Equal(t, e.Error(), "err", "error message should be equal")
}
func TestSpawnWorkerWithPanic(t *testing.T) {
	c := make(chan struct{})
	w := newWorker(func() (interface{}, error) {
		return nil, nil
	}, c)
	w.init()

	work := func(interface{}) (interface{}, error) {
		panic("panic")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	r, e := w.do(ctx, work)

	assert.NotNil(t, e, "error should be returned")
	assert.Nil(t, r, "no result to return")
	assert.EqualError(t, e, "panic", "error message should be equal")
}

func TestSpawnWorkerWithWorkCanceled(t *testing.T) {
	c := make(chan struct{})
	w := newWorker(func() (interface{}, error) {
		return nil, nil
	}, c)
	w.init()

	work := func(interface{}) (interface{}, error) {
		close(c)
		time.Sleep(2 * time.Second)
		return "ok", nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	r, e := w.do(ctx, work)

	assert.NotNil(t, e, "error should be returned")
	assert.Nil(t, r, "no result to return")
	assert.EqualError(t, e, ErrorCanceled.Error(), "error message should be equal")
}
func TestSpawnWorkerWithoutWorkCanceled(t *testing.T) {
	c := make(chan struct{})
	w := newWorker(func() (interface{}, error) {
		return nil, nil
	}, c)
	w.init()
	time.Sleep(1 * time.Second)
	close(c)
	time.Sleep(2 * time.Second)
	assert.Equal(t, true, w.canceled, "worker should be canceled")
}
func TestSpawnWorkerToMuchWorkCanceled(t *testing.T) {
	c := make(chan struct{})
	w := newWorker(func() (interface{}, error) {
		return nil, nil
	}, c)
	w.init()

	work := func(interface{}) (interface{}, error) {
		time.Sleep(5 * time.Second)
		return "ok", nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	go w.do(ctx, work)
	// the in worker channel only have place for 1 message this will block and trigger the cancel
	// this will block when trying to push more work
	w.do(ctx, work)
	close(c)
	time.Sleep(2 * time.Second)
	assert.Equal(t, true, w.canceled, "worker should be canceled")
}
