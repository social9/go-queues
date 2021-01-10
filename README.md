<p align="center">
  <a href="https://pkg.go.dev/github.com/social9/go-queues/?tab=doc">
    <img src="https://img.shields.io/badge/%F0%9F%93%9A%20godoc-pkg-00ACD7.svg?color=00ACD7&style=flat">
  </a>
  <a href="https://goreportcard.com/report/github.com/social9/go-queues">
    <img src="https://img.shields.io/badge/%F0%9F%93%9D%20goreport-A%2B-75C46B">
  </a>
  <a href="https://gocover.io/github.com/social9/go-queues">
    <img src="https://img.shields.io/badge/coverage-0%25-orange">
  </a>
</p>

# go-queues

A generic producer-consumer service with pluggable queues written in Go

It is designed to be inherently scalable, apply concurrent processing using goroutines and has pluggable queue sources such as `SQS`, `Kafka`, etc.

## Quick Start

```go
package main

import (
	"log"
	"strconv"
	"time"

	"github.com/social9/go-queues/streams/sqs"

	"github.com/aws/aws-sdk-go/aws"
	awsSqs "github.com/aws/aws-sdk-go/service/sqs"
)

func main() {
	// Instantiate the queue with service connection
	queue, _ := sqs.NewSQS(sqs.Config{
		// aws config
		AWSRegion:  "us-east-2",
		MaxRetries: 10,

		// aws creds - if provided, auto added to env. Or you can add manually as well
		AWSKey:    "<AWS Access Key>",
		AWSSecret: "<AWS Secret>",

		// sqs config
		URL:               "https://sqs.us-east-2.amazonaws.com/..../MyQueue.fifo",
		BatchSize:         10,  // fetch 10 messages per batch
		VisibilityTimeout: 120, // hide for 2 minutes from other consumers 
		WaitSeconds:       5,   // poll for 5 seconds per batch

		// misc config
		RunInterval: 20,    // poll every 20 seconds
		RunOnce:     false, // if set to true, polled only once
		MaxHandlers: 10,    // maximum number of messages to process at a time
		BusyTimeout: 30,    // wait for 30 seconds before rechecking if handlers are freed (when max handlers reached)
	})

	// Simlulate sending the messages in batch
	queue.Enqueue(getMessagesToEnque())

	// simulate processing a request for 2 seconds
	queue.RegisterPollHandler(func(msg *awsSqs.Message) {
		log.Println("Wait 2 seconds for:", *msg.MessageId)
		wait := time.Duration(2) * time.Second
		<-time.After(wait)

		log.Println("Processing:", *msg.MessageId, *msg.Body)

		// Simulate processing time as 10 seconds
		time.Sleep(10 * time.Second)
		log.Println("Finished:", *msg.MessageId)

		// Send back to the queue
		queue.ChangeVisibilityTimeout(msg, 0)
	})

	// Poll from the SQS queue
	queue.Poll()

}

func getMessagesToEnque() []*awsSqs.SendMessageBatchRequestEntry {
	msgs := []string{"Test message 1-1", "Test Message 2-1", "Test Message 3-1"}

	var msgBatch []*awsSqs.SendMessageBatchRequestEntry
	for i := 0; i < len(msgs); i++ {
		message := &awsSqs.SendMessageBatchRequestEntry{
			Id:                     aws.String(`test_` + strconv.Itoa(i)),
			MessageBody:            aws.String(msgs[i]),
			MessageDeduplicationId: aws.String(`dedup_` + strconv.Itoa(i)),
			MessageGroupId:         aws.String("test_group"),
		}
		msgBatch = append(msgBatch, message)
	}

	return msgBatch
}

```

## Setup

- Clone the repo `git clone https://github.com/social9/go-queues`
- Configurable environment variables
  | Parameter              | Description  | Default | Allowed |
  |------------------------|--------------|---------|---------|
  |`AWS_ACCESS_KEY_ID`     |AWS Access Key|`""`|`string`|
  |`AWS_SECRET_ACCESS_KEY` |AWS Access Secret|`""`|`string`|
  |`AWS_REGION`            |The AWS region to establish service connection|`us-east-2`|`A valid AWS region`|
  |`SQS_URL`               |The SQS endpoint to poll|`""`|`string`|
  |`SQS_BATCH_SIZE`        |The maximum number of messages to receive per request|`10`|`1-10`|
  |`SQS_WAIT_TIME`         |The maximum polling wait time in seconds|`20`|`0-20`|
  |`SQS_VISIBILITY_TIMEOUT`|Visiblity timeout for a message after received in seconds|`20`|`0 - 12*60*60`|
  |`RUN_ONCE`              |Run the service one-time or in intervals|`true`|`boolean`|
  |`RUN_INTERVAL`          |Run the service in the defined interval. Works only when `RUN_ONCE` is set to `true`|`10`|`>0`|

> Note: You can also load the env values from a file named `.env` stored in the root path

## Contribution Guidelines

- Fork this repo to your GitHub account
- You can either create an issue or pick from the existing and seek maintainers' attention before development
- Your _Pull Request_ branch must be rebased with the `dev` branch i.e. have a linear history
- One or more maintainers will review your PR once associated to an issue.

> Do append the issue ID in the pull request title e.g. **Implemented a functionality closes #20** where **20** is the issue number

## License

MIT
