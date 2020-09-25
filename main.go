package main

import (
	"fmt"
	"go-consumer/config"
	"go-consumer/streams"
	"sync"
	"time"
)

// simulate processing a request for 2 seconds
func handler(wg *sync.WaitGroup, job streams.SQSJob) {
	wait := job.ScheduledAt.Sub(time.Now())

	fmt.Println(job.ID, "will run after", wait)
	<-time.After(wait)

	fmt.Println("Running", job)

	// Post business logic goes here
	time.Sleep(2 * time.Second)
	fmt.Println("Finished ", job)

	wg.Done()
}

func main() {

	env := config.Env()

	sqs := streams.NewSQS(env.SQSLimit, env.SQSWaitTime)

	run := func(wg *sync.WaitGroup) {
		start := time.Now()
		sqs.Poll()
		sqs.Read(handler)
		fmt.Println("\nBatch took: ", time.Now().Sub(start))

		wg.Done()
	}

	wg := sync.WaitGroup{}
	i := 0
	for {
		i++
		fmt.Println("\nBatch", i, "\n")

		wg.Add(1)
		go run(&wg)

		if !env.RunOnce {
			<-time.After(time.Duration(env.RunInterval) * time.Second)
		} else {
			break
		}
	}

	wg.Wait()

}
