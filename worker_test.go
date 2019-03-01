package pool

import (
	"context"
	"errors"

	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSpawnWorkerOK(t *testing.T) {
	w := NewWorker(func() (interface{}, error) {
		return nil, nil
	})
	w.init()

	work := func(*Worker) (interface{}, error) {
		return "ok", nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	r, e := w.do(ctx, work)

	assert.Nil(t, e, "no error should be returned")
	assert.Equal(t, r, "ok", "Result should be equal")
}

func TestSpawnWorkerWithTimeout(t *testing.T) {
	w := NewWorker(func() (interface{}, error) {
		return nil, nil
	})
	w.init()

	work := func(*Worker) (interface{}, error) {
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
	w := NewWorker(func() (interface{}, error) {
		return nil, nil
	})
	work := func(*Worker) (interface{}, error) {
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
	w := NewWorker(func() (interface{}, error) {
		return nil, nil
	})
	w.init()
	work := func(*Worker) (interface{}, error) {
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
	w := NewWorker(func() (interface{}, error) {
		return nil, nil
	})
	w.init()
	work := func(*Worker) (interface{}, error) {
		panic("panic")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	r, e := w.do(ctx, work)

	assert.NotNil(t, e, "error should be returned")
	assert.Nil(t, r, "no result to return")
	assert.EqualError(t, e, "panic", "error message should be equal")
}
