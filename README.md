# go-queues (In Development)

A generic producer-consumer service with pluggable queues written in Go

It is designed to be inherently scalable, apply concurrent processing using goroutines and has pluggable queue sources such as `SQS`, `Kafka`, etc.

## Setup

- Clone the repo `git clone https://github.com/cnp96/go-consumer`
- Configurable environment variables
  - `AWS_ACCESS_KEY_ID` : AWS Access Key, **Required**
  - `AWS_SECRET_ACCESS_KEY` : AWS Access Secret, **Required**
  - `AWS_REGION` : The AWS region to establish service connection, _(default **us-east-2**)_
  - `SQS_URL` : The SQS endpoint to poll, _(default **""**)_
  - `SQS_BATCH_SIZE` : The maximum number of messages to receive per request, _(default **10**)_, _(accepted **1-10)**_
  - `SQS_WAIT_TIME` : The maximum polling wait time in seconds, _(default **20 seconds**)_, _(accepted **0-20)**_
  - `SQS_VISIBILITY_TIMEOUT` : Visiblity timout for a message after received in seconds, _(default **20 seconds**)_
  - `RUN_ONCE` : Run the service one-time or in intervals, _(default **true**)_
  - `RUN_INTERVAL` : Run the service in the defined interval, _(default **10 seconds**)_. Works only when `RUN_ONCE` is set to `true`

> Note: You can also load the env values from a file named `.env` stored in the root path

## Quick Start

```go

package main

import (
	"log"
	"sync"
	"time"

	"go-queues/config"
	"go-queues/streams/sqs"

	awsSqs "github.com/aws/aws-sdk-go/service/sqs"
)

func main() {
	env := config.Env()

	// Instantiate the queue with service connection
	queue, _ := sqs.NewSQS(sqs.SQSConfig{
		Verbosity: 0,

		// aws config
		AWSRegion:  env.AWSRegion,
		MaxRetries: 10,

		// aws creds - Env uis updated when provided
		// Or you can set the following keys your self
		// AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY
		// Read more about the credential chain [here](https://docs.aws.amazon.com/sdk-for-go/api/aws/credentials/).
		AWSKey:    env.AWSKey,
		AWSSecret: env.AWSSecret,

		// sqs config
		URL:               env.SQSURL,
		BatchSize:         env.SQSBatchSize,
		VisibilityTimeout: 120,
		WaitSeconds:       5,

		// run config
		RunInterval: 20,
		RunOnce:     env.RunOnce,
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

## Contribution Guidelines

_To be added soon_

## License

MIT
