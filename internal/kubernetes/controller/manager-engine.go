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
	managerBuilder *ControllerManagerBuilder
	started        bool
	cancelCtx      context.CancelFunc
}

func CreateManagerEngine(scheme *runtime.Scheme, config *rest.Config, client *grpc.RemoteWorkProcessorGrpcClient) *ManagerEngine {
	builder := CreateManagerBuilder(client).
		WithConfig(config).
		WithOptions(scheme).
		WithoutLeaderElection()
	return &ManagerEngine{
		managerBuilder: builder,
	}
}

func (e *ManagerEngine) SetWatchConfiguration(wc *pb.UpdateConfigRequestMessage) {
	e.managerBuilder.SetWatchConfiguration(wc)
}

func (e *ManagerEngine) StartManager(ctx context.Context, isEnabled func() bool) error {
	log.Println("starting manager...")
	cm := e.managerBuilder.Build()

	if err := cm.CreateControllers(isEnabled); err != nil {
		return fmt.Errorf("unable to create controllers: %v", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	e.started = true
	e.cancelCtx = cancel
	return cm.manager.Start(ctx)
}

func (e *ManagerEngine) StopManager() {
	log.Println("stopping controller manager...")
	e.cancelCtx()
}

func (e *ManagerEngine) IsStarted() bool {
	return e.started
}
