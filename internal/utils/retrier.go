package utils

import (
	"context"
	"log"
	"time"
)

type RetryStrategy string

const (
	RetryStrategyFixed       RetryStrategy = "fixed"
	RetryStrategyIncremental               = "incr"
)

type RetryConfig struct {
	retryInterval time.Duration
	retryStrategy RetryStrategy
	signalChan    chan<- struct{}
	attempts      uint
}

func CreateDefaultRetryConfig(signalChan chan<- struct{}) *RetryConfig {
	return CreateRetryConfig(10*time.Second, RetryStrategyFixed, signalChan)
}

func CreateRetryConfig(interval time.Duration, strategy RetryStrategy, signalChan chan<- struct{}) *RetryConfig {
	return &RetryConfig{
		retryInterval: interval,
		retryStrategy: strategy,
		signalChan:    signalChan,
	}
}

func (conf *RetryConfig) GetAttempts() uint {
	return conf.attempts
}

func (conf *RetryConfig) getNextRetryInterval() time.Duration {
	attempts := conf.attempts
	conf.attempts++
	if conf.retryStrategy == RetryStrategyFixed {
		return conf.retryInterval
	}
	if conf.retryStrategy == RetryStrategyIncremental {
		return time.Duration(float32(attempts+1)*1.75) * conf.retryInterval
	}
	return 0 //unreachable
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
