/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/SAP/remote-work-processor/internal/grpc"
	"github.com/SAP/remote-work-processor/internal/grpc/processors"
	"github.com/SAP/remote-work-processor/internal/kubernetes/controller"
	meta "github.com/SAP/remote-work-processor/internal/kubernetes/metadata"
	"io"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"log"
	"os"
	"os/signal"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"strconv"
	"syscall"
	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	//+kubebuilder:scaffold:imports
)

const (
	standaloneModeOpt = "standalone-mode"
	instanceIdOpt     = "instance-id"
	connRetriesOpt    = "connection-retries"
)

var Version string

func main() {
	setupFlagsAndLogger()

	rootCtx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	instanceIDFlag := flag.Lookup(instanceIdOpt).Value.String()
	rwpMetadata := meta.LoadMetadata(instanceIDFlag, Version)
	grpcClient := grpc.NewClient(rwpMetadata)
	drainChan := make(chan struct{}, 1)

	var factory processors.ProcessorFactory

	isInStandaloneMode, _ := strconv.ParseBool(flag.Lookup(standaloneModeOpt).Value.String())
	if isInStandaloneMode {
		factory = processors.NewStandaloneProcessorFactory()
	} else {
		config := getKubeConfig()
		scheme := runtime.NewScheme()
		utilruntime.Must(clientgoscheme.AddToScheme(scheme))
		//+kubebuilder:scaffold:scheme

		engine := controller.CreateManagerEngine(scheme, config, grpcClient)
		factory = processors.NewKubernetesProcessorFactory(engine, drainChan)
	}

	connAttemptChan := make(chan struct{}, 1)
	connAttemptChan <- struct{}{}
	connAttempts := uint64(0)
	maxRetries, _ := strconv.ParseUint(flag.Lookup(connRetriesOpt).Value.String(), 10, 64)
	for connAttempts < maxRetries {
		select {
		case <-rootCtx.Done():
			break
		case <-connAttemptChan:
			err := grpcClient.InitSession(rootCtx, rwpMetadata.SessionID())
			if err != nil {
				signalRetry(&connAttempts, connAttemptChan, err)
			}
		default:
			operation, err := grpcClient.ReceiveMsg()
			if err != nil {
				signalRetry(&connAttempts, connAttemptChan, err)
				continue
			}
			if operation == nil {
				// this flow is only when the backend closes the gRPC connection
				connAttemptChan <- struct{}{}
				// do not increment the retries, as this isn't a failure
				continue
			}

			processor, err := factory.CreateProcessor(operation)
			if err != nil {
				signalRetry(&connAttempts, connAttemptChan, fmt.Errorf("error creating operation processor: %v", err))
				continue
			}

			msg, err := processor.Process(rootCtx)
			//TODO: not every error needs session reestablishment; make a custom error struct and only
			// recreation the session based on error type
			if err != nil {
				//TODO: check how the backed handles the case when the client doesn't send a "confirm" message
				signalRetry(&connAttempts, connAttemptChan, fmt.Errorf("error processing operation: %v", err))
				continue
			}
			if msg == nil {
				continue
			}

			if err = grpcClient.Send(msg); err != nil {
				signalRetry(&connAttempts, connAttemptChan, err)
			}
		}
	}

	if !isInStandaloneMode {
		// wait for context cancellation to be propagated to the k8s manager
		<-drainChan
	}
}

func setupFlagsAndLogger() {
	hostname := getHashedHostname()

	flag.Bool(standaloneModeOpt, false, "Whether to run the Remote Work Processor in Standalone mode")
	flag.String(instanceIdOpt, hostname, "Instance Identifier for the Remote Work Processor (only applicable for Standalone mode)")
	flag.Uint(connRetriesOpt, 3, "Number of retries for gRPC connection to AutoPi server")

	opts := zap.Options{}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
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

func getKubeConfig() *rest.Config {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalln("Could not create kubeconfig:", err)
	}
	return config
}

func signalRetry(attempts *uint64, retryChan chan<- struct{}, err error) {
	if err != nil {
		log.Println(err)
	}
	retryChan <- struct{}{}
	*attempts++
}
