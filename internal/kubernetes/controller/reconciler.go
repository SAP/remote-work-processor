package controller

import (
	"context"
	"encoding/json"

	pb "github.com/SAP/remote-work-processor/build/proto/generated"
	"github.com/SAP/remote-work-processor/internal/grpc"
	"github.com/SAP/remote-work-processor/internal/kubernetes/dynamic"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dyn "k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	FINALIZER = "automation.pilot.sap.com/finalizer"
)

type WatchConfigReconciler struct {
	*dynamic.DynamicClient
	*runtime.Scheme
	mapping    *meta.RESTMapping
	reconciler string
}

func createReconciler(scheme *runtime.Scheme, client *dynamic.DynamicClient, mapping *meta.RESTMapping, reconciler string) reconcile.Reconciler {
	return &WatchConfigReconciler{
		Scheme:        scheme,
		DynamicClient: client,
		mapping:       mapping,
		reconciler:    reconciler,
	}
}

func (r *WatchConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info(req.Name)

	var resource dyn.ResourceInterface
	if r.mapping.Scope.Name() == meta.RESTScopeNameNamespace { // This is the case when reconciled resource is namespaced
		resource = r.Client.Resource(r.mapping.Resource).Namespace(req.Namespace)
	} else {
		resource = r.Client.Resource(r.mapping.Resource)
	}

	u, err := resource.Get(ctx, req.Name, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("resource not found. Ignoring the reconciliation, because object could be deleted")
			return ctrl.Result{}, nil
		}

		logger.Error(err, "failed to get the resource for reconciliation")
		return ctrl.Result{}, err
	}

	if u.GetDeletionTimestamp().IsZero() {
		if !controllerutil.ContainsFinalizer(u, FINALIZER) {
			controllerutil.AddFinalizer(u, FINALIZER)
			if _, err := resource.Update(ctx, u, v1.UpdateOptions{}); err != nil {
				logger.Error(err, "failed to add resource finalizer")
				return ctrl.Result{}, err
			}
		}
	} else {
		if controllerutil.ContainsFinalizer(u, FINALIZER) {
			if err := r.sendReconciliationEvent(u, pb.ReconcileEventMessage_RECONCILE_TYPE_DELETE, logger); err != nil {
				return ctrl.Result{}, err
			}

			controllerutil.RemoveFinalizer(u, FINALIZER)
			if _, err := resource.Update(ctx, u, v1.UpdateOptions{}); err != nil {
				logger.Error(err, "failed to remove resource finalizer")
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	if err := r.sendReconciliationEvent(u, pb.ReconcileEventMessage_RECONCILE_TYPE_CREATE_OR_UPDATE, logger); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *WatchConfigReconciler) sendReconciliationEvent(u *unstructured.Unstructured, t pb.ReconcileEventMessage_ReconcileType, logger logr.Logger) error {
	b, err := json.Marshal(u)
	if err != nil {
		logger.Error(err, "failed to marshal resource to JSON")
		return err
	}

	grpc.Client.Send(newReconciliationEvent(
		ofType(t),
		withContent(string(b)),
		withResourceVersion(u.GetResourceVersion()),
		withReconcilerName(r.reconciler),
	).wrap())

	return nil
}
