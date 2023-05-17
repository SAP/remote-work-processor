package processors

import (
	"time"

	pb "github.com/SAP/remote-work-processor/build/proto/generated"
)

type ProbeSessionProcessor struct {
}

func NewProbeSessionProcessor() ProbeSessionProcessor {
	return ProbeSessionProcessor{}
}

func (p ProbeSessionProcessor) Process() <-chan *ProcessorResult {
	c := make(chan *ProcessorResult)
	done := make(chan struct{})
	go p.buildProbeSession(c, done)

	return c
}

func (p ProbeSessionProcessor) buildProbeSession(c chan<- *ProcessorResult, done chan struct{}) {
	t := time.NewTicker(30 * time.Second)

	for {
		select {
		case <-t.C:
			op := &pb.ClientMessage{
				Body: &pb.ClientMessage_ProbeSession{
					ProbeSession: &pb.ProbeSessionMessage{},
				},
			}

			c <- NewProcessorResult(Result(op), OnChannel(done))
		case <-done:
			t.Stop()
		}
	}
}
