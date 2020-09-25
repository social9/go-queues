package streams

import (
	"fmt"
	"sync"
	"time"
)

// SQSJob The schema representing messages in the queue
type SQSJob struct {
	ID          int
	ScheduledAt time.Time
}

// SQS Wrapper for SQS methods
type SQS struct {
	jobs         chan SQSJob
	Limit        int
	WaitSeconds  int
	VerboseLevel int
}

// SQSOps An interface for SQS operations
type SQSOps interface {
	Poll()
	Read(func(wg *sync.WaitGroup, job SQSJob))
}

// NewSQS Initialise a SQS instance
func NewSQS(limit, waitSeconds int) SQSOps {
	return &SQS{make(chan SQSJob, limit), limit, waitSeconds, 0}
}

// Poll Poll for messages in the SQS
func (s *SQS) Poll() {

	s.jobs = make(chan SQSJob, s.Limit)

	// Add business logic to poll from SQS here
	for i := 1; i <= s.Limit; i++ {
		s.jobs <- SQSJob{
			i,
			time.Now().Add(time.Duration(1*i) * time.Second),
		}
	}
	close(s.jobs)
}

// Read Read from the poll and spawn workers for received messages in a worker group.
//
// The caller must write to `done` channel in the `run` fn to mark the successful completion of a request.
func (s *SQS) Read(run func(wg *sync.WaitGroup, job SQSJob)) {
	wg := sync.WaitGroup{}

	fmt.Println("Listening")
	for job := range s.jobs {

		fmt.Printf("Adding job %v\n", job)
		wg.Add(1)
		go run(&wg, job)
	}

	wg.Wait()
}
