package main

import (
	"log"
	"strconv"
	"sync"
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

		// misc config
		RunInterval: 20,
		RunOnce:     env.RunOnce,
		MaxHandlers: 10,
	})

	// Simlulate sending the messages in batch
	queue.Enque(SendMessageBatch())

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

// SendMessageBatch - Simlulate sending the messages in batch
func SendMessageBatch() []*awsSqs.SendMessageBatchRequestEntry {
	msgs := []string{"Test message 1", "Test Message 2", "Test Message 3"}

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
