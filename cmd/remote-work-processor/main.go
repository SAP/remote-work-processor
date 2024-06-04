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
	"flag"
	"fmt"
	"github.com/SAP/remote-work-processor/internal/grpc"
	"github.com/SAP/remote-work-processor/internal/grpc/processors"
	"github.com/SAP/remote-work-processor/internal/kubernetes/controller"
	meta "github.com/SAP/remote-work-processor/internal/kubernetes/metadata"
	"github.com/SAP/remote-work-processor/internal/opt"
	"github.com/SAP/remote-work-processor/internal/utils"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"log"
	"os"
	"os/signal"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"syscall"
	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	//+kubebuilder:scaffold:imports
)

var (
	// Version of the Remote Work Processor.
	// Injected at linking time via ldflags.
	Version string
	// BuildDate of the Remote Work Processor.
	// Injected at linking time via ldflags.
	BuildDate string
)

func main() {
	opts := setupFlagsAndLogger()

	if opts.DisplayVersion {
		fmt.Printf("rwp-%s Built: %s\n", Version, BuildDate)
		return
	}

	rootCtx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	rwpMetadata := meta.LoadMetadata(opts.InstanceId, Version)
	grpcClient := grpc.NewClient(rwpMetadata, opts.StandaloneMode)
	var drainChan chan struct{}

	var factory processors.ProcessorFactory

	if opts.StandaloneMode {
		factory = processors.NewStandaloneProcessorFactory()
	} else {
		config := getKubeConfig()
		scheme := runtime.NewScheme()
		utilruntime.Must(clientgoscheme.AddToScheme(scheme))
		//+kubebuilder:scaffold:scheme

		drainChan = make(chan struct{}, 1)
		engine := controller.CreateManagerEngine(scheme, config, grpcClient)
		factory = processors.NewKubernetesProcessorFactory(engine, drainChan)
	}

	connAttemptChan := make(chan struct{}, 1)
	connAttemptChan <- struct{}{}
	retryConfig := utils.CreateRetryConfig(opts.RetryInterval, opts.RetryStrategy.Unmarshall(), opts.MaxConnRetries, connAttemptChan)

Loop:
	for retryConfig.CanRetry() {
		select {
		case <-rootCtx.Done():
			log.Println("Received cancellation signal. Stopping Remote Work Processor...")
			break Loop
		case <-connAttemptChan:
			err := grpcClient.InitSession(rootCtx, rwpMetadata.SessionID())
			if err != nil {
				utils.Retry(rootCtx, retryConfig, err)
			}
		default:
			operation, err := grpcClient.ReceiveMsg()
			if err != nil {
				utils.Retry(rootCtx, retryConfig, err)
				continue
			}
			if operation == nil {
				// this flow is when the gRPC connection is closed by the server
				break Loop
			}

			log.Printf("Creating processor for operation: %T\n", operation.Body)
			processor, err := factory.CreateProcessor(operation)
			if err != nil {
				log.Printf("error creating operation processor: %v\n", err)
				continue
			}

			msg, err := processor.Process(rootCtx)
			if err != nil {
				utils.Retry(rootCtx, retryConfig, fmt.Errorf("error processing operation: %v", err))
				continue
			}
			if msg == nil {
				continue
			}

			if err = grpcClient.Send(msg); err != nil {
				utils.Retry(rootCtx, retryConfig, err)
			}
		}
	}

	if !opts.StandaloneMode {
		// wait for context cancellation to be propagated to the k8s manager
		<-drainChan
	}
}

func setupFlagsAndLogger() *opt.Options {
	opts := &opt.Options{}
	opts.BindFlags(flag.CommandLine)

	zapOpts := zap.Options{}
	zapOpts.BindFlags(flag.CommandLine)

	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zapOpts)))
	return opts
}

func getKubeConfig() *rest.Config {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalln("Could not create kubeconfig:", err)
	}
	return config
}
