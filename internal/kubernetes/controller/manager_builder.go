package controller

import (
	"github.com/SAP/remote-work-processor/internal/grpc"
	"github.com/SAP/remote-work-processor/internal/kubernetes/dynamic"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"log"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"time"
)

type ManagerBuilder struct {
	delegate      manager.Manager
	dynamicClient *dynamic.Client
	grpcClient    *grpc.RemoteWorkProcessorGrpcClient
}

func NewManagerBuilder() *ManagerBuilder {
	return &ManagerBuilder{}
}

func (b *ManagerBuilder) SetGrpcClient(client *grpc.RemoteWorkProcessorGrpcClient) *ManagerBuilder {
	b.grpcClient = client
	return b
}

func (b *ManagerBuilder) BuildDynamicClient(config *rest.Config) *ManagerBuilder {
	dc, err := dynamic.NewDynamicClient(config)
	if err != nil {
		log.Panicln("Failed to create dynamic client:", err)
	}
	b.dynamicClient = dc
	return b
}

func (b *ManagerBuilder) BuildInternalManager(config *rest.Config, scheme *runtime.Scheme) *ManagerBuilder {
	t := time.Duration(0)
	options := manager.Options{
		Scheme:                  scheme,
		GracefulShutdownTimeout: &t,
		WebhookServer:           nil,
		LeaderElection:          false,
		HealthProbeBindAddress:  "localhost:8811",
		MetricsBindAddress:      "0",
	}

	mgr, err := ctrl.NewManager(config, options)
	if err != nil {
		log.Panicln("Failed to create manager:", err)
	}

	// these can only fail if the manager has been started prior to calling them
	mgr.AddHealthzCheck("healthz", healthz.Ping)
	mgr.AddReadyzCheck("readyz", healthz.Ping)
	b.delegate = mgr
	return b
}

func (b *ManagerBuilder) Build() Manager {
	return Manager{
		delegate:      b.delegate,
		dynamicClient: b.dynamicClient,
		grpcClient:    b.grpcClient,
	}
}
