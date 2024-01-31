package processors

import (
	"context"
	"log"

	pb "github.com/SAP/remote-work-processor/build/proto/generated"
	"github.com/SAP/remote-work-processor/internal/kubernetes/engine"
)

type UpdateWatchConfigurationProcessor struct {
	op        *pb.ServerMessage_UpdateConfigRequest
	engine    engine.ManagerEngine
	drainChan chan struct{}
	isEnabled func() bool
}

func NewUpdateWatchConfigurationProcessor(op *pb.ServerMessage_UpdateConfigRequest, engine engine.ManagerEngine,
	drainChan chan struct{}, isEnabled func() bool) UpdateWatchConfigurationProcessor {
	return UpdateWatchConfigurationProcessor{
		op:        op,
		engine:    engine,
		drainChan: drainChan,
		isEnabled: isEnabled,
	}
}

func (p UpdateWatchConfigurationProcessor) Process(ctx context.Context) (*pb.ClientMessage, error) {
	if !p.isEnabled() || p.engine == nil {
		return nil, nil
	}

	if p.engine.IsStarted() {
		log.Println("Stopping Manager....")
		p.engine.StopManager()
		<-p.drainChan
	}

	go func() {
		select {
		case <-p.drainChan:
			//drain in case the manager failed to start previously
		default:
		}

		log.Println("New watch config received. Starting manager....")
		p.engine.SetWatchConfiguration(p.op.UpdateConfigRequest)

		if err := p.engine.StartManager(ctx, p.isEnabled); err != nil {
			log.Printf("unable to start manager: %v\n", err)
			//TODO: send an error message to the server
		}
		p.drainChan <- struct{}{}
	}()

	return &pb.ClientMessage{
		Body: &pb.ClientMessage_ConfirmConfigUpdate{
			ConfirmConfigUpdate: &pb.ConfirmConfigUpdateMessage{
				ConfigVersion: p.op.UpdateConfigRequest.GetConfigVersion(),
			},
		},
	}, nil
}
