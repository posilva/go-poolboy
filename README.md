# go-poolboy
Simple pool of workers to run tasks/functions in a bounded number of workers

[![Go Report Card](https://goreportcard.com/badge/github.com/posilva/go-poolboy)](https://goreportcard.com/report/github.com/posilva/go-poolboy)  [![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/posilva/go-poolboy/blob/master/LICENSE) [![Build Status](https://travis-ci.org/posilva/go-poolboy.svg?branch=master)](https://travis-ci.org/posilva/go-poolboy)[![codecov.io Code Coverage](https://img.shields.io/codecov/c/github/posilva/go-poolboy?maxAge=2592000)](https://codecov.io/github/posilva/go-poolboy?branch=master)

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
	"github.com/posilva/go-mfa/session"
	"log"
)

func main() {
	accountsIds := []string{
		"123456789012",
	}

	roleName := "MyAdminRole"
	p := session.Params{
		Profile:      "default",
		SerialDevice: "arn:aws:iam::098765432109:mfa/posilva",
		MFAToken:     session.AskMFA(),
	}

	mfaSession := session.NewMFASession(p)
	sessionsMap, err := mfaSession.AssumeBulk(roleName, accountsIds)
	if err != nil {
		log.Fatal(err)
    }
    
	useSession(sessionsMap, "123456789012", "eu-west-2")
	printAccountSessions(sessionsMap)
}

// printAccountSessions uses the ForEach function to run a given function in
// all the cached AWSSessions
func printAccountSessions(mfaSession *session.MFASession) {
	mfaSession.ForEachSession(func(a string, r string, s *session.AWSSession) error {
		c, _ := s.Get().Config.Credentials.Get()
		fmt.Printf("%v - %v - %v \n", a, r, c)
		return nil
	})

}

// useSession is an example of requiring a cached session and execute normal
// AWS SDK
func useSession(mfaSession *session.MFASession, account string, region string) {
	s, err := mfaSession.Get(account, region)
	if err != nil {
		log.Fatal(err)
	}
	svc := s3.New(s.Get())
	input := &s3.ListBucketsInput{}
	output, err := svc.ListBuckets(input)
	if err != nil {
		log.Fatal(err)
	}
	for _, b := range output.Buckets {
		fmt.Println(*b.Name)
	}
}

```