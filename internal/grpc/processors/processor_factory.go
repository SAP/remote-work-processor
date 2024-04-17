package processors

import (
	"fmt"
	pb "github.com/SAP/remote-work-processor/build/proto/generated"
	"github.com/SAP/remote-work-processor/internal/kubernetes/engine"
	"sync/atomic"
)

type ProcessorFactory struct {
	engine     engine.ManagerEngine
	drainChan  chan struct{}
	rwpEnabled *atomic.Bool
}

func NewKubernetesProcessorFactory(engine engine.ManagerEngine, drainChan chan struct{}) ProcessorFactory {
	enabled := &atomic.Bool{}
	enabled.Store(true)
	// ensure the channel does not deadlock main() in case no watch config is ever set
	drainChan <- struct{}{}
	return ProcessorFactory{
		engine:     engine,
		drainChan:  drainChan,
		rwpEnabled: enabled,
	}
}

func NewStandaloneProcessorFactory() ProcessorFactory {
	enabled := &atomic.Bool{}
	enabled.Store(true)
	return ProcessorFactory{
		rwpEnabled: enabled,
	}
}

func (pf *ProcessorFactory) CreateProcessor(op *pb.ServerMessage) (Processor, error) {
	// NOTE: The NextEventRequestMessage message changes the current k8s reconciliation flow.
	// 	Instead of sending an event message to the server on every reconcile loop,
	//  push these events to a queue (in a separate goroutine).
	//  That routine will listen for the NextEventRequestMessage and only send messages when it receives it.
	//  This queue will send reconcilliation event messages to the backend when either:
	//  - the queue is empty;
	//  - the queue has elements and there is a NextEventRequestMessage.
	//  Since this logic hasn't been implemented in the backend yet, it's not present here either.
	switch b := op.Body.(type) {
	case *pb.ServerMessage_TaskExecutionRequest:
		return NewRemoteTaskProcessor(b, pf.rwpEnabled.Load), nil
	case *pb.ServerMessage_UpdateConfigRequest:
		return NewUpdateWatchConfigurationProcessor(b, pf.engine, pf.drainChan, pf.rwpEnabled.Load), nil
	case *pb.ServerMessage_DisableRequest:
		return NewDisableProcessor(func() { pf.rwpEnabled.Store(false) }), nil
	case *pb.ServerMessage_EnableRequest:
		return NewEnableProcessor(func() { pf.rwpEnabled.Store(true) }), nil
	default:
		return nil, fmt.Errorf("unrecognized request type %+v", op.Body)
	}
}
