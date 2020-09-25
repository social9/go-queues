# go-consumer (In Development)

A generic consumer service written in Go

It is designed to be inherently scalable with concurrent processing using goroutines and has pluggable stream sources such as `SQS`, `Kafka`, etc.

## Setup

- Clone the repo `git clone https://github.com/cnp96/go-consumer`
- Configurable environment variables
  - `SQS_BATCH_SIZE` : The maximum number of messages to receive in a polling request, _(default **5**)_
  - `SQS_WAIT_TIME` : The maximum polling wait time in seconds, _(default **20 seconds**)_
  - `RUN_ONCE` : Run the service one-time or in intervals, _(default **true**)_
  - `RUN_INTERVAL` : Run the service in the defined interval, _(default **10 seconds**)_. Works only when `RUN_ONCE` is set to `true`

## Contribution Guidelines

_To be added soon_

## License

MIT
