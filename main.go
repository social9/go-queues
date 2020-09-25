package main

import (
	"fmt"
	"go-consumer/streams"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/service/sqs"
)

func main() {
	// Instantiate the queue with service connection
	queue := streams.NewSQS()

	// simulate processing a request for 2 seconds
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

	// Poll from the SQS queue
	queue.Poll(handler)
}
