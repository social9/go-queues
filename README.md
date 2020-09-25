# go-consumer (In Development)

A generic consumer service written in Go

It is designed to be inherently scalable with concurrent processing using goroutines and has pluggable stream sources such as `SQS`, `Kafka`, etc.

## Setup

- Clone the repo `git clone https://github.com/cnp96/go-consumer`
- Configurable environment variables
  - `MAX_PROC` : The maximum number of processor cores to use, _(default **all available**)_
  - `SQS_BATCH_SIZE` : The maximum number of messages to receive in a polling request, _(default **1000**)_
  - `SQS_WAIT_TIME` : The maximum polling wait time in seconds, _(default **20 seconds**)_


## Contribution Guidelines

_To be added soon_

## License

MIT
