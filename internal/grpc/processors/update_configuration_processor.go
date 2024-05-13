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
		// this corner case can't happen unless there is a proxy between the AutoPi and the RWP
		// because the AutoPi won't send any messages to disabled Operators
		return nil, nil
	}

	if len(p.op.UpdateConfigRequest.Resources) == 0 {
		// handle session auto-config
		return &pb.ClientMessage{Body: p.getConfirmUpdateMessage()}, nil
	}

	if p.engine == nil {
		log.Println("Unable to process watch config: Remote Worker is running in standalone mode.")
		// this corner case can't happen unless there is a proxy between the AutoPi and the RWP
		// because the AutoPi won't send UpdateConfig messages to Standalone Operators
		return nil, nil
	}

	if p.engine.IsRunning() {
		log.Println("Stopping Manager...")
		p.engine.Stop()
		<-p.drainChan
	}

	go func() {
		select {
		case <-p.drainChan:
			//drain in case the manager hasn't been started yet (the processor factory signals this channel)
		default:
		}

		log.Println("New watch config received...")
		p.engine.SetWatchConfiguration(p.op.UpdateConfigRequest)

		if err := p.engine.WatchResources(ctx, p.isEnabled); err != nil {
			log.Fatalln("failed to watch resources:", err)
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
