package processors

import (
	"log"

	pb "github.com/SAP/remote-work-processor/build/proto/generated"
	"github.com/SAP/remote-work-processor/internal/kubernetes/engine"
)

type UpdateWatchConfigurationProcessor struct {
	op     *pb.ServerMessage
	engine engine.ManagerEngine
	wcc    chan *pb.UpdateConfigRequestMessage
}

func NewUpdateWatchConfigurationProcessor(op *pb.ServerMessage, engine engine.ManagerEngine) UpdateWatchConfigurationProcessor {
	return UpdateWatchConfigurationProcessor{
		op:     op,
		engine: engine,
		wcc:    make(chan *pb.UpdateConfigRequestMessage),
	}
}

func (p UpdateWatchConfigurationProcessor) Process() <-chan *ProcessorResult {
	c := make(chan *ProcessorResult)

	go func() {
		if p.engine.ManagerStartedAtLeastOnce() {
			log.Print("Stopping Manager....")
			p.engine.StopManager()
		}

		go func() {
			for {
				wc := <-p.wcc
				log.Print("New watch config received. Starting manager....")

				p.engine.WithWatchConfiguration(wc)
				p.engine.WithContext()

				if err := p.engine.StartManager(); err != nil {
					log.Fatalf("unable to start manager: %v\n", err)
				}
			}
		}()

		uc, ok := p.op.Body.(*pb.ServerMessage_UpdateConfigRequest)
		if !ok {
			c <- NewProcessorResult(Error(ProcessorError{}))
		}

		p.wcc <- uc.UpdateConfigRequest
		c <- NewProcessorResult(Result(&pb.ClientMessage{
			Body: &pb.ClientMessage_ConfirmConfigUpdate{
				ConfirmConfigUpdate: &pb.ConfirmConfigUpdateMessage{
					ConfigVersion: uc.UpdateConfigRequest.GetConfigVersion(),
				},
			},
		}))
	}()

	return c
}
