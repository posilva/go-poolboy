package pool

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestPoolCheckoutWithTimeout(t *testing.T) {
	p, err := NewPoolWithInit(func() (interface{}, error) {
		return nil, nil
	}, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_, err = p.checkout(ctx)

	ctx1, cancel1 := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel1()
	_, err = p.checkout(ctx1)

	assert.Error(t, err, "failed to checkout")

}
func TestPoolInitWithError(t *testing.T) {
	p, e := NewPoolWithInit(func() (interface{}, error) {
		return nil, errors.New("err")
	}, 10)
	assert.NotNil(t, e, "error should be returned")
	assert.Nil(t, p, "result should be nil")
	assert.Equal(t, e.Error(), "err", "return error")
}

func TestPoolSimple(t *testing.T) {
	p, err := NewPoolWithInit(func() (interface{}, error) {
		return nil, nil
	}, 10)
	assert.Nil(t, err, "no init error should be returned")
	var wg sync.WaitGroup
	wg.Add(10)
	for index := 0; index < 10; index++ {
		go func() {
			defer wg.Done()
			r, e := p.Execute(func(*Worker) (interface{}, error) {
				return "ok", nil
			}, 5000)
			assert.Nil(t, e, "no error should be returned")
			assert.Equal(t, r, "ok", "result should be equal")
		}()
	}
	wg.Wait()
}

func TestPoolMoreWorkThanWorkers(t *testing.T) {
	p, err := NewPoolWithInit(func() (interface{}, error) {
		return nil, nil
	}, 10)
	assert.Nil(t, err, "no init error should be returned")
	var wg sync.WaitGroup
	wg.Add(100)
	for index := 0; index < 100; index++ {
		go func() {
			defer wg.Done()
			r, e := p.Execute(func(*Worker) (interface{}, error) {
				return "ok", nil
			}, 5000)
			assert.Nil(t, e, "no error should be returned")
			assert.Equal(t, r, "ok", "result should be equal")
		}()
	}
	wg.Wait()
}

func TestPoolMoreWorkThanWorkersWithTimeouts(t *testing.T) {
	p, err := NewPoolWithInit(func() (interface{}, error) {
		return nil, nil
	}, 10)
	assert.Nil(t, err, "no init error should be returned")
	var wg sync.WaitGroup
	wg.Add(5)
	for index := 0; index < 5; index++ {
		go func() {
			defer wg.Done()
			r, e := p.Execute(func(*Worker) (interface{}, error) {
				time.Sleep(2 * time.Second)
				return "ok", nil
			}, 1000)
			assert.NotNil(t, e, " error should be returned")
			assert.Nil(t, r, " no result")
			assert.Equal(t, e.Error(), "timeout", "error was return")
		}()
	}
	wg.Wait()
}
