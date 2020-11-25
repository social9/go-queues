package sqs

import (
	"errors"
	"os"
	"strconv"
	"sync"
	"time"

	libLogger "github.com/social9/go-queues/lib/logger"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// Config Wrapper for Config methods
type Config struct {
	AWSKey    string
	AWSSecret string
	AWSRegion string

	// Poll from this SQS URL
	URL string

	// Maximum number of time to attempt AWS service connection
	MaxRetries int

	// Maximum number to retrieve of messages per batch
	BatchSize int64

	// The maximum poll time (0 <= 20)
	WaitSeconds int64

	// Once received a consumer, the maximum time in seconds till others can see this message
	VisibilityTimeout int64

	// Poll only once
	RunOnce bool

	// Poll every X seconds defined by this value
	RunInterval int

	// 0-5 : debug, info, warn, error, fatal
	Verbosity int

	// Maximum number of handlers to spawn
	MaxHandlers int

	// BusyTimeout in seconds
	BusyTimeout int

	svc    *sqs.SQS
	logger libLogger.Logger

	// First class function register poll handler
	PollHandler func(wg *sync.WaitGroup, msg *sqs.Message)
}

// SQS An interface for SQS operations
type SQS interface {
	Poll()
	Delete(msg *sqs.Message) error
	Enqueue(msgBatch []*sqs.SendMessageBatchRequestEntry) error
	RegisterPollHandler(pollHandler func(wg *sync.WaitGroup, msg *sqs.Message))
	ChangeVisibilityTimeout(msg *sqs.Message, seconds int64) bool
}

// NewSQS Initialise a SQS instance
func NewSQS(opts Config) (SQS, error) {
	logger := libLogger.NewLogger(libLogger.Config{Level: opts.Verbosity})

	// Validate parameters
	validateErr := validateOpts(opts)
	if validateErr != nil {
		logger.Debug(validateErr)
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
		logger.Debug("Creds error", err)
		return nil, errors.New("Invalid AWS credentials. Please make sure that `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` is present in the env")
	}

	// Create AWS Config
	awsConfig := aws.NewConfig().WithRegion(opts.AWSRegion).WithMaxRetries(opts.MaxRetries).WithCredentials(creds)
	if awsConfig == nil {
		logger.Debug("Invalid AWS Config")
		return nil, errors.New("Something is wrong with your AWS config parameters")
	}

	// Establish a session
	newSession := session.Must(session.NewSession(awsConfig))
	if newSession == nil {
		logger.Debug("Unable to create session")
		return nil, errors.New("Unable to create session")
	}

	// Create a service connection
	svc := sqs.New(newSession)
	if svc == nil {
		logger.Debug("Unable to connect to SQS")
		return nil, errors.New("Unable to create a service connection with AWS SQS")
	}

	logger.Info("Fetching queue attributes")
	if _, err := svc.GetQueueAttributes(&sqs.GetQueueAttributesInput{
		QueueUrl: &opts.URL,
	}); err != nil {
		logger.Debug("Unable to fetch queue attributes", err)
		return nil, errors.New("Unable to get queue attributes")
	}
	logger.Info("Connected to Queue")

	opts.logger = logger
	opts.svc = svc
	return &opts, nil
}

// Poll for messages in the queue
func (s *Config) Poll() {
	if s.svc == nil {
		s.logger.Fatal("No service connection")
	}
	handlerCount := 0
	wg := sync.WaitGroup{}
	batch := 0

	for {
		batch++
		logger := s.logger.Child(libLogger.Config{Name: "batch-" + strconv.Itoa(batch)})

		logger.Info("Start receiving messages")

		maxMsgs := s.BatchSize
		// Is running at capacity?
		if s.MaxHandlers > 0 && handlerCount >= s.MaxHandlers {
			logger.Info("Running at full capacity with ", handlerCount, "handlers")

			// Since all handlers are busy, let's wait for BusyTimeout seconds
			logger.Info("Going to wait state for", s.BusyTimeout, "seconds")
			<-time.After(time.Duration(s.BusyTimeout) * time.Second)
			continue
		} else {
			maxMsgs = int64(s.MaxHandlers - handlerCount)
			logger.Info("Can accept a maximum of ", maxMsgs, "messages")
		}

		result, err := s.svc.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            &s.URL,
			MaxNumberOfMessages: &maxMsgs,
			WaitTimeSeconds:     &s.WaitSeconds,
			VisibilityTimeout:   &s.VisibilityTimeout,
		})

		// Retrieve error?
		if err != nil {
			logger.Error("ReceiveMessageError:", err)
			break
		}

		// Message log
		if len(result.Messages) == 0 {
			logger.Info("Queue is empty")
		} else {
			logger.Info("Fetched", len(result.Messages), "messages")
		}

		// Process messages
		for _, msg := range result.Messages {
			handlerCount++
			wg.Add(1)
			if s.PollHandler == nil {
				logger.Error("No Poll handler registered. Register a handler for custom handling")
			} else {
				go s.PollHandler(&wg, msg)
			}

			logger.Debug("Spawned handler for", msg.MessageId)
		}

		if s.RunOnce == true {
			logger.Info(`Exiting since confugured to run once`)
			break
		} else {
			logger.Info("Waiting for ", s.RunInterval, "seconds before polling for next batch")
			<-time.After(time.Duration(s.RunInterval) * time.Second)
		}

		logger.Info("Finished polling")
	}

	wg.Wait()
}

// Enqueue messages to SQS
func (s *Config) Enqueue(msgBatch []*sqs.SendMessageBatchRequestEntry) error {
	if s.svc == nil {
		s.logger.Fatal("No service connection")
	}

	s.logger.Info(len(msgBatch), `messages are processing`)

	result, err := s.svc.SendMessageBatch(&sqs.SendMessageBatchInput{
		QueueUrl: &s.URL,
		Entries:  msgBatch,
	})

	s.logger.Info(len(result.Successful), ": Successfully Processed")
	s.logger.Info(len(result.Failed), ": Failed to process")

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
func (s *Config) RegisterPollHandler(pollHandler func(wg *sync.WaitGroup, msg *sqs.Message)) {
	s.PollHandler = pollHandler
}

// ChangeVisibilityTimeout : Method to change visibility timeout of a message.
func (s *Config) ChangeVisibilityTimeout(msg *sqs.Message, seconds int64) bool {
	retVal := false
	s.logger.Info("ChangeVisibilityTimeout : ", sqs.Message.String(*msg))

	if s.svc == nil {
		s.logger.Fatal("SQS Connection failed")
		return retVal
	}

	strURL := &s.URL
	receiptHandle := *msg.ReceiptHandle

	changeMessageVisibilityInput := sqs.ChangeMessageVisibilityInput{}

	changeMessageVisibilityInput.SetQueueUrl(*strURL)
	changeMessageVisibilityInput.SetReceiptHandle(receiptHandle)
	changeMessageVisibilityInput.SetVisibilityTimeout(seconds)

	out, err := s.svc.ChangeMessageVisibility(&changeMessageVisibilityInput)

	if err != nil {
		s.logger.Fatal("Change visibility timeout failed", (*out).GoString())
	} else {
		s.logger.Info(" Return :", (*out).GoString())
		retVal = true
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
