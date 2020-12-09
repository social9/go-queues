# go-queues (In Development)

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
		Verbosity: 0,

		// aws config
		AWSRegion:  "us-east-2",
		MaxRetries: 10,

		// aws creds - if provided, env is temporarily updated. Or you can add to env yourself
		AWSKey:    "...",
		AWSSecret: "...",

		// sqs config
		URL:               "https://sqs.us-east-2.amazonaws.com/1234567/MyQueue.fifo",
		BatchSize:         10,
		VisibilityTimeout: 120,
		WaitSeconds:       5,

		// misc config
		RunInterval: 20,
		RunOnce:     env.RunOnce,
		MaxHandlers: 10,
		BusyTimeout: 30,
	})

	// Simlulate sending the messages in batch
	queue.Enqueue(EnqueueMsgs())

	// simulate processing a request for 2 seconds
	queue.RegisterPollHandler(func(msg *awsSqs.Message) {
		log.Println("Waiting:", *msg.MessageId)
		wait := time.Duration(2) * time.Second
		<-time.After(wait)

		log.Println("Processing:", *msg.MessageId, *msg.Body)

		time.Sleep(60 * time.Second) // Processing time 60 seconds
		log.Println("Finished:", *msg.MessageId)

		queue.ChangeVisibilityTimeout(msg, 0) // Shall comeback to the queue

		// err := queue.Delete(msg)
		// log.Println("Delete Error:", err)
	},
	)
	time.Sleep(60 * time.Second) // wait, go to console and see if there are some messages visible, Hit "Poll for messages"
	// Poll from the SQS queue
	queue.Poll()

}

// EnqueueMsgs - Simlulate sending the messages in batch
func EnqueueMsgs() []*awsSqs.SendMessageBatchRequestEntry {
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

- Clone the repo `git clone https://github.com/social9/go-consumer`
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
- You can either create an issue or pick from the existing and seek maintainers' attention before developement
- Your _Pull Request_ branch must be rebased with the `dev` branch i.e. have a linear history
- One or more maintainers will review your PR once associated to an issue.

> Do append the issue ID in the pull request title e.g. **Implemented a functionality closes #20** where **20** is the issue number

## License

MIT
