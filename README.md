# go-queues (In Development)

A generic producer-consumer service with pluggable queues written in Go

It is designed to be inherently scalable, apply concurrent processing using goroutines and has pluggable queue sources such as `SQS`, `Kafka`, etc.

## Quick Start

```go

package main

import (
	"log"
	"sync"
	"time"

	"github.com/social9/go-queues/config"
	"github.com/social9/go-queues/streams/sqs"

	awsSqs "github.com/aws/aws-sdk-go/service/sqs"
)

func main() {

	// Instantiate a SQS instance
	queue, _ := sqs.NewSQS(sqs.SQSConfig{
		Verbosity: 0,

		// aws config
		AWSRegion:  "us-east-2",
		MaxRetries: 10,

		// aws creds - Env uis updated when provided
		// Or you can set the following keys your self `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`
		// Read more about the credential chain here, https://docs.aws.amazon.com/sdk-for-go/api/aws/credentials/
		AWSKey:    "<YOUR AWS ACCESS KEY>",
		AWSSecret: "<YOUR AWS ACCESS SECRET>",

		// sqs config
		URL:               "https://sqs.us-east-2.amazonaws.com/1234567/MyQueue.fifo",
		BatchSize:         10,
		VisibilityTimeout: 120,
		WaitSeconds:       5,

		// run config
		RunInterval: 20,
		RunOnce:     true,
	})

	// simulate processing a request for 2 seconds
	handler := func(wg *sync.WaitGroup, msg *awsSqs.Message) {

		log.Println("Waiting:", *msg.MessageId)
		wait := time.Duration(1) * time.Second
		<-time.After(wait)

		log.Println("Processing:", *msg.MessageId, *msg.Body)
		time.Sleep(2 * time.Second)
		log.Println("Finished:", *msg.MessageId)

		err := queue.Delete(msg)
		log.Println("Delete Error:", err)

		wg.Done()
	}

	// Poll from the SQS queue
	queue.Poll(handler)
}

```

## Setup

- Clone the repo `git clone https://github.com/social9/go-consumer`
- Configurable environment variables
  | Parameter | Description | Required | Default | Allowed |
  |-----------|-------------|----------|---------|---------|
  |`AWS_ACCESS_KEY_ID`|AWS Access Key|yes|`""`|`string`|
  |`AWS_SECRET_ACCESS_KEY`|AWS Access Secret|yes|`""`|`string`|
  |`AWS_REGION`|The AWS region to establish service connection|`us-east-2`|A valid AWS region|
  |`SQS_URL`|The SQS endpoint to poll|""|`string`|
  |`SQS_BATCH_SIZE`|The maximum number of messages to receive per request|`10`|`1-10`|
  |`SQS_WAIT_TIME`|The maximum polling wait time in seconds|`20`|`0-20`|
  |`SQS_VISIBILITY_TIMEOUT`|Visiblity timeout for a message after received in seconds|`20`|`0 - 12*60*60`|
  |`RUN_ONCE`|Run the service one-time or in intervals|`true`|`boolean`|
  |`RUN_INTERVAL`|Run the service in the defined interval. Works only when `RUN_ONCE` is set to `true`|`10`|`>0`|

> Note: You can also load the env values from a file named `.env` stored in the root path

## Contribution Guidelines

_To be added soon_

## License

MIT
