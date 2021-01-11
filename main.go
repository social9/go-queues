package main

import (
	"log"
	"strconv"
	"time"

	"github.com/social9/go-queues/config"
	"github.com/social9/go-queues/streams/sqs"

	"github.com/aws/aws-sdk-go/aws"
	awsSqs "github.com/aws/aws-sdk-go/service/sqs"
)

func main() {
	env := config.Env()

	// Instantiate the queue with service connection
	queue, _ := sqs.NewSQS(sqs.Config{
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
		WaitSeconds:       20,

		// misc config
		RunInterval: 20,
		RunOnce:     env.RunOnce,
		MaxHandlers: 10,
		BusyTimeout: 30,
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
		queue.Delete(msg)
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
