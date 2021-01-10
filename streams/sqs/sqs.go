package sqs

import (
	"errors"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "(gq-sqs)", log.Lshortfile)
}

// Config Wrapper for Config methods
type Config struct {
	AWSKey    string
	AWSSecret string
	AWSRegion string

	// Poll from this SQS URL
	URL string

	// Maximum number of time to attempt AWS service connection
	MaxRetries int

	// Maximum number of messages to retrieve per batch
	BatchSize int64

	// The maximum poll time (0 <= 20)
	WaitSeconds int64

	// Once a message is received by a consumer, the maximum time in seconds till others can see this
	VisibilityTimeout int64

	// Poll only once and exit
	RunOnce bool

	// Poll every X seconds defined by this value
	RunInterval int

	// Maximum number of handlers to spawn for batch processing
	MaxHandlers int

	// BusyTimeout in seconds
	BusyTimeout int

	svc          *sqs.SQS
	handlerCount int
	pollHandler  func(msg *sqs.Message)
}

// SQS An interface for SQS operations
type SQS interface {
	Poll()
	Delete(msg *sqs.Message) error
	Enqueue(msgBatch []*sqs.SendMessageBatchRequestEntry) error
	RegisterPollHandler(pollHandler func(msg *sqs.Message))
	ChangeVisibilityTimeout(msg *sqs.Message, seconds int64) bool
}

// NewSQS Instantiate a SQS instance
func NewSQS(opts Config) (SQS, error) {
	// Validate parameters
	validateErr := validateOpts(opts)
	if validateErr != nil {
		logger.Println(validateErr)
		return nil, validateErr
	}

	// Validate creds
	if opts.AWSKey != "" {
		os.Setenv("AWS_ACCESS_KEY_ID", opts.AWSKey)
	}
	if opts.AWSSecret != "" {
		os.Setenv("AWS_SECRET_ACCESS_KEY", opts.AWSSecret)
	}

	creds := credentials.NewEnvCredentials()
	if _, err := creds.Get(); err != nil {
		logger.Println("AWS Credential error", err)
		return nil, errors.New("Invalid AWS credentials. Please make sure that `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` is present in the env")
	}

	// Create AWS Config
	awsConfig := aws.NewConfig().WithRegion(opts.AWSRegion).WithMaxRetries(opts.MaxRetries).WithCredentials(creds)
	if awsConfig == nil {
		logger.Println("Invalid AWS Config")
		return nil, errors.New("Something is wrong with your AWS config parameters")
	}

	// Establish a session
	newSession := session.Must(session.NewSession(awsConfig))
	if newSession == nil {
		logger.Println("Unable to create session")
		return nil, errors.New("Unable to create session")
	}

	// Create a service connection
	svc := sqs.New(newSession)
	if svc == nil {
		logger.Println("Unable to connect to SQS")
		return nil, errors.New("Unable to create a service connection with AWS SQS")
	}

	logger.Println("Fetching queue attributes")
	if _, err := svc.GetQueueAttributes(&sqs.GetQueueAttributesInput{
		QueueUrl: &opts.URL,
	}); err != nil {
		logger.Println("Unable to fetch queue attributes", err)
		return nil, errors.New("Unable to get queue attributes")
	}
	logger.Println("Connected to Queue")

	opts.svc = svc
	return &opts, nil
}

// Poll for messages in the queue
func (s *Config) Poll() {
	if s.svc == nil {
		logger.Fatalln("No service connection")
	}

	wg := sync.WaitGroup{}
	batch := 0

	for {
		batch++
		childLogger := log.New(os.Stdout, "(gq-sqs) batch-"+strconv.Itoa(batch), log.Lshortfile)

		childLogger.Println("Start receiving messages")

		maxMsgs := s.BatchSize
		// Is running at capacity?
		if s.MaxHandlers > 0 && s.handlerCount >= s.MaxHandlers {
			childLogger.Printf("Running at full capacity with %d handlers", s.handlerCount)

			// Since all handlers are busy, let's wait for BusyTimeout seconds
			childLogger.Printf("Going to wait state for %d seconds", s.BusyTimeout)
			<-time.After(time.Duration(s.BusyTimeout) * time.Second)
			continue
		} else {
			maxMsgs = int64(s.MaxHandlers - s.handlerCount)
			childLogger.Printf("Can accept a maximum of %d messages", maxMsgs)
		}

		result, err := s.svc.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            &s.URL,
			MaxNumberOfMessages: &maxMsgs,
			WaitTimeSeconds:     &s.WaitSeconds,
			VisibilityTimeout:   &s.VisibilityTimeout,
		})

		// Retrieve error?
		if err != nil {
			childLogger.Println("ReceiveMessageError:", err)
			break
		}

		// Message log
		if len(result.Messages) == 0 {
			childLogger.Println("Queue is empty")
		} else {
			childLogger.Printf("Fetched %d messages", len(result.Messages))
		}

		// Process messages
		for _, msg := range result.Messages {
			if s.pollHandler == nil {
				childLogger.Println("No Poll handler registered. Register a handler for custom handling")
			} else {
				s.handlerCount++
				wg.Add(1)

				go func(w *sync.WaitGroup, inst *Config) {
					inst.pollHandler(msg)
					w.Done()
					inst.handlerCount--
				}(&wg, s)
			}

			childLogger.Printf("Spawned handler for %s", *msg.MessageId)
		}

		if s.RunOnce == true {
			childLogger.Println(`Exiting since confugured to run once`)
			break
		} else {
			childLogger.Printf("Waiting for %d seconds before polling for next batch", s.RunInterval)
			<-time.After(time.Duration(s.RunInterval) * time.Second)
		}

		childLogger.Println("Finished polling")
	}

	wg.Wait()
}

// Enqueue messages to SQS
func (s *Config) Enqueue(msgBatch []*sqs.SendMessageBatchRequestEntry) error {
	if s.svc == nil {
		logger.Fatal("No service connection")
	}

	logger.Printf(`%d messages are processing`, len(msgBatch))

	result, err := s.svc.SendMessageBatch(&sqs.SendMessageBatchInput{
		QueueUrl: &s.URL,
		Entries:  msgBatch,
	})

	logger.Printf("%d: Successfully Processed", len(result.Successful))
	logger.Printf("%d: Failed to process", len(result.Failed))

	return err
}

// Delete a SQS message from the queue
func (s *Config) Delete(msg *sqs.Message) error {

	_, err := s.svc.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      &s.URL,
		ReceiptHandle: msg.ReceiptHandle,
	})

	return err
}

// RegisterPollHandler : A method to register a custom Poll Handling method
func (s *Config) RegisterPollHandler(pollHandler func(msg *sqs.Message)) {
	s.pollHandler = pollHandler
}

// ChangeVisibilityTimeout : Method to change visibility timeout of a message.
func (s *Config) ChangeVisibilityTimeout(msg *sqs.Message, seconds int64) bool {
	retVal := false
	logger.Printf("change visibility timeout for message ID %s", *msg.MessageId)

	if s.svc == nil {
		logger.Fatal("SQS Connection failed")
		return retVal
	}

	strURL := &s.URL
	receiptHandle := *msg.ReceiptHandle

	changeMessageVisibilityInput := sqs.ChangeMessageVisibilityInput{}

	changeMessageVisibilityInput.SetQueueUrl(*strURL)
	changeMessageVisibilityInput.SetReceiptHandle(receiptHandle)
	changeMessageVisibilityInput.SetVisibilityTimeout(seconds)

	out, err := s.svc.ChangeMessageVisibility(&changeMessageVisibilityInput)

	if err == nil {
		logger.Printf("changed visibility timeout success for %s", (*out).GoString())
		retVal = true
	} else {
		logger.Printf("change visibility timeout failed: %s", (*out).GoString())
	}

	return retVal
}

func validateOpts(opts Config) error {
	if opts.AWSRegion == "" {
		return errors.New("AWSRegion is required")
	}

	if opts.URL == "" {
		return errors.New("A valid SQS URL is required")
	}

	if opts.BatchSize < 0 || opts.BatchSize > 10 {
		return errors.New("BatchSize should be between 1-10")
	}

	if opts.WaitSeconds < 0 || opts.WaitSeconds > 20 {
		return errors.New("WaitSecond should be between 1-20")
	}

	if opts.VisibilityTimeout < 0 || opts.VisibilityTimeout > 12*60*60 {
		return errors.New("WaitSecond should be between 1-43200")
	}

	return nil
}
