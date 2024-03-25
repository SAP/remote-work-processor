package controller

import (
	pb "github.com/SAP/remote-work-processor/build/proto/generated"
	"github.com/SAP/remote-work-processor/internal/functional"
)

type ReconciliationEvent struct {
	msg *pb.ClientMessage_ReconcileEvent
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

func (re *ReconciliationEvent) toProtoMessage() *pb.ClientMessage {
	return &pb.ClientMessage{
		Body: re.msg,
	}
}

func ofType(t pb.ReconcileEventMessage_ReconcileType) functional.Option[ReconciliationEvent] {
	return func(re *ReconciliationEvent) {
		re.msg.ReconcileEvent.Type = t
	}
}

func withContent(c string) functional.Option[ReconciliationEvent] {
	return func(re *ReconciliationEvent) {
		re.msg.ReconcileEvent.Content = c
	}
}

func withReconcilerName(c string) functional.Option[ReconciliationEvent] {
	return func(re *ReconciliationEvent) {
		re.msg.ReconcileEvent.ReconcilerName = c
	}
}

func withResourceVersion(c string) functional.Option[ReconciliationEvent] {
	return func(re *ReconciliationEvent) {
		re.msg.ReconcileEvent.ResourceVersion = c
	}
}

func withReconciliationRequest(name, namespace string) functional.Option[ReconciliationEvent] {
	return func(re *ReconciliationEvent) {
		re.msg.ReconcileEvent.ReconciliationRequest = &pb.ReconciliationRequest{
			ResourceName:      name,
			ResourceNamespace: &namespace,
		}
	}
}
