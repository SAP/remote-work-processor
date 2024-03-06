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
	if !p.isEnabled() {
		log.Println("Unable to process watch config: Remote Worker is disabled.")
		return nil, nil
	}

	if len(p.op.UpdateConfigRequest.Resources) == 0 {
		// handle session auto-config
		return &pb.ClientMessage{Body: p.getConfirmUpdateMessage()}, nil
	}

	if p.engine == nil {
		log.Println("Unable to process watch config: Remote Worker is running in standalone mode.")
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
			//drain in case the manager hasn't been started yet (the processor factory signals this channel)
		default:
		}

		log.Println("New watch config received. Starting manager....")
		p.engine.SetWatchConfiguration(p.op.UpdateConfigRequest)

		if err := p.engine.StartManager(ctx, p.isEnabled); err != nil {
			log.Fatalln("failed to start manager:", err)
		}
		p.drainChan <- struct{}{}
	}()

	return &pb.ClientMessage{Body: p.getConfirmUpdateMessage()}, nil
}

func (p UpdateWatchConfigurationProcessor) getConfirmUpdateMessage() *pb.ClientMessage_ConfirmConfigUpdate {
	return &pb.ClientMessage_ConfirmConfigUpdate{
		ConfirmConfigUpdate: &pb.ConfirmConfigUpdateMessage{
			ConfigVersion: p.op.UpdateConfigRequest.GetConfigVersion(),
		},
	}
}
