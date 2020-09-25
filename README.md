# go-consumer (In Development)

A generic consumer service written in Go

It is designed to be inherently scalable, apply concurrent processing using goroutines and has pluggable stream sources such as `SQS`, `Kafka`, etc.

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

## Running

```go
package main

import (
	"fmt"
	"go-consumer/streams"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/service/sqs"
)

func main() {
	// simulate processing a request for 2 seconds and then delete it
	handler := func(wg *sync.WaitGroup, msg *sqs.Message) {
		fmt.Println("Waiting:", *msg.MessageId)
		wait := time.Duration(1) * time.Second
		<-time.After(wait)

		fmt.Println("Processing:", *msg.MessageId, *msg.Body)

		time.Sleep(2 * time.Second)
		fmt.Println("Finished:", *msg.MessageId)

		err := queue.Delete(msg)
		fmt.Println("Delete Error:", err)

		wg.Done()
	}

  // Instantiate the queue with service connection
	queue := streams.NewSQS()

	// Poll from the SQS queue
	queue.Poll(handler)
}

```

## Contribution Guidelines

_To be added soon_

## License

MIT
