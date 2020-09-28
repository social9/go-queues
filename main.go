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
	env := config.Env()

	// Instantiate the queue with service connection
	queue, _ := sqs.NewSQS(sqs.SQSConfig{
		Verbosity: 0,

		// aws config
		AWSRegion:  env.AWSRegion,
		MaxRetries: 10,

		// aws creds - if provided, env is temporarily updated. Or you can add to env yourself
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
