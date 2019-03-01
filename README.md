# go-poolboy
Simple pool of workers to run tasks/functions in a bounded number of workers

[![Go Report Card](https://goreportcard.com/badge/github.com/posilva/go-poolboy)](https://goreportcard.com/report/github.com/posilva/go-poolboy)  [![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/posilva/go-poolboy/blob/master/LICENSE) [![Build Status](https://travis-ci.org/posilva/go-poolboy.svg?branch=master)](https://travis-ci.org/posilva/go-poolboy)[![codecov.io Code Coverage](https://img.shields.io/codecov/c/github/posilva/go-poolboy.svg)](https://codecov.io/github/posilva/go-poolboy?branch=master)

## Import
```bash
$ go get github.com/posilva/go-poolboy
```
...
## Usage
```go
package main

import (
	"fmt"
	poolboy "github.com/posilva/go-poolboy"
	"strings"
)

func main() {

	pool, err := poolboy.NewPoolWithInit(func() (interface{}, error) {
		// lets init the workers with some state
		state := "some state data"
		return state, nil
	}, 10)
	if err != nil {
		panic(fmt.Errorf("failed to create the pool: %v", err))
	}
	defer pool.Cancel()
	result, err := pool.ExecuteWithTimeout(func(s interface{}) (interface{}, error) {
		// get the state
		state := s.(string)
		// do something with state and return
		return strings.ToUpper(state), nil

	}, 2000)
	if err != nil {
		panic(err)
	}

	fmt.Printf("result upper case: %v \n", result)
}
```