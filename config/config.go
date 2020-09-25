package config

import (
	"log"

	"github.com/caarlos0/env"
)

// Config a struct containing env variables
type Config struct {
	SQSURL      string `env:"SQS_URL" envDefault:""`
	SQSLimit    int    `env:"SQS_BATCH_SIZE" envDefault:"5"`
	SQSWaitTime int    `env:"SQS_WAIT_TIME" envDefault:"20"`
	RunInterval int    `env:"RUN_INTERVAL" envDefault:"10"`
	RunOnce     bool   `env:"RUN_ONCE" envDefault:"true"`
}

var instance Config

func init() {
	instance = Config{}
	if err := env.Parse(&instance); err != nil {
		log.Fatal(err)
	}
}

// Env returns a config instance
func Env() Config {
	return instance
}
