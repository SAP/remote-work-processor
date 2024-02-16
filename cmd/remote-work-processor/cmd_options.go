package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"io"
	"log"
	"os"
)

type Options struct {
	DisplayVersion bool
	StandaloneMode bool
	InstanceId     string
	MaxConnRetries uint
}

const (
	standaloneModeOpt = "standalone-mode"
	instanceIdOpt     = "instance-id"
	connRetriesOpt    = "conn-retries"
	versionOpt        = "version"
)

func (opts *Options) BindFlags(fs *flag.FlagSet) {
	hostname := getHashedHostname()

	fs.BoolVar(&opts.StandaloneMode, standaloneModeOpt, false,
		"Whether to run the Remote Work Processor in Standalone mode")
	fs.StringVar(&opts.InstanceId, instanceIdOpt, hostname,
		"Instance Identifier for the Remote Work Processor (only applicable for Standalone mode)")
	fs.UintVar(&opts.MaxConnRetries, connRetriesOpt, 3, "Number of retries for gRPC connection to AutoPi server")
	fs.BoolVar(&opts.DisplayVersion, versionOpt, false, "Display binary version and exit")
}

func getHashedHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("could not get hostname: %v\n", err)
	} else {
		hasher := sha256.New()
		io.WriteString(hasher, hostname)
		hostname = hex.EncodeToString(hasher.Sum(nil))
	}
	return hostname
}
