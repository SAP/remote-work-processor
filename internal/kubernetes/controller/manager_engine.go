package controller

import (
	"context"
	"fmt"
	"log"

	pb "github.com/SAP/remote-work-processor/build/proto/generated"
	"github.com/SAP/remote-work-processor/internal/grpc"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
)

type ManagerEngine struct {
	watchedResources map[string]*pb.Resource
	grpcClient       *grpc.RemoteWorkProcessorGrpcClient
	scheme           *runtime.Scheme
	config           *rest.Config
	started          bool
	cancelCtx        context.CancelFunc
}

func CreateManagerEngine(scheme *runtime.Scheme, config *rest.Config, client *grpc.RemoteWorkProcessorGrpcClient) *ManagerEngine {
	return &ManagerEngine{
		grpcClient: client,
		scheme:     scheme,
		config:     config,
	}
}

func (e *ManagerEngine) SetWatchConfiguration(wc *pb.UpdateConfigRequestMessage) {
	e.watchedResources = wc.Resources
}

func (e *ManagerEngine) StartManager(ctx context.Context, isEnabled func() bool) error {
	if len(e.watchedResources) == 0 {
		return fmt.Errorf("no resources to watch")
	}

	log.Println("Creating manager...")
	manager := NewManagerBuilder().
		SetGrpcClient(e.grpcClient).
		BuildDynamicClient(e.config).
		BuildInternalManager(e.config, e.scheme).
		Build()

	log.Println("Creating controllers...")
	if err := manager.CreateControllersFor(e.watchedResources, isEnabled); err != nil {
		return fmt.Errorf("failed to create controllers: %v", err)
	}

	log.Println("Starting manager...")
	ctx, cancel := context.WithCancel(ctx)
	e.started = true
	e.cancelCtx = cancel
	return manager.Start(ctx)
}

func (e *ManagerEngine) StopManager() {
	log.Println("stopping controller manager...")
	e.cancelCtx()
}

func (e *ManagerEngine) IsStarted() bool {
	return e.started
}
