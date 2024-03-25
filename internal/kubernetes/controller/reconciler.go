package controller

import (
	"context"
	"encoding/json"
	stdLog "log"
	"time"

	pb "github.com/SAP/remote-work-processor/build/proto/generated"
	"github.com/SAP/remote-work-processor/internal/grpc"
	"github.com/SAP/remote-work-processor/internal/kubernetes/dynamic"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
	*dynamic.Client
	mapping                        *meta.RESTMapping
	reconciler                     string
	reconcilicationPeriodInMinutes time.Duration
	grpcClient                     *grpc.RemoteWorkProcessorGrpcClient
	isEnabled                      func() bool
}

func createReconciler(client *dynamic.Client, mapping *meta.RESTMapping, reconciler string,
	grpcClient *grpc.RemoteWorkProcessorGrpcClient, reconcilicationPeriodInMinutes int32, isEnabled func() bool) reconcile.Reconciler {
	return &WatchConfigReconciler{
		Client:                         client,
		mapping:                        mapping,
		reconciler:                     reconciler,
		grpcClient:                     grpcClient,
		reconcilicationPeriodInMinutes: time.Duration(reconcilicationPeriodInMinutes) * time.Minute,
		isEnabled:                      isEnabled,
	}
}

func (r *WatchConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if !r.isEnabled() {
		return ctrl.Result{RequeueAfter: r.reconcilicationPeriodInMinutes}, nil
	}

	var resource dyn.ResourceInterface
	if r.mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		resource = r.GetResourceInterface(r.mapping.Resource).Namespace(req.Namespace)
	} else {
		resource = r.GetResourceInterface(r.mapping.Resource)
	}

	logger := log.FromContext(ctx)

	object, err := resource.Get(ctx, req.Name, v1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			logger.Info("resource not found. Ignoring the reconciliation, because object could be deleted")
			return ctrl.Result{}, nil
		}

		logger.Error(err, "failed to get the resource for reconciliation")
		return ctrl.Result{}, err
	}

	if object.GetDeletionTimestamp().IsZero() {
		if !controllerutil.ContainsFinalizer(object, FINALIZER) {
			controllerutil.AddFinalizer(object, FINALIZER)
			if _, err := resource.Update(ctx, object, v1.UpdateOptions{}); err != nil {
				logger.Error(err, "failed to add resource finalizer")
				return ctrl.Result{}, err
			}
		}
	} else {
		if err := r.sendReconciliationEvent(object, pb.ReconcileEventMessage_RECONCILE_TYPE_DELETE); err != nil {
			return ctrl.Result{}, err
		}

		if controllerutil.RemoveFinalizer(object, FINALIZER) {
			if _, err := resource.Update(ctx, object, v1.UpdateOptions{}); err != nil {
				logger.Error(err, "failed to remove resource finalizer")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{RequeueAfter: r.reconcilicationPeriodInMinutes}, nil
	}

	if err := r.sendReconciliationEvent(object, pb.ReconcileEventMessage_RECONCILE_TYPE_CREATE_OR_UPDATE); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: r.reconcilicationPeriodInMinutes}, nil
}

func (r *WatchConfigReconciler) sendReconciliationEvent(object *unstructured.Unstructured,
	reconcileType pb.ReconcileEventMessage_ReconcileType) error {
	serialized, err := json.Marshal(object)
	if err != nil {
		return err
	}

	msg := newReconciliationEvent(
		ofType(reconcileType),
		withContent(string(serialized)),
		withResourceVersion(object.GetResourceVersion()),
		withReconcilerName(r.reconciler),
		withReconciliationRequest(object.GetName(), object.GetNamespace()),
	).toProtoMessage()

	err = r.grpcClient.Send(msg)
	if err != nil {
		// the gRPC connection has broken down, need to reestablish or restart the process
		stdLog.Printf("could not send reconciliation event message: %v\n", err)
	}
	return nil
}
