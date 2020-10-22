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
	AWSKey            string
	AWSSecret         string
	AWSRegion         string
	URL               string
	BatchSize         int64
	WaitSeconds       int64
	VisibilityTimeout int64
	RunOnce           bool
	RunInterval       int
	Verbosity         int
	MaxRetries        int
	MaxHandlers       int
	svc               *sqs.SQS
	logger            libLogger.Logger
}

// SQS An interface for SQS operations
type SQS interface {
	Poll(handler func(wg *sync.WaitGroup, msg *sqs.Message))
	Delete(msg *sqs.Message) error
	Enque(msgBatch []*sqs.SendMessageBatchRequestEntry) error
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

// Poll Poll for messages in the SQS
func (s *Config) Poll(handler func(wg *sync.WaitGroup, msg *sqs.Message)) {
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
		result, err := s.svc.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            &s.URL,
			MaxNumberOfMessages: &s.BatchSize,
			WaitTimeSeconds:     &s.WaitSeconds,
			VisibilityTimeout:   &s.VisibilityTimeout,
		})

		if err != nil {
			logger.Error("ReceiveMessageError:", err)
			break
		}

		if len(result.Messages) == 0 {
			logger.Info("Queue is empty")
		} else {
			logger.Info("Fetched", len(result.Messages), "messages")
		}
		for _, msg := range result.Messages {
			handlerCount++
			wg.Add(1)
			go handler(&wg, msg)
			logger.Debug("Spawned handler for", msg.MessageId)

			// Is running at capacity?
			if s.MaxHandlers > 0 && handlerCount >= s.MaxHandlers {
				logger.Info("Reached max handlers", handlerCount)
				// Send all messages back to the queue
				// --
				// Since all handlers are busy, let's wait for 30 seconds
				logger.Info("Going to wait state for 30 seconds")
				<-time.After(30 * time.Second)
				break
			}
		}

		if s.RunOnce == true {
			logger.Info(`Exiting since RUN_ONCE is set to "true"`)
			break
		} else {
			logger.Info("Waiting for ", s.RunInterval, "seconds before polling for next batch")
			<-time.After(time.Duration(s.RunInterval) * time.Second)
		}

		logger.Info("Finished polling")
	}

	wg.Wait()
}

// SendBatch messages to SQS
func (s *Config) Enque(msgBatch []*sqs.SendMessageBatchRequestEntry) error {
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
