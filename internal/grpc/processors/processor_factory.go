package processors

import (
	"sync"

	pb "github.com/SAP/remote-work-processor/build/proto/generated"
	"github.com/SAP/remote-work-processor/internal/kubernetes/engine"
)

var (
	once    sync.Once
	Factory ProcessorFactory
)

type ProcessorFactory struct {
	engine engine.ManagerEngine
}

func InitProcessorFactory(engine engine.ManagerEngine) {
	once.Do(func() {
		Factory = ProcessorFactory{
			engine: engine,
		}
	})
}

func (pf *ProcessorFactory) CreateProcessor(op *pb.ServerMessage) (Processor, error) {
	switch b := op.Body.(type) {
	case *pb.ServerMessage_TaskExecutionRequest:
		return NewRemoteTaskProcessor(b), nil
	case *pb.ServerMessage_UpdateConfigRequest:
		return NewUpdateWatchConfigurationProcessor(op, pf.engine), nil
	case *pb.ServerMessage_DisableRequest:
		return NewDisableProcessor(), nil
	case *pb.ServerMessage_EnableRequest:
		return NewEnableProcessor(), nil
	default:
		return nil, ProcessorError{}
	}
}

func (pf *ProcessorFactory) CreateProbeSessionProcessor() Processor {
	return NewProbeSessionProcessor()
}
