package utils

import (
	"context"
	"time"
)

type RetryStrategy string

const (
	RetryStrategyFixed       RetryStrategy = "fixed"
	RetryStrategyExponential               = "expnt"
)

type RetryConfig struct {
	retryInterval             time.Duration
	retryStrategy             RetryStrategy
	retryMultiplicationFactor float32
	signalChan                chan<- struct{}
	attempts                  uint
}

func CreateDefaultRetryConfig(signalChan chan<- struct{}) *RetryConfig {
	return CreateRetryConfig(10*time.Second, RetryStrategyFixed, 1, signalChan)
}

func CreateRetryConfig(interval time.Duration, strategy RetryStrategy, factor float32, signalChan chan<- struct{}) *RetryConfig {
	return &RetryConfig{
		retryInterval:             interval,
		retryStrategy:             strategy,
		retryMultiplicationFactor: factor,
		signalChan:                signalChan,
	}
}

func (conf *RetryConfig) GetAttempts() uint {
	return conf.attempts
}

func (conf *RetryConfig) getNextRetryInterval() time.Duration {
	attempts := conf.attempts
	conf.attempts++
	if conf.retryStrategy == RetryStrategyFixed {
		return time.Duration(attempts+1) * conf.retryInterval
	}
	return 0
}

func Retry(ctx context.Context, config *RetryConfig, err error) {

}
