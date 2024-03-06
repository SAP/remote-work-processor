package controller

import (
	"fmt"
	pb "github.com/SAP/remote-work-processor/build/proto/generated"
	"github.com/SAP/remote-work-processor/internal/grpc"
	"github.com/SAP/remote-work-processor/internal/kubernetes/selector"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type ControllerBuilder struct {
	resource                      *pb.Resource
	selector                      *selector.Selector
	manager                       *Manager
	reconciliationPeriodInMinutes int32
}

func NewControllerFor(r *pb.Resource) *ControllerBuilder {
	return &ControllerBuilder{
		resource: r,
	}
}

func (c *ControllerBuilder) WithReconcilicationPeriodInMinutes(period int32) *ControllerBuilder {
	c.reconciliationPeriodInMinutes = period
	return c
}

func (c *ControllerBuilder) WithSelector(selector *selector.Selector) *ControllerBuilder {
	c.selector = selector
	return c
}

func (c *ControllerBuilder) ManagedBy(manager *Manager) *ControllerBuilder {
	c.manager = manager
	return c
}

func (c *ControllerBuilder) Create(reconciler string, grpcClient *grpc.RemoteWorkProcessorGrpcClient,
	isEnabled func() bool) error {
	gvk := schema.FromAPIVersionAndKind(c.resource.ApiVersion, c.resource.Kind)
	mapping, err := c.manager.dynamicClient.GetGVR(&gvk)
	if err != nil {
		return fmt.Errorf("failed to resolve resource type from kind %+v: %v", gvk, err)
	}

	object := &unstructured.Unstructured{}
	object.SetGroupVersionKind(gvk)

	err = ctrl.NewControllerManagedBy(c.manager.delegate).
		For(object).
		WithEventFilter(c.shouldWatchResource(gvk)).
		Complete(createReconciler(c.manager.dynamicClient, mapping, reconciler, grpcClient,
			c.reconciliationPeriodInMinutes, isEnabled))
	if err != nil {
		return fmt.Errorf("failed to create controller: %v", err)
	}
	return nil
}

func (c *ControllerBuilder) shouldWatchResource(gvk schema.GroupVersionKind) predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return c.isWatchedResource(e.Object, gvk)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return c.isWatchedResource(e.ObjectNew, gvk)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return c.isWatchedResource(e.Object, gvk)
		},
	}
}

func (c *ControllerBuilder) isWatchedResource(o client.Object, gvk schema.GroupVersionKind) bool {
	return o != nil &&
		o.GetObjectKind().GroupVersionKind() == gvk &&
		o.GetNamespace() == c.resource.GetNamespace().GetValue() &&
		c.selector.LabelSelector.Matches(labels.Set(o.GetLabels())) &&
		c.selector.FieldSelector.Matches(o)
}
