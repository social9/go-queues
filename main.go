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
