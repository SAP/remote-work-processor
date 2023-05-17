package controller

import (
	"context"

	pb "github.com/SAP/remote-work-processor/build/proto/generated"
	"github.com/SAP/remote-work-processor/internal/cache"
	"github.com/SAP/remote-work-processor/internal/kubernetes/dynamic"
	"github.com/SAP/remote-work-processor/internal/kubernetes/selector"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type ControllerManager struct {
	manager       manager.Manager
	options       manager.Options
	config        *rest.Config
	watchConfig   *pb.UpdateConfigRequestMessage
	dynamicClient *dynamic.DynamicClient
	selectorCache cache.Cache[string, selector.Selector]
}

func (m *ControllerManager) GetClient() client.Client {
	return m.manager.GetClient()
}

func (m *ControllerManager) GetScheme() *runtime.Scheme {
	return m.manager.GetScheme()
}

func (m *ControllerManager) CreateControllers(ctx context.Context) error {
	for reconciler, resource := range m.watchConfig.GetResources() {
		_, err := CreateControllerBuilder().
			For(resource).
			ManagedBy(m).
			Build(ctx, reconciler)

		if err != nil {
			return err
		}
	}

	return nil
}
