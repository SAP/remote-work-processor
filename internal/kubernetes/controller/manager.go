package controller

import (
	pb "github.com/SAP/remote-work-processor/build/proto/generated"
	"github.com/SAP/remote-work-processor/internal/grpc"
	"github.com/SAP/remote-work-processor/internal/kubernetes/dynamic"
	"github.com/SAP/remote-work-processor/internal/kubernetes/selector"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type ControllerManager struct {
	manager       manager.Manager
	options       manager.Options
	config        *rest.Config
	watchConfig   *pb.UpdateConfigRequestMessage
	dynamicClient *dynamic.Client
	grpcClient    *grpc.RemoteWorkProcessorGrpcClient
	selectors     map[string]selector.Selector
}

func (m *ControllerManager) GetScheme() *runtime.Scheme {
	return m.manager.GetScheme()
}

func (m *ControllerManager) InitSelectors(resources map[string]*pb.Resource) {
	m.selectors = make(map[string]selector.Selector, len(resources))
	for name, resource := range resources {
		m.selectors[name] = selector.NewSelector(resource.GetLabelSelectors(), resource.GetFieldSelectors())
	}
}

func (m *ControllerManager) GetSelector(reconciler string) selector.Selector {
	return m.selectors[reconciler]
}

func (m *ControllerManager) CreateControllers(isEnabled func() bool) error {
	for reconciler, resource := range m.watchConfig.GetResources() {
		_, err := CreateControllerBuilder().
			For(resource).
			ManagedBy(m).
			WithReconcilicationPeriodInMinutes(resource.ReconciliationPeriodInMinutes).
			Build(reconciler, m.grpcClient, isEnabled)

		if err != nil {
			return err
		}
	}
	return nil
}
