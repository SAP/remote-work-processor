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
	// "flag"
	// "os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	"log"
	"os"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	// "sigs.k8s.io/controller-runtime/pkg/healthz"
	// "sigs.k8s.io/controller-runtime/pkg/log/zap"
	// "github.com/SAP/remote-work-processor/kubernetes/controllers"
	"github.com/SAP/remote-work-processor/internal/grpc"
	"github.com/SAP/remote-work-processor/internal/grpc/processors"
	"github.com/SAP/remote-work-processor/internal/kubernetes/controller"
	"github.com/SAP/remote-work-processor/internal/kubernetes/metadata"
	//+kubebuilder:scaffold:imports
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	metadata.InitRemoteWorkProcessorMetadata()
	config := getKubeConfig()

	e := controller.CreateManagerEngine(scheme, config)
	processors.InitProcessorFactory(e)
	grpc.InitRemoteWorkProcessorGrpcClient()

	opc := grpc.Client.Receive()

	for {
		op := <-opc
		p, err := processors.Factory.CreateProcessor(op)
		if err != nil {
			log.Fatalf("Error occurred while creating operation processor: %v\n", err)
		}

		res := <-p.Process()
		if res.Err != nil {
			log.Fatalf("Error occurred while processing operation: %v\n", err)
		}

		if res.Result != nil {
			grpc.Client.Send(res.Result)
		}
	}
}

func getKubeConfig() *rest.Config {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	overrides := &clientcmd.ConfigOverrides{}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		os.Exit(1)
	}

	return config
}
