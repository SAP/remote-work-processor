package processors

import (
	"fmt"

	pb "github.com/SAP/remote-work-processor/build/proto/generated"
)

// Currently remote work processor ID should be in the following format - <cluster>:<app>:<instance>
type EnableProcessor struct {
}

func NewEnableProcessor() EnableProcessor {
	return EnableProcessor{}
}

func (p EnableProcessor) Process() <-chan *ProcessorResult {
	c := make(chan *ProcessorResult)
	go p.buildClientMessage(c)

	return c
}

func (p EnableProcessor) buildClientMessage(c chan<- *ProcessorResult) {
	fmt.Println("ENABLING OPERATOR...")
	c <- NewProcessorResult(Result(&pb.ClientMessage{
		Body: &pb.ClientMessage_ConfirmEnabled{
			ConfirmEnabled: &pb.ConfirmEnabledMessage{},
		},
	}))
}
