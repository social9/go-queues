package config

import (
	"log"
	"os"
	"sync"

	"github.com/caarlos0/env"
)

// Config a struct containing env variables
type Config struct {
	AWSKey               string `env:"AWS_ACCESS_KEY_ID" envDefault:""`
	AWSSecret            string `env:"AWS_SECRET_ACCESS_KEY" envDefault:""`
	SQSURL               string `env:"SQS_URL" envDefault:""`
	AWSRegion            string `env:"AWS_REGION" envDefault:"us-east-2"`
	SQSLimit             int64  `env:"SQS_BATCH_SIZE" envDefault:"10"`
	SQSWaitTime          int64  `env:"SQS_WAIT_TIME" envDefault:"20"`
	SQSVisibilityTimeout int64  `env:"SQS_VISIBILITY_TIMEOUT" envDefault:"20"`
	RunInterval          int    `env:"RUN_INTERVAL" envDefault:"10"`
	RunOnce              bool   `env:"RUN_ONCE" envDefault:"true"`
}

var instance Config
var once sync.Once

func init() {
	once.Do(func() {
		instance = Config{}
		if err := env.Parse(&instance); err != nil {
			log.Fatal(err)
		} else {

			os.Setenv("AWS_ACCESS_KEY_ID", instance.AWSKey)
			os.Setenv("AWS_SECRET_ACCESS_KEY", instance.AWSSecret)
		}
	})
}

// Env returns a config instance
func Env() Config {
	return instance
}
