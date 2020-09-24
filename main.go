package main

import (
	"fmt"
	queue "go-consumer/streams"
	"sync"
	"time"
)

// simulate processing a request for 2 seconds
func run(wg *sync.WaitGroup, job queue.SQSJob) {
	wait := job.ScheduledAt.Sub(time.Now())

	fmt.Println(job.ID, "will run after", wait)
	<-time.After(wait)

	fmt.Println("Running ", job.ID, "scheduled at", job.ScheduledAt)

	// Post business logic goes here
	time.Sleep(2 * time.Second)
	fmt.Println("Finished Job", job.ID)

	wg.Done()
}

func main() {

	limit, wait := 5, 0
	sqs := queue.NewSQS(limit, wait)

	sqs.Poll()
	sqs.Read(run)

	fmt.Println("Program exited.")
}
