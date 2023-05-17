package controller

import (
	"context"
	"fmt"
	"log"

	pb "github.com/SAP/remote-work-processor/build/proto/generated"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
)

type ManagerEngine struct {
	managerBuilder            *ControllerManagerBuilder
	context                   context.Context
	cancellation              chan struct{}
	managerStartedAtLeastOnce bool
}

func CreateManagerEngine(scheme *runtime.Scheme, config *rest.Config) *ManagerEngine {
	builder := CreateManagerBuilder().
		WithConfig(config).
		WithOptions(scheme).
		WithoutLeaderElection()

	me := &ManagerEngine{
		managerBuilder: builder,
	}

	return me
}

func (e *ManagerEngine) WithContext() {
	ctx, cancel := context.WithCancel(context.Background())
	fmt.Println("creating cancellation channel")
	e.cancellation = make(chan struct{})

	go func() {
		<-e.cancellation
		cancel()
	}()

	e.context = ctx
}

func (e *ManagerEngine) WithWatchConfiguration(wc *pb.UpdateConfigRequestMessage) {
	e.managerBuilder.WithWatchConfiguration(wc)
}

func (e *ManagerEngine) StartManager() error {
	fmt.Println("starting manager")
	cm := e.managerBuilder.Build()

	if err := cm.CreateControllers(e.context); err != nil {
		log.Fatal("unable to create controllers", err)
	}

	e.managerStartedAtLeastOnce = true
	return cm.manager.Start(e.context)
}

func (e *ManagerEngine) StopManager() {
	fmt.Println("stopping controller manager...")
	close(e.cancellation)
}

func (e *ManagerEngine) ManagerStartedAtLeastOnce() bool {
	return e.managerStartedAtLeastOnce
}
