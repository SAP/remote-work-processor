package controller

import (
	"context"
	"fmt"
	pb "github.com/SAP/remote-work-processor/build/proto/generated"
	"github.com/SAP/remote-work-processor/internal/grpc"
	"github.com/SAP/remote-work-processor/internal/kubernetes/dynamic"
	"github.com/SAP/remote-work-processor/internal/kubernetes/selector"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type Manager struct {
	delegate      manager.Manager
	dynamicClient *dynamic.Client
	grpcClient    *grpc.RemoteWorkProcessorGrpcClient
}

func (m *Manager) CreateControllersFor(resources map[string]*pb.Resource, isEnabled func() bool) error {
	for reconciler, resource := range resources {
		log.Printf("Creating controller for %s/%s watched by %s\n", resource.ApiVersion, resource.Kind, reconciler)
		err := NewControllerFor(resource).
			ManagedBy(m).
			WithReconcilicationPeriodInMinutes(resource.ReconciliationPeriodInMinutes).
			WithSelector(selector.NewSelector(resource.GetLabelSelectors(), resource.GetFieldSelectors())).
			Create(reconciler, m.grpcClient, isEnabled)
		if err != nil {
			return fmt.Errorf("failed to create controller for %s/%s: %s", resource.ApiVersion, resource.Kind, err)
		}
	}
	return nil
}

func (m *Manager) Start(ctx context.Context) error {
	return m.delegate.Start(ctx)
}
