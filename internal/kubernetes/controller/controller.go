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

type ResourceControllerBuilder interface {
	For(r *pb.Resource) *ControllerBuilder
	ManagedBy(m *ControllerManager) *ControllerBuilder
	Build() Controller
}

type Controller struct {
	resource                      *pb.Resource
	manager                       *ControllerManager
	reconciliationPeriodInMinutes int32
}

type ControllerBuilder struct {
	Controller //TODO: do not embed
}

func CreateControllerBuilder() *ControllerBuilder {
	return &ControllerBuilder{}
}

func (cb *ControllerBuilder) For(r *pb.Resource) *ControllerBuilder {
	cb.resource = r
	return cb
}

func (cb *ControllerBuilder) WithReconcilicationPeriodInMinutes(p int32) *ControllerBuilder {
	cb.reconciliationPeriodInMinutes = p
	return cb
}

func (cb *ControllerBuilder) ManagedBy(m *ControllerManager) *ControllerBuilder {
	cb.manager = m
	return cb
}

func (cb *ControllerBuilder) Build(reconciler string, grpcClient *grpc.RemoteWorkProcessorGrpcClient,
	isEnabled func() bool) (Controller, error) {
	gvk := schema.FromAPIVersionAndKind(cb.resource.ApiVersion, cb.resource.Kind)
	mapping, err := cb.manager.dynamicClient.GetGVR(&gvk)
	if err != nil {
		return Controller{}, fmt.Errorf("failed to resolve resource type from kind %+v: %v", gvk, err)
	}

	object := &unstructured.Unstructured{}
	object.SetGroupVersionKind(gvk)

	s := cb.manager.GetSelector(reconciler)

	err = ctrl.NewControllerManagedBy(cb.manager.manager).
		For(object).
		WithEventFilter(shouldWatchResource(gvk, cb.resource.GetNamespace().GetValue(), &s)).
		Complete(createReconciler(cb.manager.dynamicClient, mapping, reconciler, grpcClient,
			cb.reconciliationPeriodInMinutes, isEnabled))
	if err != nil {
		return Controller{}, fmt.Errorf("unable to create a controller: %v", err)
	}

	return cb.Controller, nil
}

func shouldWatchResource(gvk schema.GroupVersionKind, ns string, s *selector.Selector) predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return isWatchedResource(e.Object, gvk, ns, s)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return isWatchedResource(e.ObjectNew, gvk, ns, s)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return isWatchedResource(e.Object, gvk, ns, s)
		},
	}
}

func isWatchedResource(o client.Object, gvk schema.GroupVersionKind, ns string, s *selector.Selector) bool {
	return o != nil &&
		o.GetObjectKind().GroupVersionKind() == gvk &&
		o.GetNamespace() == ns &&
		s.LabelSelector.Matches(labels.Set(o.GetLabels())) &&
		s.FieldSelector.Matches(o)
}
