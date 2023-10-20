package controller

import (
	pb "github.com/SAP/remote-work-processor/build/proto/generated"
	"github.com/SAP/remote-work-processor/internal/functional"
)

const (
	OPERATOR_ID_ENV_VAR = "OPERATOR_ID"
)

type ReconciliationEvent struct {
	*pb.ClientMessage_ReconcileEvent
}

func newReconciliationEvent(opts ...functional.Option[ReconciliationEvent]) *ReconciliationEvent {
	re := &ReconciliationEvent{
		&pb.ClientMessage_ReconcileEvent{
			ReconcileEvent: &pb.ReconcileEventMessage{},
		},
	}

	for _, opt := range opts {
		opt(re)
	}

	return re
}

func (re *ReconciliationEvent) wrap() *pb.ClientMessage {
	op := &pb.ClientMessage{
		Body: re.ClientMessage_ReconcileEvent,
	}
	return op
}

func ofType(t pb.ReconcileEventMessage_ReconcileType) functional.Option[ReconciliationEvent] {
	return func(re *ReconciliationEvent) {
		re.ReconcileEvent.Type = t
	}
}

func withContent(c string) functional.Option[ReconciliationEvent] {
	return func(re *ReconciliationEvent) {
		re.ReconcileEvent.Content = c
	}
}

func withReconcilerName(c string) functional.Option[ReconciliationEvent] {
	return func(re *ReconciliationEvent) {
		re.ReconcileEvent.ReconcilerName = c
	}
}

func withResourceVersion(c string) functional.Option[ReconciliationEvent] {
	return func(re *ReconciliationEvent) {
		re.ReconcileEvent.ResourceVersion = c
	}
}

func withReconciliationRequest(name, namespace string) functional.Option[ReconciliationEvent] {
	return func(re *ReconciliationEvent) {
		re.ReconcileEvent.ReconciliationRequest = &pb.ReconciliationRequest{
			ResourceName:      name,
			ResourceNamespace: &namespace,
		}
	}
}
