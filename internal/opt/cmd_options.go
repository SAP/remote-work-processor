package opt

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"github.com/SAP/remote-work-processor/internal/utils"
	"github.com/google/uuid"
	"io"
	"log"
	"os"
	"time"
)

type Options struct {
	DisplayVersion bool
	StandaloneMode bool
	InstanceId     string
	MaxConnRetries uint
	RetryInterval  time.Duration
	RetryStrategy  StrategyOpt
}

type StrategyOpt utils.RetryStrategy

const (
	standaloneModeOpt = "standalone-mode"
	instanceIdOpt     = "instance-id"
	connRetriesOpt    = "conn-retries"
	versionOpt        = "version"
	retryIntervalOpt  = "retry-interval"
	retryStrategyOpt  = "retry-strategy"
)

func (opts *Options) BindFlags(fs *flag.FlagSet) {
	hostname := getHashedHostname()

	fs.BoolVar(&opts.StandaloneMode, standaloneModeOpt, false,
		"Whether to run the Remote Work Processor in Standalone mode")
	fs.StringVar(&opts.InstanceId, instanceIdOpt, hostname,
		"Instance Identifier for the Remote Work Processor (only applicable for Standalone mode)")
	fs.UintVar(&opts.MaxConnRetries, connRetriesOpt, 6, "Number of retries for gRPC connection to AutoPi server")
	fs.BoolVar(&opts.DisplayVersion, versionOpt, false, "Display binary version and exit")
	fs.DurationVar(&opts.RetryInterval, retryIntervalOpt, 10*time.Second, "Retry interval")
	fs.Var(&opts.RetryStrategy, retryStrategyOpt, "Retry strategy [fixed, incr]")
}

func (opt *StrategyOpt) String() string {
	if len(*opt) == 0 {
		return string(utils.RetryStrategyFixed)
	}
	return string(*opt)
}

func (opt *StrategyOpt) Get() any {
	if len(*opt) == 0 {
		return utils.RetryStrategyFixed
	}
	return utils.RetryStrategy(*opt)
}

func (opt *StrategyOpt) Set(value string) error {
	if value != string(utils.RetryStrategyFixed) && value != utils.RetryStrategyIncremental {
		return errors.New("invalid value for retry-strategy: " + value)
	}
	*opt = StrategyOpt(value)
	return nil
}

func (opt *StrategyOpt) Unmarshall() utils.RetryStrategy {
	return opt.Get().(utils.RetryStrategy)
}

func getHashedHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("could not get hostname: %v\n", err)
		return uuid.Nil.String()
	} else {
		hasher := sha256.New()
		io.WriteString(hasher, hostname)
		hostname = hex.EncodeToString(hasher.Sum(nil))
	}
	return hostname
}
