package streams

import (
	"go-consumer/config"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// SQS Wrapper for SQS methods
type SQS struct {
	Limit             int64
	WaitSeconds       int64
	VisibilityTimeout int64
	Verbosity         int
	svc               *sqs.SQS
}

// SQSOps An interface for SQS operations
type SQSOps interface {
	Poll(handler func(wg *sync.WaitGroup, msg *sqs.Message))
	Delete(msg *sqs.Message) error
}

// NewSQS Initialise a SQS instance
func NewSQS() SQSOps {
	env := config.Env()

	creds := credentials.NewEnvCredentials()
	awsConfig := aws.NewConfig().WithRegion(env.AWSRegion).WithMaxRetries(10).WithCredentials(creds)

	newSession := session.Must(session.NewSession(awsConfig))
	svc := sqs.New(newSession)

	limit := env.SQSLimit
	waitTime := env.SQSWaitTime
	visibilityTimeout := env.SQSVisibilityTimeout
	verbosity := 0

	return &SQS{limit, waitTime, visibilityTimeout, verbosity, svc}
}

// Poll Poll for messages in the SQS
func (s *SQS) Poll(handler func(wg *sync.WaitGroup, msg *sqs.Message)) {
	env := config.Env()
	wg := sync.WaitGroup{}

	for {
		result, err := s.svc.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            &env.SQSURL,
			MaxNumberOfMessages: &s.Limit,
			WaitTimeSeconds:     &s.WaitSeconds,
			VisibilityTimeout:   &s.VisibilityTimeout,
		})

		if err != nil {
			log.Fatal(err)
			break
		}

		for _, msg := range result.Messages {
			wg.Add(1)
			go handler(&wg, msg)
		}
	}

	wg.Wait()
}

// Delete a SQS message from the queue
func (s *SQS) Delete(msg *sqs.Message) error {

	env := config.Env()
	_, err := s.svc.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      &env.SQSURL,
		ReceiptHandle: msg.ReceiptHandle,
	})

	return err
}
