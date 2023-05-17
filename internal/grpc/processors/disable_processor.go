package processors

import (
	"fmt"

	pb "github.com/SAP/remote-work-processor/build/proto/generated"
)

// Currently remote work processor ID should be in the following format - <cluster>:<app>:<instance>
type DisableProcessor struct {
}

func NewDisableProcessor() DisableProcessor {
	return DisableProcessor{}
}

func (p DisableProcessor) Process() <-chan *ProcessorResult {
	c := make(chan *ProcessorResult)
	go p.buildClientMessage(c)

	return c
}

func (p DisableProcessor) buildClientMessage(c chan<- *ProcessorResult) {
	fmt.Println("DISABLE OPERATOR...")
	c <- NewProcessorResult(Result(&pb.ClientMessage{
		Body: &pb.ClientMessage_ConfirmDisabled{
			ConfirmDisabled: &pb.ConfirmDisabledMessage{},
		},
	}))
}
