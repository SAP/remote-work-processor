package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pb "github.com/SAP/remote-work-processor/build/proto/generated"
	"github.com/SAP/remote-work-processor/internal/grpc"
	"github.com/SAP/remote-work-processor/internal/kubernetes/dynamic"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dyn "k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	FINALIZER = "automation.pilot.sap.com/finalizer"
)

type WatchConfigReconciler struct {
	*dynamic.DynamicClient
	*runtime.Scheme
	mapping                        *meta.RESTMapping
	reconciler                     string
	reconcilicationPeriodInMinutes time.Duration
}

func createReconciler(scheme *runtime.Scheme, client *dynamic.DynamicClient, mapping *meta.RESTMapping, reconciler string, reconcilicationPeriodInMinutes int32) reconcile.Reconciler {
	return &WatchConfigReconciler{
		Scheme:                         scheme,
		DynamicClient:                  client,
		mapping:                        mapping,
		reconciler:                     reconciler,
		reconcilicationPeriodInMinutes: time.Duration(reconcilicationPeriodInMinutes) * time.Minute,
	}
}

func (r *WatchConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var resource dyn.ResourceInterface
	if r.mapping.Scope.Name() == meta.RESTScopeNameNamespace { // This is the case when reconciled resource is namespaced
		resource = r.Client.Resource(r.mapping.Resource).Namespace(req.Namespace)
	} else {
		resource = r.Client.Resource(r.mapping.Resource)
	}

	u, err := resource.Get(ctx, req.Name, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Printf("resource not found. Ignoring the reconciliation, because object could be deleted")
			return ctrl.Result{}, nil
		}

		fmt.Printf("failed to get the resource for reconciliation: %v", err)
		return ctrl.Result{}, err
	}

	if u.GetDeletionTimestamp().IsZero() {
		if !controllerutil.ContainsFinalizer(u, FINALIZER) {
			controllerutil.AddFinalizer(u, FINALIZER)
			if _, err := resource.Update(ctx, u, v1.UpdateOptions{}); err != nil {
				fmt.Printf("failed to add resource finalizer: %v", err)
				return ctrl.Result{}, err
			}
		}
	} else {
		if controllerutil.ContainsFinalizer(u, FINALIZER) {
			if err := r.sendReconciliationEvent(u, pb.ReconcileEventMessage_RECONCILE_TYPE_DELETE); err != nil {
				return ctrl.Result{}, err
			}

			controllerutil.RemoveFinalizer(u, FINALIZER)
			if _, err := resource.Update(ctx, u, v1.UpdateOptions{}); err != nil {
				fmt.Printf("failed to remove resource finalizer: %v", err)
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{RequeueAfter: r.reconcilicationPeriodInMinutes}, nil
	}

	if err := r.sendReconciliationEvent(u, pb.ReconcileEventMessage_RECONCILE_TYPE_CREATE_OR_UPDATE); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: r.reconcilicationPeriodInMinutes}, nil
}

func (r *WatchConfigReconciler) sendReconciliationEvent(u *unstructured.Unstructured, t pb.ReconcileEventMessage_ReconcileType) error {
	b, err := json.Marshal(u)
	if err != nil {
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
