package controller

import (
	"context"
	"log"

	pb "github.com/SAP/remote-work-processor/build/proto/generated"
	"github.com/SAP/remote-work-processor/internal/kubernetes/selector"
	"github.com/pkg/errors"
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
	Controller
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

func (cb *ControllerBuilder) Build(ctx context.Context, reconciler string) (Controller, error) {
	b := ctrl.NewControllerManagedBy(cb.manager.manager)
	u := &unstructured.Unstructured{}
	gvk := schema.FromAPIVersionAndKind(cb.resource.ApiVersion, cb.resource.Kind)
	mapper, err := cb.manager.dynamicClient.GetGVR(&gvk)
	if err != nil {
		log.Fatalf("Failed to resolve resource type from kind: %v\n", err)
	}

	u.SetGroupVersionKind(gvk)

	s := cb.manager.selectorCache.Read(reconciler)

	b.For(u).WithEventFilter(shouldWatchResource(gvk, cb.resource.GetNamespace().GetValue(), &s))

	err = b.Complete(createReconciler(cb.manager.GetScheme(), cb.manager.dynamicClient, mapper, reconciler, cb.reconciliationPeriodInMinutes))
	if err != nil {
		return Controller{}, errors.Errorf("Unable to create a controller: %s", err)
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
	var l labels.Set
	l = o.GetLabels()

	return o != nil &&
		o.GetObjectKind().GroupVersionKind() == gvk &&
		o.GetNamespace() == ns &&
		s.LabelSelector.Matches(l) &&
		s.FieldSelector.Matches(o)
}
