package processors

import (
	"context"
	"log"

	pb "github.com/SAP/remote-work-processor/build/proto/generated"
)

type DisableProcessor struct {
	disableFunc func()
}

func NewDisableProcessor(disableFunc func()) DisableProcessor {
	return DisableProcessor{
		disableFunc: disableFunc,
	}
}

func (p DisableProcessor) Process(_ context.Context) (*pb.ClientMessage, error) {
	log.Println("Disabling work processor...")

	p.disableFunc()

	return &pb.ClientMessage{
		Body: &pb.ClientMessage_ConfirmDisabled{
			ConfirmDisabled: &pb.ConfirmDisabledMessage{},
		},
	}, nil
}
