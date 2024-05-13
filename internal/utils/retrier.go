package utils

import (
	"context"
	"log"
	"time"
)

type RetryStrategy string

const (
	RetryStrategyFixed       RetryStrategy = "fixed"
	RetryStrategyIncremental RetryStrategy = "incr"
)

type RetryConfig struct {
	retryInterval time.Duration
	retryStrategy RetryStrategy
	signalChan    chan<- struct{}
	attempts      uint
	maxAttempts   uint
}

func CreateDefaultRetryConfig(signalChan chan<- struct{}) *RetryConfig {
	return CreateRetryConfig(10*time.Second, RetryStrategyFixed, 6, signalChan)
}

func CreateRetryConfig(interval time.Duration, strategy RetryStrategy, maxAttempts uint, signalChan chan<- struct{}) *RetryConfig {
	return &RetryConfig{
		retryInterval: interval,
		retryStrategy: strategy,
		signalChan:    signalChan,
		maxAttempts:   maxAttempts,
	}
}

func (conf *RetryConfig) CanRetry() bool {
	return conf.attempts < conf.maxAttempts
}

func (conf *RetryConfig) getNextRetryInterval() time.Duration {
	attempts := conf.attempts
	conf.attempts++
	if conf.retryStrategy == RetryStrategyIncremental {
		return time.Duration(float32(attempts+1)*1.75) * conf.retryInterval
	}
	// default: fixed
	return conf.retryInterval
}

func Retry(ctx context.Context, config *RetryConfig, err error) {
	log.Println(err)
	nextRetryInterval := config.getNextRetryInterval()
	log.Println("retrying after", nextRetryInterval)
	select {
	case <-ctx.Done():
		return
	case <-time.After(nextRetryInterval):
	}
	config.signalChan <- struct{}{}
}
